package main
import "github.com/dgryski/go-bitstream"
import "github.com/google/gopacket"
import "bytes"
import "io"

type ReliablePacket struct {
	IsFinal bool
	Reliability uint32
	HasSplitPacket bool
	LengthInBits uint16
	ReliableMessageNumber uint32
	SequencingIndex uint32
	OrderingIndex uint32
	OrderingChannel uint8
	SplitPacketCount uint32
	SplitPacketID uint16
	SplitPacketIndex uint32
	SelfData []byte
	FinalData []byte
	Contents []byte
	AllRakNetLayers []*RakNetLayer
	AllReliablePackets []*ReliablePacket
}

type ReliabilityLayer struct {
	Packets []*ReliablePacket
	Contents []byte
}

type CountedReliablePacketList struct {
	ReliablePackets []*ReliablePacket
	RakNetPackets []*RakNetLayer
	Count uint32
}
type SplitPacketList map[string](map[uint16](*CountedReliablePacketList))

var SplitPackets SplitPacketList

func NewCountedReliablePacketList(packet *ReliablePacket) *CountedReliablePacketList {
	reliables := make([]*ReliablePacket, int(packet.SplitPacketCount))
	raknets := make([]*RakNetLayer, int(packet.SplitPacketCount))
	return &CountedReliablePacketList{reliables, raknets, 0}
}
func (list *CountedReliablePacketList) AddPacket(packet *ReliablePacket, rakNetPacket *RakNetLayer, index uint32) uint32 {// Packets may be duplicated. At least I think so. Thanks UDP
	if list.ReliablePackets[index] == nil {
		list.ReliablePackets[index] = packet
		list.Count++
	}
	list.RakNetPackets[index] = rakNetPacket
	return list.Count
}

func AddSplitPacket(source string, packet *ReliablePacket, rakNetPacket *RakNetLayer) (bool, []*ReliablePacket, []*RakNetLayer) {
	splitPacketId := packet.SplitPacketID
	splitPacketCount := packet.SplitPacketCount
	splitPacketIndex := packet.SplitPacketIndex

	var newCount uint32
	if SplitPackets == nil {
		var currentList = NewCountedReliablePacketList(packet)
		newCount = currentList.AddPacket(packet, rakNetPacket, splitPacketIndex)

		SplitPackets = SplitPacketList{source: map[uint16]*CountedReliablePacketList{splitPacketId: currentList}}
	} else if SplitPackets[source] == nil {
		var currentList = NewCountedReliablePacketList(packet)
		newCount = currentList.AddPacket(packet, rakNetPacket, splitPacketIndex)

		SplitPackets[source] = map[uint16]*CountedReliablePacketList{splitPacketId: currentList}
	} else if SplitPackets[source][splitPacketId] == nil {
		var currentList = NewCountedReliablePacketList(packet)
		newCount = currentList.AddPacket(packet, rakNetPacket, splitPacketIndex)

		SplitPackets[source][splitPacketId] = currentList
	} else {
		newCount = SplitPackets[source][splitPacketId].AddPacket(packet, rakNetPacket, splitPacketIndex)
	}

	return newCount == splitPacketCount, SplitPackets[source][splitPacketId].ReliablePackets, SplitPackets[source][splitPacketId].RakNetPackets
}

func HandleSplitPacket(reliablePacket *ReliablePacket, rakNetPacket *RakNetLayer, context *CommunicationContext, packet gopacket.Packet) {
	source := SourceInterfaceFromPacket(packet)
	HasEnough, AllPackets, AllRakNetPackets := AddSplitPacket(source, reliablePacket, rakNetPacket)
	if HasEnough {
		reliablePacket.IsFinal = true
		reliablePacket.AllReliablePackets = AllPackets
		reliablePacket.AllRakNetLayers = AllRakNetPackets
		var totalLength uint32

		var i uint32
		for i = 0; i < reliablePacket.SplitPacketCount; i++ {
			currentPacket := SplitPackets[source][reliablePacket.SplitPacketID].ReliablePackets[i]
			totalLength += uint32((currentPacket.LengthInBits + 7)/8)
		}

		finalData := make([]byte, 0, totalLength)
		for i = 0; i < reliablePacket.SplitPacketCount; i++ {
			finalData = append(finalData, SplitPackets[source][reliablePacket.SplitPacketID].ReliablePackets[i].SelfData...)
		}
		reliablePacket.FinalData = finalData
	}
}

func NewReliabilityLayer() *ReliabilityLayer {
	return &ReliabilityLayer{Packets: make([]*ReliablePacket, 0), Contents: make([]byte, 0)}
}
func NewReliablePacket() *ReliablePacket {
	return &ReliablePacket{SelfData: make([]byte, 0), FinalData: make([]byte, 0), Contents: make([]byte, 0)}
}

func DecodeReliabilityLayer(data []byte, context *CommunicationContext, packet gopacket.Packet, rakNetPacket *RakNetLayer) (*ReliabilityLayer, error) {
	layer := NewReliabilityLayer()
	bitstream := ExtendedReader{bitstream.NewReader(bytes.NewReader(data))}

	var reliability uint64
	var err error
	for reliability, err = bitstream.Bits(3); err == nil; reliability, err = bitstream.Bits(3) {
		reliablePacket := NewReliablePacket()
		reliablePacket.Reliability = uint32(reliability)
		reliablePacket.HasSplitPacket, _ = bitstream.ReadBool()
		bitstream.Align()

		reliablePacket.LengthInBits, _ = bitstream.ReadUint16BE()
		if reliability >= 2 && reliability <= 4 {
			reliablePacket.ReliableMessageNumber, _ = bitstream.ReadUint24LE()
		}
		if reliability == 1 || reliability == 4 {
			reliablePacket.SequencingIndex, _ = bitstream.ReadUint24LE()
		}
		if reliability == 1 || reliability == 3 || reliability == 4 || reliability == 7 {
			reliablePacket.OrderingIndex, _ = bitstream.ReadUint24LE()
			reliablePacket.OrderingChannel, _ = bitstream.ReadUint8()
		}
		if reliablePacket.HasSplitPacket {
			reliablePacket.SplitPacketCount, _ = bitstream.ReadUint32BE()
			reliablePacket.SplitPacketID, _ = bitstream.ReadUint16BE()
			reliablePacket.SplitPacketIndex, _ = bitstream.ReadUint32BE()
		}
		reliablePacket.SelfData, _ = bitstream.ReadString(int((reliablePacket.LengthInBits + 7)/8))

		if reliablePacket.HasSplitPacket {
			HandleSplitPacket(reliablePacket, rakNetPacket, context, packet)
		} else {
			reliablePacket.SplitPacketCount = 1
			reliablePacket.FinalData = reliablePacket.SelfData
			reliablePacket.IsFinal = true
			reliablePacket.AllReliablePackets = []*ReliablePacket{reliablePacket}
			reliablePacket.AllRakNetLayers = []*RakNetLayer{rakNetPacket}
		}
		layer.Packets = append(layer.Packets, reliablePacket)
	}
	if err != io.EOF {
		return layer, err
	}
	return layer, nil
}
