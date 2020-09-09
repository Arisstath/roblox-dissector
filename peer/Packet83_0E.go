package peer

import (
	"errors"
	"fmt"

	"github.com/Gskartwii/roblox-dissector/datamodel"
)

// Packet83_0E represents ID_REPLIC_REGION_REMOVAL
type Packet83_0E struct {
	Region    StreamInfo
	Instances []*datamodel.Instance
}

func (thisStream *extendedReader) DecodePacket83_0E(reader PacketReader, layers *PacketLayers) (Packet83Subpacket, error) {
	inner := &Packet83_0E{}
	var err error

	inner.Region, err = thisStream.readStreamInfo()
	if err != nil {
		return inner, err
	}
	numInstances, err := thisStream.readUint32BE()
	if err != nil {
		return inner, err
	}
	if numInstances > 0x100000 {
		return inner, errors.New("too many instances in region removal")
	}
	inner.Instances = make([]*datamodel.Instance, numInstances)

	useCompression, err := thisStream.readBoolByte()
	if err != nil {
		return inner, err
	}

	if useCompression {
		thisStream, err = thisStream.RegionToZStdStream()
		if err != nil {
			return inner, err
		}
	}

	for i := uint32(0); i < numInstances; i++ {
		ref, err := thisStream.readObject(reader.Context())
		if err != nil {
			return inner, err
		}

		inner.Instances[i], err = reader.Context().InstancesByReference.TryGetInstance(ref)
		if err != nil {
			return inner, err
		}
	}

	return inner, nil
}

// Serialize implements Packet83Subpacket.Serialize()
func (layer *Packet83_0E) Serialize(writer PacketWriter, stream *extendedWriter) error {
	err := layer.Region.Serialize(stream)
	if err != nil {
		return err
	}

	err = stream.writeUint32BE(uint32(len(layer.Instances)))
	if err != nil {
		return err
	}

	err = stream.writeBoolByte(true)
	if err != nil {
		return err
	}

	zstdStream := stream.wrapZstd()
	for _, inst := range layer.Instances {
		err = zstdStream.writeObject(inst, writer.Context())
		if err != nil {
			zstdStream.Close()
			return err
		}
	}
	return zstdStream.Close()
}

// Type implements Packet83Subpacket.Type()
func (Packet83_0E) Type() uint8 {
	return 0x0E
}

// TypeString implements Packet83Subpacket.TypeString()
func (Packet83_0E) TypeString() string {
	return "ID_REPLIC_REGION_REMOVAL"
}

func (layer *Packet83_0E) String() string {
	return fmt.Sprintf("ID_REPLIC_REGION_REMOVAL: %s, %d instances", layer.Region, len(layer.Instances))
}
