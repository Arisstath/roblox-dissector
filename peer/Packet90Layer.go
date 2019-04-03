package peer

import "fmt"

// ID_PROTOCOL_SYNC - client -> server
type Packet90Layer struct {
	SchemaVersion  uint32
	RequestedFlags []string
	JoinData       string
}

func NewPacket90Layer() *Packet90Layer {
	return &Packet90Layer{}
}

func (stream *extendedReader) DecodePacket90Layer(reader PacketReader, layers *PacketLayers) (RakNetPacket, error) {
	layer := NewPacket90Layer()

	lenBytes := bitsToBytes(uint(layers.Reliability.LengthInBits)) - 1 // -1 for packet id
	thisStream, err := stream.aesDecrypt(int(lenBytes))
	if err != nil {
		return layer, err
	}
	layer.SchemaVersion, err = thisStream.readUint32BE()

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
	return layer, nil
}

func (layer *Packet90Layer) Serialize(writer PacketWriter, stream *extendedWriter) error {
	err := stream.WriteByte(0x90)
	if err != nil {
		return err
	}
	rawStream := stream.aesEncrypt()
	err = rawStream.writeUint32BE(layer.SchemaVersion)
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
	err = rawStream.Close()
	if err != nil {
		return err
	}
	return nil
}

func (layer *Packet90Layer) String() string {
	return fmt.Sprintf("ID_PROTOCOL_SYNC: Version %d, %d flags", layer.SchemaVersion, len(layer.RequestedFlags))
}

func (Packet90Layer) TypeString() string {
	return "ID_PROTOCOL_SYNC"
}
