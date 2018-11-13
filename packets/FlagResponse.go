package packets

import "fmt"
import "github.com/gskartwii/roblox-dissector/util"

// ID_DICTIONARY_FORMAT - server -> client
// Response to ID_PROTOCOL_SYNC (FlagRequest)
type FlagResponse struct {
	ProtocolSchemaSync bool
	// Use dictionary compression?
	ApiDictionaryCompression bool
	// Flags set by the server
	Params map[string]bool
}

func NewFlagResponse() *FlagResponse {
	return &FlagResponse{Params: make(map[string]bool)}
}

func (thisBitstream *PacketReaderBitstream) DecodeFlagResponse(reader util.PacketReader, layers *PacketLayers) (RakNetPacket, error) {
	layer := NewFlagResponse()

	var err error
	layer.ProtocolSchemaSync, err = thisBitstream.ReadBool()
	if err != nil {
		return layer, err
	}
	layer.ApiDictionaryCompression, err = thisBitstream.ReadBool()
	if err != nil {
		return layer, err
	}
	thisBitstream.Align()

	numParams, err := thisBitstream.ReadUint16BE()
	if err != nil {
		return layer, err
	}

	var i uint16
	for i = 0; i < numParams; i++ {
		nameLen, err := thisBitstream.ReadUint16BE()
		if err != nil {
			return layer, err
		}
		name, err := thisBitstream.ReadString(int(nameLen))
		if err != nil {
			return layer, err
		}

		valueLen, err := thisBitstream.ReadUint16BE()
		if err != nil {
			return layer, err
		}
		value, err := thisBitstream.ReadString(int(valueLen))
		if err != nil {
			return layer, err
		}
		layer.Params[string(name)] = string(value) == "true"
	}

	return layer, nil
}

func (layer *FlagResponse) Serialize(writer util.PacketWriter, stream *PacketWriterBitstream) error {
	var err error
	err = stream.WriteByte(0x93)
	if err != nil {
		return err
	}

	err = stream.WriteBool(layer.ProtocolSchemaSync)
	if err != nil {
		return err
	}
	err = stream.WriteBool(layer.ApiDictionaryCompression)
	if err != nil {
		return err
	}
	err = stream.Align()
	if err != nil {
		return err
	}
	err = stream.WriteUint16BE(uint16(len(layer.Params)))
	if err != nil {
		return err
	}

	for name, value := range layer.Params {
		err = stream.WriteUint16BE(uint16(len(name)))
		if err != nil {
			return err
		}
		err = stream.WriteASCII(name)
		if err != nil {
			return err
		}
		encodedValue := fmt.Sprintf("%v", value)
		err = stream.WriteUint16BE(uint16(len(encodedValue)))
		if err != nil {
			return err
		}
		err = stream.WriteASCII(encodedValue)
		if err != nil {
			return err
		}
	}

	return err
}
