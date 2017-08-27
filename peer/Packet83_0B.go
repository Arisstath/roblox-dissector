package peer
import "errors"
import "github.com/gskartwii/rbxfile"

type Packet83_0B struct {
	Instances []*rbxfile.Instance
}

func NewPacket83_0BLayer(length int) *Packet83_0B {
	return &Packet83_0B{make([]*rbxfile.Instance, length)}
}

func DecodePacket83_0B(packet *UDPPacket, context *CommunicationContext, instanceSchema []*InstanceSchemaItem) (interface{}, error) {
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
		layer.Instances[i], err = DecodeReplicationInstance(true, newPacket, context, instanceSchema)
		if err != nil {
			return layer, err
		}

		gzipStream.Align()
	}
	return layer, nil
}
