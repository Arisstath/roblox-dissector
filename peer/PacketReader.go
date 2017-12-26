package peer
import "errors"
import "fmt"

type DecoderFunc func(*UDPPacket, *CommunicationContext) (interface{}, error)
var PacketDecoders = map[byte]DecoderFunc{
	0x05: DecodePacket05Layer,
	0x06: DecodePacket06Layer,
	0x07: DecodePacket07Layer,
	0x08: DecodePacket08Layer,
	0x00: DecodePacket00Layer,
	0x03: DecodePacket03Layer,
	0x09: DecodePacket09Layer,
	0x10: DecodePacket10Layer,
	0x13: DecodePacket13Layer,

	0x81: DecodePacket81Layer,
	0x82: DecodePacket82Layer,
	0x83: DecodePacket83Layer,
	0x85: DecodePacket85Layer,
	0x86: DecodePacket86Layer,
	0x8F: DecodePacket8FLayer,
	0x90: DecodePacket90Layer,
	0x92: DecodePacket92Layer,
	0x93: DecodePacket93Layer,
	0x97: DecodePacket97Layer,
}

type ReceiveHandler func(byte, *UDPPacket, *PacketLayers)
type ErrorHandler func(error)

type RMNumberQueue struct {
	Index uint32
	Queue map[uint32]*PacketLayers
}
type SequenceQueue RMNumberQueue
type OrderingQueue struct {
	Index [32]uint32
	Queue [32]map[uint32]*PacketLayers
}

type Queues struct {
	RMNumberQueue *RMNumberQueue
	SequenceQueue *SequenceQueue
	OrderingQueue *OrderingQueue
}

func (q *Queues) Add(layers *PacketLayers) {
	packet := layers.Reliability
	reliability := packet.Reliability
	hasRMN := reliability >= 2 && reliability <= 4
	hasSeq := reliability == 1 || reliability == 4
	hasOrd := reliability == 1 || reliability == 4 || reliability == 3 || reliability == 7
	if hasRMN {
		q.RMNumberQueue.Queue[packet.ReliableMessageNumber] = layers
	}
	if hasSeq {
		q.SequenceQueue.Queue[packet.SequencingIndex] = layers
	}
	if hasOrd {
		q.OrderingQueue.Queue[packet.OrderingChannel][packet.OrderingIndex] = layers
	}
}

func (q *Queues) Remove(layers *PacketLayers) {
	packet := layers.Reliability
	reliability := packet.Reliability
	hasRMN := reliability >= 2 && reliability <= 4
	hasSeq := reliability == 1 || reliability == 4
	hasOrd := reliability == 1 || reliability == 4 || reliability == 3 || reliability == 7
	if hasRMN {
		delete(q.RMNumberQueue.Queue, packet.ReliableMessageNumber)
	}
	if hasSeq {
		delete(q.SequenceQueue.Queue, packet.SequencingIndex)
	}
	if hasOrd {
		delete(q.OrderingQueue.Queue[packet.OrderingChannel], packet.OrderingIndex)
	}
}

func (q *Queues) Next(channel uint8) *PacketLayers {
	rmnReadIndex := q.RMNumberQueue.Index
	packet, ok := q.RMNumberQueue.Queue[rmnReadIndex]
	if ok {
		q.RMNumberQueue.Index++
		reliability := packet.Reliability.Reliability
		hasSeq := reliability == 1 || reliability == 4
		hasOrd := reliability == 1 || reliability == 4 || reliability == 3 || reliability == 7
		if !hasSeq && !hasOrd {
			return packet
		}
	}

	ordReadIndex := q.OrderingQueue.Index[channel]
	packet, ok = q.OrderingQueue.Queue[channel][ordReadIndex]
	if !ok {
		return nil
	}
	if packet.Reliability.IsFinal {
		q.OrderingQueue.Index[channel]++
	}

	reliability := packet.Reliability.Reliability
	hasSeq := reliability == 1 || reliability == 4
	if !hasSeq {
		return packet
	}

	if packet.Reliability.SequencingIndex == q.SequenceQueue.Index {
		if packet.Reliability.IsFinal {
			q.SequenceQueue.Index++
		}
		return packet
	}
	return nil
}

type PacketReader struct {
	SimpleHandler ReceiveHandler
	ReliableHandler ReceiveHandler
	FullReliableHandler ReceiveHandler
	ACKHandler func(*UDPPacket, *RakNetLayer)
	ReliabilityLayerHandler func(*UDPPacket, *ReliabilityLayer, *RakNetLayer)
	ErrorHandler ErrorHandler
	Context *CommunicationContext

	ServerQueues *Queues
	ClientQueues *Queues
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
		ServerQueues: &Queues{
			RMNumberQueue: &RMNumberQueue{
				Index: 0,
				Queue: make(map[uint32]*PacketLayers),
			},
			SequenceQueue: &SequenceQueue{
				Index: 0,
				Queue: make(map[uint32]*PacketLayers),
			},
			OrderingQueue: &OrderingQueue{
				Queue: serverOrderingQueue,
			},
		},
		ClientQueues: &Queues{
			RMNumberQueue: &RMNumberQueue{
				Index: 0,
				Queue: make(map[uint32]*PacketLayers),
			},
			SequenceQueue: &SequenceQueue{
				Index: 0,
				Queue: make(map[uint32]*PacketLayers),
			},
			OrderingQueue: &OrderingQueue{
				Queue: clientOrderingQueue,
			},
		},
	}
}

func (this *PacketReader) ReadSimple(packetType uint8, layers *PacketLayers, packet *UDPPacket) {
	var err error
	decoder := PacketDecoders[packetType]
	if decoder != nil {
		layers.Main, err = decoder(packet, this.Context)
		if err != nil {
			this.ErrorHandler(errors.New(fmt.Sprintf("Failed to decode simple packet %02X: %s", packetType, err.Error())))
			return
		}
	}

	this.SimpleHandler(packetType, packet, layers)
}

func (this *PacketReader) ReadGeneric(packetType uint8, layers *PacketLayers, packet *UDPPacket) {
	var err error
	if packetType == 0x1B {
		tsLayer, err := DecodePacket1BLayer(packet, this.Context)
		if err != nil {
			this.ErrorHandler(errors.New(fmt.Sprintf("Failed to decode timestamped packet: %s", err.Error())))
			return
		}
		layers.Timestamp = tsLayer.(*Packet1BLayer)
		packetType, err = packet.Stream.ReadByte()
		if err != nil {
			this.ErrorHandler(errors.New(fmt.Sprintf("Failed to decode timestamped packet: %s", err.Error())))
			return
		}
		layers.Reliability.PacketType = packetType
		layers.Reliability.HasPacketType = true
	}
	if packetType == 0x8A {
		println("will read ", layers.Reliability.LengthInBits)
		data, err := packet.Stream.ReadString(int((layers.Reliability.LengthInBits+7)/8) - 1)
		if err != nil {
			println("failed while reading packet8a")
			this.ErrorHandler(errors.New(fmt.Sprintf("Failed to decode reliable packet %02X: %s", packetType, err.Error())))

			return
		}
		layers.Main, err = DecodePacket8ALayer(packet, this.Context, data)

		if err != nil {
			println("decode fail")
			this.ErrorHandler(errors.New(fmt.Sprintf("Failed to decode reliable packet %02X: %s", packetType, err.Error())))

			return
		}
	} else {
		decoder := PacketDecoders[packetType]
		if decoder != nil {
			layers.Main, err = decoder(packet, this.Context)

			if err != nil {
				this.ErrorHandler(errors.New(fmt.Sprintf("Failed to decode reliable packet %02X: %s", packetType, err.Error())))

				return
			}
		}
	}

	this.FullReliableHandler(packetType, packet, layers)
}

func (this *PacketReader) ReadOrdered(layers *PacketLayers, packet *UDPPacket) {
	var err error
	subPacket := layers.Reliability
	if subPacket.HasPacketType && !subPacket.HasBeenDecoded && subPacket.IsFinal {
		packetType := subPacket.PacketType
		_, err = subPacket.FullDataReader.ReadByte()
		if err != nil {
			this.ErrorHandler(errors.New(fmt.Sprintf("Failed to decode reliablePacket %02X: %s", packetType, err.Error())))
			return
		}
		newPacket := &UDPPacket{subPacket.FullDataReader, packet.Source, packet.Destination}

		this.ReadGeneric(packetType, layers, newPacket)
	}
}

func (this *PacketReader) ReadReliable(layers *PacketLayers, packet *UDPPacket) {
	packet.Stream = layers.RakNet.Payload
	reliabilityLayer, err := DecodeReliabilityLayer(packet, this.Context, layers.RakNet)
	if err != nil {
		this.ErrorHandler(errors.New("Failed to decode reliable packet: " + err.Error()))
		return
	}

	isClient := this.Context.IsClient(packet.Source)
	var queues *Queues

	if isClient {
		queues = this.ClientQueues
	} else {
		queues = this.ServerQueues
	}

	this.ReliabilityLayerHandler(packet, reliabilityLayer, layers.RakNet)
	for _, subPacket := range reliabilityLayer.Packets {
		reliablePacketLayers := &PacketLayers{RakNet: layers.RakNet, Reliability: subPacket}
		this.ReliableHandler(subPacket.PacketType, packet, reliablePacketLayers)
		queues.Add(reliablePacketLayers)
		if reliablePacketLayers.Reliability.Reliability == 0 {
			this.ReadOrdered(reliablePacketLayers, packet)
			queues.Remove(reliablePacketLayers)
			continue
		}

		reliablePacketLayers = queues.Next(subPacket.OrderingChannel)
		for reliablePacketLayers != nil {
			this.ReadOrdered(reliablePacketLayers, packet)
			queues.Remove(reliablePacketLayers)
			reliablePacketLayers = queues.Next(subPacket.OrderingChannel)
		}
	}
}

func (this *PacketReader) ReadPacket(payload []byte, packet *UDPPacket) {
	context := this.Context

	packet.Stream = BufferToStream(payload)
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
		this.ReadSimple(packetType, layers, packet)
		return
	} else if !rakNetLayer.IsValid {
		this.ErrorHandler(errors.New(fmt.Sprintf("Sent invalid packet (packet header %x)", payload[0])))
		return
	} else if rakNetLayer.IsACK || rakNetLayer.IsNAK {
		this.ACKHandler(packet, rakNetLayer)
	} else {
		this.ReadReliable(layers, packet)
		return
	}
}
