package peer
import "errors"
import "fmt"

type decoderFunc func(*UDPPacket, *CommunicationContext) (interface{}, error)
var packetDecoders = map[byte]decoderFunc{
	0x05: decodePacket05Layer,
	0x06: decodePacket06Layer,
	0x07: decodePacket07Layer,
	0x08: decodePacket08Layer,
	0x00: decodePacket00Layer,
	0x03: decodePacket03Layer,
	0x09: decodePacket09Layer,
	0x10: decodePacket10Layer,
	0x13: decodePacket13Layer,

	0x81: decodePacket81Layer,
	0x82: decodePacket82Layer,
	0x83: decodePacket83Layer,
	//0x85: decodePacket85Layer,
	//0x86: decodePacket86Layer,
	0x8F: decodePacket8FLayer,
	0x90: decodePacket90Layer,
	0x92: decodePacket92Layer,
	0x93: decodePacket93Layer,
	0x97: decodePacket97Layer,
}

type ReceiveHandler func(byte, *UDPPacket, *PacketLayers)
type ErrorHandler func(error)

type rmNumberQueue struct {
	index uint32
	queue map[uint32]*PacketLayers
}
type sequenceQueue rmNumberQueue
type orderingQueue struct {
	index [32]uint32
	queue [32]map[uint32]*PacketLayers
}

type queues struct {
	rmNumberQueue *rmNumberQueue
	sequenceQueue *sequenceQueue
	orderingQueue *orderingQueue
}

func (q *queues) add(layers *PacketLayers) {
	packet := layers.Reliability
	reliability := packet.Reliability
	hasRMN := reliability >= 2 && reliability <= 4
	hasSeq := reliability == 1 || reliability == 4
	hasOrd := reliability == 1 || reliability == 4 || reliability == 3 || reliability == 7
	if hasRMN {
		q.rmNumberQueue.queue[packet.ReliableMessageNumber] = layers
	}
	if hasSeq {
		q.sequenceQueue.queue[packet.SequencingIndex] = layers
	}
	if hasOrd {
		q.orderingQueue.queue[packet.OrderingChannel][packet.OrderingIndex] = layers
	}
}

func (q *queues) remove(layers *PacketLayers) {
	packet := layers.Reliability
	reliability := packet.Reliability
	hasRMN := reliability >= 2 && reliability <= 4
	hasSeq := reliability == 1 || reliability == 4
	hasOrd := reliability == 1 || reliability == 4 || reliability == 3 || reliability == 7
	if hasRMN {
		delete(q.rmNumberQueue.queue, packet.ReliableMessageNumber)
	}
	if hasSeq {
		delete(q.sequenceQueue.queue, packet.SequencingIndex)
	}
	if hasOrd {
		delete(q.orderingQueue.queue[packet.OrderingChannel], packet.OrderingIndex)
	}
}

func (q *queues) next(channel uint8) *PacketLayers {
	rmnReadIndex := q.rmNumberQueue.index
	packet, ok := q.rmNumberQueue.queue[rmnReadIndex]
	if ok {
		q.rmNumberQueue.index++
		reliability := packet.Reliability.Reliability
		hasSeq := reliability == 1 || reliability == 4
		hasOrd := reliability == 1 || reliability == 4 || reliability == 3 || reliability == 7
		if !hasSeq && !hasOrd {
			return packet
		}
	}

	ordReadIndex := q.orderingQueue.index[channel]
	packet, ok = q.orderingQueue.queue[channel][ordReadIndex]
	if !ok {
		return nil
	}
	if packet.Reliability.IsFinal {
		q.orderingQueue.index[channel]++
	}

	reliability := packet.Reliability.Reliability
	hasSeq := reliability == 1 || reliability == 4
	if !hasSeq {
		return packet
	}

	if packet.Reliability.SequencingIndex == q.sequenceQueue.index {
		if packet.Reliability.IsFinal {
			q.sequenceQueue.index++
		}
		return packet
	}
	return nil
}

// PacketReader is a struct that can be used to read packets from a source
// Pass packets in using ReadPacket() and bind to the given callbacks
// to receive the results
type PacketReader struct {
	// Callback for "simple" packets (pre-connection offline packets).
	SimpleHandler ReceiveHandler
	// Callback for ReliabilityLayer subpackets. This callback is invoked for every
	// split of every packets, possible duplicates, etc.
	ReliableHandler ReceiveHandler
	// Callback for generic packets (anything that is sent when a connection has been
	// established. You definitely want to bind to this.
	FullReliableHandler ReceiveHandler
	// Callback for ACKs and NAKs. Not very useful if you're just parsing packets.
	// However, you want to bind to this if you are writing a peer.
	ACKHandler func(*UDPPacket, *RakNetLayer)
	// Callback for ReliabilityLayer full packets. This callback is invoked for every
	// real ReliabilityLayer.
	ReliabilityLayerHandler func(*UDPPacket, *ReliabilityLayer, *RakNetLayer)
	// Simply enough, any errors encountered are reported to ErrorHandler.
	ErrorHandler ErrorHandler
	// Context is a struct representing the state of the connection. It contains
	// information such as the addresses of the peers and the state of the DataModel.
	Context *CommunicationContext

	serverQueues *queues
	clientQueues *queues
}

func NewPacketReader() *PacketReader {
	serverOrderingQueue := [32]map[uint32]*PacketLayers{}
	for i := 0; i < 32; i++ {
		serverOrderingQueue[i] = make(map[uint32]*PacketLayers)
	}
	clientOrderingQueue := [32]map[uint32]*PacketLayers{}
	for i := 0; i < 32; i++ {
		clientOrderingQueue[i] = make(map[uint32]*PacketLayers)
	}

	return &PacketReader{
		serverQueues: &queues{
			rmNumberQueue: &rmNumberQueue{
				index: 0,
				queue: make(map[uint32]*PacketLayers),
			},
			sequenceQueue: &sequenceQueue{
				index: 0,
				queue: make(map[uint32]*PacketLayers),
			},
			orderingQueue: &orderingQueue{
				queue: serverOrderingQueue,
			},
		},
		clientQueues: &queues{
			rmNumberQueue: &rmNumberQueue{
				index: 0,
				queue: make(map[uint32]*PacketLayers),
			},
			sequenceQueue: &sequenceQueue{
				index: 0,
				queue: make(map[uint32]*PacketLayers),
			},
			orderingQueue: &orderingQueue{
				queue: clientOrderingQueue,
			},
		},
	}
}

func (this *PacketReader) readSimple(packetType uint8, layers *PacketLayers, packet *UDPPacket) {
	var err error
	decoder := packetDecoders[packetType]
	if decoder != nil {
		layers.Main, err = decoder(packet, this.Context)
		if err != nil {
			this.ErrorHandler(errors.New(fmt.Sprintf("Failed to decode simple packet %02X: %s", packetType, err.Error())))
			return
		}
	}

	this.SimpleHandler(packetType, packet, layers)
}

func (this *PacketReader) readGeneric(packetType uint8, layers *PacketLayers, packet *UDPPacket) {
	var err error
	if packetType == 0x1B {
		tsLayer, err := decodePacket1BLayer(packet, this.Context)
		if err != nil {
			packet.Logger.Println("error:", err.Error())
			this.ErrorHandler(errors.New(fmt.Sprintf("Failed to decode timestamped packet: %s", err.Error())))
			return
		}
		layers.Timestamp = tsLayer.(*Packet1BLayer)
		packetType, err = packet.stream.ReadByte()
		if err != nil {
			packet.Logger.Println("error:", err.Error())
			this.ErrorHandler(errors.New(fmt.Sprintf("Failed to decode timestamped packet: %s", err.Error())))
			return
		}
		layers.Reliability.PacketType = packetType
		layers.Reliability.HasPacketType = true
	}
	if packetType == 0x8A {
		data, err := packet.stream.readString(int((layers.Reliability.LengthInBits+7)/8) - 1)
		if err != nil {
			packet.Logger.Println("error:", err.Error())
			this.ErrorHandler(errors.New(fmt.Sprintf("Failed to decode reliable packet %02X: %s", packetType, err.Error())))

			return
		}
		layers.Main, err = decodePacket8ALayer(packet, this.Context, data)

		if err != nil {
			packet.Logger.Println("error:", err.Error())
			this.ErrorHandler(errors.New(fmt.Sprintf("Failed to decode reliable packet %02X: %s", packetType, err.Error())))

			return
		}
	} else {
		decoder := packetDecoders[layers.Reliability.PacketType]
		if decoder != nil {
			layers.Main, err = decoder(packet, this.Context)

			if err != nil {
				packet.Logger.Println("error:", err.Error())
				this.ErrorHandler(errors.New(fmt.Sprintf("Failed to decode reliable packet %02X: %s", layers.Reliability.PacketType, err.Error())))

				return
			}
		}
	}

	this.FullReliableHandler(packetType, packet, layers)
}

func (this *PacketReader) readOrdered(layers *PacketLayers, packet *UDPPacket) {
	var err error
	subPacket := layers.Reliability
	if subPacket.HasPacketType && !subPacket.HasBeenDecoded && subPacket.IsFinal {
		packetType := subPacket.PacketType
		_, err = subPacket.fullDataReader.ReadByte()
		if err != nil {
			packet.Logger.Println("error:", err.Error())
			this.ErrorHandler(errors.New(fmt.Sprintf("Failed to decode reliablePacket %02X: %s", packetType, err.Error())))
			return
		}
		newPacket := &UDPPacket{subPacket.fullDataReader, packet.Source, packet.Destination}

		this.readGeneric(packetType, layers, newPacket)
	}
}

func (this *PacketReader) readReliable(layers *PacketLayers, packet *UDPPacket) {
	packet.stream = layers.RakNet.payload
	reliabilityLayer, err := DecodeReliabilityLayer(packet, this.Context, layers.RakNet)
	if err != nil {
		this.ErrorHandler(errors.New("Failed to decode reliable packet: " + err.Error()))
		return
	}

	isClient := this.Context.IsClient(packet.Source)
	var queues *queues

	if isClient {
		queues = this.clientQueues
	} else {
		queues = this.serverQueues
	}

	this.ReliabilityLayerHandler(packet, reliabilityLayer, layers.RakNet)
	for _, subPacket := range reliabilityLayer.Packets {
		reliablePacketLayers := &PacketLayers{RakNet: layers.RakNet, Reliability: subPacket}
		this.ReliableHandler(subPacket.PacketType, packet, reliablePacketLayers)
		queues.add(reliablePacketLayers)
		if reliablePacketLayers.Reliability.Reliability == 0 {
			//println("read UNRELI")
			this.readOrdered(reliablePacketLayers, packet)
			queues.remove(reliablePacketLayers)
			continue
		}

		reliablePacketLayers = queues.next(subPacket.OrderingChannel)
		for reliablePacketLayers != nil {
			this.readOrdered(reliablePacketLayers, packet)
			queues.remove(reliablePacketLayers)
			reliablePacketLayers = queues.next(subPacket.OrderingChannel)
		}
	}
}

// ReadPacket reads a single packet and invokes all according handler functions
func (this *PacketReader) ReadPacket(payload []byte, packet *UDPPacket) {
	context := this.Context

	packet.stream = bufferToStream(payload)
	rakNetLayer, err := DecodeRakNetLayer(payload[0], packet, context)
	if err != nil {
		this.ErrorHandler(err)
		return
	}
	if rakNetLayer.IsDuplicate {
		return
	}

	layers := &PacketLayers{RakNet: rakNetLayer}
	if rakNetLayer.IsSimple {
		packetType := rakNetLayer.SimpleLayerID
		this.readSimple(packetType, layers, packet)
		return
	} else if !rakNetLayer.IsValid {
		this.ErrorHandler(errors.New(fmt.Sprintf("Sent invalid packet (packet header %x)", payload[0])))
		return
	} else if rakNetLayer.IsACK || rakNetLayer.IsNAK {
		this.ACKHandler(packet, rakNetLayer)
	} else {
		this.readReliable(layers, packet)
		return
	}
}
