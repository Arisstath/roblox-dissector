package main

import (
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"math"
	"strconv"

	"github.com/Gskartwii/roblox-dissector/peer"
	"github.com/gotk3/gotk3/cairo"
	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/gtk"
)

type splitPacketRange struct {
    start uint32
    end uint32 // inclusive
}

type PacketDetailsViewer struct {
	mainWidget *gtk.Notebook
	logBox     *gtk.TextView

	hexBox     *gtk.TextView
	hexDumpData []byte

	reliablity *gtk.Label
	rmNumber *gtk.Label
	channel *gtk.Label
	index *gtk.Label
	splitPacketsLabel *gtk.Label

	missingSplitPacketRanges []splitPacketRange
	countSplitPackets uint32
}

func NewPacketDetailsViewer() (*PacketDetailsViewer, error) {
	viewer := &PacketDetailsViewer{}

	builder, err := gtk.BuilderNewFromFile("packetdetailsviewer.ui")
	if err != nil {
		return nil, err
	}
	notebook_, err := builder.GetObject("packetdetailsnotebook")
	if err != nil {
		return nil, err
	}

	notebook, ok := notebook_.(*gtk.Notebook)
	if !ok {
		return nil, invalidUi("notebook")
	}

	notebookContainer_, err := builder.GetObject("packetdetailsviewerwindow")
	if err != nil {
		return nil, err
	}
	notebookContainer, ok := notebookContainer_.(*gtk.Window)
	if !ok {
		return nil, invalidUi("notebook container")
	}
	notebookContainer.Remove(notebook)

	logBox_, err := builder.GetObject("logbox")
	if err != nil {
		return nil, err
	}
	logBox, ok := logBox_.(*gtk.TextView)
	if !ok {
		return nil, invalidUi("logbox")
	}
	hexBox_, err := builder.GetObject("hexbox")
	if err != nil {
		return nil, err
	}
	hexBox, ok := hexBox_.(*gtk.TextView)
	if !ok {
		return nil, invalidUi("hexbox")
	}
	copyHex_, err := builder.GetObject("copyashexstreambutton")
	if err != nil {
		return nil, err
	}
	copyHex, ok := copyHex_.(*gtk.Button)
	if !ok {
		return nil, invalidUi("copyhex")
	}
	copyHex.Connect("clicked", func() {
		clipboard, err := gtk.ClipboardGet(gdk.GdkAtomIntern("CLIPBOARD", true))
		if err != nil {
			println("Failed to get clipboard:", err.Error())
			return
		}
		clipboard.SetText(hex.EncodeToString(viewer.hexDumpData))
	})
	saveHex_, err := builder.GetObject("savehexbutton")
	if err != nil {
		return nil, err
	}
	saveHex, ok := saveHex_.(*gtk.Button)
	if !ok {
		return nil, invalidUi("savehex")
	}
	saveHex.Connect("clicked", func() {
		parentWindow, err := viewer.mainWidget.GetToplevel()
		if err != nil {
			println("Failed to get parent window:", err.Error())
			return
		}
		dialog, err := gtk.FileChooserNativeDialogNew(
			"Save packet as",
			parentWindow.(gtk.IWindow),
			gtk.FILE_CHOOSER_ACTION_SAVE,
			"Save",
			"Cancel",
		)
		if err != nil {
			println("Failed to make dialog:", err.Error())
			return
		}
		filter, err := gtk.FileFilterNew()
		if err != nil {
			println("Failed to make filter:", err.Error())
			return
		}
		filter.AddPattern("*.bin")
		dialog.AddFilter(filter)
		resp := dialog.Run()
		if gtk.ResponseType(resp) == gtk.RESPONSE_ACCEPT {
			filename := dialog.GetFilename()
			err := ioutil.WriteFile(filename, viewer.hexDumpData, 0644)
			if err != nil {
				ShowError(viewer.mainWidget, err, "Saving packet to file")
			}
		}
	})

	reliability_, err := builder.GetObject("reliabilitytype")
	if err != nil {
    	return nil, err
	}
	reliability, ok := reliability_.(*gtk.Label)
	if !ok {
    	return nil, invalidUi("reliabilitytype")
	}
	rmNumber_, err := builder.GetObject("rmnumber")
	if err != nil {
    	return nil, err
	}
	rmNumber, ok := rmNumber_.(*gtk.Label)
	if !ok {
    	return nil, invalidUi("rmnumber")
	}
	channel_, err := builder.GetObject("channel")
	if err != nil {
    	return nil, err
	}
	channel, ok := channel_.(*gtk.Label)
	if !ok {
    	return nil, invalidUi("channel")
	}
	index_, err := builder.GetObject("orderingidx")
	if err != nil {
    	return nil, err
	}
	index, ok := index_.(*gtk.Label)
	if !ok {
    	return nil, invalidUi("orderingidx")
	}
	splitPacketsLabel_, err := builder.GetObject("splitpacketslabel")
	if err != nil {
    	return nil, err
	}
	splitPacketsLabel, ok := splitPacketsLabel_.(*gtk.Label)
	if !ok {
    	return nil, invalidUi("splitpacketslabel")
	}

	drawingArea_, err := builder.GetObject("splitpacketsdrawarea")
	if err != nil {
    	return nil, err
	}
	drawingArea, ok := drawingArea_.(*gtk.DrawingArea)
	if !ok {
    	return nil, invalidUi("drawingarea")
	}
	drawingArea.Connect("draw", func(drawingArea *gtk.DrawingArea, ctx *cairo.Context) bool {
        viewWidth := float64(drawingArea.GetAllocatedWidth())
        viewHeight := float64(drawingArea.GetAllocatedHeight())
		ctx.Rectangle(0, 0, viewWidth, viewHeight)
		ctx.SetSourceRGB(46.0/255.0, 125.0/255.0, 50.0/255.0)
		ctx.Fill()

		ctx.SetSourceRGB(198.0/255.0, 40.0/255.0, 40.0/255.0)
		count := float64(viewer.countSplitPackets)
		for _, range_ := range viewer.missingSplitPacketRanges {
			startCoord := float64(range_.start) / count * viewWidth
			width := math.Ceil(float64(range_.end - range_.start) / count) * viewWidth
			ctx.Rectangle(startCoord, 0, width, viewHeight)
			ctx.Fill()
		}

		return false
	})

	viewer.mainWidget = notebook
	viewer.logBox = logBox
	viewer.hexBox = hexBox
	viewer.reliablity = reliability
	viewer.rmNumber = rmNumber
	viewer.channel = channel
	viewer.index = index
	viewer.splitPacketsLabel = splitPacketsLabel

	return viewer, nil
}

func (viewer *PacketDetailsViewer) updateHexTab(layers *peer.PacketLayers) error {
	if layers.OfflinePayload != nil {
		viewer.hexDumpData = layers.OfflinePayload
	} else {
		viewer.hexDumpData = layers.Reliability.SplitBuffer.Data
	}
	textBuffer, err := gtk.TextBufferNew(nil)
	if err != nil {
		return err
	}
	textBuffer.SetText(hex.Dump(viewer.hexDumpData))
	viewer.hexBox.SetBuffer(textBuffer)
	return nil
}

var reliabilityNames = []string{
	"Unreliable",
	"Unreliable, sequenced",
	"Reliable",
	"Reliable, ordered",
	"Reliable, sequenced",
}

func (viewer *PacketDetailsViewer) updateReliabilityTab(layers *peer.PacketLayers) error {
	if layers.OfflinePayload != nil {
		viewer.missingSplitPacketRanges = nil
		viewer.reliablity.SetText("N/A")
		viewer.rmNumber.SetText("N/A")
		viewer.channel.SetText("N/A")
		viewer.index.SetText("N/A")
		viewer.splitPacketsLabel.SetText("Split packets (N/A):")
	} else {
    	viewer.reliablity.SetText(reliabilityNames[layers.Reliability.Reliability])
    	if layers.Reliability.IsReliable() {
			viewer.rmNumber.SetText(strconv.FormatInt(int64(layers.Reliability.ReliableMessageNumber), 10))
    	} else {
        	viewer.rmNumber.SetText("N/A")
    	}
    	if layers.Reliability.IsOrdered() {
			viewer.channel.SetText(strconv.FormatInt(int64(layers.Reliability.OrderingChannel), 10))
			viewer.index.SetText(strconv.FormatInt(int64(layers.Reliability.OrderingIndex), 10))
    	} else {
        	viewer.channel.SetText("N/A")
        	viewer.index.SetText("N/A")
    	}

		viewer.countSplitPackets = layers.Reliability.SplitPacketCount
    	viewer.missingSplitPacketRanges = nil
    	received := 0
    	for i := uint32(0); i < viewer.countSplitPackets; i++ {
        	if layers.SplitPacket.ReliablePackets[i] == nil {
            	rangeCount := len(viewer.missingSplitPacketRanges)
				if rangeCount == 0 || viewer.missingSplitPacketRanges[rangeCount - 1].end != i - 1 {
    				viewer.missingSplitPacketRanges = append(viewer.missingSplitPacketRanges, splitPacketRange{
                         start: i,
                         end: i,
    				})
				} else {
    				viewer.missingSplitPacketRanges[rangeCount - 1].end++
				}
        	} else {
            	received++
        	}
    	}
    	viewer.splitPacketsLabel.SetText(fmt.Sprintf("Split packets (%d/%d):", received, viewer.countSplitPackets))
	}

	return nil
}

func (viewer *PacketDetailsViewer) ShowPacket(layers *peer.PacketLayers) error {
	err := viewer.updateHexTab(layers)
	if err != nil {
    	return err
	}
	err = viewer.updateReliabilityTab(layers)
	if err != nil {
    	return err
	}

	return nil
}
