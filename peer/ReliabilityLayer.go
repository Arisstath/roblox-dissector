package peer

import "io"
import "errors"

const (
	UNRELIABLE              = iota
	UNRELIABLE_SEQ          = iota
	RELIABLE                = iota
	RELIABLE_ORD            = iota
	RELIABLE_SEQ            = iota
	UNRELIABLE_ACK_RECP     = iota
	UNRELIABLE_SEQ_ACK_RECP = iota
	RELIABLE_ACK_RECP       = iota
	RELIABLE_ORD_ACK_RECP   = iota
	RELIABLE_SEQ_ACK_RECP   = iota
)

// ReliablePacket describes a packet within a ReliabilityLayer
type ReliablePacket struct {
	// Reliability ID: (un)reliable? ordered? sequenced?
	Reliability    uint32
	HasSplitPacket bool
	// Length of this split in bits
	LengthInBits uint16
	// Unique ID given to each packet. Splits of the same packet have a different ID.
	ReliableMessageNumber uint32
	// Unchannelled sequencing index
	SequencingIndex uint32
	// Channelled ordering index
	OrderingIndex   uint32
	OrderingChannel uint8
	// Count of splits this packet has
	SplitPacketCount uint32
	// Splits of the same packet have the same SplitPacketID
	SplitPacketID uint16
	// 0 <= SplitPacketIndex < SplitPacketCount
	SplitPacketIndex uint32
	// The RakNet layer containing this packet
	RakNetLayer *RakNetLayer

	// Data contained by this split
	SelfData []byte

	SplitBuffer *SplitPacketBuffer
}

type ReliabilityLayer struct {
	Packets []*ReliablePacket
}

func NewReliabilityLayer() *ReliabilityLayer {
	return &ReliabilityLayer{Packets: make([]*ReliablePacket, 0)}
}
func NewReliablePacket() *ReliablePacket {
	return &ReliablePacket{SelfData: []byte{}}
}

func (packet *ReliablePacket) GetLog() string {
	return packet.SplitBuffer.logBuffer.String()
}

func (packet *ReliablePacket) IsReliable() bool {
	return packet.Reliability == RELIABLE || packet.Reliability == RELIABLE_SEQ || packet.Reliability == RELIABLE_ORD
}
func (packet *ReliablePacket) IsSequenced() bool {
	return packet.Reliability == UNRELIABLE_SEQ || packet.Reliability == RELIABLE_SEQ
}
func (packet *ReliablePacket) IsOrdered() bool {
	return packet.Reliability == UNRELIABLE_SEQ || packet.Reliability == RELIABLE_SEQ || packet.Reliability == RELIABLE_ORD || packet.Reliability == RELIABLE_ORD_ACK_RECP
}

func DecodeReliabilityLayer(packet *UDPPacket, context *CommunicationContext, rakNetPacket *RakNetLayer) (*ReliabilityLayer, error) {
	layer := NewReliabilityLayer()
	thisBitstream := packet.stream

	var reliability uint64
	var err error
	for reliability, err = thisBitstream.bits(3); err == nil; reliability, err = thisBitstream.bits(3) {
		reliablePacket := NewReliablePacket()
		reliablePacket.RakNetLayer = rakNetPacket

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

		if reliablePacket.IsReliable() {
			reliablePacket.ReliableMessageNumber, err = thisBitstream.readUint24LE()
			if err != nil {
				return layer, err
			}
		}
		if reliablePacket.IsSequenced() {
			reliablePacket.SequencingIndex, err = thisBitstream.readUint24LE()
			if err != nil {
				return layer, err
			}
		}
		if reliablePacket.IsOrdered() {
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
		reliablePacket.SelfData, err = thisBitstream.readString(int((reliablePacket.LengthInBits + 7) / 8))
		if err != nil {
			return layer, err
		}
		if !reliablePacket.HasSplitPacket {
			reliablePacket.SplitPacketCount = 1
		}

		layer.Packets = append(layer.Packets, reliablePacket)
	}
	if err != io.EOF {
		return layer, err
	}
	return layer, nil
}

func (layer *ReliabilityLayer) Serialize(writer PacketWriter, outputStream *extendedWriter) error {
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
