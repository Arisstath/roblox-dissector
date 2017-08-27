package peer

type Packet93Layer struct {
	UnknownBool1 bool
	UnknownBool2 bool
	Params map[string]bool
}

func NewPacket93Layer() Packet93Layer {
	return Packet93Layer{Params: make(map[string]bool)}
}

func DecodePacket93Layer(packet *UDPPacket, context *CommunicationContext) (interface{}, error) {
	layer := NewPacket93Layer()
	thisBitstream := packet.Stream

	var err error
	layer.UnknownBool1, err = thisBitstream.ReadBool()
	if err != nil {
		return layer, err
	}
	layer.UnknownBool2, err = thisBitstream.ReadBool()
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
