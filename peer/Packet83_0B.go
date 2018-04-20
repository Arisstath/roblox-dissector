package peer
import "errors"
import "github.com/gskartwii/rbxfile"
import "bytes"
import "github.com/gskartwii/go-bitstream"
import "github.com/DataDog/zstd"

// ID_JOINDATA
type Packet83_0B struct {
	// Instances replicated by the server
	Instances []*rbxfile.Instance
}

func NewPacket83_0BLayer(length int) *Packet83_0B {
	return &Packet83_0B{make([]*rbxfile.Instance, length)}
}

func decodePacket83_0B(packet *UDPPacket, context *CommunicationContext) (interface{}, error) {
	var layer *Packet83_0B
	thisBitstream := packet.stream
	thisBitstream.Align()
	arrayLen, err := thisBitstream.readUint32BE()
	if err != nil {
		return layer, err
	}
	if arrayLen == 0 {
		return layer, nil // no joindata? why is the subpacket sent then?
	}

	if arrayLen > 0x100000 {
		return layer, errors.New("sanity check: array len too long")
	}

	layer = NewPacket83_0BLayer(int(arrayLen))

	zstdStream, err := thisBitstream.RegionToZStdStream()
	if err != nil {
		return layer, err
	}

	newPacket := &UDPPacket{zstdStream, packet.Source, packet.Destination}

	var i uint32
	for i = 0; i < arrayLen; i++ {
		layer.Instances[i], err = decodeReplicationInstance(context.IsClient(newPacket.Source), true, newPacket, context)
		if err != nil {
			return layer, err
		}

		zstdStream.Align()
	}
	return layer, nil
}

func (layer *Packet83_0B) serialize(isClient bool, context *CommunicationContext, stream *extendedWriter) error {
    var err error
    err = stream.Align()
    if err != nil {
        return err
    }

    err = stream.writeUint32BE(uint32(len(layer.Instances)))
    uncompressedBuf := bytes.NewBuffer([]byte{})
	zstdBuf := bytes.NewBuffer([]byte{})
    middleStream := zstd.NewWriter(zstdBuf)
    defer middleStream.Close()
    zstdStream := &extendedWriter{bitstream.NewWriter(uncompressedBuf)}

    for i := 0; i < len(layer.Instances); i++ {
        err = serializeReplicationInstance(isClient, layer.Instances[i], true, context, zstdStream)
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
