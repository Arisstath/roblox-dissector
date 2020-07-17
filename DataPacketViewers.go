package main

import (
	"errors"
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
	case 0x01, 0x04, 0x0B, 0x10, 0x13, 0x14:
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
	case 0x07:
		event := packet.(*peer.Packet83_07)
		viewer, err := NewPropertyEventViewer()
		if err != nil {
			return nil, err
		}
		viewer.ViewEvent(event.Instance, event.Schema.Name, event.Event.Arguments)
		viewer.mainWidget.ShowAll()
		return viewer.mainWidget, nil
	}
	return nil, errors.New("unimplemented")
}
