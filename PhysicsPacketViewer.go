package main

import (
	"fmt"
	"strconv"

	"github.com/Gskartwii/roblox-dissector/peer"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
	"github.com/robloxapi/rbxfile"
)

const (
	COL_PHYSICS_NAME = iota
	COL_PHYSICS_VALUE
)

type PhysicsPacketViewer struct {
	mainWidget *gtk.ScrolledWindow
	model      *gtk.TreeStore
	treeView   *gtk.TreeView
}

var NetworkHumanoidStates = [...]string{
	"Falling down",
	"Ragdoll",
	"Getting up",
	"Jumping",
	"Swimming",
	"Freefall",
	"Flying",
	"Landed",
	"Running",
	"Unknown 9",
	"Running, no physics",
	"Strafing, no physics",
	"Climbing",
	"Seated",
	"Standing on a platform",
	"Dead",
	"Pure physics",
	"Unknown 17",
	"None",
}

func humanoidStateName(id uint8) string {
	if int(id) < len(NetworkHumanoidStates) {
		return NetworkHumanoidStates[id]
	}
	return fmt.Sprintf("Unknown %d", id)
}

func appendPhysicsFloatRow(model *gtk.TreeStore, parent *gtk.TreeIter, name string, val float32) {
	floatRow := model.Append(parent)
	model.SetValue(floatRow, COL_PHYSICS_NAME, name)
	model.SetValue(floatRow, COL_PHYSICS_VALUE, strconv.FormatFloat(float64(val), 'g', -1, 32))
}

func appendPhysicsVelocityRow(model *gtk.TreeStore, parent *gtk.TreeIter, name string, val rbxfile.ValueVector3) {
	veloRow := model.Append(parent)
	model.SetValue(veloRow, COL_PHYSICS_NAME, name)
	model.SetValue(veloRow, COL_PHYSICS_VALUE, val.String())
	appendPhysicsFloatRow(model, veloRow, "X", val.X)
	appendPhysicsFloatRow(model, veloRow, "Y", val.Y)
	appendPhysicsFloatRow(model, veloRow, "Z", val.Z)
}

func appendPhysicsCFrameRow(model *gtk.TreeStore, parent *gtk.TreeIter, name string, val rbxfile.ValueCFrame) {
	cframeRow := model.Append(parent)
	model.SetValue(cframeRow, COL_PHYSICS_NAME, name)
	model.SetValue(cframeRow, COL_PHYSICS_VALUE, val.String())
	appendPhysicsFloatRow(model, cframeRow, "X", val.Position.X)
	appendPhysicsFloatRow(model, cframeRow, "Y", val.Position.Y)
	appendPhysicsFloatRow(model, cframeRow, "Z", val.Position.Z)
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			appendPhysicsFloatRow(model, cframeRow, fmt.Sprintf("R%d%d", i, j), val.Rotation[3*i+j])
		}
	}
}

func appendPhysicsDataRows(model *gtk.TreeStore, parent *gtk.TreeIter, data *peer.PhysicsData) {
	appendPhysicsCFrameRow(model, parent, "CFrame", data.CFrame)
	appendPhysicsVelocityRow(model, parent, "Linear velocity", data.LinearVelocity)
	appendPhysicsVelocityRow(model, parent, "Angular velocity", data.RotationalVelocity)

	if data.PlatformChild != nil {
		pcRow := model.Append(parent)
		model.SetValue(pcRow, COL_PHYSICS_NAME, "Platform child")
		model.SetValue(pcRow, COL_PHYSICS_VALUE, fmt.Sprintf("%s: %s", data.PlatformChild.Ref, data.PlatformChild.Name()))
	}

	for i, motor := range data.Motors {
		appendPhysicsCFrameRow(model, parent, "Motor "+strconv.Itoa(i), rbxfile.ValueCFrame(motor))
	}
}

func (viewer *PhysicsPacketViewer) ViewPacket(packet *peer.Packet85LayerSubpacket) {
	viewer.model.Clear()
	humStateRow := viewer.model.Append(nil)
	viewer.model.SetValue(humStateRow, COL_PHYSICS_NAME, "Humanoid State")
	viewer.model.SetValue(humStateRow, COL_PHYSICS_VALUE, humanoidStateName(packet.NetworkHumanoidState))

	if len(packet.History) == 0 {
		appendPhysicsDataRows(viewer.model, nil, &packet.Data)
	} else {
		for i, motor := range packet.Data.Motors {
			appendPhysicsCFrameRow(viewer.model, nil, "Motor "+strconv.Itoa(i), rbxfile.ValueCFrame(motor))
		}

		for _, data := range packet.History {
			intervalRow := viewer.model.Append(nil)
			viewer.model.SetValue(intervalRow, COL_PHYSICS_NAME, fmt.Sprintf("%+f", data.Interval))
			appendPhysicsDataRows(viewer.model, intervalRow, data)
		}
	}
	for i, child := range packet.Children {
		childRow := viewer.model.Append(nil)
		viewer.model.SetValue(childRow, COL_PHYSICS_NAME, "Assembly child "+strconv.Itoa(i))
		viewer.model.SetValue(childRow, COL_PHYSICS_VALUE, child.Instance.Ref.String()+": "+child.Instance.Name())
		appendPhysicsDataRows(viewer.model, childRow, child)
	}
}

func NewPhysicsPacketViewer() (*PhysicsPacketViewer, error) {
	viewer := &PhysicsPacketViewer{}

	mainWidget, err := gtk.ScrolledWindowNew(nil, nil)
	if err != nil {
		return nil, err
	}

	model, err := gtk.TreeStoreNew(
		glib.TYPE_STRING, // COL_PHYSICS_NAME
		glib.TYPE_STRING, // COL_PHYSICS_VALUE
	)
	if err != nil {
		return nil, err
	}

	treeView, err := gtk.TreeViewNewWithModel(model)
	if err != nil {
		return nil, err
	}

	for i, colName := range []string{"Name", "Value"} {
		colRenderer, err := gtk.CellRendererTextNew()
		if err != nil {
			return nil, err
		}
		col, err := gtk.TreeViewColumnNewWithAttribute(
			colName,
			colRenderer,
			"text",
			i,
		)
		if err != nil {
			return nil, err
		}
		col.SetSortColumnID(i)
		treeView.AppendColumn(col)
	}

	mainWidget.Add(treeView)

	viewer.mainWidget = mainWidget
	viewer.model = model
	viewer.treeView = treeView

	return viewer, nil
}
