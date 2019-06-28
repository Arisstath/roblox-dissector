package peer

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/Gskartwii/roblox-dissector/datamodel"
	"github.com/robloxapi/rbxfile"

	"github.com/olebedev/emitter"
)

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

// ContextualHandler is a generic interface for structs that
// provide a CommunicationContext and Caches
type ContextualHandler interface {
	SetContext(*CommunicationContext)
	Context() *CommunicationContext
	SetCaches(*Caches)
	Caches() *Caches
	SharedStrings() map[string]rbxfile.ValueSharedString
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
	// sharedStrings contains a map of deferred strings indexed by their MD5 hash
	sharedStrings map[string]rbxfile.ValueSharedString
}

func (handler *contextualHandler) Context() *CommunicationContext {
	return handler.context
}
func (handler *contextualHandler) Caches() *Caches {
	return handler.caches
}
func (handler *contextualHandler) SharedStrings() map[string]rbxfile.ValueSharedString {
	return handler.sharedStrings
}

func (handler *contextualHandler) SetCaches(val *Caches) {
	handler.caches = val
}
func (handler *contextualHandler) SetContext(val *CommunicationContext) {
	handler.context = val
}

// DefaultPacketReader is a struct that can be used to read packets from a source
// Pass packets in using ReadPacket() and bind to the given callbacks
// to receive the results
type DefaultPacketReader struct {
	contextualHandler
	// LayerEmitter provides a low-level interface for receiving packets
	// Topics: full-reliable, offline, reliable, reliability, ack
	LayerEmitter *emitter.Emitter
	// ErrorEmitter is the same as LayerEmitter, except invoked when layers.Error != nil ErrorEmitter *emitter.Emitter
	ErrorEmitter *emitter.Emitter

	// PacketEmitter provides a high-level interface for receiving offline and reliable packets
	// Topics correspond to TypeString() return values
	PacketEmitter *emitter.Emitter

	// DataEmitter provides a high-level interface for receiving ID_DATA subpackets
	// These topics correspond to TypeString() return values
	DataEmitter *emitter.Emitter

	isClient bool

	rmState      *reliableMessageState
	sqState      *sequenceState
	ordQueue     *orderingQueue
	splitPackets splitPacketList
}

// IsClient implements PacketReader.IsClient()
func (reader *DefaultPacketReader) IsClient() bool {
	return reader.isClient
}

// SetIsClient implements PacketReader.SetIsClient()
func (reader *DefaultPacketReader) SetIsClient(val bool) {
	reader.isClient = val
}

// NewPacketReader initializes a new DefaultPacketReader
func NewPacketReader() *DefaultPacketReader {
	var thisQ [32]map[uint32]*PacketLayers
	for i := 0; i < 32; i++ {
		thisQ[i] = make(map[uint32]*PacketLayers)
	}

	reader := &DefaultPacketReader{
		LayerEmitter:  emitter.New(0),
		ErrorEmitter:  emitter.New(0),
		PacketEmitter: emitter.New(0),
		DataEmitter:   emitter.New(0),
		contextualHandler: contextualHandler{
			caches:        new(Caches),
			sharedStrings: make(map[string]rbxfile.ValueSharedString),
		},

		rmState: &reliableMessageState{
			hasHandled: make(map[uint32]bool),
		},
		sqState: &sequenceState{},
		ordQueue: &orderingQueue{
			queue: thisQ,
		},
	}

	reader.bindBasicPacketHandler()
	reader.bindDataPacketHandler()

	return reader
}

func (reader *DefaultPacketReader) emitLayers(topic string, layers *PacketLayers) {
	if layers.Error != nil {
		<-reader.ErrorEmitter.Emit(topic, layers)
	} else {
		<-reader.LayerEmitter.Emit(topic, layers)
	}
}

func (reader *DefaultPacketReader) readOffline(stream *extendedReader, packetType uint8, layers *PacketLayers) {
	var err error
	layers.Root.logBuffer = new(strings.Builder)
	layers.Root.Logger = log.New(layers.Root.logBuffer, "", log.Lmicroseconds|log.Ltime)
	decoder := packetDecoders[packetType]
	if decoder != nil {
		layers.Main, err = decoder(stream, reader, layers)
		if err != nil {
			layers.Error = fmt.Errorf("Failed to decode offline packet %02X: %s", packetType, err.Error())
		}
	}

	reader.emitLayers("offline", layers)
}

func (reader *DefaultPacketReader) readGeneric(stream *extendedReader, layers *PacketLayers) {
	var err error
	if layers.PacketType == 0x1B { // ID_TIMESTAMP
		tsLayer, err := packetDecoders[0x1B](stream, reader, layers)
		if err != nil {
			layers.Reliability.SplitBuffer.Logger.Println("error:", err.Error())
			layers.Error = fmt.Errorf("Failed to decode timestamped packet: %s", err.Error())
			return
		}
		layers.Timestamp = tsLayer.(*Packet1BLayer)
		packetType, err := stream.ReadByte()
		if err != nil {
			layers.Reliability.SplitBuffer.Logger.Println("error:", err.Error())
			layers.Error = fmt.Errorf("Failed to decode timestamped packet: %s", err.Error())
			return
		}
		layers.Reliability.SplitBuffer.PacketType = packetType
		layers.Reliability.SplitBuffer.HasPacketType = true
		layers.PacketType = packetType
	}
	decoder := packetDecoders[layers.PacketType]
	// TODO: Should we really void partial deserializations?
	if decoder != nil {
		layers.Main, err = decoder(stream, reader, layers)

		if err != nil {
			layers.Main = nil
			layers.Reliability.SplitBuffer.Logger.Println("error:", err.Error())
			layers.Error = fmt.Errorf("Failed to decode reliable packet %02X: %s", layers.PacketType, err.Error())
		}
	} else {
		layers.Error = fmt.Errorf("Unknown packetType %d", layers.PacketType)
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
			subPacket.SplitBuffer.Logger.Println("error:", err.Error())
			layers.Error = fmt.Errorf("Failed to decode reliablePacket type %d: %s", packetType, err.Error())
		} else {
			layers.PacketType = packetType
			reader.readGeneric(buffer.dataReader, layers)
			byteReader := buffer.byteReader

			// The parser "successfully" parsed a packet, but there is still data remaining?
			if byteReader.Len() != 0 && layers.Error == nil {
				layers.Error = fmt.Errorf("parsed packet %02X but still have %d bytes remaining", layers.PacketType, byteReader.Len())
			}
		}

		// fullreliablehandler, regardless of whether the parsing succeeded or not!
		// this is because readGeneric will have set the Error and Main layers accordingly
		reader.emitLayers("full-reliable", layers)
	}
}

func (reader *DefaultPacketReader) readReliable(layers *PacketLayers) {
	stream := layers.RakNet.payload
	reliabilityLayer, err := stream.DecodeReliabilityLayer(reader, layers)
	if err != nil {
		layers.Error = errors.New("Failed to decode reliable packet: " + err.Error())
		reader.emitLayers("reliability", layers)
		return
	}

	reader.emitLayers("reliability", layers)
	for _, subPacket := range reliabilityLayer.Packets {
		reliablePacketLayers := &PacketLayers{Root: layers.Root, RakNet: layers.RakNet, Reliability: subPacket}

		buffer, err := reader.handleSplitPacket(reliablePacketLayers)
		reliablePacketLayers.SplitPacket = buffer
		subPacket.SplitBuffer = buffer
		reliablePacketLayers.PacketType = buffer.PacketType

		if err != nil {
			subPacket.SplitBuffer.Logger.Println("error while handling split:", err.Error())
			reliablePacketLayers.Error = fmt.Errorf("Error while handling split packet: %s", err.Error())
			reader.emitLayers("reliable", reliablePacketLayers)
			return
		}

		reader.emitLayers("reliable", reliablePacketLayers)
		thisRelPacket := reliablePacketLayers.Reliability
		hasHandled := reader.rmState.hasHandled
		switch thisRelPacket.Reliability {
		case Unreliable:
			reader.readOrdered(reliablePacketLayers)
		case Reliable:
			if !hasHandled[thisRelPacket.ReliableMessageNumber] {
				hasHandled[thisRelPacket.ReliableMessageNumber] = true
				reader.readOrdered(reliablePacketLayers)
			}
		case ReliableSequenced:
			if !hasHandled[thisRelPacket.ReliableMessageNumber] {
				hasHandled[thisRelPacket.ReliableMessageNumber] = true
				if reader.sqState.highestIndex >= thisRelPacket.SequencingIndex {
					reader.sqState.highestIndex = thisRelPacket.SequencingIndex
					reader.readOrdered(reliablePacketLayers)
				}
			}
		case ReliableOrdered:
			if !hasHandled[thisRelPacket.ReliableMessageNumber] {
				hasHandled[thisRelPacket.ReliableMessageNumber] = true
				reader.ordQueue.add(reliablePacketLayers)

				reliablePacketLayers = reader.ordQueue.next(subPacket.OrderingChannel)
				for reliablePacketLayers != nil {
					reader.readOrdered(reliablePacketLayers)
					reliablePacketLayers = reader.ordQueue.next(subPacket.OrderingChannel)
				}
			}
		default:
			reliablePacketLayers.Error = fmt.Errorf("Unknown reliability: %d", reliablePacketLayers.Reliability.Reliability)
			// TODO: Is it legal to emit the reliable packet twice?
			reader.emitLayers("reliable", reliablePacketLayers)
			return
		}
	}
}

// ReadPacket reads a single packet and invokes all according handler functions
func (reader *DefaultPacketReader) ReadPacket(payload []byte, layers *PacketLayers) {
	var err error
	if IsOfflineMessage(payload) {
		layers.OfflinePayload = payload

		layers.PacketType = payload[0]
		layers.OfflinePayload = payload
		reader.readOffline(bufferToStream(payload[1+0x10:]), layers.PacketType, layers)
		return
	}

	stream := bufferToStream(payload)

	// TODO: Should not create RakNetLayer if it's an offline message
	rakNetLayer, err := stream.DecodeRakNetLayer(reader, payload[0], layers)
	layers.PacketType = payload[0]
	if err != nil {
		layers.Error = err
		reader.emitLayers("offline", layers)
		return
	}
	layers.RakNet = rakNetLayer
	if rakNetLayer.Flags.IsACK || rakNetLayer.Flags.IsNAK {
		reader.emitLayers("ack", layers)
	} else {
		reader.readReliable(layers)
	}
}

// HandlePacket01 is the default handler for ID_REPLIC_DELETE_INSTANCE packets
func (reader *DefaultPacketReader) HandlePacket01(e *emitter.Event) {
	packet := e.Args[0].(*Packet83_01)
	err := packet.Instance.SetParent(nil)
	if err != nil {
		e.Args[1].(*PacketLayers).Root.Logger.Println("delete error:", err.Error())
	}
}

func (reader *DefaultPacketReader) handleReplicationInstance(inst *ReplicationInstance) error {
	// First, assign the properties
	inst.Instance.PropertiesMutex.Lock()
	for name, val := range inst.Properties {
		// Do not call Set() here. Nothing should be listening to PropertyEmitter
		// and we have the lock
		inst.Instance.Properties[name] = val
	}
	inst.Instance.PropertiesMutex.Unlock()

	// Once they are assigned, we can release this instance to be used by the DataModel
	return inst.Instance.SetParent(inst.Parent)
}

// HandlePacket02 is the default handler for ID_REPLIC_NEW_INSTANCE packets
func (reader *DefaultPacketReader) HandlePacket02(e *emitter.Event) {
	packet := e.Args[0].(*Packet83_02)

	err := reader.handleReplicationInstance(packet.ReplicationInstance)
	if err != nil {
		e.Args[1].(*PacketLayers).Error = err
	}
}

// HandlePacket03 is the default handler for ID_REPLIC_PROP packets
func (reader *DefaultPacketReader) HandlePacket03(e *emitter.Event) {
	packet := e.Args[0].(*Packet83_03)
	if packet.Schema == nil {
		// Parent handler
		err := packet.Value.(datamodel.ValueReference).Instance.AddChild(packet.Instance)
		if err != nil {
			e.Args[1].(*PacketLayers).Error = err
		}
		return
	}
	packet.Instance.Set(packet.Schema.Name, packet.Value)
}

// HandlePacket07 is the default handler fo ID_REPLIC_EVENT packets
func (reader *DefaultPacketReader) HandlePacket07(e *emitter.Event) {
	packet := e.Args[0].(*Packet83_07)
	packet.Instance.FireEvent(packet.Schema.Name, packet.Event.Arguments...)
}

// HandlePacket0B is the default handler for ID_REPLIC_JOIN_DATA packets
func (reader *DefaultPacketReader) HandlePacket0B(e *emitter.Event) {
	packet := e.Args[0].(*Packet83_0B)
	for _, inst := range packet.Instances {
		err := reader.handleReplicationInstance(inst)
		if err != nil {
			e.Args[1].(*PacketLayers).Error = err
			return
		}
	}
}

// HandlePacket13 is the default handler for ID_REPLIC_ATOMIC packets
func (reader *DefaultPacketReader) HandlePacket13(e *emitter.Event) {
	packet := e.Args[0].(*Packet83_13)
	packet.Instance.SetParent(packet.Parent)
}

// HandlePacket81 is the default handler for ID_SET_GLOBALS packets
func (reader *DefaultPacketReader) HandlePacket81(e *emitter.Event) {
	packet := e.Args[0].(*Packet81Layer)
	for _, item := range packet.Items {
		reader.context.DataModel.AddService(item.Instance)
	}
}

// BindDataModelHandlers binds the default handlers so that the PacketReader
// will update the DataModel based on what it reads
func (reader *DefaultPacketReader) BindDataModelHandlers() {
	reader.PacketEmitter.On("ID_SET_GLOBALS", reader.HandlePacket81, emitter.Void)
	reader.DataEmitter.On("ID_REPLIC_DELETE_INSTANCE", reader.HandlePacket01, emitter.Void)
	reader.DataEmitter.On("ID_REPLIC_NEW_INSTANCE", reader.HandlePacket02, emitter.Void)
	reader.DataEmitter.On("ID_REPLIC_PROP", reader.HandlePacket03, emitter.Void)
	reader.DataEmitter.On("ID_REPLIC_EVENT", reader.HandlePacket07, emitter.Void)
	reader.DataEmitter.On("ID_REPLIC_JOIN_DATA", reader.HandlePacket0B, emitter.Void)
	reader.DataEmitter.On("ID_REPLIC_ATOMIC", reader.HandlePacket13, emitter.Void)
}

func (reader *DefaultPacketReader) bindBasicPacketHandler() {
	// important: sync!
	reader.LayerEmitter.On("full-reliable", func(e *emitter.Event) {
		layers := e.Args[0].(*PacketLayers)
		<-reader.PacketEmitter.Emit(layers.Main.TypeString(), layers.Main, layers)
	}, emitter.Void)
	reader.LayerEmitter.On("offline", func(e *emitter.Event) {
		layers := e.Args[0].(*PacketLayers)
		<-reader.PacketEmitter.Emit(layers.Main.TypeString(), layers.Main, layers)
	}, emitter.Void)
}

func (reader *DefaultPacketReader) bindDataPacketHandler() {
	// important: sync!
	reader.PacketEmitter.On("ID_DATA", func(e *emitter.Event) {
		layers := e.Args[1].(*PacketLayers)
		for _, sub := range layers.Main.(*Packet83Layer).SubPackets {
			subLayers := &PacketLayers{
				Root:        layers.Root,
				RakNet:      layers.RakNet,
				Reliability: layers.Reliability,
				SplitPacket: layers.SplitPacket,
				Timestamp:   layers.Timestamp,
				Main:        layers.Main,
				Error:       layers.Error,
				PacketType:  sub.Type(),
				Subpacket:   sub,
			}
			<-reader.DataEmitter.Emit(sub.TypeString(), sub, subLayers)
		}
	}, emitter.Void)
}

// Layers returns the emitter for successfully parsed packets
func (reader *DefaultPacketReader) Layers() *emitter.Emitter {
	return reader.LayerEmitter
}

// Errors returns the emitter for parser errors
func (reader *DefaultPacketReader) Errors() *emitter.Emitter {
	return reader.ErrorEmitter
}
