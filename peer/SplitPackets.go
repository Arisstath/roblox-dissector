package peer
import "github.com/gskartwii/roblox-dissector/packets"

func (reader *DefaultPacketReader) addSplitPacket(layers *packets.PacketLayers) *packets.SplitPacketBuffer {
	packet := layers.Reliability
	splitPacketId := packet.SplitPacketID
	splitPacketIndex := packet.SplitPacketIndex

	if !packet.HasSplitPacket {
		buffer := packets.NewSplitPacketBuffer(packet, reader.context)
		buffer.AddPacket(packet, layers.RakNet, 0)

		return buffer
	}

	var buffer *SplitPacketBuffer
	if reader.splitPackets == nil {
		buffer = packets.NewSplitPacketBuffer(packet, reader.context)

		reader.splitPackets = map[uint16]*packets.SplitPacketBuffer{splitPacketId: buffer}
	} else if reader.splitPackets[splitPacketId] == nil {
		buffer = packets.NewSplitPacketBuffer(packet, reader.context)

		reader.splitPackets[splitPacketId] = buffer
	} else {
		buffer = reader.splitPackets[splitPacketId]
	}
	buffer.addPacket(packet, layers.RakNet, splitPacketIndex)
	packet.SplitBuffer = buffer

	return buffer
}

func (reader *DefaultPacketReader) handleSplitPacket(layers *packets.PacketLayers) (*SplitPacketBuffer, error) {
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
		packetBuffer.DataReader = &packets.PacketReaderBitstream{&packets.BitstreamReader{bitstream.NewReader(bytes.NewReader(packetBuffer.Data))}}
		if reliablePacket.HasSplitPacket {
			// TODO: Use a linked list
			reader.splitPackets.Delete(layers)
		}
	}
	packetBuffer.NumReceivedSplits = expectedPacket

	if reliablePacket.SplitPacketIndex == 0 {
		packetBuffer.PacketType = reliablePacket.SelfData[0]
		packetBuffer.HasPacketType = true
	}

	layers.Root.Logger = packetBuffer.Logger
	layers.Root.LogBuffer = packetBuffer.LogBuffer

	return packetBuffer, nil
}
