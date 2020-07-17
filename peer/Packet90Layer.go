package peer

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
)

// Packet90VersionID represents a ID_PROTOCOL_SYNC version id
type Packet90VersionID [5]int32

// Packet90Layer represents ID_PROTOCOL_SYNC - client -> server
type Packet90Layer struct {
	SchemaVersion  uint32
	Int1           uint8
	Int2           uint8
	RequestedFlags []string
	JoinData       string
	PubKeyData     []byte
	VersionID      Packet90VersionID
}

func (stream *extendedReader) DecodePacket90Layer(reader PacketReader, layers *PacketLayers) (RakNetPacket, error) {
	layer := &Packet90Layer{}

	lenBytes := uint(layers.SplitPacket.RealLength) - 1 // -1 for packet id
	thisStream, err := stream.aesDecrypt(int(lenBytes), packet90AESKey)
	if err != nil {
		return layer, err
	}
	layer.Int2, err = thisStream.ReadByte()
	if err != nil {
		return layer, err
	}
	layer.SchemaVersion, err = thisStream.readUint32BE()
	if err != nil {
		return layer, err
	}
	layer.Int1, err = thisStream.ReadByte()
	if err != nil {
		return layer, err
	}

	flagsLen, err := thisStream.readUint16BE()
	if err != nil {
		return layer, err
	}

	layer.RequestedFlags = make([]string, flagsLen)
	for i := 0; i < int(flagsLen); i++ {
		layer.RequestedFlags[i], err = thisStream.readVarLengthString()
		if err != nil {
			return layer, err
		}
	}
	layer.JoinData, err = thisStream.readVarLengthString()
	if err != nil {
		return layer, err
	}

	placeIDRegex := regexp.MustCompile(`placeId=(\d+)`)
	placeIDMatches := placeIDRegex.FindStringSubmatch(layer.JoinData)
	if placeIDMatches != nil {
		placeID, _ := strconv.Atoi(placeIDMatches[1])
		reader.Context().PlaceID = int64(placeID)
	} else {
		return layer, errors.New("Could not match placeId regex (malformed JoinData)")
	}

	layer.PubKeyData, err = thisStream.readString(32)
	if err != nil {
		return layer, err
	}

	id, err := thisStream.readUint32BE()
	layer.VersionID[0] = int32(id)
	if err != nil {
		return layer, err
	}
	if layer.VersionID[0]&0xC == 0 {
		_, err = thisStream.readUint32BE()
		if err != nil {
			return layer, err
		}
	}
	id, err = thisStream.readUint32BE()
	layer.VersionID[1] = int32(id)
	if err != nil {
		return layer, err
	}
	if layer.VersionID[0]&0x50 == 0 {
		_, err = thisStream.readUint32BE()
		if err != nil {
			return layer, err
		}
	}
	id, err = thisStream.readUint32BE()
	layer.VersionID[2] = int32(id)
	if err != nil {
		return layer, err
	}
	if layer.VersionID[0]&0xA0 == 0 {
		_, err = thisStream.readUint32BE()
		if err != nil {
			return layer, err
		}
	}
	id, err = thisStream.readUint32BE()
	layer.VersionID[3] = int32(id)
	if err != nil {
		return layer, err
	}
	if layer.VersionID[0]&0x900 == 0 {
		_, err = thisStream.readUint32BE()
		if err != nil {
			return layer, err
		}
	}
	id, err = thisStream.readUint32BE()
	layer.VersionID[4] = int32(id)
	if err != nil {
		return layer, err
	}

	reader.Context().VersionID = layer.VersionID

	return layer, nil
}

// Serialize implements RakNetPacket.Serialize
func (layer *Packet90Layer) Serialize(writer PacketWriter, stream *extendedWriter) error {
	var err error
	rawStream := stream.aesEncrypt(packet90AESKey)
	err = rawStream.WriteByte(layer.Int2)
	if err != nil {
		return err
	}
	err = rawStream.writeUint32BE(layer.SchemaVersion)
	if err != nil {
		return err
	}
	err = rawStream.WriteByte(layer.Int1)
	if err != nil {
		return err
	}
	err = rawStream.writeUint16BE(uint16(len(layer.RequestedFlags)))
	if err != nil {
		return err
	}
	for i := 0; i < len(layer.RequestedFlags); i++ {
		err = rawStream.writeVarLengthString(layer.RequestedFlags[i])
		if err != nil {
			return err
		}
	}
	err = rawStream.writeVarLengthString(layer.JoinData)
	if err != nil {
		return err
	}

	err = rawStream.allBytes(layer.PubKeyData)
	if err != nil {
		return err
	}

	err = rawStream.writeUint32BE(uint32(layer.VersionID[0]))
	if err != nil {
		return err
	}
	if layer.VersionID[0]&0xC == 0 {
		err = rawStream.writeUint32BE(uint32(layer.VersionID[0]))
		if err != nil {
			return err
		}
	}
	err = rawStream.writeUint32BE(uint32(layer.VersionID[1]))
	if err != nil {
		return err
	}
	if layer.VersionID[0]&0x50 == 0 {
		err = rawStream.writeUint32BE(uint32(layer.VersionID[0]))
		if err != nil {
			return err
		}
	}
	err = rawStream.writeUint32BE(uint32(layer.VersionID[2]))
	if err != nil {
		return err
	}
	if layer.VersionID[0]&0xA0 == 0 {
		err = rawStream.writeUint32BE(uint32(layer.VersionID[0]))
		if err != nil {
			return err
		}
	}
	err = rawStream.writeUint32BE(uint32(layer.VersionID[3]))
	if err != nil {
		return err
	}
	if layer.VersionID[0]&0x900 == 0 {
		err = rawStream.writeUint32BE(uint32(layer.VersionID[0]))
		if err != nil {
			return err
		}
	}
	err = rawStream.writeUint32BE(uint32(layer.VersionID[4]))
	if err != nil {
		return err
	}

	err = rawStream.Close()
	if err != nil {
		return err
	}
	return nil
}

func (layer *Packet90Layer) String() string {
	return fmt.Sprintf("ID_PROTOCOL_SYNC: Version %d, %d flags", layer.SchemaVersion, len(layer.RequestedFlags))
}

// TypeString implements RakNetPacket.TypeString()
func (Packet90Layer) TypeString() string {
	return "ID_PROTOCOL_SYNC"
}

// Type implements RakNetPacket.Type()
func (Packet90Layer) Type() byte {
	return 0x90
}
