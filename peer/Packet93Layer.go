package peer

import "fmt"

// Packet93Layer represents ID_DICTIONARY_FORMAT - server -> client
// Response to ID_PROTOCOL_SYNC (Packet90Layer)
type Packet93Layer struct {
	ProtocolSchemaSync bool
	// Use dictionary compression?
	APIDictionaryCompression bool
	// Flags set by the server
	Params map[string]bool
}

func (thisStream *extendedReader) DecodePacket93Layer(reader PacketReader, layers *PacketLayers) (RakNetPacket, error) {
	layer := &Packet93Layer{}
	layer.Params = make(map[string]bool)

	flags, err := thisStream.ReadByte()
	if err != nil {
		return layer, err
	}
	layer.ProtocolSchemaSync = flags&1 == 1
	layer.APIDictionaryCompression = flags&2 == 2
	if err != nil {
		return layer, err
	}

	numParams, err := thisStream.readUint16BE()
	if err != nil {
		return layer, err
	}

	var i uint16
	for i = 0; i < numParams; i++ {
		nameLen, err := thisStream.readUint16BE()
		if err != nil {
			return layer, err
		}
		name, err := thisStream.readString(int(nameLen))
		if err != nil {
			return layer, err
		}

		valueLen, err := thisStream.readUint16BE()
		if err != nil {
			return layer, err
		}
		value, err := thisStream.readString(int(valueLen))
		if err != nil {
			return layer, err
		}
		layer.Params[string(name)] = string(value) == "true"
	}

	return layer, nil
}

// Serialize implements RakNetPacket.Serialize
func (layer *Packet93Layer) Serialize(writer PacketWriter, stream *extendedWriter) error {
	var err error

	var flags byte
	if layer.ProtocolSchemaSync {
		flags |= 1
	}
	if layer.APIDictionaryCompression {
		flags |= 2
	}
	err = stream.WriteByte(flags)
	if err != nil {
		return err
	}
	err = stream.writeUint16BE(uint16(len(layer.Params)))
	if err != nil {
		return err
	}

	for name, value := range layer.Params {
		err = stream.writeUint16BE(uint16(len(name)))
		if err != nil {
			return err
		}
		err = stream.writeASCII(name)
		if err != nil {
			return err
		}
		encodedValue := fmt.Sprintf("%v", value)
		err = stream.writeUint16BE(uint16(len(encodedValue)))
		if err != nil {
			return err
		}
		err = stream.writeASCII(encodedValue)
		if err != nil {
			return err
		}
	}

	return err
}

func (layer *Packet93Layer) String() string {
	return fmt.Sprintf("ID_DICTIONARY_FORMAT: %d flags", len(layer.Params))
}

// TypeString implements RakNetPacket.TypeString()
func (Packet93Layer) TypeString() string {
	return "ID_DICTIONARY_FORMAT"
}

// Type implements RakNetPacket.Type()
func (Packet93Layer) Type() byte {
	return 0x93
}
