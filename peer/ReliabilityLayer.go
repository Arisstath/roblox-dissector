package peer
import "github.com/gskartwii/go-bitstream"
import "bytes"
import "io"
import "errors"

// ReliablePacket describes a packet within a ReliabilityLayer
type ReliablePacket struct {
	// Have all splits been received?
	IsFinal bool
	// Is this the first split?
	IsFirst bool
	// Unique ID given to each packet. Splits of the same packet have the same ID.
	UniqueID uint32
	// Reliability ID: (un)reliable? ordered? sequenced?
	Reliability uint32
	HasSplitPacket bool
	// Length of this split in bits
	LengthInBits uint16
	// Total length received so far, in bytes
	RealLength uint32
	// Unique ID given to each packet. Splits of the same packet have a different ID.
	ReliableMessageNumber uint32
	// Unchannelled sequencing index
	SequencingIndex uint32
	// Channelled ordering index
	OrderingIndex uint32
	OrderingChannel uint8
	// Count of splits this packet has
	SplitPacketCount uint32
	// Splits of the same packet have the same SplitPacketID
	SplitPacketID uint16
	// 0 <= SplitPacketIndex < SplitPacketCount
	SplitPacketIndex uint32
	// Number of _ordered_ splits we have received so far
	NumReceivedSplits uint32
	AllRakNetLayers []*RakNetLayer
	AllReliablePackets []*ReliablePacket

	// Has a decoder routine started decoding this packet yet?
	HasBeenDecoded bool

	fullDataReader *extendedReader
	// Data contained by this split
	SelfData []byte

	// Has received packet type yet? Set to true when the first split of this packet
	// is received
	HasPacketType bool
	PacketType byte

	buffer *splitPacketBuffer
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
	thisBitstream := packet.stream

	var reliability uint64
	var err error
	for reliability, err = thisBitstream.bits(3); err == nil; reliability, err = thisBitstream.bits(3) {
		reliablePacket := NewReliablePacket()
		reliablePacket.Reliability = uint32(reliability)
		reliablePacket.HasSplitPacket, err = thisBitstream.readBool()
		if err != nil {
			return layer, err
		}
		thisBitstream.Align()

		reliablePacket.LengthInBits, err = thisBitstream.readUint16BE()
		if err != nil {
			return layer, err
		}
		if reliablePacket.LengthInBits < 8 {
			return layer, errors.New("Invalid length of 0!")
		}

		reliablePacket.RealLength = uint32((reliablePacket.LengthInBits + 7) / 8)
		if reliability >= 2 && reliability <= 4 {
			reliablePacket.ReliableMessageNumber, err = thisBitstream.readUint24LE()
			if err != nil {
				return layer, err
			}
		}
		if reliability == 1 || reliability == 4 {
			reliablePacket.SequencingIndex, err = thisBitstream.readUint24LE()
			if err != nil {
				return layer, err
			}
		}
		if reliability == 1 || reliability == 3 || reliability == 4 || reliability == 7 {
			reliablePacket.OrderingIndex, err = thisBitstream.readUint24LE()
			if err != nil {
				return layer, err
			}
			reliablePacket.OrderingChannel, err = thisBitstream.readUint8()
			if err != nil {
				return layer, err
			}
		}
		if reliablePacket.HasSplitPacket {
			reliablePacket.SplitPacketCount, err = thisBitstream.readUint32BE()
			if err != nil {
				return layer, err
			}
			reliablePacket.SplitPacketID, err = thisBitstream.readUint16BE()
			if err != nil {
				return layer, err
			}
			reliablePacket.SplitPacketIndex, err = thisBitstream.readUint32BE()
			if err != nil {
				return layer, err
			}
		}
		reliablePacket.SelfData, err = thisBitstream.readString(int((reliablePacket.LengthInBits + 7)/8))
		if err != nil {
			return layer, err
		}
		if reliablePacket.SplitPacketIndex == 0 {
			reliablePacket.PacketType = reliablePacket.SelfData[0]
			reliablePacket.HasPacketType = true
		}
		if !reliablePacket.HasPacketType {
			reliablePacket.PacketType = 0xFF
		}

		reliablePacket.UniqueID = context.UniqueID
		context.UniqueID++

		if reliablePacket.HasSplitPacket {
			reliablePacket, err = context.handleSplitPacket(reliablePacket, rakNetPacket, packet)
			if err != nil {
				return layer, err
			}
		} else {
			reliablePacket.SplitPacketCount = 1
			reliablePacket.NumReceivedSplits = 1
			reliablePacket.fullDataReader = &extendedReader{bitstream.NewReader(bytes.NewReader(reliablePacket.SelfData))}
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

func (layer *ReliabilityLayer) serialize(isClient bool, context *CommunicationContext, outputStream *extendedWriter) error {
	var err error
	for _, packet := range layer.Packets {
		reliability := uint64(packet.Reliability)
		err = outputStream.bits(3, reliability)
		if err != nil {
			return err
		}
		err = outputStream.writeBool(packet.HasSplitPacket)
		if err != nil {
			return err
		}
		err = outputStream.Align()
		if err != nil {
			return err
		}
        packet.LengthInBits = uint16(len(packet.SelfData) * 8)
		err = outputStream.writeUint16BE(packet.LengthInBits)
		if err != nil {
			return err
		}
		if reliability >= 2 && reliability <= 4 {
			err = outputStream.writeUint24LE(packet.ReliableMessageNumber)
			if err != nil {
				return err
			}
		}
		if reliability == 1 || reliability == 4 {
			err = outputStream.writeUint24LE(packet.SequencingIndex)
			if err != nil {
				return err
			}
		}
		if reliability == 1 || reliability == 4 || reliability == 3 || reliability == 7 {
			err = outputStream.writeUint24LE(packet.OrderingIndex)
			if err != nil {
				return err
			}
			err = outputStream.WriteByte(byte(packet.OrderingChannel))
			if err != nil {
				return err
			}
		}
		if packet.HasSplitPacket {
			err = outputStream.writeUint32BE(packet.SplitPacketCount)
			if err != nil {
				return err
			}
			err = outputStream.writeUint16BE(packet.SplitPacketID)
			if err != nil {
				return err
			}
			err = outputStream.writeUint32BE(packet.SplitPacketIndex)
			if err != nil {
				return err
			}
		}
		err = outputStream.allBytes(packet.SelfData)
		if err != nil {
			return err
		}
	}
	return nil
}
