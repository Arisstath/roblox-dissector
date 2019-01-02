package peer

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/DataDog/zstd"
	bitstream "github.com/gskartwii/go-bitstream"
	"github.com/gskartwii/roblox-dissector/datamodel"
)

// ID_JOINDATA
type Packet83_0B struct {
	// Instances replicated by the server
	Instances []*datamodel.Instance
}

func NewPacket83_0BLayer() *Packet83_0B {
	return &Packet83_0B{}
}

func (thisBitstream *extendedReader) DecodePacket83_0B(reader PacketReader, layers *PacketLayers) (Packet83Subpacket, error) {
	layer := NewPacket83_0BLayer()

	thisBitstream.Align()
	arrayLen, err := thisBitstream.readUint32BE()
	if err != nil {
		return layer, err
	}

	if arrayLen > 0x100000 {
		return layer, errors.New("sanity check: array len too long")
	}

	layer.Instances = make([]*datamodel.Instance, arrayLen)
	if arrayLen == 0 {
		return layer, nil
	}

	zstdStream, err := thisBitstream.RegionToZStdStream()
	if err != nil {
		return layer, err
	}

	var i uint32
	for i = 0; i < arrayLen; i++ {
		layer.Instances[i], err = decodeReplicationInstance(reader, &JoinSerializeReader{zstdStream}, layers)
		if err != nil {
			return layer, err
		}

		zstdStream.Align()
	}
	return layer, nil
}

func (layer *Packet83_0B) Serialize(writer PacketWriter, stream *extendedWriter) error {
	var err error
	err = stream.Align()
	if err != nil {
		return err
	}
	if layer.Instances == nil || len(layer.Instances) == 0 { // bail
		err = stream.writeUint32BE(0)
		return err
	}

	err = stream.writeUint32BE(uint32(len(layer.Instances)))
	uncompressedBuf := bytes.NewBuffer([]byte{})
	zstdBuf := bytes.NewBuffer([]byte{})
	middleStream := zstd.NewWriter(zstdBuf)
	zstdStream := &extendedWriter{bitstream.NewWriter(uncompressedBuf)}

	for i := 0; i < len(layer.Instances); i++ {
		err = serializeReplicationInstance(layer.Instances[i], writer, &JoinSerializeWriter{zstdStream})
		if err != nil {
			middleStream.Close()
			return err
		}
		err = zstdStream.Align()
		if err != nil {
			middleStream.Close()
			return err
		}
	}

	err = zstdStream.Flush(bitstream.Zero)
	if err != nil {
		middleStream.Close()
		return err
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

func (Packet83_0B) Type() uint8 {
	return 0xB
}
func (Packet83_0B) TypeString() string {
	return "ID_REPLIC_JOIN_DATA"
}

func (layer *Packet83_0B) String() string {
	return fmt.Sprintf("ID_REPLIC_JOIN_DATA: %d items", len(layer.Instances))
}
