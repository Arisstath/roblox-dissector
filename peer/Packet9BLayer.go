package peer

import (
	"fmt"
)

// Packet9BLayer represents ID_LUAU_CHALLENGE
type Packet9BLayer struct {
	Int1      uint32
	Challenge uint32
	Response  uint32
	Script    []byte
	Signature []byte
}

func (thisStream *extendedReader) DecodePacket9BLayer(reader PacketReader, layers *PacketLayers) (RakNetPacket, error) {
	var err error
	layer := &Packet9BLayer{}

	if !reader.IsClient() {
		layer.Int1, err = thisStream.readUint32BE()
		if err != nil {
			return layer, err
		}
		layer.Challenge, err = thisStream.readUint32BE()
		if err != nil {
			return layer, err
		}
		length, err := thisStream.readUint32BE()
		if err != nil {
			return layer, err
		}
		layer.Script, err = thisStream.readString(int(length))
		if err != nil {
			return layer, err
		}
		layer.Signature, err = thisStream.readString(32)
		if err != nil {
			return layer, err
		}
		return layer, nil
	}

	layer.Challenge, err = thisStream.readUint32BE()
	if err != nil {
		return layer, err
	}
	layer.Response, err = thisStream.readUint32BE()
	if err != nil {
		return layer, err
	}
	return layer, nil
}

// Serialize implements RakNetPacket.Serialize
func (layer *Packet9BLayer) Serialize(writer PacketWriter, stream *extendedWriter) error {
	if writer.ToClient() {
		err := stream.writeUint32BE(layer.Int1)
		if err != nil {
			return err
		}
		err = stream.writeUint32BE(layer.Challenge)
		if err != nil {
			return err
		}
		err = stream.writeUint32BE(uint32(len(layer.Script)))
		if err != nil {
			return err
		}
		err = stream.allBytes(layer.Script)
		if err != nil {
			return err
		}
		return stream.allBytes(layer.Signature)
	}

	err := stream.writeUint32BE(layer.Challenge)
	if err != nil {
		return err
	}
	return stream.writeUint32BE(layer.Response)
}

func (layer *Packet9BLayer) String() string {
	return fmt.Sprintf("ID_LUAU_CHALLENGE: %08X", layer.Challenge)
}

// TypeString implements RakNetPacket.TypeString()
func (Packet9BLayer) TypeString() string {
	return "ID_LUAU_CHALLENGE"
}

// Type implements RakNetPacket.Type()
func (Packet9BLayer) Type() byte {
	return 0x9B
}
