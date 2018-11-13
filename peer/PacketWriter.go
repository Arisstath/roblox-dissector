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
	SetContext(*CommunicationContext)
	Context() *CommunicationContext
	SetToClient(bool)
	ToClient() bool
	SetCaches(*Caches)
	Caches() *Caches
}

// TODO: Make an interface "Writable" that is implemented by Layers, RakNetPacket, Packet83Subpacket?

// PacketWriter is a struct used to write packets to a peer
// Pass packets in using WriteSimple/WriteGeneric/etc.
// and bind to the given callbacks
type DefaultPacketWriter struct {
	contextualHandler
	// OutputHandler sends the data for all packets to be written.
	OutputHandler   func([]byte)
	orderingIndex   uint32
	sequencingIndex uint32
	splitPacketID   uint16
	reliableNumber  uint32
	datagramNumber  uint32
	// Set this to true if the packets produced by this writer are sent to a client.
	toClient bool
	caches   *Caches
	context  *CommunicationContext
}

func NewPacketWriter() *DefaultPacketWriter {
	return &DefaultPacketWriter{}
}
func (writer *DefaultPacketWriter) ToClient() bool {
	return writer.toClient
}
func (writer *DefaultPacketWriter) SetToClient(val bool) {
	writer.toClient = val
}

// WriteSimple is used to write pre-connection packets (IDs 5-8). It doesn't use a
// ReliabilityLayer.
func (writer *DefaultPacketWriter) WriteSimple(packet RakNetPacket) error {
	output := make([]byte, 0, 1492)
	buffer := bytes.NewBuffer(output)
	stream := &extendedWriter{bitstream.NewWriter(buffer)}
	err := packet.Serialize(writer, stream)
	if err != nil {
		return err
	}

	stream.Flush(bitstream.Bit(false))
	writer.OutputHandler(buffer.Bytes())
	return nil
}

// WriteRakNet is used to write raw RakNet packets to the client. You aren't probably
// going to need it.
func (writer *DefaultPacketWriter) WriteRakNet(packet *RakNetLayer) error {
	output := make([]byte, 0, 1492)
	buffer := bytes.NewBuffer(output)
	stream := &extendedWriter{bitstream.NewWriter(buffer)}
	err := packet.Serialize(writer, stream)
	if err != nil {
		return err
	}

	stream.Flush(bitstream.Bit(false))
	writer.OutputHandler(buffer.Bytes())
	return nil
}
func (writer *DefaultPacketWriter) writeReliable(packet *ReliabilityLayer) error {
	output := make([]byte, 0, 1492)
	buffer := bytes.NewBuffer(output)
	stream := &extendedWriter{bitstream.NewWriter(buffer)}
	err := packet.Serialize(writer, stream)
	if err != nil {
		return err
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
	return nil
}
func (writer *DefaultPacketWriter) writeReliableWithDN(packet *ReliabilityLayer, dn uint32) error {
	output := make([]byte, 0, 1492)
	buffer := bytes.NewBuffer(output)
	stream := &extendedWriter{bitstream.NewWriter(buffer)}
	err := packet.Serialize(writer, stream)
	if err != nil {
		return err
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
	return nil
}

func (writer *DefaultPacketWriter) WriteReliablePacket(data []byte, packet *ReliablePacket) error {
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

		return writer.writeReliable(&ReliabilityLayer{[]*ReliablePacket{packet}})
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
		err := writer.writeReliable(&ReliabilityLayer{[]*ReliablePacket{packet}})
		if err != nil {
			return err
		}

		for i := 1; i < requiredSplits; i++ {
			packet.SplitPacketIndex = uint32(i)
			if reliability >= 2 && reliability <= 4 {
				packet.ReliableMessageNumber = writer.reliableNumber
				writer.reliableNumber++
			}

			packet.SelfData = data[splitBandwidth*i : min(uint(realLen), uint(splitBandwidth*(i+1)))]
			err = writer.writeReliable(&ReliabilityLayer{[]*ReliablePacket{packet}})
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (writer *DefaultPacketWriter) WriteTimestamped(timestamp *Packet1BLayer, generic RakNetPacket, reliability uint32) ([]byte, error) {
	output := make([]byte, 0, 1492)
	buffer := bytes.NewBuffer(output) // Will allocate more if needed
	stream := &extendedWriter{bitstream.NewWriter(buffer)}
	err := timestamp.Serialize(writer, stream)
	if err != nil {
		return nil, err
	}
	err = generic.Serialize(writer, stream)
	if err != nil {
		return nil, err
	}

	stream.Flush(bitstream.Bit(false))
	result := buffer.Bytes()

	packet := &ReliablePacket{
		Reliability: reliability,
	}

	err = writer.WriteReliablePacket(result, packet)

	return result, err
}

// WriteGeneric is used to write packets after the pre-connection. You want to use it
// for most of your packets.
func (writer *DefaultPacketWriter) WriteGeneric(generic RakNetPacket, reliability uint32) ([]byte, error) {
	output := make([]byte, 0, 1492)
	buffer := bytes.NewBuffer(output) // Will allocate more if needed
	stream := &extendedWriter{bitstream.NewWriter(buffer)}
	err := generic.Serialize(writer, stream)
	if err != nil {
		return nil, err
	}

	stream.Flush(bitstream.Bit(false))
	result := buffer.Bytes()

	packet := &ReliablePacket{
		Reliability: reliability,
	}

	err = writer.WriteReliablePacket(result, packet)

	return result, err
}

func (writer *DefaultPacketWriter) WritePacket(generic RakNetPacket) ([]byte, error) {
	return writer.WriteGeneric(generic, RELIABLE_ORD)
}
func (writer *DefaultPacketWriter) WritePhysics(timestamp *Packet1BLayer, generic RakNetPacket) ([]byte, error) {
	return writer.WriteTimestamped(timestamp, generic, UNRELIABLE)
}
