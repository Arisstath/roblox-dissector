package main
import "github.com/google/gopacket"
import "github.com/davecgh/go-spew/spew"
import "github.com/therecipe/qt/widgets"

type Packet83_01 struct {
	Object1 Object
}

func (this Packet83_01) Show() widgets.QWidget_ITF {
	return NewQLabelF("Init referent: %s", this.Object1.Show())
}

func DecodePacket83_01(thisBitstream *ExtendedReader, context *CommunicationContext, packet gopacket.Packet) (interface{}, error) {
	var err error
	inner := &Packet83_01{}
	inner.Object1, err = thisBitstream.ReadObject(false, context)
	return inner, err
}
