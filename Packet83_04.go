package main
import "github.com/google/gopacket"

type Packet83_04 struct {
	MarkerId uint32
}

func DecodePacket83_04(thisBitstream *ExtendedReader, context *CommunicationContext, packet gopacket.Packet) (interface{}, error) {
	var err error
	inner := &Packet83_04{}
	inner.MarkerId, err = thisBitstream.ReadUint32LE()
	if err != nil {
		return inner, err
	}
	println(DebugInfo(context, packet), "Receive marker", inner.MarkerId)

	return inner, err
}
