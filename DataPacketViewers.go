package main

import (
	"errors"
	"strconv"

	"github.com/Gskartwii/roblox-dissector/peer"
	"github.com/gotk3/gotk3/gtk"
)

func blanketViewer(content string) (gtk.IWidget, error) {
	viewer, err := gtk.LabelNew(content)
	if err != nil {
		return nil, err
	}
	viewer.SetHAlign(gtk.ALIGN_START)
	viewer.SetVAlign(gtk.ALIGN_START)
	viewer.SetMarginTop(8)
	viewer.SetMarginBottom(8)
	viewer.SetMarginStart(8)
	viewer.SetMarginEnd(8)
	viewer.ShowAll()

	return viewer, nil
}

func viewerForDataPacket(packet peer.Packet83Subpacket) (gtk.IWidget, error) {
	switch packet.Type() {
	case 0x01, 0x04, 0x0B, 0x0C, 0x0F, 0x10, 0x13, 0x14:
		return blanketViewer(packet.String())
	case 0x02:
		newInst := packet.(*peer.Packet83_02)
		viewer, err := NewInstanceViewer()
		if err != nil {
			return nil, err
		}
		viewer.ViewInstance(newInst.ReplicationInstance)
		viewer.mainWidget.ShowAll()

		return viewer.mainWidget, nil
	case 0x03:
		prop := packet.(*peer.Packet83_03)
		viewer, err := NewPropertyEventViewer()
		if err != nil {
			return nil, err
		}
		var version = int32(-1)
		if prop.HasVersion {
			version = prop.Version
		}

		viewer.ViewPropertyUpdate(prop.Instance, prop.Schema.Name, prop.Value, version)
		viewer.mainWidget.ShowAll()
		return viewer.mainWidget, nil
	case 0x05:
		ping := packet.(*peer.Packet83_05)
		box, err := boxWithMargin()
		if err != nil {
			return nil, err
		}
		timestamp, err := newLabelF("Timestamp: %d", ping.Timestamp)
		if err != nil {
			return nil, err
		}
		box.Add(timestamp)
		sendStats, err := newLabelF("Send stats: %08X", ping.SendStats)
		if err != nil {
			return nil, err
		}
		box.Add(sendStats)
		extraStats, err := newLabelF("Extra stats: %08X", ping.ExtraStats)
		if err != nil {
			return nil, err
		}
		box.Add(extraStats)
		if ping.PacketVersion == 2 {
			int1, err := newLabelF("Int 1: %d", ping.Int1)
			if err != nil {
				return nil, err
			}
			box.Add(int1)
			fps1, err := newLabelF("Fps 1: %f", ping.Fps1)
			if err != nil {
				return nil, err
			}
			box.Add(fps1)
			fps2, err := newLabelF("Fps 2: %f", ping.Fps2)
			if err != nil {
				return nil, err
			}
			box.Add(fps2)
			fps3, err := newLabelF("Fps 3: %f", ping.Fps3)
			if err != nil {
				return nil, err
			}
			box.Add(fps3)
		}
		box.ShowAll()
		return box, nil
	case 0x06:
		ping := packet.(*peer.Packet83_06)
		box, err := boxWithMargin()
		if err != nil {
			return nil, err
		}
		timestamp, err := newLabelF("Timestamp: %d", ping.Timestamp)
		if err != nil {
			return nil, err
		}
		box.Add(timestamp)
		sendStats, err := newLabelF("Send stats: %08X", ping.SendStats)
		if err != nil {
			return nil, err
		}
		box.Add(sendStats)
		extraStats, err := newLabelF("Extra stats: %08X", ping.ExtraStats)
		if err != nil {
			return nil, err
		}
		box.Add(extraStats)
		box.ShowAll()
		return box, nil
	case 0x07:
		event := packet.(*peer.Packet83_07)
		viewer, err := NewPropertyEventViewer()
		if err != nil {
			return nil, err
		}
		viewer.ViewEvent(event.Instance, event.Schema.Name, event.Event.Arguments)
		viewer.mainWidget.ShowAll()
		return viewer.mainWidget, nil
	case 0x0A:
		ack := packet.(*peer.Packet83_0A)
		box, err := boxWithMargin()
		if err != nil {
			return nil, err
		}
		instance, err := gtk.LabelNew("Instance: " + ack.Instance.Name())
		if err != nil {
			return nil, err
		}
		instance.SetHAlign(gtk.ALIGN_START)
		box.Add(instance)
		ref, err := gtk.LabelNew("Reference: " + ack.Instance.Ref.String())
		if err != nil {
			return nil, err
		}
		ref.SetHAlign(gtk.ALIGN_START)
		box.Add(ref)
		property, err := gtk.LabelNew("Property: " + ack.Schema.Name)
		if err != nil {
			return nil, err
		}
		property.SetHAlign(gtk.ALIGN_START)
		box.Add(property)
		versionsString := "Versions: "
		for _, ver := range ack.Versions {
    		versionsString += strconv.Itoa(int(ver)) + ", "
		}
		versions, err := gtk.LabelNew(versionsString[:len(versionsString) - 2])
		if err != nil {
			return nil, err
		}
		versions.SetHAlign(gtk.ALIGN_START)
		box.Add(versions)
		box.ShowAll()

		return box, nil
	}
	return nil, errors.New("unimplemented")
}
