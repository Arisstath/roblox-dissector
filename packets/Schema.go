package packets

import (
	"bytes"
	"compress/gzip"
	"errors"

	"github.com/gskartwii/go-bitstream"
    "github.com/gskartwii/roblox-dissector/schema"
)
// ID_NEW_SCHEMA - server -> client
// Negotiates a network schema with the client
type SchemaPacket struct {
	Schema schema.StaticSchema
}

func NewSchemaPacket() *SchemaPacket {
	return &SchemaPacket{}
}

func (thisBitstream *PacketReaderBitstream) DecodeSchemaPacket(reader PacketReader, layers *PacketLayers) (RakNetPacket, error) {
	layer := NewSchemaPacket()

	var err error
	stream, err := thisBitstream.RegionToZStdStream()
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
	layer.Schema.Enums = make([]schema.StaticEnumSchema, enumArrayLen)
	layer.Schema.EnumsByName = make(map[string]int, enumArrayLen)
	for i := 0; i < int(enumArrayLen); i++ {
		stringLen, err := stream.readUintUTF8()
		if err != nil {
			return layer, err
		}
		layer.Schema.Enums[i].Name, err = stream.readASCII(int(stringLen))
		if err != nil {
			return layer, err
		}
		layer.Schema.Enums[i].BitSize, err = stream.readUint8()
		if err != nil {
			return layer, err
		}
		layer.Schema.EnumsByName[layer.Schema.Enums[i].Name] = i
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
	layer.Schema.Instances = make([]schema.StaticInstanceSchema, classArrayLen)
	layer.Schema.Properties = make([]schema.StaticPropertySchema, propertyArrayLen)
	layer.Schema.Events = make([]schema.StaticEventSchema, eventArrayLen)
	layer.Schema.ClassesByName = make(map[string]int, classArrayLen)
	layer.Schema.PropertiesByName = make(map[string]int, propertyArrayLen)
	layer.Schema.EventsByName = make(map[string]int, eventArrayLen)
	propertyGlobalIndex := 0
	classGlobalIndex := 0
	eventGlobalIndex := 0
	for i := 0; i < int(classArrayLen); i++ {
		classNameLen, err := stream.readUintUTF8()
		if err != nil {
			return layer, err
		}
		thisInstance := layer.Schema.Instances[i]
		thisInstance.Name, err = stream.readASCII(int(classNameLen))
		if err != nil {
			return layer, err
		}
		layer.Schema.ClassesByName[thisInstance.Name] = i

		propertyCount, err := stream.readUintUTF8()
		if err != nil {
			return layer, err
		}
		if propertyCount > 0x10000 {
			return layer, errors.New("sanity check: exceeded maximum property count")
		}
		thisInstance.Properties = make([]schema.StaticPropertySchema, propertyCount)

		for j := 0; j < int(propertyCount); j++ {
			thisProperty := thisInstance.Properties[j]
			propNameLen, err := stream.readUintUTF8()
			if err != nil {
				return layer, err
			}
			thisProperty.Name, err = stream.readASCII(int(propNameLen))
			if err != nil {
				return layer, err
			}
			propertyGlobalName := make([]byte, propNameLen+1+classNameLen)
			copy(propertyGlobalName, thisInstance.Name)
			propertyGlobalName[classNameLen] = byte('.')
			copy(propertyGlobalName[classNameLen+1:], thisProperty.Name)
			layer.Schema.PropertiesByName[string(propertyGlobalName)] = propertyGlobalIndex

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

			thisProperty.InstanceSchema = &thisInstance
			layer.Schema.Properties[propertyGlobalIndex] = thisProperty
			thisInstance.Properties[j] = thisProperty
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
		thisInstance.Events = make([]schema.StaticEventSchema, eventCount)

		for j := 0; j < int(eventCount); j++ {
			thisEvent := thisInstance.Events[j]
			eventNameLen, err := stream.readUintUTF8()
			if err != nil {
				return layer, err
			}
			thisEvent.Name, err = stream.readASCII(int(eventNameLen))
			if err != nil {
				return layer, err
			}
			eventGlobalName := make([]byte, eventNameLen+1+classNameLen)
			copy(eventGlobalName, thisInstance.Name)
			eventGlobalName[classNameLen] = byte('.')
			copy(eventGlobalName[classNameLen+1:], thisEvent.Name)
			layer.Schema.EventsByName[string(eventGlobalName)] = eventGlobalIndex

			countArguments, err := stream.readUintUTF8()
			if err != nil {
				return layer, err
			}
			if countArguments > 0x10000 {
				return layer, errors.New("sanity check: exceeded maximum argument count")
			}
			thisEvent.Arguments = make([]schema.StaticArgumentSchema, countArguments)

			for k := 0; k < int(countArguments); k++ {
				thisArgument := thisEvent.Arguments[k]
				thisArgument.Type, err = stream.readUint8()
				if err != nil {
					return layer, err
				}
				thisArgument.TypeString = TypeNames[thisArgument.Type]
				thisArgument.EnumID, err = stream.readUint16BE()
				if err != nil {
					return layer, err
				}
				thisEvent.Arguments[k] = thisArgument
			}
			thisEvent.InstanceSchema = &thisInstance
			layer.Schema.Events[eventGlobalIndex] = thisEvent
			thisInstance.Events[j] = thisEvent
			eventGlobalIndex++
		}
		layer.Schema.Instances[classGlobalIndex] = thisInstance
		classGlobalIndex++
	}
	reader.Context().StaticSchema = &layer.Schema

	return layer, err
}

func (layer *SchemaPacket) Serialize(writer PacketWriter, stream *PacketWriterBitstream) error {
	var err error

	err = stream.WriteByte(0x97)
	if err != nil {
		return err
	}
	gzipBuf := bytes.NewBuffer([]byte{})
	middleStream := gzip.NewWriter(gzipBuf)
	defer middleStream.Close()
	gzipStream := &PacketWriterBitstream{bitstream.NewWriter(middleStream)}

	schema := layer.Schema
	err = gzipStream.writeUintUTF8(uint32(len(schema.Enums)))
	if err != nil {
		return err
	}
	for _, enum := range schema.Enums {
		err = gzipStream.writeUintUTF8(uint32(len(enum.Name)))
		err = gzipStream.writeASCII(enum.Name)
		err = gzipStream.WriteByte(enum.BitSize)
	}

	err = gzipStream.writeUintUTF8(uint32(len(schema.Instances)))
	if err != nil {
		return err
	}
	err = gzipStream.writeUintUTF8(uint32(len(schema.Properties)))
	if err != nil {
		return err
	}
	err = gzipStream.writeUintUTF8(uint32(len(schema.Events)))
	if err != nil {
		return err
	}
	for _, instance := range schema.Instances {
		err = gzipStream.writeUintUTF8(uint32(len(instance.Name)))
		if err != nil {
			return err
		}
		err = gzipStream.writeASCII(instance.Name)
		if err != nil {
			return err
		}
		err = gzipStream.writeUintUTF8(uint32(len(instance.Properties)))
		if err != nil {
			return err
		}

		for _, property := range instance.Properties {
			err = gzipStream.writeUintUTF8(uint32(len(property.Name)))
			if err != nil {
				return err
			}
			err = gzipStream.writeASCII(property.Name)
			if err != nil {
				return err
			}
			err = gzipStream.WriteByte(property.Type)
			if err != nil {
				return err
			}
			if property.Type == 7 {
				err = gzipStream.writeUint16BE(property.EnumID)
				if err != nil {
					return err
				}
			}
		}

		err = gzipStream.writeUint16BE(instance.Unknown)
		if err != nil {
			return err
		}
		err = gzipStream.writeUintUTF8(uint32(len(instance.Events)))
		if err != nil {
			return err
		}
		for _, event := range instance.Events {
			err = gzipStream.writeUintUTF8(uint32(len(event.Name)))
			if err != nil {
				return err
			}
			err = gzipStream.writeASCII(event.Name)
			if err != nil {
				return err
			}

			err = gzipStream.writeUintUTF8(uint32(len(event.Arguments)))
			if err != nil {
				return err
			}
			for _, argument := range event.Arguments {
				err = gzipStream.WriteByte(argument.Type)
				if err != nil {
					return err
				}
				err = gzipStream.writeUint16BE(argument.EnumID)
				if err != nil {
					return err
				}
			}
		}
	}

	err = gzipStream.Flush(bitstream.Zero)
	if err != nil {
		return err
	}
	err = middleStream.Flush()
	if err != nil {
		return err
	}
	err = middleStream.Close()
	if err != nil {
		return err
	}
	err = stream.writeUint32BE(uint32(gzipBuf.Len()))
	if err != nil {
		return err
	}
	err = stream.allBytes(gzipBuf.Bytes())
	return err
}
