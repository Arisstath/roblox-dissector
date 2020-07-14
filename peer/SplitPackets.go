package peer

import (
	"bytes"
	"log"
	"strings"
)

// SplitPacketBuffer represents a structure that accumulates every
// layer that is used to transmit the split packet.
type SplitPacketBuffer struct {
	// All ReliabilityLayer packets for this packet received so far
	ReliablePackets []*ReliablePacket
	// All RakNet layers for this packet received so far
	// IN RECEIVE ORDER, NOT SPLIT ORDER!!
	// Use ReliablePackets[i].RakNetLayer to access them in that order.
	RakNetPackets []*RakNetLayer
	// Next expected index
	NextExpectedPacket uint32
	// Number of _ordered_ splits we have received so far
	NumReceivedSplits uint32
	// Has received packet type yet? Set to true when the first split of this packet
	// is received
	HasPacketType bool
	PacketType    byte

	byteReader *bytes.Reader
	dataReader *extendedReader
	Data       []byte

	// Have all splits been received?
	IsFinal bool
	// Total length received so far, in bytes
	RealLength uint32
	UniqueID   uint64

	logBuffer *strings.Builder // must be a pointer because it may be copied!
	Logger    *log.Logger
}
type splitPacketList map[uint16](*SplitPacketBuffer)

func newSplitPacketBuffer(packet *ReliablePacket, context *CommunicationContext) (*SplitPacketBuffer, uint64) {
	reliables := make([]*ReliablePacket, int(packet.SplitPacketCount))
	raknets := make([]*RakNetLayer, 0, int(packet.SplitPacketCount))

	list := &SplitPacketBuffer{
		ReliablePackets: reliables,
		RakNetPackets:   raknets,
	}
	list.UniqueID = context.uniqueID
	context.uniqueID++
	list.logBuffer = new(strings.Builder)
	list.Logger = log.New(list.logBuffer, "", log.Lmicroseconds|log.Ltime)

	return list, list.UniqueID
}

func (list *SplitPacketBuffer) addPacket(packet *ReliablePacket, rakNetPacket *RakNetLayer, index uint32) {
	// Packets may be duplicated. At least I think so. Thanks UDP
	list.ReliablePackets[index] = packet
	list.RakNetPackets = append(list.RakNetPackets, rakNetPacket)
}

func (list splitPacketList) delete(layers *PacketLayers) {
	delete(list, layers.Reliability.SplitPacketID)
}

func (reader *DefaultPacketReader) addSplitPacket(layers *PacketLayers) *SplitPacketBuffer {
	packet := layers.Reliability
	splitPacketID := packet.SplitPacketID
	splitPacketIndex := packet.SplitPacketIndex

	if !packet.HasSplitPacket {
		buffer, id := newSplitPacketBuffer(packet, reader.context)
		layers.UniqueID = id
		buffer.addPacket(packet, layers.RakNet, 0)

		return buffer
	}

	var buffer *SplitPacketBuffer
	var id uint64
	if reader.splitPackets == nil {
		buffer, id = newSplitPacketBuffer(packet, reader.context)

		reader.splitPackets = map[uint16]*SplitPacketBuffer{splitPacketID: buffer}
	} else if reader.splitPackets[splitPacketID] == nil {
		buffer, id = newSplitPacketBuffer(packet, reader.context)

		reader.splitPackets[splitPacketID] = buffer
	} else {
		buffer = reader.splitPackets[splitPacketID]
		id = buffer.UniqueID
	}
	buffer.addPacket(packet, layers.RakNet, splitPacketIndex)
	packet.SplitBuffer = buffer
	layers.UniqueID = id

	return buffer
}

func (reader *DefaultPacketReader) handleSplitPacket(layers *PacketLayers) (*SplitPacketBuffer, error) {
	reliablePacket := layers.Reliability
	packetBuffer := reader.addSplitPacket(layers)
	expectedPacket := packetBuffer.NextExpectedPacket

	packetBuffer.RealLength += uint32(len(reliablePacket.SelfData))

	var shouldClose bool
	for len(packetBuffer.ReliablePackets) > int(expectedPacket) && packetBuffer.ReliablePackets[expectedPacket] != nil {
		packetBuffer.Data = append(packetBuffer.Data, packetBuffer.ReliablePackets[expectedPacket].SelfData...)

		expectedPacket++
		shouldClose = len(packetBuffer.ReliablePackets) == int(expectedPacket)
		packetBuffer.NextExpectedPacket = expectedPacket
	}
	if shouldClose {
		packetBuffer.IsFinal = true
		packetBuffer.byteReader = bytes.NewReader(packetBuffer.Data)
		packetBuffer.dataReader = &extendedReader{packetBuffer.byteReader}
		if reliablePacket.HasSplitPacket {
			// TODO: Use a linked list
			reader.splitPackets.delete(layers)
		}
	}
	packetBuffer.NumReceivedSplits = expectedPacket

	if reliablePacket.SplitPacketIndex == 0 {
		packetBuffer.PacketType = reliablePacket.SelfData[0]
		packetBuffer.HasPacketType = true
	}

	layers.Root.Logger = packetBuffer.Logger
	layers.Root.logBuffer = packetBuffer.logBuffer

	return packetBuffer, nil
}
