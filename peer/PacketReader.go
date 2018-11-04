package peer

import "errors"
import "fmt"

import "strings"
import "log"

type decoderFunc func(PacketReader, *UDPPacket) (RakNetPacket, error)

var packetDecoders = map[byte]decoderFunc{
	0x05: DecodePacket05Layer,
	0x06: DecodePacket06Layer,
	0x07: DecodePacket07Layer,
	0x08: DecodePacket08Layer,
	0x00: DecodePacket00Layer,
	0x03: DecodePacket03Layer,
	0x09: DecodePacket09Layer,
	0x10: DecodePacket10Layer,
	0x13: DecodePacket13Layer,
	0x15: DecodePacket15Layer,
	0x1B: DecodePacket1BLayer,

	0x81: DecodePacket81Layer,
	//0x82: DecodePacket82Layer,
	0x83: DecodePacket83Layer,
	0x85: DecodePacket85Layer,
	0x86: DecodePacket86Layer,
	0x8D: DecodePacket8DLayer,
	0x8F: DecodePacket8FLayer,
	0x90: DecodePacket90Layer,
	0x92: DecodePacket92Layer,
	0x93: DecodePacket93Layer,
	0x97: DecodePacket97Layer,
}

type ReceiveHandler func(byte, *UDPPacket, *PacketLayers)
type ErrorHandler func(error, *UDPPacket)

// PacketReader is an interface that can be passed to packet decoders
type PacketReader interface {
	Context() *CommunicationContext
	Caches() *Caches
	IsClient() bool
}

// PacketReader is a struct that can be used to read packets from a source
// Pass packets in using ReadPacket() and bind to the given callbacks
// to receive the results
type DefaultPacketReader struct {
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
	ValContext  *CommunicationContext
	ValCaches   *Caches
	ValIsClient bool

	SkipParsing map[byte]struct{}

	queues *queues
}

func (reader *DefaultPacketReader) Context() *CommunicationContext {
	return reader.ValContext
}
func (reader *DefaultPacketReader) Caches() *Caches {
	return reader.ValCaches
}
func (reader *DefaultPacketReader) IsClient() bool {
	return reader.ValIsClient
}

func NewPacketReader() *DefaultPacketReader {
	return &DefaultPacketReader{
		queues:      newPeerQueues(),
		SkipParsing: map[byte]struct{}{
			//0x85: struct{}{}, // Skip physics! they don't work very well
		},
	}
}

func (reader *DefaultPacketReader) readSimple(packetType uint8, layers *PacketLayers, packet *UDPPacket) {
	var err error
	decoder := packetDecoders[packetType]
	_, skip := reader.SkipParsing[packetType]
	if decoder != nil && !skip {
		layers.Main, err = decoder(reader, packet)
		if err != nil {
			reader.ErrorHandler(errors.New(fmt.Sprintf("Failed to decode simple packet %02X: %s", packetType, err.Error())), packet)
			return
		}
	}
	packet.logBuffer = new(strings.Builder)
	packet.Logger = log.New(packet.logBuffer, "", log.Lmicroseconds|log.Ltime)

	reader.SimpleHandler(packetType, packet, layers)
}

func (reader *DefaultPacketReader) readGeneric(packetType uint8, layers *PacketLayers, packet *UDPPacket) {
	var err error
	if packetType == 0x1B { // ID_TIMESTAMP
		tsLayer, err := packetDecoders[0x1B](reader, packet)
		if err != nil {
			layers.Reliability.SplitBuffer.Logger.Println("error:", err.Error())
			reader.ErrorHandler(errors.New(fmt.Sprintf("Failed to decode timestamped packet: %s", err.Error())), packet)
			return
		}
		layers.Timestamp = tsLayer.(*Packet1BLayer)
		packetType, err = packet.stream.ReadByte()
		if err != nil {
			layers.Reliability.SplitBuffer.Logger.Println("error:", err.Error())
			reader.ErrorHandler(errors.New(fmt.Sprintf("Failed to decode timestamped packet: %s", err.Error())), packet)
			return
		}
		layers.Reliability.SplitBuffer.PacketType = packetType
		layers.Reliability.SplitBuffer.HasPacketType = true
	}
	if packetType == 0x8A { // ID_ROBLOX_AUTH -- ID_SUBMIT_TICKET
		data, err := packet.stream.readString(int((layers.Reliability.LengthInBits+7)/8) - 1)
		if err != nil {
			layers.Reliability.SplitBuffer.Logger.Println("error:", err.Error())
			reader.ErrorHandler(errors.New(fmt.Sprintf("Failed to decode reliable packet %02X: %s", packetType, err.Error())), packet)

			return
		}
		_, skip := reader.SkipParsing[0x8A]
		if !skip {
			layers.Main, err = DecodePacket8ALayer(reader, packet, data)

			if err != nil {
				layers.Main = nil
				layers.Reliability.SplitBuffer.Logger.Println("error:", err.Error())
				reader.ErrorHandler(errors.New(fmt.Sprintf("Failed to decode reliable packet %02X: %s", packetType, err.Error())), packet)

				return
			}
		}
	} else {
		decoder := packetDecoders[layers.Reliability.SplitBuffer.PacketType]
		_, skip := reader.SkipParsing[layers.Reliability.SplitBuffer.PacketType]
		if decoder != nil && !skip {
			layers.Main, err = decoder(reader, packet)

			if err != nil {
				layers.Main = nil
				layers.Reliability.SplitBuffer.Logger.Println("error:", err.Error())
				reader.ErrorHandler(fmt.Errorf("Failed to decode reliable packet %02X: %s", layers.Reliability.SplitBuffer.PacketType, err.Error()), packet)

				return
			}
		}
	}
}

func (reader *DefaultPacketReader) readOrdered(layers *PacketLayers, packet *UDPPacket) {
	var err error
	subPacket := layers.Reliability
	buffer := subPacket.SplitBuffer
	if !buffer.HasBeenDecoded && buffer.IsFinal {
		var packetType uint8
		packetType, err = buffer.dataReader.ReadByte()
		if err != nil {
			subPacket.SplitBuffer.Logger.Println("error:", err.Error())
			reader.ErrorHandler(errors.New(fmt.Sprintf("Failed to decode reliablePacket type: %s", packetType, err.Error())), packet)
			return
		}
		newPacket := &UDPPacket{}
		newPacket.stream = buffer.dataReader
		newPacket.Source = packet.Source
		newPacket.Destination = packet.Destination
		newPacket.logBuffer = buffer.logBuffer
		newPacket.Logger = buffer.Logger

		reader.readGeneric(packetType, layers, newPacket)
		reader.FullReliableHandler(layers.Reliability.SplitBuffer.PacketType, newPacket, layers) // fullreliablehandler, regardless of whether the parsing succeeded or not!
	}
}

func (reader *DefaultPacketReader) readReliable(layers *PacketLayers, packet *UDPPacket) {
	packet.stream = layers.RakNet.payload
	reliabilityLayer, err := DecodeReliabilityLayer(packet, reader.ValContext, layers.RakNet)
	if err != nil {
		reader.ErrorHandler(errors.New("Failed to decode reliable packet: "+err.Error()), packet)
		return
	}

	queues := reader.queues

	reader.ReliabilityLayerHandler(packet, reliabilityLayer, layers.RakNet)
	for _, subPacket := range reliabilityLayer.Packets {
		reliablePacketLayers := &PacketLayers{RakNet: layers.RakNet, Reliability: subPacket}

		buffer, err := reader.ValContext.handleSplitPacket(subPacket, layers.RakNet, packet)
		if err != nil {
			subPacket.SplitBuffer.Logger.Println("error while handling split:", err.Error())
			reader.ErrorHandler(errors.New(fmt.Sprintf("Error while handling split packet: %s", err.Error())), packet)
			return
		}
		subPacket.SplitBuffer = buffer

		reader.ReliableHandler(buffer.PacketType, packet, reliablePacketLayers)
		queues.add(reliablePacketLayers)
		if reliablePacketLayers.Reliability.Reliability == 0 {
			reader.readOrdered(reliablePacketLayers, packet)
			queues.remove(reliablePacketLayers)
			continue
		}

		reliablePacketLayers = queues.next(subPacket.OrderingChannel)
		for reliablePacketLayers != nil {
			reader.readOrdered(reliablePacketLayers, packet)
			queues.remove(reliablePacketLayers)
			reliablePacketLayers = queues.next(subPacket.OrderingChannel)
		}
	}
}

// ReadPacket reads a single packet and invokes all according handler functions
func (reader *DefaultPacketReader) ReadPacket(payload []byte, packet *UDPPacket) {
	context := reader.ValContext

	packet.stream = bufferToStream(payload)
	rakNetLayer, err := DecodeRakNetLayer(payload[0], packet, context)
	if err != nil {
		reader.ErrorHandler(err, packet)
		return
	}
	if rakNetLayer.IsDuplicate {
		return
	}

	layers := &PacketLayers{RakNet: rakNetLayer}
	if rakNetLayer.IsSimple {
		packetType := rakNetLayer.SimpleLayerID
		reader.readSimple(packetType, layers, packet)
		return
	} else if !rakNetLayer.IsValid {
		reader.ErrorHandler(fmt.Errorf("Sent invalid packet (packet header %x)", payload[0]), packet)
		return
	} else if rakNetLayer.IsACK || rakNetLayer.IsNAK {
		reader.ACKHandler(packet, rakNetLayer)
	} else {
		reader.readReliable(layers, packet)
		return
	}
}
