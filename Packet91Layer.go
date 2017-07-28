package main
import "github.com/google/gopacket"
import "compress/gzip"
import "github.com/dgryski/go-bitstream"

type EnumSchemaItem struct {
	Name string
	BitSize uint32
}

type PropertySchemaItem struct {
	CommonID uint32
	Name string
	DictionaryType string
	Type string
	Bool1 bool
	IsEnum bool
	BitSize uint32
}
type EventSchemaItem struct {
	CommonID uint32
	Name string
	ArgumentTypes []string
}

type InstanceSchemaItem struct {
	CommonID uint32
	Name string
	IsCreatable bool
	PropertySchema []*PropertySchemaItem
	EventSchema []*EventSchemaItem
}

type Packet91Layer struct {
	EnumSchema []*EnumSchemaItem
	InstanceSchema []*InstanceSchemaItem
}

func NewPacket91Layer() Packet91Layer {
	return Packet91Layer{}
}

func DecodePacket91Layer(thisBitstream *ExtendedReader, context *CommunicationContext, packet gopacket.Packet) (interface{}, error) {
	typeDescriptor := context.TypeDescriptor

	layer := NewPacket91Layer()

	_, err := thisBitstream.Bits(32) // Void compressed len
	if err != nil {
		return layer, err
	}
	var decompressedStream *ExtendedReader
	gzipStream, err := gzip.NewReader(thisBitstream.GetReader())
	if err != nil {
		return layer, err
	}
	decompressedStream = &ExtendedReader{bitstream.NewReader(gzipStream)}
	thisLen, err := decompressedStream.ReadUint32BE()
	if err != nil {
		return layer, err
	}
	layer.EnumSchema = make([]*EnumSchemaItem, thisLen)
	var i, j, k uint32
	for i = 0; i < thisLen; i++ {
		name, err := decompressedStream.ReadLengthAndString()
		if err != nil {
			return layer, err
		}
		bitSize, err := decompressedStream.ReadUint32BE()
		if err != nil {
			return layer, err
		}
		layer.EnumSchema[i] = &EnumSchemaItem{string(name), bitSize}
	}

	thisLen, err = decompressedStream.ReadUint32BE()
	if err != nil {
		return layer, err
	}
	layer.InstanceSchema = make([]*InstanceSchemaItem, thisLen)

	for i = 0; i < thisLen; i++ {
		thisInstance := &InstanceSchemaItem{}
		thisInstance.CommonID, err = decompressedStream.ReadUint32BE()
		if err != nil {
			return layer, err
		}
		thisInstance.Name, err = decompressedStream.ReadLengthAndString()
		if err != nil {
			return layer, err
		}
		thisInstance.IsCreatable, err = decompressedStream.ReadBoolByte()
		if err != nil {
			return layer, err
		}

		len2, err := decompressedStream.ReadUint32BE()
		if err != nil {
			return layer, err
		}

		thisInstance.PropertySchema = make([]*PropertySchemaItem, len2)
		for j = 0; j < len2; j++ {
			thisProperty := &PropertySchemaItem{}
			thisProperty.CommonID, err = decompressedStream.ReadUint32BE()
			if err != nil {
				return layer, err
			}
			thisProperty.Name, err = decompressedStream.ReadLengthAndString()
			if err != nil {
				return layer, err
			}
			typeIDx, err := decompressedStream.ReadUint32BE()
			if err != nil {
				return layer, err
			}

			thisType := typeDescriptor[typeIDx]
			thisProperty.DictionaryType = thisType

			thisProperty.Type, err = decompressedStream.ReadLengthAndString()
			if err != nil {
				return layer, err
			}
			thisProperty.Bool1, err = decompressedStream.ReadBool()
			if err != nil {
				return layer, err
			}
			thisProperty.IsEnum, err = decompressedStream.ReadBool()
			if err != nil {
				return layer, err
			}
			if thisProperty.IsEnum {
				thisProperty.BitSize, err = decompressedStream.ReadUint32BE()
				if err != nil {
					return layer, err
				}
			}
			thisInstance.PropertySchema[j] = thisProperty
		}

		len3, err := decompressedStream.ReadUint32BE()
		if err != nil {
			return layer, err
		}
		thisInstance.EventSchema = make([]*EventSchemaItem, len3)
		for j = 0; j < len3; j++ {
			thisEvent := &EventSchemaItem{}
			thisEvent.CommonID, err = decompressedStream.ReadUint32BE()
			if err != nil {
				return layer, err
			}
			thisEvent.Name, err = decompressedStream.ReadLengthAndString()
			if err != nil {
				return layer, err
			}
			len4, err := decompressedStream.ReadUint32BE()
			if err != nil {
				return layer, err
			}

			thisEvent.ArgumentTypes = make([]string, len4)
			
			for k = 0; k < len4; k++ {
				typeIDx, err := decompressedStream.ReadUint32BE()
				if err != nil {
					return layer, err
				}

				thisType := typeDescriptor[typeIDx] 
				thisEvent.ArgumentTypes[k] = thisType
			}
			thisInstance.EventSchema[j] = thisEvent
		}
		layer.InstanceSchema[i] = thisInstance
	}

	return layer, nil
}
