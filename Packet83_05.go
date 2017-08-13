package main
import "github.com/therecipe/qt/widgets"
import "github.com/google/gopacket"

type Packet83_05 struct {
	Bool1 bool
	Int1 uint64
	Int2 uint32
	Int3 uint32
}

func (this Packet83_05) Show() widgets.QWidget_ITF {
	widget := widgets.NewQWidget(nil, 0)
	layout := widgets.NewQVBoxLayout()
	layout.AddWidget(NewQLabelF("Unknown bool: %v", this.Bool1), 0, 0)
	layout.AddWidget(NewQLabelF("Int 1: %d", this.Int1), 0, 0)
	layout.AddWidget(NewQLabelF("Int 2: %d", this.Int2), 0, 0)
	layout.AddWidget(NewQLabelF("Int 3: %d", this.Int3), 0, 0)
	widget.SetLayout(layout)

	return widget
}

func DecodePacket83_05(thisBitstream *ExtendedReader, context *CommunicationContext, packet gopacket.Packet) (interface{}, error) {
	var err error
	inner := &Packet83_05{}
	inner.Bool1, err = thisBitstream.ReadBool()
	if err != nil {
		return inner, err
	}

	inner.Int1, err = thisBitstream.Bits(64)
	if err != nil {
		return inner, err
	}
	inner.Int2, err = thisBitstream.ReadUint32BE()
	if err != nil {
		return inner, err
	}
	inner.Int3, err = thisBitstream.ReadUint32BE()
	if err != nil {
		return inner, err
	}
	//println(DebugInfo(context, packet), "Receive 0x05", inner.Bool1, ",", inner.Int1, ",", inner.Int2, ",", inner.Int3)

	return inner, err
}
