package main
import "github.com/google/gopacket"
import "github.com/therecipe/qt/widgets"

type packet83_09Subpacket_ packet83Subpacket_
type Packet83_09Subpacket struct {
	child packet83_09Subpacket_
}

type Packet83_09 struct {
	Subpacket Packet83_09Subpacket
	Type uint8
}

type Packet83_09_01 struct {
	Int1 uint8
	Int2 uint32
	Int3 uint32
	Int4 uint32
	Int5 uint64
}

type Packet83_09_05 struct {
	Int uint32
}

type Packet83_09_default struct {
	Int1 uint8
	Int2 uint32
}

type Packet83_09_07 struct{}

func (this Packet83_09Subpacket) Show() widgets.QWidget_ITF {
	return this.child.Show()
}

func (this Packet83_09_01) Show() widgets.QWidget_ITF {
	widget := widgets.NewQWidget(nil, 0)
	layout := widgets.NewQVBoxLayout()
	layout.AddWidget(NewQLabelF("Int 1: %d", this.Int1), 0, 0)
	layout.AddWidget(NewQLabelF("Int 2: %d", this.Int2), 0, 0)
	layout.AddWidget(NewQLabelF("Int 3: %d", this.Int3), 0, 0)
	layout.AddWidget(NewQLabelF("Int 4: %d", this.Int4), 0, 0)
	layout.AddWidget(NewQLabelF("Int 5: %d", this.Int5), 0, 0)
	widget.SetLayout(layout)

	return widget
}

func (this Packet83_09_05) Show() widgets.QWidget_ITF {
	return NewQLabelF("Int: %d", this.Int)
}

func (this Packet83_09_07) Show() widgets.QWidget_ITF {
	return NewQLabelF("(no values)")
}

func (this Packet83_09_default) Show() widgets.QWidget_ITF {
	widget := widgets.NewQWidget(nil, 0)
	layout := widgets.NewQVBoxLayout()
	layout.AddWidget(NewQLabelF("Int 1: %d", this.Int1), 0, 0)
	layout.AddWidget(NewQLabelF("Int 2: %d", this.Int2), 0, 0)
	widget.SetLayout(layout)

	return widget
}

func (this Packet83_09) Show() widgets.QWidget_ITF {
	widget := widgets.NewQWidget(nil, 0)
	layout := widgets.NewQVBoxLayout()
	layout.AddWidget(NewQLabelF("Type: %d", this.Type), 0, 0)
	layout.AddWidget(this.Subpacket.Show(), 0, 0)
	widget.SetLayout(layout)

	return widget
}

func DecodePacket83_09(thisBitstream *ExtendedReader, context *CommunicationContext, packet gopacket.Packet) (interface{}, error) {
	var err error
	inner := &Packet83_09{}
	inner.Type, err = thisBitstream.ReadUint8()
	if err != nil {
		return inner, err
	}
	var subpacket interface{}
	switch inner.Type {
	case 1:
		thisSubpacket := Packet83_09_01{}
		thisSubpacket.Int1, err = thisBitstream.ReadUint8()
		if err != nil {
			return inner, err
		}
		thisSubpacket.Int2, err = thisBitstream.ReadUint32BE()
		if err != nil {
			return inner, err
		}
		thisSubpacket.Int3, err = thisBitstream.ReadUint32BE()
		if err != nil {
			return inner, err
		}
		thisSubpacket.Int4, err = thisBitstream.ReadUint32BE()
		if err != nil {
			return inner, err
		}
		thisSubpacket.Int5, err = thisBitstream.ReadUint64BE()
		if err != nil {
			return inner, err
		}
		subpacket = thisSubpacket
	case 5:
		thisSubpacket := Packet83_09_05{}
		thisSubpacket.Int, err = thisBitstream.ReadUint32BE()
		if err != nil {
			return inner, err
		}
		subpacket = thisSubpacket
	case 7:
		thisSubpacket := Packet83_09_07{}
		subpacket = thisSubpacket
	default:
		thisSubpacket := Packet83_09_default{}
		thisSubpacket.Int1, err = thisBitstream.ReadUint8()
		if err != nil {
			return inner, err
		}
		thisSubpacket.Int2, err = thisBitstream.ReadUint32BE()
		if err != nil {
			return inner, err
		}
		subpacket = thisSubpacket
	}
	inner.Subpacket = Packet83_09Subpacket{subpacket.(packet83_09Subpacket_)}

	return inner, err
}
