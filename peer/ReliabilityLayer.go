package peer
import "github.com/gskartwii/go-bitstream"
import "bytes"
import "io"
import "errors"

type ReliablePacket struct {
	IsFinal bool
	IsFirst bool
	UniqueID uint32
	Reliability uint32
	HasSplitPacket bool
	LengthInBits uint16
	RealLength uint32
	ReliableMessageNumber uint32
	SequencingIndex uint32
	OrderingIndex uint32
	OrderingChannel uint8
	SplitPacketCount uint32
	SplitPacketID uint16
	SplitPacketIndex uint32
	NumReceivedSplits uint32
	AllRakNetLayers []*RakNetLayer
	AllReliablePackets []*ReliablePacket

	HasBeenDecoded bool

	FullDataReader *ExtendedReader
	SelfData []byte

	HasPacketType bool
	PacketType byte

	Buffer *SplitPacketBuffer
}

type ReliabilityLayer struct {
	Packets []*ReliablePacket
}

func NewReliabilityLayer() *ReliabilityLayer {
	return &ReliabilityLayer{Packets: make([]*ReliablePacket, 0)}
}
func NewReliablePacket() *ReliablePacket {
	return &ReliablePacket{SelfData: make([]byte, 0)}
}

func DecodeReliabilityLayer(packet *UDPPacket, context *CommunicationContext, rakNetPacket *RakNetLayer) (*ReliabilityLayer, error) {
	layer := NewReliabilityLayer()
	thisBitstream := packet.Stream

	var reliability uint64
	var err error
	for reliability, err = thisBitstream.Bits(3); err == nil; reliability, err = thisBitstream.Bits(3) {
		reliablePacket := NewReliablePacket()
		reliablePacket.Reliability = uint32(reliability)
		reliablePacket.HasSplitPacket, err = thisBitstream.ReadBool()
		if err != nil {
			return layer, err
		}
		thisBitstream.Align()

		reliablePacket.LengthInBits, err = thisBitstream.ReadUint16BE()
		if err != nil {
			return layer, err
		}
		if reliablePacket.LengthInBits == 0 {
			return layer, errors.New("Invalid length of 0!")
		}

		reliablePacket.RealLength = uint32((reliablePacket.LengthInBits + 7) / 8)
		if reliability >= 2 && reliability <= 4 {
			reliablePacket.ReliableMessageNumber, err = thisBitstream.ReadUint24LE()
			if err != nil {
				return layer, err
			}
		}
		if reliability == 1 || reliability == 4 {
			reliablePacket.SequencingIndex, err = thisBitstream.ReadUint24LE()
			if err != nil {
				return layer, err
			}
		}
		if reliability == 1 || reliability == 3 || reliability == 4 || reliability == 7 {
			reliablePacket.OrderingIndex, err = thisBitstream.ReadUint24LE()
			if err != nil {
				return layer, err
			}
			reliablePacket.OrderingChannel, err = thisBitstream.ReadUint8()
			if err != nil {
				return layer, err
			}
		}
		if reliablePacket.HasSplitPacket {
			reliablePacket.SplitPacketCount, err = thisBitstream.ReadUint32BE()
			if err != nil {
				return layer, err
			}
			reliablePacket.SplitPacketID, err = thisBitstream.ReadUint16BE()
			if err != nil {
				return layer, err
			}
			reliablePacket.SplitPacketIndex, err = thisBitstream.ReadUint32BE()
			if err != nil {
				return layer, err
			}
		}
		reliablePacket.SelfData, err = thisBitstream.ReadString(int((reliablePacket.LengthInBits + 7)/8))
		if err != nil {
			return layer, err
		}
		if reliablePacket.SplitPacketIndex == 0 {
			reliablePacket.HasPacketType = true
			reliablePacket.PacketType = reliablePacket.SelfData[0]
		} else if !reliablePacket.HasPacketType {
			reliablePacket.PacketType = 0xFF
		}
		reliablePacket.UniqueID = context.UniqueID
		context.UniqueID++

		if reliablePacket.HasSplitPacket {
			reliablePacket, err = context.HandleSplitPacket(reliablePacket, rakNetPacket, packet)
			if err != nil {
				return layer, err
			}
		} else {
			reliablePacket.SplitPacketCount = 1
			reliablePacket.NumReceivedSplits = 1
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

func (layer *ReliabilityLayer) Serialize(outputStream *ExtendedWriter) error {
	var err error
	for _, packet := range layer.Packets {
		reliability := uint64(packet.Reliability)
		err = outputStream.Bits(3, reliability)
		if err != nil {
			return err
		}
		err = outputStream.WriteBool(packet.HasSplitPacket)
		if err != nil {
			return err
		}
		err = outputStream.Align()
		if err != nil {
			return err
		}
        packet.LengthInBits = uint16(len(packet.SelfData) * 8)
		err = outputStream.WriteUint16BE(packet.LengthInBits)
		if err != nil {
			return err
		}
		if reliability >= 2 && reliability <= 4 {
			err = outputStream.WriteUint24LE(packet.ReliableMessageNumber)
			if err != nil {
				return err
			}
		}
		if reliability == 1 || reliability == 4 {
			err = outputStream.WriteUint24LE(packet.SequencingIndex)
			if err != nil {
				return err
			}
		}
		if reliability == 1 || reliability == 4 || reliability == 3 || reliability == 7 {
			err = outputStream.WriteUint24LE(packet.OrderingIndex)
			if err != nil {
				return err
			}
			err = outputStream.WriteByte(byte(packet.OrderingChannel))
			if err != nil {
				return err
			}
		}
		if packet.HasSplitPacket {
			err = outputStream.WriteUint32BE(packet.SplitPacketCount)
			if err != nil {
				return err
			}
			err = outputStream.WriteUint16BE(packet.SplitPacketID)
			if err != nil {
				return err
			}
			err = outputStream.WriteUint32BE(packet.SplitPacketIndex)
			if err != nil {
				return err
			}
		}
		err = outputStream.AllBytes(packet.SelfData)
		if err != nil {
			return err
		}
	}
	return nil
}
