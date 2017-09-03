package peer
import "github.com/gskartwii/go-bitstream"
import "bytes"

type PacketWriter struct {
	ErrorHandler func(error)
	OutputHandler func([]byte)
	reliableQueue chan [1492]byte
}

func NewPacketWriter() *PacketWriter {
	return &PacketWriter{reliableQueue: make(chan [1492]byte)}
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
	err = packet.Serialize(stream)
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
	err := packet.Serialize(stream)
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
	err := packet.Serialize(stream)
	if err != nil {
		this.ErrorHandler(err)
		return
	}

	stream.Flush(bitstream.Bit(false))
	this.OutputHandler(buffer.Bytes())
}

func (this *PacketWriter) WriteGeneric(packetType byte, packet RakNetPacket) error {

	return nil
}
