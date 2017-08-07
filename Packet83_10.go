package main
import "github.com/google/gopacket"

type Packet83_10 struct {
	TagId uint32
}

func DecodePacket83_10(thisBitstream *ExtendedReader, context *CommunicationContext, packet gopacket.Packet) (interface{}, error) {
	var err error
	inner := &Packet83_10{}
	inner.TagId, err = thisBitstream.ReadUint32BE()
	if err != nil {
		return inner, err
	}
	println(DebugInfo(context, packet), "Receive tag", inner.TagId)

	return inner, err
}
