package peer
import "github.com/gskartwii/go-bitstream"
import "bytes"

func min(x, y uint) uint {
	if x < y {
		return x
	}
	return y
}

type PacketWriter struct {
	ErrorHandler func(error)
	OutputHandler func([]byte)
	OrderingIndex uint32
	SequencingIndex uint32
	SplitPacketID uint16
	ReliableNumber uint32
	DatagramNumber uint32
}

func NewPacketWriter() *PacketWriter {
	return &PacketWriter{}
}

func (this *PacketWriter) WriteSimple(packetType byte, packet RakNetPacket) {
	output := make([]byte, 0, 1492)
	buffer := bytes.NewBuffer(output)
	stream := &ExtendedWriter{bitstream.NewWriter(buffer)}
	err := stream.WriteByte(packetType)
	if err != nil {
		this.ErrorHandler(err)
		return
	}
	err = packet.Serialize(nil, stream)
	if err != nil {
		this.ErrorHandler(err)
		return
	}

	stream.Flush(bitstream.Bit(false))
	this.OutputHandler(buffer.Bytes())
}
func (this *PacketWriter) WriteRakNet(packet *RakNetLayer) {
	output := make([]byte, 0, 1492)
	buffer := bytes.NewBuffer(output)
	stream := &ExtendedWriter{bitstream.NewWriter(buffer)}
	err := packet.Serialize(nil, stream)
	if err != nil {
		this.ErrorHandler(err)
		return
	}

	stream.Flush(bitstream.Bit(false))
	this.OutputHandler(buffer.Bytes())
}
func (this *PacketWriter) WriteReliable(packet *ReliabilityLayer) {
	output := make([]byte, 0, 1492)
	buffer := bytes.NewBuffer(output)
	stream := &ExtendedWriter{bitstream.NewWriter(buffer)}
	err := packet.Serialize(nil, stream)
	if err != nil {
		this.ErrorHandler(err)
		return
	}

	stream.Flush(bitstream.Bit(false))

	payload := buffer.Bytes()
	raknet := &RakNetLayer{
		Payload: BufferToStream(payload),
		IsValid: true,
		DatagramNumber: this.DatagramNumber,
	}
	this.DatagramNumber++

	this.WriteRakNet(raknet)
}

func (this *PacketWriter) WriteGeneric(context *CommunicationContext, packetType byte, generic RakNetPacket, reliability uint32) {
	output := make([]byte, 0, 1492)
	buffer := bytes.NewBuffer(output) // Will allocate more if needed
	stream := &ExtendedWriter{bitstream.NewWriter(buffer)}
	err := generic.Serialize(context, stream)
	if err != nil {
		this.ErrorHandler(err)
		return
	}

	stream.Flush(bitstream.Bit(false))
	result := buffer.Bytes()
	realLen := len(result)

	packet := &ReliablePacket{
		IsFinal: true,
		IsFirst: true,
		Reliability: reliability,
		RealLength: uint32(realLen),
	}
	estHeaderLength := 0x1C // UDP
	estHeaderLength += 4 // RakNet
	estHeaderLength += 1 // Reliability, has split
	estHeaderLength += 2 // len

	if reliability >= 2 && reliability <= 4 {
		packet.ReliableMessageNumber = this.ReliableNumber
		this.ReliableNumber++
		estHeaderLength += 3
	}
	if reliability == 1 || reliability == 4 {
		packet.SequencingIndex = this.SequencingIndex
		this.SequencingIndex++
		estHeaderLength += 3
	}
	if reliability == 1 || reliability == 3 || reliability == 4 || reliability == 7 {
		packet.OrderingChannel = 0
		packet.OrderingIndex = this.OrderingIndex
		this.OrderingIndex++
		estHeaderLength += 7
	}

	if realLen <= 1492 - estHeaderLength { // Don't need to split
		println("Writing normal packet")
		packet.SelfData = result
		packet.LengthInBits = uint16(realLen * 8)

		this.WriteReliable(&ReliabilityLayer{[]*ReliablePacket{packet}})
	} else {
		packet.HasSplitPacket = true
		packet.SplitPacketID = this.SplitPacketID
		this.SplitPacketID++
		packet.SplitPacketIndex = 0
		estHeaderLength += 10

		splitBandwidth := 1472 - estHeaderLength
		requiredSplits := (realLen + splitBandwidth - 1) / splitBandwidth
		packet.SplitPacketCount = uint32(requiredSplits)
		println("Writing split", 0, "/", requiredSplits)
		packet.SelfData = result[:splitBandwidth]
		this.WriteReliable(&ReliabilityLayer{[]*ReliablePacket{packet}})

		for i := 1; i < requiredSplits; i++ {
			println("Writing split", i, "/", requiredSplits)
			packet.SplitPacketIndex = uint32(i)
			if reliability >= 2 && reliability <= 4 {
				packet.ReliableMessageNumber = this.ReliableNumber
				this.ReliableNumber++
			}

			packet.SelfData = result[splitBandwidth*i:min(uint(realLen), uint(splitBandwidth*(i + 1)))]
			this.WriteReliable(&ReliabilityLayer{[]*ReliablePacket{packet}})
		}
	}
}
