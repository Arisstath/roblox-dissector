package main
import "github.com/google/gopacket"
import "github.com/davecgh/go-spew/spew"

type Packet83_01 struct {
	Object1 Object
}

func DecodePacket83_01(thisBitstream *ExtendedReader, context *CommunicationContext, packet gopacket.Packet) (interface{}, error) {
	var err error
	inner := &Packet83_01{}
	inner.Object1, err = thisBitstream.ReadObject(false, context)
	println("Read init referent", spew.Sdump(inner))
	return inner, err
}
