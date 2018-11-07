package peer

import "github.com/gskartwii/go-bitstream"
import "bytes"

func min(x, y uint) uint {
	if x < y {
		return x
	}
	return y
}

type PacketWriter interface {
	Context() *CommunicationContext
	ToClient() bool
	Caches() *Caches
}

// TODO: Make an interface "Writable" that is implemented by Layers, RakNetPacket, Packet83Subpacket?

// PacketWriter is a struct used to write packets to a peer
// Pass packets in using WriteSimple/WriteGeneric/etc.
// and bind to the given callbacks
type DefaultPacketWriter struct {
	// Any errors that are encountered are passed to ErrorHandler.
	// TODO: Get rid of ErrorHandler, just _return_ errors when packets are written?
	ErrorHandler func(error)
	// OutputHandler sends the data for all packets to be written.
	OutputHandler   func([]byte)
	orderingIndex   uint32
	sequencingIndex uint32
	splitPacketID   uint16
	reliableNumber  uint32
	datagramNumber  uint32
	// Set this to true if the packets produced by this writer are sent to a client.
	ValToClient bool
	ValCaches   *Caches
	ValContext  *CommunicationContext
}

func NewPacketWriter() *DefaultPacketWriter {
	return &DefaultPacketWriter{}
}
func (writer *DefaultPacketWriter) ToClient() bool {
	return writer.ValToClient
}
func (writer *DefaultPacketWriter) Caches() *Caches {
	return writer.ValCaches
}
func (writer *DefaultPacketWriter) Context() *CommunicationContext {
	return writer.ValContext
}

// WriteSimple is used to write pre-connection packets (IDs 5-8). It doesn't use a
// ReliabilityLayer.
func (writer *DefaultPacketWriter) WriteSimple(packet RakNetPacket) {
	output := make([]byte, 0, 1492)
	buffer := bytes.NewBuffer(output)
	stream := &extendedWriter{bitstream.NewWriter(buffer)}
	err := packet.Serialize(writer, stream)
	if err != nil {
		writer.ErrorHandler(err)
		return
	}

	stream.Flush(bitstream.Bit(false))
	writer.OutputHandler(buffer.Bytes())
}

// WriteRakNet is used to write raw RakNet packets to the client. You aren't probably
// going to need it.
func (writer *DefaultPacketWriter) WriteRakNet(packet *RakNetLayer) {
	output := make([]byte, 0, 1492)
	buffer := bytes.NewBuffer(output)
	stream := &extendedWriter{bitstream.NewWriter(buffer)}
	err := packet.Serialize(writer, stream)
	if err != nil {
		writer.ErrorHandler(err)
		return
	}

	stream.Flush(bitstream.Bit(false))
	writer.OutputHandler(buffer.Bytes())
}
func (writer *DefaultPacketWriter) writeReliable(packet *ReliabilityLayer) {
	output := make([]byte, 0, 1492)
	buffer := bytes.NewBuffer(output)
	stream := &extendedWriter{bitstream.NewWriter(buffer)}
	err := packet.Serialize(writer, stream)
	if err != nil {
		writer.ErrorHandler(err)
		return
	}

	stream.Flush(bitstream.Bit(false))

	payload := buffer.Bytes()
	raknet := &RakNetLayer{
		payload:        bufferToStream(payload),
		IsValid:        true,
		DatagramNumber: writer.datagramNumber,
	}
	writer.datagramNumber++

	writer.WriteRakNet(raknet)
}
func (writer *DefaultPacketWriter) writeReliableWithDN(packet *ReliabilityLayer, dn uint32) {
	output := make([]byte, 0, 1492)
	buffer := bytes.NewBuffer(output)
	stream := &extendedWriter{bitstream.NewWriter(buffer)}
	err := packet.Serialize(writer, stream)
	if err != nil {
		writer.ErrorHandler(err)
		return
	}

	stream.Flush(bitstream.Bit(false))

	payload := buffer.Bytes()
	raknet := &RakNetLayer{
		payload:        bufferToStream(payload),
		IsValid:        true,
		DatagramNumber: dn,
	}
	if dn >= writer.datagramNumber {
		writer.datagramNumber = dn + 1
	}

	writer.WriteRakNet(raknet)
}

func (writer *DefaultPacketWriter) WriteReliablePacket(data []byte, packet *ReliablePacket) {
	reliability := packet.Reliability
	realLen := len(data)
	estHeaderLength := 0x1C // UDP
	estHeaderLength += 4    // RakNet
	estHeaderLength += 1    // Reliability, has split
	estHeaderLength += 2    // len

	if reliability >= 2 && reliability <= 4 {
		packet.ReliableMessageNumber = writer.reliableNumber
		writer.reliableNumber++
		estHeaderLength += 3
	}
	if reliability == 1 || reliability == 4 {
		packet.SequencingIndex = writer.sequencingIndex
		writer.sequencingIndex++
		estHeaderLength += 3
	}
	if reliability == 1 || reliability == 3 || reliability == 4 || reliability == 7 {
		packet.OrderingChannel = 0
		packet.OrderingIndex = writer.orderingIndex
		writer.orderingIndex++
		estHeaderLength += 7
	}

	if realLen <= 1492-estHeaderLength { // Don't need to split
		packet.SelfData = data
		packet.LengthInBits = uint16(realLen * 8)

		writer.writeReliable(&ReliabilityLayer{[]*ReliablePacket{packet}})
	} else {
		packet.HasSplitPacket = true
		packet.SplitPacketID = writer.splitPacketID
		writer.splitPacketID++
		packet.SplitPacketIndex = 0
		estHeaderLength += 10

		splitBandwidth := 1472 - estHeaderLength
		requiredSplits := (realLen + splitBandwidth - 1) / splitBandwidth
		packet.SplitPacketCount = uint32(requiredSplits)
		packet.SelfData = data[:splitBandwidth]
		writer.writeReliable(&ReliabilityLayer{[]*ReliablePacket{packet}})

		for i := 1; i < requiredSplits; i++ {
			packet.SplitPacketIndex = uint32(i)
			if reliability >= 2 && reliability <= 4 {
				packet.ReliableMessageNumber = writer.reliableNumber
				writer.reliableNumber++
			}

			packet.SelfData = data[splitBandwidth*i : min(uint(realLen), uint(splitBandwidth*(i+1)))]
			writer.writeReliable(&ReliabilityLayer{[]*ReliablePacket{packet}})
		}
	}
}

func (writer *DefaultPacketWriter) WriteTimestamped(timestamp *Packet1BLayer, generic RakNetPacket, reliability uint32) []byte {
	output := make([]byte, 0, 1492)
	buffer := bytes.NewBuffer(output) // Will allocate more if needed
	stream := &extendedWriter{bitstream.NewWriter(buffer)}
	err := timestamp.Serialize(writer, stream)
	if err != nil {
		writer.ErrorHandler(err)
		return nil
	}
	err = generic.Serialize(writer, stream)
	if err != nil {
		writer.ErrorHandler(err)
		return nil
	}

	stream.Flush(bitstream.Bit(false))
	result := buffer.Bytes()

	packet := &ReliablePacket{
		Reliability: reliability,
	}

	writer.WriteReliablePacket(result, packet)

	return result
}

// WriteGeneric is used to write packets after the pre-connection. You want to use it
// for most of your packets.
func (writer *DefaultPacketWriter) WriteGeneric(generic RakNetPacket, reliability uint32) []byte {
	output := make([]byte, 0, 1492)
	buffer := bytes.NewBuffer(output) // Will allocate more if needed
	stream := &extendedWriter{bitstream.NewWriter(buffer)}
	err := generic.Serialize(writer, stream)
	if err != nil {
		writer.ErrorHandler(err)
		return nil
	}

	stream.Flush(bitstream.Bit(false))
	result := buffer.Bytes()

	packet := &ReliablePacket{
		Reliability: reliability,
	}

	writer.WriteReliablePacket(result, packet)

	return result
}
