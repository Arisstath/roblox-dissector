package peer

import "errors"
import "fmt"

import "strings"
import "log"

type decoderFunc func(*extendedReader, PacketReader, *PacketLayers) (RakNetPacket, error)

var packetDecoders = map[byte]decoderFunc{
	0x05: (*extendedReader).DecodePacket05Layer,
	0x06: (*extendedReader).DecodePacket06Layer,
	0x07: (*extendedReader).DecodePacket07Layer,
	0x08: (*extendedReader).DecodePacket08Layer,
	0x00: (*extendedReader).DecodePacket00Layer,
	0x03: (*extendedReader).DecodePacket03Layer,
	0x09: (*extendedReader).DecodePacket09Layer,
	0x10: (*extendedReader).DecodePacket10Layer,
	0x13: (*extendedReader).DecodePacket13Layer,
	0x15: (*extendedReader).DecodePacket15Layer,
	0x1B: (*extendedReader).DecodePacket1BLayer,

	0x81: (*extendedReader).DecodePacket81Layer,
	//0x82: DecodePacket82Layer,
	0x83: (*extendedReader).DecodePacket83Layer,
	0x84: (*extendedReader).DecodePacket84Layer,
	0x85: (*extendedReader).DecodePacket85Layer,
	0x86: (*extendedReader).DecodePacket86Layer,
	0x8A: (*extendedReader).DecodePacket8ALayer,
	0x8D: (*extendedReader).DecodePacket8DLayer,
	0x8F: (*extendedReader).DecodePacket8FLayer,
	0x90: (*extendedReader).DecodePacket90Layer,
	0x92: (*extendedReader).DecodePacket92Layer,
	0x93: (*extendedReader).DecodePacket93Layer,
	0x96: (*extendedReader).DecodePacket96Layer,
	0x97: (*extendedReader).DecodePacket97Layer,
}

type ContextualHandler interface {
	SetContext(*CommunicationContext)
	Context() *CommunicationContext
	SetCaches(*Caches)
	Caches() *Caches
}

// PacketReader is an interface that can be passed to packet decoders
type PacketReader interface {
	ContextualHandler
	SetIsClient(bool)
	IsClient() bool
}

type contextualHandler struct {
	context *CommunicationContext
	caches  *Caches
}

func (handler *contextualHandler) Context() *CommunicationContext {
	return handler.context
}
func (handler *contextualHandler) Caches() *Caches {
	return handler.caches
}
func (handler *contextualHandler) SetCaches(val *Caches) {
	handler.caches = val
}
func (handler *contextualHandler) SetContext(val *CommunicationContext) {
	handler.context = val
}

// PacketReader is a struct that can be used to read packets from a source
// Pass packets in using ReadPacket() and bind to the given callbacks
// to receive the results
type DefaultPacketReader struct {
	contextualHandler
	// Callback for "simple" packets (pre-connection offline packets).
	SimpleHandler RawPacketHandlerMap
	// Callback for ReliabilityLayer subpackets. This callback is invoked for every
	// split of every packets, possible duplicates, etc.
	ReliableHandler RawPacketHandlerMap
	// Callback for generic packets (anything that is sent when a connection has been
	// established. You definitely want to bind to this.
	FullReliableHandler RawPacketHandlerMap
	// Callback for ACKs and NAKs. Not very useful if you're just parsing packets.
	// However, you want to bind to this if you are writing a peer.
	// ACKHandler uses a packet type of 0, but you can just use BindAll()
	ACKHandler RawPacketHandlerMap
	// Callback for ReliabilityLayer full packets. This callback is invoked for every
	// real ReliabilityLayer.
	// ReliabilityLayerHandler uses a packet type of 0, but you can just use BindAll()
	ReliabilityLayerHandler RawPacketHandlerMap

	DataHandler DataPacketHandlerMap
	isClient    bool

	queues       *queues
	splitPackets splitPacketList
}

func (reader *DefaultPacketReader) IsClient() bool {
	return reader.isClient
}
func (reader *DefaultPacketReader) SetIsClient(val bool) {
	reader.isClient = val
}

func NewPacketReader() *DefaultPacketReader {
	reader := &DefaultPacketReader{
		queues:                  newPeerQueues(),
		SimpleHandler:           NewRawPacketHandlerMap(),
		ACKHandler:              NewRawPacketHandlerMap(),
		FullReliableHandler:     NewRawPacketHandlerMap(),
		ReliabilityLayerHandler: NewRawPacketHandlerMap(),
		ReliableHandler:         NewRawPacketHandlerMap(),
		DataHandler:             NewDataHandlerMap(),
	}

	reader.FullReliableHandler.Bind(0x83, func(_ uint8, layers *PacketLayers) {
		for _, sub := range layers.Main.(*Packet83Layer).SubPackets {
			reader.DataHandler.Fire(sub.Type(), sub)
		}
	})

	return reader
}

func (reader *DefaultPacketReader) readSimple(stream *extendedReader, packetType uint8, layers *PacketLayers) {
	var err error
	layers.Root.logBuffer = new(strings.Builder)
	layers.Root.Logger = log.New(layers.Root.logBuffer, "", log.Lmicroseconds|log.Ltime)
	decoder := packetDecoders[packetType]
	if decoder != nil {
		layers.Main, err = decoder(stream, reader, layers)
		if err != nil {
			layers.Error = fmt.Errorf("Failed to decode simple packet %02X: %s", packetType, err.Error())
		}
	}

	reader.SimpleHandler.Fire(packetType, layers)
}

func (reader *DefaultPacketReader) readGeneric(stream *extendedReader, packetType uint8, layers *PacketLayers) {
	var err error
	if packetType == 0x1B { // ID_TIMESTAMP
		tsLayer, err := packetDecoders[0x1B](stream, reader, layers)
		if err != nil {
			println("timestamp fail")
			layers.Reliability.SplitBuffer.Logger.Println("error:", err.Error())
			layers.Error = fmt.Errorf("Failed to decode timestamped packet: %s", err.Error())
			return
		}
		layers.Timestamp = tsLayer.(*Packet1BLayer)
		packetType, err = stream.ReadByte()
		if err != nil {
			println("timestamp type fail")
			layers.Reliability.SplitBuffer.Logger.Println("error:", err.Error())
			layers.Error = fmt.Errorf("Failed to decode timestamped packet: %s", err.Error())
			return
		}
		layers.Reliability.SplitBuffer.PacketType = packetType
		layers.Reliability.SplitBuffer.HasPacketType = true
	}
	decoder := packetDecoders[layers.Reliability.SplitBuffer.PacketType]
	// TODO: Should we really void partial deserializations?
	if decoder != nil {
		layers.Main, err = decoder(stream, reader, layers)

		if err != nil {
			println("parser error, setting main to nil", err.Error())
			layers.Main = nil
			layers.Reliability.SplitBuffer.Logger.Println("error:", err.Error())
			layers.Error = fmt.Errorf("Failed to decode reliable packet %02X: %s", layers.Reliability.SplitBuffer.PacketType, err.Error())
		}
	}
}

func (reader *DefaultPacketReader) readOrdered(layers *PacketLayers) {
	var err error
	subPacket := layers.Reliability
	buffer := subPacket.SplitBuffer
	if buffer.IsFinal {
		var packetType uint8
		packetType, err = buffer.dataReader.ReadByte()
		if err != nil {
			println("ouch, my packetType", err.Error())
			subPacket.SplitBuffer.Logger.Println("error:", err.Error())
			layers.Error = fmt.Errorf("Failed to decode reliablePacket type %d: %s", packetType, err.Error())
		} else {
			reader.readGeneric(buffer.dataReader, packetType, layers)
		}

		// fullreliablehandler, regardless of whether the parsing succeeded or not!
		// this is because readGeneric will have set the Error and Main layers accordingly
		reader.FullReliableHandler.Fire(layers.Reliability.SplitBuffer.PacketType, layers)
	}
}

func (reader *DefaultPacketReader) readReliable(layers *PacketLayers) {
	stream := layers.RakNet.payload
	reliabilityLayer, err := stream.DecodeReliabilityLayer(reader, layers)
	if err != nil {
		layers.Error = errors.New("Failed to decode reliable packet: " + err.Error())
		reader.ReliabilityLayerHandler(0, layers)
		return
	}

	queues := reader.queues

	reader.ReliabilityLayerHandler(0, layers)
	for _, subPacket := range reliabilityLayer.Packets {
		reliablePacketLayers := &PacketLayers{Root: layers.Root, RakNet: layers.RakNet, Reliability: subPacket}

		buffer, err := reader.handleSplitPacket(reliablePacketLayers)
		subPacket.SplitBuffer = buffer
		if err != nil {
			println("error while handling split")
			subPacket.SplitBuffer.Logger.Println("error while handling split:", err.Error())
			reliablePacketLayers.Error = fmt.Errorf("Error while handling split packet: %s", err.Error())
			reader.ReliableHandler.Fire(buffer.PacketType, reliablePacketLayers)
			return
		}

		reader.ReliableHandler.Fire(buffer.PacketType, reliablePacketLayers)
		queues.add(reliablePacketLayers)
		if reliablePacketLayers.Reliability.Reliability == 0 {
			reader.readOrdered(reliablePacketLayers)
			queues.remove(reliablePacketLayers)
			// We can skip the code below: unreliable packets can't have released
			// any pending packets that are on the queue
			continue
		}

		reliablePacketLayers = queues.next(subPacket.OrderingChannel)
		for reliablePacketLayers != nil {
			reader.readOrdered(reliablePacketLayers)
			queues.remove(reliablePacketLayers)
			reliablePacketLayers = queues.next(subPacket.OrderingChannel)
		}
	}
}

// ReadPacket reads a single packet and invokes all according handler functions
func (reader *DefaultPacketReader) ReadPacket(payload []byte, layers *PacketLayers) {
	stream := bufferToStream(payload)
	rakNetLayer, err := stream.DecodeRakNetLayer(reader, payload[0], layers)
	if err != nil {
		layers.Error = err
		reader.SimpleHandler.Fire(payload[0], layers)
		return
	}

	layers.RakNet = rakNetLayer
	if rakNetLayer.IsSimple {
		packetType := rakNetLayer.SimpleLayerID
		reader.readSimple(stream, packetType, layers)
	} else if !rakNetLayer.IsValid {
		layers.Error = fmt.Errorf("Sent invalid packet (packet header %x)", payload[0])
		reader.SimpleHandler.Fire(payload[0], layers)
	} else if rakNetLayer.IsACK || rakNetLayer.IsNAK {
		reader.ACKHandler.Fire(0, layers)
	} else {
		reader.readReliable(layers)
	}
}

// Deletion handler
func (reader *DefaultPacketReader) HandlePacket01(packet *Packet83_01) error {
	return packet.Instance.SetParent(nil)
}

// New instance handler
func (reader *DefaultPacketReader) HandlePacket02(packet *Packet83_02) error {
	return packet.Instance.SetParent(packet.Parent)
}

// Prop update handler
func (reader *DefaultPacketReader) HandlePacket03(packet *Packet83_03) error {
	packet.Instance.Set(packet.Schema.Name, packet.Value)
	return nil
}

// event handler
func (reader *DefaultPacketReader) HandlePacket07(packet *Packet83_07) error {
	packet.Instance.FireEvent(packet.Schema.Name, packet.Event.Arguments)
	return nil
}

// Joindata handler
func (reader *DefaultPacketReader) HandlePacket0B(packet *Packet83_0B) error {
	for _, inst := range packet.Instances {
		err := inst.Instance.SetParent(inst.Parent)
		if err != nil {
			return err
		}
	}
	return nil
}

// Top replic handler
func (reader *DefaultPacketReader) HandlePacket81(packet Packet81Layer) error {
	for _, item := range packet.Items {
		err := reader.context.DataModel.AddService(item.Instance)
		if err != nil {
			return err
		}
	}
	return nil
}

func (reader *DefaultPacketReader) BindDefaultHandlers() {
	reader.FullReliableHandler.Bind(0x81, reader.HandlePacket81)
	reader.DataHandler.Bind(1, reader.HandlePacket01)
	reader.DataHandler.Bind(2, reader.HandlePacket02)
	reader.DataHandler.Bind(3, reader.HandlePacket03)
	reader.DataHandler.Bind(7, reader.HandlePacket07)
	reader.DataHandler.Bind(0xB, reader.HandlePacket0B)
}
