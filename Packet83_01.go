package main
import "github.com/google/gopacket"
import "github.com/therecipe/qt/widgets"
import "github.com/gskartwii/rbxfile"

type Packet83_01 struct {
	Instance *rbxfile.Instance
}

func (this Packet83_01) Show() widgets.QWidget_ITF {
	return NewQLabelF("Init referent: %s", this.Instance.Reference)
}

func DecodePacket83_01(thisBitstream *ExtendedReader, context *CommunicationContext, packet gopacket.Packet) (interface{}, error) {
	var err error
	inner := &Packet83_01{}
    referent, err := thisBitstream.ReadObject(false, context)
    inner.Instance = context.InstancesByReferent[referent]

	return inner, err
}
