package packets

import "github.com/gskartwii/roblox-dissector/util"
import "log"
import "strings"

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

	DataReader *PacketReaderBitstream
	Data       []byte

	// Have all splits been received?
	IsFinal bool
	// Unique ID given to each packet. Splits of the same packet have the same ID.
	UniqueID uint32
	// Total length received so far, in bytes
	RealLength uint32

	LogBuffer *strings.Builder // must be a pointer because it may be copied!
	Logger    *log.Logger
}
type SplitPacketList map[uint16](*SplitPacketBuffer)

func NewSplitPacketBuffer(packet *ReliablePacket, context *util.CommunicationContext) *SplitPacketBuffer {
	reliables := make([]*ReliablePacket, int(packet.SplitPacketCount))
	raknets := make([]*RakNetLayer, 0, int(packet.SplitPacketCount))

	list := &SplitPacketBuffer{
		ReliablePackets: reliables,
		RakNetPackets:   raknets,
	}
	list.Data = make([]byte, 0, uint32(packet.LengthInBits)*packet.SplitPacketCount*8)
	list.PacketType = 0xFF
	list.UniqueID = context.UniqueID
	context.UniqueID++
	list.logBuffer = new(strings.Builder)
	list.Logger = log.New(list.logBuffer, "", log.Lmicroseconds|log.Ltime)

	return list
}

func (list *SplitPacketBuffer) AddPacket(packet *ReliablePacket, rakNetPacket *RakNetLayer, index uint32) {
	// Packets may be duplicated. At least I think so. Thanks UDP
	list.ReliablePackets[index] = packet
	list.RakNetPackets = append(list.RakNetPackets, rakNetPacket)
}

func (list SplitPacketList) Delete(layers *PacketLayers) {
	delete(list, layers.Reliability.SplitPacketID)
}
