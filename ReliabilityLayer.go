package main
import "github.com/gskartwii/go-bitstream"
import "github.com/google/gopacket"
import "bytes"
import "io"

type ReliablePacket struct {
	IsFinal bool
	IsFirst bool
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
	AllRakNetLayers []*RakNetLayer
	AllReliablePackets []*ReliablePacket

	FullDataReader *ExtendedReader
	SelfData []byte
}

type ReliabilityLayer struct {
	Packets []*ReliablePacket
}

type SplitPacketBuffer struct {
	ReliablePackets []*ReliablePacket
	RakNetPackets []*RakNetLayer
	NextExpectedPacket uint32
	WrittenPackets uint32

	DataReader *ExtendedReader
	DataWriter *io.PipeWriter
	DataQueue chan []byte
}
type SplitPacketList map[string](map[uint16](*SplitPacketBuffer))

var SplitPackets SplitPacketList

func NewSplitPacketBuffer(packet *ReliablePacket) *SplitPacketBuffer {
	reliables := make([]*ReliablePacket, int(packet.SplitPacketCount))
	raknets := make([]*RakNetLayer, int(packet.SplitPacketCount))

	list := &SplitPacketBuffer{reliables, raknets, 0, 0, nil, nil, make(chan []byte, packet.SplitPacketCount)}
	reader, writer := io.Pipe()
	list.DataReader = &ExtendedReader{bitstream.NewReader(reader)}
	list.DataWriter = writer

	go func() {
		for list.WrittenPackets < packet.SplitPacketCount {
			writer.Write(<- list.DataQueue)
			list.WrittenPackets++
			if list.WrittenPackets == packet.SplitPacketCount {
				err := writer.Close()
				if err != nil {
					println("Warning: failed to close split packet stream:", err.Error())
				}
			}
		}
	}()

	return list
}
func (list *SplitPacketBuffer) AddPacket(packet *ReliablePacket, rakNetPacket *RakNetLayer, index uint32)  {
	// Packets may be duplicated. At least I think so. Thanks UDP
	list.ReliablePackets[index] = packet
	list.RakNetPackets[index] = rakNetPacket
}

func AddSplitPacket(source string, packet *ReliablePacket, rakNetPacket *RakNetLayer) *SplitPacketBuffer {
	splitPacketId := packet.SplitPacketID
	splitPacketIndex := packet.SplitPacketIndex

	if SplitPackets == nil {
		var currentList = NewSplitPacketBuffer(packet)
		currentList.AddPacket(packet, rakNetPacket, splitPacketIndex)

		SplitPackets = SplitPacketList{source: map[uint16]*SplitPacketBuffer{splitPacketId: currentList}}
	} else if SplitPackets[source] == nil {
		var currentList = NewSplitPacketBuffer(packet)
		currentList.AddPacket(packet, rakNetPacket, splitPacketIndex)

		SplitPackets[source] = map[uint16]*SplitPacketBuffer{splitPacketId: currentList}
	} else if SplitPackets[source][splitPacketId] == nil {
		var currentList = NewSplitPacketBuffer(packet)
		currentList.AddPacket(packet, rakNetPacket, splitPacketIndex)

		SplitPackets[source][splitPacketId] = currentList
	} else {
		SplitPackets[source][splitPacketId].AddPacket(packet, rakNetPacket, splitPacketIndex)
	}

	return SplitPackets[source][splitPacketId]
}

func (reliablePacket *ReliablePacket) HandleSplitPacket(rakNetPacket *RakNetLayer, context *CommunicationContext, packet gopacket.Packet) error {
	source := SourceInterfaceFromPacket(packet)
	packetBuffer := AddSplitPacket(source, reliablePacket, rakNetPacket)
	reliablePacket.AllReliablePackets = packetBuffer.ReliablePackets
	reliablePacket.AllRakNetLayers = packetBuffer.RakNetPackets
	expectedPacket := packetBuffer.NextExpectedPacket

	reliablePacket.FullDataReader = packetBuffer.DataReader

	if expectedPacket == 0 {
		reliablePacket.IsFirst = true
	}

	var shouldClose bool
	for len(packetBuffer.ReliablePackets) > int(expectedPacket) && packetBuffer.ReliablePackets[expectedPacket] != nil {
		shouldClose = len(packetBuffer.ReliablePackets) == int(expectedPacket + 1)
		packetBuffer.DataQueue <- packetBuffer.ReliablePackets[expectedPacket].SelfData

		expectedPacket++
		packetBuffer.NextExpectedPacket = expectedPacket
	}
	if shouldClose {
		reliablePacket.IsFinal = true
	}
	return nil
}

func NewReliabilityLayer() *ReliabilityLayer {
	return &ReliabilityLayer{Packets: make([]*ReliablePacket, 0)}
}
func NewReliablePacket() *ReliablePacket {
	return &ReliablePacket{SelfData: make([]byte, 0)}
}

func DecodeReliabilityLayer(thisBitstream *ExtendedReader, context *CommunicationContext, packet gopacket.Packet, rakNetPacket *RakNetLayer) (*ReliabilityLayer, error) {
	layer := NewReliabilityLayer()

	var reliability uint64
	var err error
	for reliability, err = thisBitstream.Bits(3); err == nil; reliability, err = thisBitstream.Bits(3) {
		reliablePacket := NewReliablePacket()
		reliablePacket.Reliability = uint32(reliability)
		reliablePacket.HasSplitPacket, _ = thisBitstream.ReadBool()
		thisBitstream.Align()

		reliablePacket.LengthInBits, _ = thisBitstream.ReadUint16BE()
		if reliability >= 2 && reliability <= 4 {
			reliablePacket.ReliableMessageNumber, _ = thisBitstream.ReadUint24LE()
		}
		if reliability == 1 || reliability == 4 {
			reliablePacket.SequencingIndex, _ = thisBitstream.ReadUint24LE()
		}
		if reliability == 1 || reliability == 3 || reliability == 4 || reliability == 7 {
			reliablePacket.OrderingIndex, _ = thisBitstream.ReadUint24LE()
			reliablePacket.OrderingChannel, _ = thisBitstream.ReadUint8()
		}
		if reliablePacket.HasSplitPacket {
			reliablePacket.SplitPacketCount, _ = thisBitstream.ReadUint32BE()
			reliablePacket.SplitPacketID, _ = thisBitstream.ReadUint16BE()
			reliablePacket.SplitPacketIndex, _ = thisBitstream.ReadUint32BE()
		}
		reliablePacket.SelfData, _ = thisBitstream.ReadString(int((reliablePacket.LengthInBits + 7)/8))

		if reliablePacket.HasSplitPacket {
			err = reliablePacket.HandleSplitPacket(rakNetPacket, context, packet)
			if err != nil {
				return layer, err
			}
		} else {
			reliablePacket.SplitPacketCount = 1
			reliablePacket.FullDataReader = &ExtendedReader{bitstream.NewReader(bytes.NewReader(reliablePacket.SelfData))}
			reliablePacket.IsFirst = true
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
