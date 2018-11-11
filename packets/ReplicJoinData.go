package peer

import "errors"
import "github.com/gskartwii/rbxfile"
import "bytes"
import "github.com/gskartwii/go-bitstream"
import "github.com/DataDog/zstd"

// ID_JOINDATA
type ReplicateJoinData struct {
	// Instances replicated by the server
	Instances []*rbxfile.Instance
}

func NewReplicateJoinDataLayer() *ReplicateJoinData {
	return &ReplicateJoinData{}
}

func (thisBitstream *extendedReader) DecodeReplicateJoinData(reader PacketReader, layers *PacketLayers) (ReplicationSubpacket, error) {
	layer := NewReplicateJoinDataLayer()

	thisBitstream.Align()
	arrayLen, err := thisBitstream.readUint32BE()
	if err != nil {
		return layer, err
	}

	if arrayLen > 0x100000 {
		return layer, errors.New("sanity check: array len too long")
	}

	layer.Instances = make([]*rbxfile.Instance, arrayLen)
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

func (layer *ReplicateJoinData) Serialize(writer PacketWriter, stream *extendedWriter) error {
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
	defer middleStream.Close()
	zstdStream := &extendedWriter{bitstream.NewWriter(uncompressedBuf)}

	for i := 0; i < len(layer.Instances); i++ {
		err = serializeReplicationInstance(layer.Instances[i], writer, &JoinSerializeWriter{zstdStream})
		if err != nil {
			return err
		}
		err = zstdStream.Align()
		if err != nil {
			return err
		}
	}

	err = zstdStream.Flush(bitstream.Zero)
	if err != nil {
		return err
	}
	_, err = middleStream.Write(uncompressedBuf.Bytes())
	if err != nil {
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

func (ReplicateJoinData) Type() uint8 {
	return 0xB
}
func (ReplicateJoinData) TypeString() string {
	return "ID_REPLIC_JOI_NDATA"
}
