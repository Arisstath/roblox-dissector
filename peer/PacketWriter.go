package peer
import "github.com/gskartwii/go-bitstream"
import "bytes"
import "net"

func min(x, y uint) uint {
	if x < y {
		return x
	}
	return y
}

// PacketWriter is a struct used to write packets to a peer
// Pass packets in using WriteSimple/WriteGeneric/etc.
// and bind to the given callbacks
type PacketWriter struct {
	// Any errors that are encountered are passed to ErrorHandler.
	ErrorHandler func(error)
	// OutputHandler sends the data for all packets to be written.
	OutputHandler func([]byte, *net.UDPAddr)
	orderingIndex uint32
	sequencingIndex uint32
	splitPacketID uint16
	reliableNumber uint32
	datagramNumber uint32
	// Set this to true if you're a server.
	ToClient bool
}

func NewPacketWriter() *PacketWriter {
	return &PacketWriter{}
}

// WriteSimple is used to write pre-connection packets (IDs 5-8). It doesn't use a
// ReliabilityLayer.
func (this *PacketWriter) WriteSimple(packetType byte, packet RakNetPacket, dest *net.UDPAddr) {
	output := make([]byte, 0, 1492)
	buffer := bytes.NewBuffer(output)
	stream := &extendedWriter{bitstream.NewWriter(buffer)}
	err := stream.WriteByte(packetType)
	if err != nil {
		this.ErrorHandler(err)
		return
	}
	err = packet.serialize(this.ToClient, nil, stream)
	if err != nil {
		this.ErrorHandler(err)
		return
	}

	stream.Flush(bitstream.Bit(false))
	this.OutputHandler(buffer.Bytes(), dest)
}
// Write RakNet is used to write raw RakNet packets to the client. You aren't probably
// going to need it.
func (this *PacketWriter) WriteRakNet(packet *RakNetLayer, dest *net.UDPAddr) {
	output := make([]byte, 0, 1492)
	buffer := bytes.NewBuffer(output)
	stream := &extendedWriter{bitstream.NewWriter(buffer)}
	err := packet.serialize(this.ToClient, nil, stream)
	if err != nil {
		this.ErrorHandler(err)
		return
	}

	stream.Flush(bitstream.Bit(false))
	this.OutputHandler(buffer.Bytes(), dest)
}
func (this *PacketWriter) writeReliable(packet *ReliabilityLayer, dest *net.UDPAddr) {
	output := make([]byte, 0, 1492)
	buffer := bytes.NewBuffer(output)
	stream := &extendedWriter{bitstream.NewWriter(buffer)}
	err := packet.serialize(this.ToClient, nil, stream)
	if err != nil {
		this.ErrorHandler(err)
		return
	}

	stream.Flush(bitstream.Bit(false))

	payload := buffer.Bytes()
	raknet := &RakNetLayer{
		payload: bufferToStream(payload),
		IsValid: true,
		DatagramNumber: this.datagramNumber,
	}
	this.datagramNumber++

	this.WriteRakNet(raknet, dest)
}
func (this *PacketWriter) writeReliableWithDN(packet *ReliabilityLayer, dest *net.UDPAddr, dn uint32) {
	output := make([]byte, 0, 1492)
	buffer := bytes.NewBuffer(output)
	stream := &extendedWriter{bitstream.NewWriter(buffer)}
	err := packet.serialize(this.ToClient, nil, stream)
	if err != nil {
		this.ErrorHandler(err)
		return
	}

	stream.Flush(bitstream.Bit(false))

	payload := buffer.Bytes()
	raknet := &RakNetLayer{
		payload: bufferToStream(payload),
		IsValid: true,
		DatagramNumber: dn,
	}
	if dn >= this.datagramNumber {
		this.datagramNumber = dn + 1
	}

	this.WriteRakNet(raknet, dest)
}

// WriteGeneric is used to write packets after the pre-connection. You want to use it
// for most of your packets.
func (this *PacketWriter) WriteGeneric(context *CommunicationContext, packetType byte, generic RakNetPacket, reliability uint32, dest *net.UDPAddr) {
	output := make([]byte, 0, 1492)
	buffer := bytes.NewBuffer(output) // Will allocate more if needed
	stream := &extendedWriter{bitstream.NewWriter(buffer)}
	err := generic.serialize(this.ToClient, context, stream)
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
		packet.ReliableMessageNumber = this.reliableNumber
		this.reliableNumber++
		estHeaderLength += 3
	}
	if reliability == 1 || reliability == 4 {
		packet.SequencingIndex = this.sequencingIndex
		this.sequencingIndex++
		estHeaderLength += 3
	}
	if reliability == 1 || reliability == 3 || reliability == 4 || reliability == 7 {
		packet.OrderingChannel = 0
		packet.OrderingIndex = this.orderingIndex
		this.orderingIndex++
		estHeaderLength += 7
	}

	if realLen <= 1492 - estHeaderLength { // Don't need to split
		println("Writing normal packet")
		packet.SelfData = result
		packet.LengthInBits = uint16(realLen * 8)

		this.writeReliable(&ReliabilityLayer{[]*ReliablePacket{packet}}, dest)
	} else {
		packet.HasSplitPacket = true
		packet.SplitPacketID = this.splitPacketID
		this.splitPacketID++
		packet.SplitPacketIndex = 0
		estHeaderLength += 10

		splitBandwidth := 1472 - estHeaderLength
		requiredSplits := (realLen + splitBandwidth - 1) / splitBandwidth
		packet.SplitPacketCount = uint32(requiredSplits)
		println("Writing split", 0, "/", requiredSplits)
		packet.SelfData = result[:splitBandwidth]
		this.writeReliable(&ReliabilityLayer{[]*ReliablePacket{packet}}, dest)

		for i := 1; i < requiredSplits; i++ {
			println("Writing split", i, "/", requiredSplits)
			packet.SplitPacketIndex = uint32(i)
			if reliability >= 2 && reliability <= 4 {
				packet.ReliableMessageNumber = this.reliableNumber
				this.reliableNumber++
			}

			packet.SelfData = result[splitBandwidth*i:min(uint(realLen), uint(splitBandwidth*(i + 1)))]
			this.writeReliable(&ReliabilityLayer{[]*ReliablePacket{packet}}, dest)
		}
	}
}
