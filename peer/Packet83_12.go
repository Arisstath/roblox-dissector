package peer

import (
	"fmt"
)

// Packet83_12 represents ID_HASH
type Packet83_12 struct {
	HashList          []uint32
	SecurityTokens    [3]uint64
	Nonce             uint32
	HasSecurityTokens bool
}

func (stream *extendedReader) DecodePacket83_12(reader PacketReader, layers *PacketLayers) (Packet83Subpacket, error) {
	var err error
	inner := &Packet83_12{}
	numItems, err := stream.readUint8()
	if err != nil {
		return inner, err
	}

	if numItems != 0xFF {
		//println("noextranumitem")
		inner.HasSecurityTokens = true
	} else {
		numItems, err = stream.readUint8()
		if err != nil {
			return inner, err
		}
	}

	inner.Nonce, err = stream.readUint32BE()
	if err != nil {
		return inner, err
	}
	hashList := make([]uint32, numItems)
	for i := 0; i < int(numItems); i++ {
		hashList[i], err = stream.readUint32BE()
		if err != nil {
			return inner, err
		}
	}

	if inner.HasSecurityTokens {
		for i := 0; i < 3; i++ {
			inner.SecurityTokens[i], err = stream.readUint64BE()
			if err != nil {
				return inner, err
			}
		}
	}

	inner.HashList = hashList

	return inner, nil
}

// Serialize implements Packet83Subpacket.Serialize()
func (layer *Packet83_12) Serialize(writer PacketWriter, stream *extendedWriter) error {
	if !layer.HasSecurityTokens {
		err := stream.WriteByte(0xFF)
		if err != nil {
			return err
		}
	}
	err := stream.WriteByte(uint8(len(layer.HashList)))
	if err != nil {
		return err
	}
	err = stream.writeUint32BE(layer.Nonce)
	if err != nil {
		return err
	}
	for _, hash := range layer.HashList {
		err = stream.writeUint32BE(hash)
		if err != nil {
			return err
		}
	}
	if layer.HasSecurityTokens {
		for _, token := range layer.SecurityTokens {
			err = stream.writeUint64BE(token)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// Type implements Packet83Subpacket.Type()
func (Packet83_12) Type() uint8 {
	return 0x12
}

// TypeString implements Packet83Subpacket.TypeString()
func (Packet83_12) TypeString() string {
	return "ID_REPLIC_HASH"
}
func (layer *Packet83_12) String() string {
	return fmt.Sprintf("ID_REPLIC_HASH: %d hashes", len(layer.HashList))
}
