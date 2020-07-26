package main

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/Gskartwii/roblox-dissector/peer"
	"github.com/dustin/go-humanize"
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
	case 0x11:
		stats := packet.(*peer.Packet83_11)
		scrolled, err := gtk.ScrolledWindowNew(nil, nil)
		if err != nil {
			return nil, err
		}
		box, err := boxWithMargin()
		if err != nil {
			return nil, err
		}

		n, pref := humanize.ComputeSI(stats.MemoryStats.TotalServerMemory)
		memLabel, err := newLabelF("Total server memory: %f %sB", n, pref)
		if err != nil {
			return nil, err
		}
		box.Add(memLabel)
		categoriesLabel, err := newLabelF("Memory usage by categories")
		if err != nil {
			return nil, err
		}
		box.Add(categoriesLabel)

		model, err := gtk.TreeStoreNew(
			glib.TYPE_STRING, // name
			glib.TYPE_STRING, // usage
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
		for i, title := range []string{"Name", "Memory usage"} {
			col, err := gtk.TreeViewColumnNewWithAttribute(title, renderer, "text", i)
			if err != nil {
				return nil, err
			}
			view.AppendColumn(col)
		}

		var devCategoryIter gtk.TreeIter
		model.InsertWithValues(&devCategoryIter, nil, -1, []int{0}, []interface{}{"Developer categories"})
		for _, category := range stats.MemoryStats.DeveloperTags {
			n, pref := humanize.ComputeSI(category.Memory)
			model.InsertWithValues(nil, &devCategoryIter, -1, []int{0, 1}, []interface{}{category.Name, fmt.Sprintf("%G %sB", n, pref)})
		}
		var internalCategoryIter gtk.TreeIter
		model.InsertWithValues(&internalCategoryIter, nil, -1, []int{0}, []interface{}{"Internal categories"})
		for _, category := range stats.MemoryStats.InternalCategories {
			n, pref := humanize.ComputeSI(category.Memory)
			model.InsertWithValues(nil, &internalCategoryIter, -1, []int{0, 1}, []interface{}{category.Name, fmt.Sprintf("%G %sB", n, pref)})
		}
		memoryStatsScroller, err := gtk.ScrolledWindowNew(nil, nil)
		if err != nil {
			return nil, err
		}
		memoryStatsScroller.Add(view)
		memoryStatsScroller.SetVExpand(true)
		box.Add(memoryStatsScroller)

		if stats.DataStoreStats.Enabled {
			getAsyncLabel, err := newLabelF("GetAsync: %d", stats.DataStoreStats.GetAsync)
			if err != nil {
				return nil, err
			}
			box.Add(getAsyncLabel)
			setAndIncrementAsyncLabel, err := newLabelF("SetAndIncrementAsync: %d", stats.DataStoreStats.SetAndIncrementAsync)
			if err != nil {
				return nil, err
			}
			box.Add(setAndIncrementAsyncLabel)
			updateAsyncLabel, err := newLabelF("UpdateAsync: %d", stats.DataStoreStats.UpdateAsync)
			if err != nil {
				return nil, err
			}
			box.Add(updateAsyncLabel)
			getSortedAsyncLabel, err := newLabelF("GetSortedAsync: %d", stats.DataStoreStats.GetSortedAsync)
			if err != nil {
				return nil, err
			}
			box.Add(getSortedAsyncLabel)
			setIncrementSortedAsyncLabel, err := newLabelF("SetIncrementSortedAsync: %d", stats.DataStoreStats.SetIncrementSortedAsync)
			if err != nil {
				return nil, err
			}
			box.Add(setIncrementSortedAsyncLabel)
			onUpdateLabel, err := newLabelF("OnUpdate: %d", stats.DataStoreStats.OnUpdate)
			if err != nil {
				return nil, err
			}
			box.Add(onUpdateLabel)
		}

		jobStatsLabel, err := newLabelF("Job stats:")
		if err != nil {
			return nil, err
		}
		box.Add(jobStatsLabel)
		jobStatsModel, err := gtk.TreeStoreNew(
			glib.TYPE_STRING, // name
			glib.TYPE_FLOAT,  // duty cycle
			glib.TYPE_FLOAT,  // op/s
			glib.TYPE_FLOAT,  // time/op
		)
		if err != nil {
			return nil, err
		}
		jobStatsView, err := gtk.TreeViewNewWithModel(jobStatsModel)
		if err != nil {
			return nil, err
		}
		jobStatsRenderer, err := gtk.CellRendererTextNew()
		if err != nil {
			return nil, err
		}
		for i, title := range []string{"Name", "Duty cycle (%)", "Op/s", "Time/op (ms)"} {
			col, err := gtk.TreeViewColumnNewWithAttribute(title, jobStatsRenderer, "text", i)
			if err != nil {
				return nil, err
			}
			jobStatsView.AppendColumn(col)
		}
		for _, job := range stats.JobStats {
			jobStatsModel.InsertWithValues(nil, nil, -1, []int{0, 1, 2, 3}, []interface{}{job.Name, job.Stat1, job.Stat2, job.Stat3})
		}
		jobStatsScroller, err := gtk.ScrolledWindowNew(nil, nil)
		if err != nil {
			return nil, err
		}
		jobStatsScroller.Add(jobStatsView)
		jobStatsScroller.SetVExpand(true)
		box.Add(jobStatsScroller)

		scriptStatsLabel, err := newLabelF("Script stats:")
		if err != nil {
			return nil, err
		}
		box.Add(scriptStatsLabel)
		scriptStatsModel, err := gtk.TreeStoreNew(
			glib.TYPE_STRING, // name
			glib.TYPE_FLOAT,  // activity
			glib.TYPE_FLOAT,  // rate/s
		)
		if err != nil {
			return nil, err
		}
		scriptStatsView, err := gtk.TreeViewNewWithModel(scriptStatsModel)
		if err != nil {
			return nil, err
		}
		scriptStatsRenderer, err := gtk.CellRendererTextNew()
		if err != nil {
			return nil, err
		}
		for i, title := range []string{"Name", "Activity (%)", "Rate/s"} {
			col, err := gtk.TreeViewColumnNewWithAttribute(title, scriptStatsRenderer, "text", i)
			if err != nil {
				return nil, err
			}
			scriptStatsView.AppendColumn(col)
		}
		for _, script := range stats.ScriptStats {
			scriptStatsModel.InsertWithValues(nil, nil, -1, []int{0, 1, 2}, []interface{}{script.Name, script.Stat1, script.Stat2})
		}
		scriptStatsScroller, err := gtk.ScrolledWindowNew(nil, nil)
		if err != nil {
			return nil, err
		}
		scriptStatsScroller.Add(scriptStatsView)
		scriptStatsScroller.SetVExpand(true)
		box.Add(scriptStatsScroller)

		avgPingMsLabel, err := newLabelF("Average ping (ms): %g", stats.AvgPingMs)
		if err != nil {
			return nil, err
		}
		box.Add(avgPingMsLabel)
		avgPhysicsSenderPktPSLabel, err := newLabelF("Average physics sender packets/s: %g", stats.AvgPhysicsSenderPktPS)
		if err != nil {
			return nil, err
		}
		box.Add(avgPhysicsSenderPktPSLabel)
		totalDataKBPSLabel, err := newLabelF("Total data kb/s: %g", stats.TotalDataKBPS)
		if err != nil {
			return nil, err
		}
		box.Add(totalDataKBPSLabel)
		totalPhysicsKBPSLabel, err := newLabelF("Total physics kb/s: %g", stats.TotalPhysicsKBPS)
		if err != nil {
			return nil, err
		}
		box.Add(totalPhysicsKBPSLabel)
		dataThroughputRatioLabel, err := newLabelF("Data throughput ratio: %g", stats.DataThroughputRatio)
		if err != nil {
			return nil, err
		}
		box.Add(dataThroughputRatioLabel)

		scrolled.SetMarginTop(8)
		scrolled.SetMarginBottom(8)
		scrolled.SetMarginStart(8)
		scrolled.SetMarginEnd(8)
		scrolled.Add(box)
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
