package main
import "github.com/google/gopacket"
import "github.com/therecipe/qt/widgets"

type Packet83_11 struct {
	SkipStats1 bool
	Stats_1_1 []byte
	Stats_1_2 float32
	Stats_1_3 float32
	Stats_1_4 float32
	Stats_1_5 bool

	SkipStats2 bool
	Stats_2_1 []byte
	Stats_2_2 float32
	Stats_2_3 uint32
	Stats_2_4 bool
	
	AvgPingMs float32
	AvgPhysicsSenderPktPS float32
	TotalDataKBPS float32
	TotalPhysicsKBPS float32
	DataThroughputRatio float32
}

func (this Packet83_11) Show() widgets.QWidget_ITF {
	widget := widgets.NewQWidget(nil, 0)
	layout := widgets.NewQVBoxLayout()
	layout.AddWidget(NewQLabelF("Skip stat set 1: %v", this.SkipStats1), 0, 0)
	if !this.SkipStats1 {
		layout.AddWidget(NewQLabelF("Stat 1/1: %s", this.Stats_1_1), 0, 0)
		layout.AddWidget(NewQLabelF("Stat 1/2: %G", this.Stats_1_2), 0, 0)
		layout.AddWidget(NewQLabelF("Stat 1/3: %G", this.Stats_1_3), 0, 0)
		layout.AddWidget(NewQLabelF("Stat 1/4: %G", this.Stats_1_4), 0, 0)
		layout.AddWidget(NewQLabelF("Stat 1/5: %v", this.Stats_1_5), 0, 0)
	}
	layout.AddWidget(NewQLabelF("Skip stat set 2: %v", this.SkipStats2), 0, 0)
	if !this.SkipStats2 {
		layout.AddWidget(NewQLabelF("Stat 2/1: %s", this.Stats_2_1), 0, 0)
		layout.AddWidget(NewQLabelF("Stat 2/2: %G", this.Stats_2_2), 0, 0)
		layout.AddWidget(NewQLabelF("Stat 2/3: %d", this.Stats_2_3), 0, 0)
		layout.AddWidget(NewQLabelF("Stat 2/4: %v", this.Stats_2_4), 0, 0)
	}
	layout.AddWidget(NewQLabelF("Average ping ms: %G", this.AvgPingMs), 0, 0)
	layout.AddWidget(NewQLabelF("Average physics sender Pkt/s: %G", this.AvgPhysicsSenderPktPS), 0, 0)
	layout.AddWidget(NewQLabelF("Total data KB/s: %G", this.TotalDataKBPS), 0, 0)
	layout.AddWidget(NewQLabelF("Total physics KB/s: %G", this.TotalPhysicsKBPS), 0, 0)
	layout.AddWidget(NewQLabelF("Data throughput ratio: %G", this.DataThroughputRatio), 0, 0)
	widget.SetLayout(layout)

	return widget
}

func DecodePacket83_11(thisBitstream *ExtendedReader, context *CommunicationContext, packet gopacket.Packet) (interface{}, error) {
	var err error
	inner := &Packet83_11{}
	inner.SkipStats1, err = thisBitstream.ReadBool()
	if err != nil {
		return inner, err
	}
	println(DebugInfo(context, packet), "Skip stats 1:", inner.SkipStats1)
	if !inner.SkipStats1 {
		stringLen, err := thisBitstream.ReadUint32BE()
		if err != nil {
			return inner, err
		}
		inner.Stats_1_1, err = thisBitstream.ReadString(int(stringLen))
		if err != nil {
			return inner, err
		}

		inner.Stats_1_2, err = thisBitstream.ReadFloat32BE()
		if err != nil {
			return inner, err
		}
		inner.Stats_1_3, err = thisBitstream.ReadFloat32BE()
		if err != nil {
			return inner, err
		}
		inner.Stats_1_4, err = thisBitstream.ReadFloat32BE()
		if err != nil {
			return inner, err
		}
		inner.Stats_1_5, err = thisBitstream.ReadBool()
		if err != nil {
			return inner, err
		}
		print("Receive stats1", inner.Stats_1_1, ",", inner.Stats_1_2, ",", inner.Stats_1_3, ",", inner.Stats_1_4, ",", inner.Stats_1_5)
	}

	inner.SkipStats2, err = thisBitstream.ReadBool()
	if err != nil {
		return inner, err
	}
	println(DebugInfo(context, packet), "Skip stats 2:", inner.SkipStats2)
	if !inner.SkipStats2 {
		stringLen, err := thisBitstream.ReadUint32BE()
		if err != nil {
			return inner, err
		}
		inner.Stats_2_1, err = thisBitstream.ReadString(int(stringLen))
		if err != nil {
			return inner, err
		}

		inner.Stats_2_2, err = thisBitstream.ReadFloat32BE()
		if err != nil {
			return inner, err
		}
		inner.Stats_2_3, err = thisBitstream.ReadUint32BE()
		if err != nil {
			return inner, err
		}
		inner.Stats_2_4, err = thisBitstream.ReadBool()
		if err != nil {
			return inner, err
		}
		print("Receive stats2", inner.Stats_2_1, ",", inner.Stats_2_2, ",", inner.Stats_2_3, ",", inner.Stats_2_4)
	}

	inner.AvgPingMs, err = thisBitstream.ReadFloat32BE()
	if err != nil {
		return inner, err
	}
	inner.AvgPhysicsSenderPktPS, err = thisBitstream.ReadFloat32BE()
	if err != nil {
		return inner, err
	}
	inner.TotalDataKBPS, err = thisBitstream.ReadFloat32BE()
	if err != nil {
		return inner, err
	}
	inner.TotalPhysicsKBPS, err = thisBitstream.ReadFloat32BE()
	if err != nil {
		return inner, err
	}
	inner.DataThroughputRatio, err = thisBitstream.ReadFloat32BE()
	if err != nil {
		return inner, err
	}
	println(DebugInfo(context, packet), "receive stats: %#v", inner)

	return inner, nil
}
