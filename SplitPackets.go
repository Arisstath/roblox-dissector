package main
import "io"
import "github.com/gskartwii/go-bitstream"
import "github.com/google/gopacket"

type SplitPacketBuffer struct {
	ReliablePackets []*ReliablePacket
	RakNetPackets []*RakNetLayer
	NextExpectedPacket uint32
	WrittenPackets uint32

	DataReader *ExtendedReader
	DataWriter *io.PipeWriter
	DataQueue chan []byte
}
type SplitPacketList map[string](map[uint16](*ReliablePacket))

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

func (SplitPackets SplitPacketList) AddSplitPacket(source string, packet *ReliablePacket, rakNetPacket *RakNetLayer) *ReliablePacket {
	splitPacketId := packet.SplitPacketID
	splitPacketIndex := packet.SplitPacketIndex

	if SplitPackets == nil {
		packet.Buffer = NewSplitPacketBuffer(packet)
		packet.IsFirst = true
		packet.FullDataReader = packet.Buffer.DataReader
		packet.Buffer.AddPacket(packet, rakNetPacket, splitPacketIndex)

		SplitPackets = SplitPacketList{source: map[uint16]*ReliablePacket{splitPacketId: packet}}
	} else if SplitPackets[source] == nil {
		packet.Buffer = NewSplitPacketBuffer(packet)
		packet.IsFirst = true
		packet.FullDataReader = packet.Buffer.DataReader
		packet.Buffer.AddPacket(packet, rakNetPacket, splitPacketIndex)

		SplitPackets[source] = map[uint16]*ReliablePacket{splitPacketId: packet}
	} else if SplitPackets[source][splitPacketId] == nil {
		packet.Buffer = NewSplitPacketBuffer(packet)
		packet.IsFirst = true
		packet.FullDataReader = packet.Buffer.DataReader
		packet.Buffer.AddPacket(packet, rakNetPacket, splitPacketIndex)

		SplitPackets[source][splitPacketId] = packet
	} else {
		SplitPackets[source][splitPacketId].IsFirst = false
		SplitPackets[source][splitPacketId].Buffer.AddPacket(packet, rakNetPacket, splitPacketIndex)
	}

	return SplitPackets[source][splitPacketId]
}

func (context *CommunicationContext) HandleSplitPacket(reliablePacket *ReliablePacket, rakNetPacket *RakNetLayer, packet gopacket.Packet) (*ReliablePacket, error) {
	source := SourceInterfaceFromPacket(packet)

	fullPacket := context.SplitPackets.AddSplitPacket(source, reliablePacket, rakNetPacket)
	packetBuffer := fullPacket.Buffer
	expectedPacket := packetBuffer.NextExpectedPacket

	fullPacket.AllReliablePackets = append(fullPacket.AllReliablePackets, reliablePacket)
	fullPacket.AllRakNetLayers = append(fullPacket.AllRakNetLayers, rakNetPacket)

	fullPacket.RealLength += uint32(len(reliablePacket.SelfData))

	var shouldClose bool
	for len(packetBuffer.ReliablePackets) > int(expectedPacket) && packetBuffer.ReliablePackets[expectedPacket] != nil {
		shouldClose = len(packetBuffer.ReliablePackets) == int(expectedPacket + 1)
		packetBuffer.DataQueue <- packetBuffer.ReliablePackets[expectedPacket].SelfData

		expectedPacket++
		packetBuffer.NextExpectedPacket = expectedPacket
	}
	if shouldClose {
		fullPacket.IsFinal = true
	}
	fullPacket.NumReceivedSplits = expectedPacket

	if reliablePacket.HasPacketType {
		fullPacket.HasPacketType = true
		fullPacket.PacketType = reliablePacket.PacketType
	}

	return fullPacket, nil
}
