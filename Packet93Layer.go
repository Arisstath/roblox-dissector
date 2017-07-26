package main
import "github.com/google/gopacket"
import "github.com/dgryski/go-bitstream"
import "bytes"

type Packet93Layer struct {
	UnknownBool1 bool
	UnknownBool2 bool
	Params map[string]bool
}

func NewPacket93Layer() Packet93Layer {
	return Packet93Layer{Params: make(map[string]bool)}
}

func DecodePacket93Layer(data []byte, context *CommunicationContext, packet gopacket.Packet) (interface{}, error) {
	layer := NewPacket93Layer()
	thisBitstream := ExtendedReader{bitstream.NewReader(bytes.NewReader(data[1:]))}

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

	return layer, nil
}
