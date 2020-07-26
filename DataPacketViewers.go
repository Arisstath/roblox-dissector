package main

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/Gskartwii/roblox-dissector/peer"
	"github.com/gotk3/gotk3/glib"
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
	case 0x01, 0x04, 0x09, 0x0B, 0x0C, 0x0D, 0x0F, 0x10, 0x13, 0x14:
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
		versions, err := gtk.LabelNew(versionsString[:len(versionsString)-2])
		if err != nil {
			return nil, err
		}
		versions.SetHAlign(gtk.ALIGN_START)
		box.Add(versions)
		box.ShowAll()

		return box, nil
	case 0x0E:
		removal := packet.(*peer.Packet83_0E)
		model, err := gtk.ListStoreNew(
			glib.TYPE_INT,    // id
			glib.TYPE_STRING, // ref
			glib.TYPE_STRING, // name
		)
		if err != nil {
			return nil, err
		}
		view, err := gtk.TreeViewNewWithModel(model)
		if err != nil {
			return nil, err
		}
		renderer, err := gtk.CellRendererTextNew()
		if err != nil {
			return nil, err
		}
		for i, title := range []string{"ID", "Reference", "Name"} {
			col, err := gtk.TreeViewColumnNewWithAttribute(title, renderer, "text", i)
			if err != nil {
				return nil, err
			}
			view.AppendColumn(col)
		}
		for i, instance := range removal.Instances {
			row := model.Append()
			model.SetValue(row, 0, i)
			model.SetValue(row, 1, instance.Ref.String())
			model.SetValue(row, 2, instance.Name())
		}
		scrolled, err := gtk.ScrolledWindowNew(nil, nil)
		if err != nil {
			return nil, err
		}
		scrolled.SetMarginTop(8)
		scrolled.SetMarginBottom(8)
		scrolled.SetMarginStart(8)
		scrolled.SetMarginEnd(8)
		scrolled.Add(view)
		scrolled.ShowAll()
		return scrolled, nil
	case 0x12:
		hashes := packet.(*peer.Packet83_12)
		model, err := gtk.ListStoreNew(
			glib.TYPE_STRING, // index
			glib.TYPE_STRING, // hash
		)
		if err != nil {
			return nil, err
		}
		view, err := gtk.TreeViewNewWithModel(model)
		if err != nil {
			return nil, err
		}
		renderer, err := gtk.CellRendererTextNew()
		if err != nil {
			return nil, err
		}
		for i, title := range []string{"Index", "Hash"} {
			col, err := gtk.TreeViewColumnNewWithAttribute(title, renderer, "text", i)
			if err != nil {
				return nil, err
			}
			view.AppendColumn(col)
		}
		for i, hash := range hashes.HashList {
			row := model.Append()
			model.SetValue(row, 0, strconv.Itoa(i))
			model.SetValue(row, 1, fmt.Sprintf("%08X", hash))
		}
		if hashes.HasSecurityTokens {
			for i, st := range hashes.SecurityTokens {
				row := model.Append()
				model.SetValue(row, 0, "ST"+strconv.Itoa(i))
				model.SetValue(row, 1, fmt.Sprintf("%08X", st))
			}
		}
		scrolled, err := gtk.ScrolledWindowNew(nil, nil)
		if err != nil {
			return nil, err
		}
		scrolled.SetMarginTop(8)
		scrolled.SetMarginBottom(8)
		scrolled.SetMarginStart(8)
		scrolled.SetMarginEnd(8)
		scrolled.Add(view)
		scrolled.ShowAll()
		return scrolled, nil

	}
	return nil, errors.New("unimplemented")
}
