package peer
import "errors"
import "github.com/gskartwii/rbxfile"
import "bytes"
import "compress/gzip"
import "github.com/gskartwii/go-bitstream"

type Packet83_0B struct {
	Instances []*rbxfile.Instance
}

func NewPacket83_0BLayer(length int) *Packet83_0B {
	return &Packet83_0B{make([]*rbxfile.Instance, length)}
}

func DecodePacket83_0B(packet *UDPPacket, context *CommunicationContext) (interface{}, error) {
	var layer *Packet83_0B
	thisBitstream := packet.Stream
	thisBitstream.Align()
	arrayLen, err := thisBitstream.ReadUint32BE()
	if err != nil {
		return layer, err
	}
	if arrayLen > 0x100000 {
		return layer, errors.New("sanity check: array len too long")
	}

	layer = NewPacket83_0BLayer(int(arrayLen))

	gzipStream, err := thisBitstream.RegionToGZipStream()
	if err != nil {
		return layer, err
	}

	newPacket := &UDPPacket{gzipStream, packet.Source, packet.Destination}

	var i uint32
	for i = 0; i < arrayLen; i++ {
		layer.Instances[i], err = DecodeReplicationInstance(context.IsClient(newPacket.Source), true, newPacket, context)
		if err != nil {
			return layer, err
		}

		gzipStream.Align()
	}
	return layer, nil
}

func (layer *Packet83_0B) Serialize(isClient bool, context *CommunicationContext, stream *ExtendedWriter) error {
    var err error
    err = stream.Align()
    if err != nil {
        return err
    }

    err = stream.WriteUint32BE(uint32(len(layer.Instances)))
    gzipBuf := bytes.NewBuffer([]byte{})
    middleStream := gzip.NewWriter(gzipBuf)
    defer middleStream.Close()
    gzipStream := &ExtendedWriter{bitstream.NewWriter(middleStream)}

    for i := 0; i < len(layer.Instances); i++ {
        err = SerializeReplicationInstance(isClient, layer.Instances[i], true, context, gzipStream)
        if err != nil {
            return err
        }
        err = gzipStream.Align()
        if err != nil {
            return err
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
    err = stream.WriteUint32BE(uint32(gzipBuf.Len()))
    if err != nil {
        return err
    }
    err = stream.AllBytes(gzipBuf.Bytes())
    return err
}
