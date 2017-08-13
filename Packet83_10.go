package main
import "github.com/google/gopacket"
import "github.com/therecipe/qt/widgets"

type Packet83_10 struct {
	TagId uint32
}

func (this Packet83_10) Show() widgets.QWidget_ITF {
	return NewQLabelF("Replication tag: %d", this.TagId)
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
