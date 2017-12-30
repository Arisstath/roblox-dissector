package peer
import "fmt"

type Packet93Layer struct {
	ProtocolSchemaSync bool
	ApiDictionaryCompression bool
	Params map[string]bool
}

func NewPacket93Layer() *Packet93Layer {
	return &Packet93Layer{Params: make(map[string]bool)}
}

func DecodePacket93Layer(packet *UDPPacket, context *CommunicationContext) (interface{}, error) {
	layer := NewPacket93Layer()
	thisBitstream := packet.Stream

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
	if val, ok := layer.Params["UseNetworkSchema2"]; val && ok {
		context.MSchema.Lock()
		context.UseStaticSchema = true
		context.ESchemaParsed.Broadcast()
		context.MSchema.Unlock()
	}

	return layer, nil
}

func (layer *Packet93Layer) Serialize(isClient bool,context *CommunicationContext, stream *ExtendedWriter) error {
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
