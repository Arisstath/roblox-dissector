package peer
import "bytes"
import "github.com/gskartwii/go-bitstream"

type SplitPacketBuffer struct {
	ReliablePackets []*ReliablePacket
	RakNetPackets []*RakNetLayer
	NextExpectedPacket uint32
	WrittenPackets uint32

	DataReader *ExtendedReader
	DataWriter *bytes.Buffer
}
type SplitPacketList map[string](map[uint16](*ReliablePacket))

func NewSplitPacketBuffer(packet *ReliablePacket) *SplitPacketBuffer {
	reliables := make([]*ReliablePacket, int(packet.SplitPacketCount))
	raknets := make([]*RakNetLayer, int(packet.SplitPacketCount))

	list := &SplitPacketBuffer{reliables, raknets, 0, 0, nil, nil}
	buffer := &bytes.Buffer{}
	list.DataReader = &ExtendedReader{bitstream.NewReader(buffer)}
	list.DataWriter = buffer

	return list
}

func (list *SplitPacketBuffer) AddPacket(packet *ReliablePacket, rakNetPacket *RakNetLayer, index uint32)  {
	// Packets may be duplicated. At least I think so. Thanks UDP
	list.ReliablePackets[index] = packet
	list.RakNetPackets[index] = rakNetPacket
}

func (context *CommunicationContext) AddSplitPacket(source string, packet *ReliablePacket, rakNetPacket *RakNetLayer) *ReliablePacket {
	splitPacketId := packet.SplitPacketID
	splitPacketIndex := packet.SplitPacketIndex

	if context.SplitPackets == nil {
		packet.Buffer = NewSplitPacketBuffer(packet)
		packet.IsFirst = true
		packet.FullDataReader = packet.Buffer.DataReader
		packet.Buffer.AddPacket(packet, rakNetPacket, splitPacketIndex)

		context.SplitPackets = SplitPacketList{source: map[uint16]*ReliablePacket{splitPacketId: packet}}
	} else if context.SplitPackets[source] == nil {
		packet.Buffer = NewSplitPacketBuffer(packet)
		packet.IsFirst = true
		packet.FullDataReader = packet.Buffer.DataReader
		packet.Buffer.AddPacket(packet, rakNetPacket, splitPacketIndex)

		context.SplitPackets[source] = map[uint16]*ReliablePacket{splitPacketId: packet}
	} else if context.SplitPackets[source][splitPacketId] == nil {
		packet.Buffer = NewSplitPacketBuffer(packet)
		packet.IsFirst = true
		packet.FullDataReader = packet.Buffer.DataReader
		packet.Buffer.AddPacket(packet, rakNetPacket, splitPacketIndex)

		context.SplitPackets[source][splitPacketId] = packet
	} else {
		context.SplitPackets[source][splitPacketId].IsFirst = false
		context.SplitPackets[source][splitPacketId].Buffer.AddPacket(packet, rakNetPacket, splitPacketIndex)
	}

	return context.SplitPackets[source][splitPacketId]
}

func (context *CommunicationContext) HandleSplitPacket(reliablePacket *ReliablePacket, rakNetPacket *RakNetLayer, packet *UDPPacket) (*ReliablePacket, error) {
	source := packet.Source.String()

	fullPacket := context.AddSplitPacket(source, reliablePacket, rakNetPacket)
	packetBuffer := fullPacket.Buffer
	expectedPacket := packetBuffer.NextExpectedPacket

	fullPacket.AllReliablePackets = append(fullPacket.AllReliablePackets, reliablePacket)
	fullPacket.AllRakNetLayers = append(fullPacket.AllRakNetLayers, rakNetPacket)
	reliablePacket.AllReliablePackets = fullPacket.AllReliablePackets
	reliablePacket.AllRakNetLayers = fullPacket.AllRakNetLayers

	fullPacket.RealLength += uint32(len(reliablePacket.SelfData))

	var shouldClose bool
	for len(packetBuffer.ReliablePackets) > int(expectedPacket) && packetBuffer.ReliablePackets[expectedPacket] != nil {
		shouldClose = len(packetBuffer.ReliablePackets) == int(expectedPacket + 1)
		packetBuffer.DataWriter.Write(packetBuffer.ReliablePackets[expectedPacket].SelfData)
		packetBuffer.WrittenPackets++

		expectedPacket++
		packetBuffer.NextExpectedPacket = expectedPacket
	}
	if shouldClose {
		fullPacket.IsFinal = true
		reliablePacket.IsFinal = true
	}
	fullPacket.NumReceivedSplits = expectedPacket
	reliablePacket.NumReceivedSplits = fullPacket.NumReceivedSplits
	fullPacket.ReliableMessageNumber = reliablePacket.ReliableMessageNumber
	reliablePacket.FullDataReader = packetBuffer.DataReader

	if reliablePacket.HasPacketType {
		fullPacket.HasPacketType = true
		fullPacket.PacketType = reliablePacket.PacketType
	} else {
		reliablePacket.HasPacketType = fullPacket.HasPacketType
		reliablePacket.PacketType = fullPacket.PacketType
	}

	return reliablePacket, nil
}
