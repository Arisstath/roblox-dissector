package peer

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/DataDog/zstd"
)

// Packet97Layer represents ID_NEW_SCHEMA - server -> client
// Negotiates a network schema with the client
type Packet97Layer struct {
	Schema *NetworkSchema
}

func (thisStream *extendedReader) DecodePacket97Layer(reader PacketReader, layers *PacketLayers) (RakNetPacket, error) {
	layer := &Packet97Layer{}

	var err error
	stream, err := thisStream.RegionToZStdStream()
	if err != nil {
		return layer, err
	}

	enumArrayLen, err := stream.readUintUTF8()
	if err != nil {
		return layer, err
	}
	if enumArrayLen > 0x10000 {
		return layer, errors.New("sanity check: exceeded maximum enum array len")
	}
	layer.Schema.Enums = make([]*NetworkEnumSchema, enumArrayLen)
	for i := 0; i < int(enumArrayLen); i++ {
		stringLen, err := stream.readUintUTF8()
		if err != nil {
			return layer, err
		}
		layer.Schema.Enums[i] = new(NetworkEnumSchema)
		layer.Schema.Enums[i].Name, err = stream.readASCII(int(stringLen))
		if err != nil {
			return layer, err
		}
		layer.Schema.Enums[i].BitSize, err = stream.readUint8()
		if err != nil {
			return layer, err
		}
		layer.Schema.Enums[i].NetworkID = uint16(i)
	}

	classArrayLen, err := stream.readUintUTF8()
	if err != nil {
		return layer, err
	}
	propertyArrayLen, err := stream.readUintUTF8()
	if err != nil {
		return layer, err
	}
	eventArrayLen, err := stream.readUintUTF8()
	if err != nil {
		return layer, err
	}
	if classArrayLen > 0x10000 {
		return layer, errors.New("sanity check: exceeded maximum class array len")
	}
	if propertyArrayLen > 0x10000 {
		return layer, errors.New("sanity check: exceeded maximum property array len")
	}
	if eventArrayLen > 0x10000 {
		return layer, errors.New("sanity check: exceeded maximum event array len")
	}
	layer.Schema.Instances = make([]*NetworkInstanceSchema, classArrayLen)
	layer.Schema.Properties = make([]*NetworkPropertySchema, propertyArrayLen)
	layer.Schema.Events = make([]*NetworkEventSchema, eventArrayLen)
	propertyGlobalIndex := 0
	classGlobalIndex := 0
	eventGlobalIndex := 0
	for i := 0; i < int(classArrayLen); i++ {
		classNameLen, err := stream.readUintUTF8()
		if err != nil {
			return layer, err
		}
		thisInstance := new(NetworkInstanceSchema)
		layer.Schema.Instances[i] = thisInstance
		thisInstance.Name, err = stream.readASCII(int(classNameLen))
		if err != nil {
			return layer, err
		}
		thisInstance.NetworkID = uint16(i)

		propertyCount, err := stream.readUintUTF8()
		if err != nil {
			return layer, err
		}
		if propertyCount > 0x10000 {
			return layer, errors.New("sanity check: exceeded maximum property count")
		}
		thisInstance.Properties = make([]*NetworkPropertySchema, propertyCount)

		for j := 0; j < int(propertyCount); j++ {
			thisProperty := new(NetworkPropertySchema)
			thisInstance.Properties[j] = thisProperty
			propNameLen, err := stream.readUintUTF8()
			if err != nil {
				return layer, err
			}
			thisProperty.Name, err = stream.readASCII(int(propNameLen))
			if err != nil {
				return layer, err
			}
			thisProperty.NetworkID = uint16(propertyGlobalIndex)

			thisProperty.Type, err = stream.readUint8()
			if err != nil {
				return layer, err
			}
			thisProperty.TypeString = TypeNames[thisProperty.Type]

			if thisProperty.Type == 7 {
				thisProperty.EnumID, err = stream.readUint16BE()
				if err != nil {
					return layer, err
				}
			}

			if int(propertyGlobalIndex) >= int(propertyArrayLen) {
				return layer, errors.New("property global index too high")
			}

			thisProperty.InstanceSchema = thisInstance
			layer.Schema.Properties[propertyGlobalIndex] = thisProperty
			propertyGlobalIndex++
		}

		thisInstance.Unknown, err = stream.readUint16BE()
		if err != nil {
			return layer, err
		}
		if int(classGlobalIndex) >= int(classArrayLen) {
			return layer, errors.New("class global index too high")
		}

		eventCount, err := stream.readUintUTF8()
		if err != nil {
			return layer, err
		}
		if eventCount > 0x10000 {
			return layer, errors.New("sanity check: exceeded maximum event count")
		}
		thisInstance.Events = make([]*NetworkEventSchema, eventCount)

		for j := 0; j < int(eventCount); j++ {
			thisEvent := new(NetworkEventSchema)
			thisInstance.Events[j] = thisEvent
			eventNameLen, err := stream.readUintUTF8()
			if err != nil {
				return layer, err
			}
			thisEvent.Name, err = stream.readASCII(int(eventNameLen))
			if err != nil {
				return layer, err
			}
			thisEvent.NetworkID = uint16(eventGlobalIndex)

			countArguments, err := stream.readUintUTF8()
			if err != nil {
				return layer, err
			}
			if countArguments > 0x10000 {
				return layer, errors.New("sanity check: exceeded maximum argument count")
			}
			thisEvent.Arguments = make([]*NetworkArgumentSchema, countArguments)

			for k := 0; k < int(countArguments); k++ {
				thisArgument := new(NetworkArgumentSchema)
				thisEvent.Arguments[k] = thisArgument
				thisArgument.Type, err = stream.readUint8()
				if err != nil {
					return layer, err
				}
				thisArgument.TypeString = TypeNames[thisArgument.Type]
				thisArgument.EnumID, err = stream.readUint16BE()
				if err != nil {
					return layer, err
				}
			}
			thisEvent.InstanceSchema = thisInstance
			layer.Schema.Events[eventGlobalIndex] = thisEvent
			eventGlobalIndex++
		}
		layer.Schema.Instances[classGlobalIndex] = thisInstance
		classGlobalIndex++
	}

	reader.Context().NetworkSchema = layer.Schema

	return layer, err
}

// Serialize implements RakNetPacket.Serialize()
func (layer *Packet97Layer) Serialize(writer PacketWriter, stream *extendedWriter) error {
	var err error

	err = stream.WriteByte(0x97)
	if err != nil {
		return err
	}
	// TODO: General NewZStdBuf() method?
	uncompressedBuf := bytes.NewBuffer([]byte{})
	zstdBuf := bytes.NewBuffer([]byte{})
	middleStream := zstd.NewWriter(zstdBuf)
	zstdStream := &extendedWriter{uncompressedBuf}

	schema := layer.Schema
	err = zstdStream.writeUintUTF8(uint32(len(schema.Enums)))
	if err != nil {
		middleStream.Close()
		return err
	}
	for _, enum := range schema.Enums {
		err = zstdStream.writeUintUTF8(uint32(len(enum.Name)))
		if err != nil {
			middleStream.Close()
			return err
		}
		err = zstdStream.writeASCII(enum.Name)
		if err != nil {
			middleStream.Close()
			return err
		}
		err = zstdStream.WriteByte(enum.BitSize)
		if err != nil {
			middleStream.Close()
			return err
		}
	}

	err = zstdStream.writeUintUTF8(uint32(len(schema.Instances)))
	if err != nil {
		middleStream.Close()
		return err
	}
	err = zstdStream.writeUintUTF8(uint32(len(schema.Properties)))
	if err != nil {
		middleStream.Close()
		return err
	}
	err = zstdStream.writeUintUTF8(uint32(len(schema.Events)))
	if err != nil {
		middleStream.Close()
		return err
	}
	for _, instance := range schema.Instances {
		err = zstdStream.writeUintUTF8(uint32(len(instance.Name)))
		if err != nil {
			middleStream.Close()
			return err
		}
		err = zstdStream.writeASCII(instance.Name)
		if err != nil {
			middleStream.Close()
			return err
		}
		err = zstdStream.writeUintUTF8(uint32(len(instance.Properties)))
		if err != nil {
			middleStream.Close()
			return err
		}

		for _, property := range instance.Properties {
			err = zstdStream.writeUintUTF8(uint32(len(property.Name)))
			if err != nil {
				middleStream.Close()
				return err
			}
			err = zstdStream.writeASCII(property.Name)
			if err != nil {
				middleStream.Close()
				return err
			}
			err = zstdStream.WriteByte(property.Type)
			if err != nil {
				middleStream.Close()
				return err
			}
			if property.Type == 7 {
				err = zstdStream.writeUint16BE(property.EnumID)
				if err != nil {
					middleStream.Close()
					return err
				}
			}
		}

		err = zstdStream.writeUint16BE(instance.Unknown)
		if err != nil {
			middleStream.Close()
			return err
		}
		err = zstdStream.writeUintUTF8(uint32(len(instance.Events)))
		if err != nil {
			middleStream.Close()
			return err
		}
		for _, event := range instance.Events {
			err = zstdStream.writeUintUTF8(uint32(len(event.Name)))
			if err != nil {
				middleStream.Close()
				return err
			}
			err = zstdStream.writeASCII(event.Name)
			if err != nil {
				middleStream.Close()
				return err
			}

			err = zstdStream.writeUintUTF8(uint32(len(event.Arguments)))
			if err != nil {
				middleStream.Close()
				return err
			}
			for _, argument := range event.Arguments {
				err = zstdStream.WriteByte(argument.Type)
				if err != nil {
					middleStream.Close()
					return err
				}
				err = zstdStream.writeUint16BE(argument.EnumID)
				if err != nil {
					middleStream.Close()
					return err
				}
			}
		}
	}

	_, err = middleStream.Write(uncompressedBuf.Bytes())
	if err != nil {
		middleStream.Close()
		return err
	}
	err = middleStream.Close()
	if err != nil {
		return err
	}
	err = stream.writeUint32BE(uint32(zstdBuf.Len()))
	if err != nil {
		return err
	}
	err = stream.writeUint32BE(uint32(uncompressedBuf.Len()))
	if err != nil {
		return err
	}
	err = stream.allBytes(zstdBuf.Bytes())
	return err
}

func (layer *Packet97Layer) String() string {
	return fmt.Sprintf("ID_NEW_SCHEMA: %d enums, %d instances", len(layer.Schema.Enums), len(layer.Schema.Instances))
}

// TypeString implements RakNetPacket.TypeString()
func (Packet97Layer) TypeString() string {
	return "ID_NEW_SCHEMA"
}

// Type implements RakNetPacket.Type()
func (Packet97Layer) Type() byte {
	return 0x97
}
