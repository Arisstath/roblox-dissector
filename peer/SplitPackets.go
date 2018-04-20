package peer
import "bytes"
import "github.com/gskartwii/go-bitstream"
import "log"
import "strings"

type splitPacketBuffer struct {
	reliablePackets []*ReliablePacket
	rakNetPackets []*RakNetLayer
	nextExpectedPacket uint32
	writtenPackets uint32

	dataReader *extendedReader
	dataWriter *bytes.Buffer
}
type splitPacketList map[string](map[uint16](*ReliablePacket))

func newSplitPacketBuffer(packet *ReliablePacket) *splitPacketBuffer {
	reliables := make([]*ReliablePacket, int(packet.SplitPacketCount))
	raknets := make([]*RakNetLayer, int(packet.SplitPacketCount))

	list := &splitPacketBuffer{reliables, raknets, 0, 0, nil, nil}
	buffer := &bytes.Buffer{}
	list.dataReader = &extendedReader{bitstream.NewReader(buffer)}
	list.dataWriter = buffer

	return list
}

func (list *splitPacketBuffer) addPacket(packet *ReliablePacket, rakNetPacket *RakNetLayer, index uint32)  {
	// Packets may be duplicated. At least I think so. Thanks UDP
	list.reliablePackets[index] = packet
	list.rakNetPackets[index] = rakNetPacket
}

func initPacketForSplitPackets(packet *ReliablePacket, rakNetPacket *RakNetLayer, splitPacketIndex uint32) {
	packet.buffer = newSplitPacketBuffer(packet)
	packet.IsFirst = true
	packet.fullDataReader = packet.buffer.dataReader
	packet.buffer.addPacket(packet, rakNetPacket, splitPacketIndex)
	packet.logBuffer = new(strings.Builder)
	packet.Logger = log.New(packet.logBuffer, "", log.Lmicroseconds | log.Ltime)
}

func (context *CommunicationContext) addSplitPacket(source string, packet *ReliablePacket, rakNetPacket *RakNetLayer) *ReliablePacket {
	splitPacketId := packet.SplitPacketID
	splitPacketIndex := packet.SplitPacketIndex

	if context.splitPackets == nil {
		initPacketForSplitPackets(packet, rakNetPacket, splitPacketIndex)

		context.splitPackets = splitPacketList{source: map[uint16]*ReliablePacket{splitPacketId: packet}}
	} else if context.splitPackets[source] == nil {
		initPacketForSplitPackets(packet, rakNetPacket, splitPacketIndex)

		context.splitPackets[source] = map[uint16]*ReliablePacket{splitPacketId: packet}
	} else if context.splitPackets[source][splitPacketId] == nil {
		initPacketForSplitPackets(packet, rakNetPacket, splitPacketIndex)

		context.splitPackets[source][splitPacketId] = packet
	} else {
		context.splitPackets[source][splitPacketId].IsFirst = false
		context.splitPackets[source][splitPacketId].buffer.addPacket(packet, rakNetPacket, splitPacketIndex)
	}

	return context.splitPackets[source][splitPacketId]
}

func (context *CommunicationContext) handleSplitPacket(reliablePacket *ReliablePacket, rakNetPacket *RakNetLayer, packet *UDPPacket) (*ReliablePacket, error) {
	source := packet.Source.String()

	fullPacket := context.addSplitPacket(source, reliablePacket, rakNetPacket)
	packetBuffer := fullPacket.buffer
	expectedPacket := packetBuffer.nextExpectedPacket

	packet.logBuffer = fullPacket.logBuffer
	packet.Logger = fullPacket.Logger

	fullPacket.AllReliablePackets = append(fullPacket.AllReliablePackets, reliablePacket)
	fullPacket.AllRakNetLayers = append(fullPacket.AllRakNetLayers, rakNetPacket)
	reliablePacket.AllReliablePackets = fullPacket.AllReliablePackets
	reliablePacket.AllRakNetLayers = fullPacket.AllRakNetLayers

	fullPacket.RealLength += uint32(len(reliablePacket.SelfData))

	var shouldClose bool
	for len(packetBuffer.reliablePackets) > int(expectedPacket) && packetBuffer.reliablePackets[expectedPacket] != nil {
		shouldClose = len(packetBuffer.reliablePackets) == int(expectedPacket + 1)
		packetBuffer.dataWriter.Write(packetBuffer.reliablePackets[expectedPacket].SelfData)
		packetBuffer.writtenPackets++

		expectedPacket++
		packetBuffer.nextExpectedPacket = expectedPacket
	}
	if shouldClose {
		fullPacket.IsFinal = true
		reliablePacket.IsFinal = true
	}
	fullPacket.NumReceivedSplits = expectedPacket
	reliablePacket.NumReceivedSplits = fullPacket.NumReceivedSplits
	fullPacket.ReliableMessageNumber = reliablePacket.ReliableMessageNumber
	reliablePacket.fullDataReader = packetBuffer.dataReader

	if reliablePacket.HasPacketType {
		fullPacket.HasPacketType = true
		fullPacket.PacketType = reliablePacket.PacketType
	} else {
		reliablePacket.HasPacketType = fullPacket.HasPacketType
		reliablePacket.PacketType = fullPacket.PacketType
	}

	return reliablePacket, nil
}
