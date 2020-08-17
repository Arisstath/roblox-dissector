package peer

import "io"
import "errors"

const (
	// Unreliable is the RakNet UNRELIABLE reliability
	Unreliable = iota
	// UnreliableSequenced is the RakNet UNRELIABLE_SEQUENCED reliability
	UnreliableSequenced
	// Reliable is the RakNet RELIABLE reliability
	Reliable
	// ReliableOrdered is the RakNet RELIABLE_ORDERED reliability
	ReliableOrdered
	// ReliableSequenced is the RakNet RELIABLE_SEQUENCED reliability
	ReliableSequenced
)

// ReliablePacket describes a packet within a ReliabilityLayer
type ReliablePacket struct {
	// Reliability ID: (un)reliable? ordered? sequenced?
	Reliability    uint8
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

// ReliabilityLayer describes a RakNet connection packet container
type ReliabilityLayer struct {
	Packets []*ReliablePacket
}

// IsReliable returns a bool describing whether the packet's reliability is considered "Reliable"
func (packet *ReliablePacket) IsReliable() bool {
	return packet.Reliability == Reliable || packet.Reliability == ReliableSequenced || packet.Reliability == ReliableOrdered
}

// IsSequenced returns a bool descibing whether the packet's reliability is considered "Sequenced"
func (packet *ReliablePacket) IsSequenced() bool {
	return packet.Reliability == UnreliableSequenced || packet.Reliability == ReliableSequenced
}

// IsOrdered returns a bool descibing whether the packet's reliability is considered "Ordered"
func (packet *ReliablePacket) IsOrdered() bool {
	return packet.Reliability == UnreliableSequenced || packet.Reliability == ReliableSequenced || packet.Reliability == ReliableOrdered
}

func (thisStream *extendedReader) DecodeReliabilityLayer(reader PacketReader, layers *PacketLayers) (*ReliabilityLayer, error) {
	layer := &ReliabilityLayer{}

	var reliability uint8
	var hasSplitPacket bool
	var err error
	for reliability, hasSplitPacket, err = thisStream.readReliabilityFlags(); err == nil; reliability, hasSplitPacket, err = thisStream.readReliabilityFlags() {
		reliablePacket := &ReliablePacket{}
		reliablePacket.RakNetLayer = layers.RakNet

		reliablePacket.Reliability = reliability
		reliablePacket.HasSplitPacket = hasSplitPacket

		reliablePacket.LengthInBits, err = thisStream.readUint16BE()
		if err != nil {
			return layer, err
		}
		if reliablePacket.LengthInBits < 8 {
			return layer, errors.New("invalid length of 0")
		}

		if reliablePacket.IsReliable() {
			reliablePacket.ReliableMessageNumber, err = thisStream.readUint24LE()
			if err != nil {
				return layer, err
			}
		}
		if reliablePacket.IsSequenced() {
			reliablePacket.SequencingIndex, err = thisStream.readUint24LE()
			if err != nil {
				return layer, err
			}
		}
		if reliablePacket.IsOrdered() {
			reliablePacket.OrderingIndex, err = thisStream.readUint24LE()
			if err != nil {
				return layer, err
			}
			reliablePacket.OrderingChannel, err = thisStream.readUint8()
			if err != nil {
				return layer, err
			}
		}
		if reliablePacket.HasSplitPacket {
			reliablePacket.SplitPacketCount, err = thisStream.readUint32BE()
			if err != nil {
				return layer, err
			}
			reliablePacket.SplitPacketID, err = thisStream.readUint16BE()
			if err != nil {
				return layer, err
			}
			reliablePacket.SplitPacketIndex, err = thisStream.readUint32BE()
			if err != nil {
				return layer, err
			}
		}
		reliablePacket.SelfData, err = thisStream.readString(int((reliablePacket.LengthInBits + 7) / 8))
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

// Serialize serializes the packet to its network format
func (layer *ReliabilityLayer) Serialize(writer PacketWriter, outputStream *extendedWriter) error {
	var err error
	for _, packet := range layer.Packets {
		reliability := packet.Reliability
		err = outputStream.writeReliabilityFlags(reliability, packet.HasSplitPacket)
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

// Copy creates a safe copy of a ReliablePacket.
// It doesn't retain the data or its length.
func (packet *ReliablePacket) Copy() *ReliablePacket {
	newPacket := new(ReliablePacket)
	newPacket.Reliability = packet.Reliability
	newPacket.HasSplitPacket = packet.HasSplitPacket
	newPacket.OrderingChannel = packet.OrderingChannel
	newPacket.OrderingIndex = packet.OrderingIndex
	newPacket.ReliableMessageNumber = packet.ReliableMessageNumber
	newPacket.SequencingIndex = packet.SequencingIndex
	newPacket.SplitPacketCount = packet.SplitPacketCount
	newPacket.SplitPacketID = packet.SplitPacketID
	newPacket.SplitPacketIndex = packet.SplitPacketIndex
	newPacket.SplitBuffer = packet.SplitBuffer

	return newPacket
}
