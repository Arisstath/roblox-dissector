package main

import (
	"github.com/Gskartwii/roblox-dissector/datamodel"
	"github.com/Gskartwii/roblox-dissector/peer"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
	"github.com/gotk3/gotk3/pango"
	"github.com/robloxapi/rbxfile"

	"fmt"
	"strconv"
)

const (
	COL_PROP_NAME = iota
	COL_PROP_TYPE
	COL_PROP_VALUE
)

type InstanceViewer struct {
	mainWidget *gtk.Box
	model      *gtk.TreeStore

	class  *gtk.Label
	id     *gtk.Label
	parent *gtk.Label
}

func appendValueRow(model *gtk.TreeStore, parent *gtk.TreeIter, name string, value rbxfile.Value) {
	newRow := model.Append(parent)
	model.SetValue(newRow, COL_PROP_NAME, name)
	model.SetValue(newRow, COL_PROP_TYPE, datamodel.TypeString(value))
	model.SetValue(newRow, COL_PROP_VALUE, value.String())

	switch value.Type() {
	case rbxfile.TypeCFrame:
		cf := value.(rbxfile.ValueCFrame)
		appendValueRow(model, newRow, "X", rbxfile.ValueFloat(cf.Position.X))
		appendValueRow(model, newRow, "Y", rbxfile.ValueFloat(cf.Position.Y))
		appendValueRow(model, newRow, "Z", rbxfile.ValueFloat(cf.Position.Z))

		for i := 0; i < 3; i++ {
			for j := 0; j < 3; j++ {
				appendValueRow(model, newRow, fmt.Sprintf("R%d%d", i, j), rbxfile.ValueFloat(cf.Rotation[3*i+j]))
			}
		}
	case rbxfile.TypeColorSequence, datamodel.TypeColorSequence:
		cs := value.(datamodel.ValueColorSequence)
		for i, keypoint := range cs {
			appendValueRow(model, newRow, fmt.Sprintf("Keypoint %d", i), keypoint)
		}
	case datamodel.TypeColorSequenceKeypoint:
		kp := value.(datamodel.ValueColorSequenceKeypoint)
		appendValueRow(model, newRow, "Color", kp.Value)
		appendValueRow(model, newRow, "Time", rbxfile.ValueFloat(kp.Time))
		appendValueRow(model, newRow, "Envelope", rbxfile.ValueFloat(kp.Envelope))
	case rbxfile.TypeNumberRange:
		ra := value.(rbxfile.ValueNumberRange)
		appendValueRow(model, newRow, "Min", rbxfile.ValueFloat(ra.Min))
		appendValueRow(model, newRow, "Max", rbxfile.ValueFloat(ra.Max))
	case rbxfile.TypeNumberSequence, datamodel.TypeNumberSequence:
		ns := value.(datamodel.ValueNumberSequence)
		for i, keypoint := range ns {
			appendValueRow(model, newRow, fmt.Sprintf("Keypoint %d", i), keypoint)
		}
	case datamodel.TypeNumberSequenceKeypoint:
		kp := value.(datamodel.ValueNumberSequenceKeypoint)
		appendValueRow(model, newRow, "Value", rbxfile.ValueFloat(kp.Value))
		appendValueRow(model, newRow, "Time", rbxfile.ValueFloat(kp.Time))
		appendValueRow(model, newRow, "Envelope", rbxfile.ValueFloat(kp.Envelope))
	case rbxfile.TypePhysicalProperties:
		pp := value.(rbxfile.ValuePhysicalProperties)
		if pp.CustomPhysics {
			appendValueRow(model, newRow, "Density", rbxfile.ValueFloat(pp.Density))
			appendValueRow(model, newRow, "Friction", rbxfile.ValueFloat(pp.Friction))
			appendValueRow(model, newRow, "Elasticity", rbxfile.ValueFloat(pp.Elasticity))
			appendValueRow(model, newRow, "Friction weight", rbxfile.ValueFloat(pp.FrictionWeight))
			appendValueRow(model, newRow, "Elasticity weight", rbxfile.ValueFloat(pp.ElasticityWeight))
		}
	case rbxfile.TypeRay:
		ray := value.(rbxfile.ValueRay)
		appendValueRow(model, newRow, "Origin", ray.Origin)
		appendValueRow(model, newRow, "Direction", ray.Direction)
	case rbxfile.TypeRect2D:
		rect := value.(rbxfile.ValueRect2D)
		appendValueRow(model, newRow, "Min", rect.Min)
		appendValueRow(model, newRow, "Max", rect.Max)
	case rbxfile.TypeUDim:
		ud := value.(rbxfile.ValueUDim)
		appendValueRow(model, newRow, "Scale", rbxfile.ValueFloat(ud.Scale))
		appendValueRow(model, newRow, "Offset", rbxfile.ValueInt(ud.Offset))
	case rbxfile.TypeUDim2:
		ud2 := value.(rbxfile.ValueUDim2)
		appendValueRow(model, newRow, "X", ud2.X)
		appendValueRow(model, newRow, "Y", ud2.Y)
	case rbxfile.TypeVector2:
		v2 := value.(rbxfile.ValueVector2)
		appendValueRow(model, newRow, "X", rbxfile.ValueFloat(v2.X))
		appendValueRow(model, newRow, "Y", rbxfile.ValueFloat(v2.Y))
	case rbxfile.TypeVector2int16:
		v2 := value.(rbxfile.ValueVector2int16)
		appendValueRow(model, newRow, "X", rbxfile.ValueInt(v2.X))
		appendValueRow(model, newRow, "Y", rbxfile.ValueInt(v2.Y))
	case rbxfile.TypeVector3:
		v3 := value.(rbxfile.ValueVector3)
		appendValueRow(model, newRow, "X", rbxfile.ValueFloat(v3.X))
		appendValueRow(model, newRow, "Y", rbxfile.ValueFloat(v3.Y))
		appendValueRow(model, newRow, "Z", rbxfile.ValueFloat(v3.Z))
	case rbxfile.TypeVector3int16:
		v3 := value.(rbxfile.ValueVector3int16)
		appendValueRow(model, newRow, "X", rbxfile.ValueInt(v3.X))
		appendValueRow(model, newRow, "Y", rbxfile.ValueInt(v3.Y))
		appendValueRow(model, newRow, "Y", rbxfile.ValueInt(v3.Z))

	case datamodel.TypeArray:
		arr := value.(datamodel.ValueArray)
		for i, v := range arr {
			appendValueRow(model, newRow, strconv.Itoa(i), v)
		}
	case datamodel.TypeDictionary:
		dict := value.(datamodel.ValueDictionary)
		for i, v := range dict {
			appendValueRow(model, newRow, i, v)
		}
	case datamodel.TypeMap:
		dict := value.(datamodel.ValueMap)
		for i, v := range dict {
			appendValueRow(model, newRow, i, v)
		}
	case datamodel.TypeRegion3:
		r3 := value.(datamodel.ValueRegion3)
		appendValueRow(model, newRow, "Start", r3.Start)
		appendValueRow(model, newRow, "End", r3.End)
	case datamodel.TypeRegion3int16:
		r3 := value.(datamodel.ValueRegion3int16)
		appendValueRow(model, newRow, "Start", r3.Start)
		appendValueRow(model, newRow, "End", r3.End)
	case datamodel.TypeTuple:
		arr := value.(datamodel.ValueTuple)
		for i, v := range arr {
			appendValueRow(model, newRow, strconv.Itoa(i), v)
		}
	}
}

func (viewer *InstanceViewer) ViewInstance(instance *peer.ReplicationInstance) {
	viewer.class.SetText(instance.Instance.ClassName)
	viewer.id.SetText("ID: " + instance.Instance.Ref.String())
	if instance.Parent != nil {
		viewer.parent.SetText("Parent: " + instance.Parent.Ref.String())
	} else {
		viewer.parent.SetText("Parent: nil")
	}

	viewer.model.Clear()
	for name, value := range instance.Properties {
		appendValueRow(viewer.model, nil, name, value)
	}
}

func NewInstanceViewer() (*InstanceViewer, error) {
	viewer := &InstanceViewer{}

	builder, err := gtk.BuilderNewFromFile("instancebrowser.ui")
	if err != nil {
		return nil, err
	}
	mainWidget_, err := builder.GetObject("instanceinfobox")
	if err != nil {
		return nil, err
	}
	mainWidget, ok := mainWidget_.(*gtk.Box)
	if !ok {
		return nil, invalidUi("instanceinfbox")
	}

	mainContainer_, err := builder.GetObject("instanceviewpanes")
	if err != nil {
		return nil, err
	}
	mainContainer, ok := mainContainer_.(*gtk.Paned)
	if !ok {
		return nil, invalidUi("instance viewer container")
	}
	mainContainer.Remove(mainWidget)

	propertiesViewContainer_, err := builder.GetObject("propertiescontainer")
	if err != nil {
		return nil, err
	}
	propertiesContainer, ok := propertiesViewContainer_.(*gtk.ScrolledWindow)
	if !ok {
		return nil, invalidUi("properties viewer container")
	}

	model, err := gtk.TreeStoreNew(
		glib.TYPE_STRING, // COL_PROP_NAME
		glib.TYPE_STRING, // COL_PROP_TYPE
		glib.TYPE_STRING, // COL_PROP_VALUE
	)
	if err != nil {
		return nil, err
	}
	treeView, err := gtk.TreeViewNewWithModel(model)
	if err != nil {
		return nil, err
	}
	treeView.SetHExpand(true)
	propertiesContainer.Add(treeView)

	for i, colName := range []string{"Name", "Type", "Value"} {
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

		if i == COL_PROP_VALUE {
			colRenderer.Set("ellipsize", int(pango.ELLIPSIZE_END))
		}

		treeView.AppendColumn(col)
	}

	id_, err := builder.GetObject("instanceidlabel")
	if err != nil {
		return nil, err
	}
	id, ok := id_.(*gtk.Label)
	if !ok {
		return nil, invalidUi("instanceidlabel")
	}

	class_, err := builder.GetObject("instanceclasslabel")
	if err != nil {
		return nil, err
	}
	class, ok := class_.(*gtk.Label)
	if !ok {
		return nil, invalidUi("instanceclasslabel")
	}

	parent_, err := builder.GetObject("instanceparentlabel")
	if err != nil {
		return nil, err
	}
	parent, ok := parent_.(*gtk.Label)
	if !ok {
		return nil, invalidUi("instanceparentlabel")
	}

	viewer.mainWidget = mainWidget
	viewer.model = model
	viewer.id = id
	viewer.class = class
	viewer.parent = parent

	return viewer, nil
}

type PropEventViewer struct {
	mainWidget *gtk.Box
	model      *gtk.TreeStore

	name         *gtk.Label
	id           *gtk.Label
	instancename *gtk.Label
	version      *gtk.Label
}

func (viewer *PropEventViewer) ViewPropertyUpdate(instance *datamodel.Instance, name string, newValue rbxfile.Value, version int32) {
	viewer.name.SetText(name)
	viewer.id.SetText("ID: " + instance.Ref.String())
	viewer.version.SetVisible(true)
	if version == -1 {
		viewer.version.SetText("Version: N/A")
	} else {
		viewer.version.SetText("Version: " + strconv.FormatInt(int64(version), 10))
	}
	viewer.instancename.SetText("Instance name: " + instance.Name())

	viewer.model.Clear()
	appendValueRow(viewer.model, nil, name, newValue)
}
func (viewer *PropEventViewer) ViewEvent(instance *datamodel.Instance, name string, arguments []rbxfile.Value) {
	viewer.name.SetText(name)
	viewer.id.SetText("ID: " + instance.Ref.String())
	viewer.version.SetVisible(false)
	viewer.instancename.SetText("Instance name: " + instance.Name())

	viewer.model.Clear()
	for i, val := range arguments {
		appendValueRow(viewer.model, nil, "Argument "+strconv.Itoa(i), val)
	}
}

func NewPropertyEventViewer() (*PropEventViewer, error) {
	viewer := &PropEventViewer{}

	builder, err := gtk.BuilderNewFromFile("propeventviewer.ui")
	if err != nil {
		return nil, err
	}
	mainWidget_, err := builder.GetObject("propertyinfobox")
	if err != nil {
		return nil, err
	}
	mainWidget, ok := mainWidget_.(*gtk.Box)
	if !ok {
		return nil, invalidUi("propertyinfobox")
	}

	mainContainer_, err := builder.GetObject("propertiesviewcontainer")
	if err != nil {
		return nil, err
	}
	mainContainer, ok := mainContainer_.(*gtk.Window)
	if !ok {
		return nil, invalidUi("propertiesviewcontainer")
	}
	mainContainer.Remove(mainWidget)

	valuesContainer_, err := builder.GetObject("valuescontainer")
	if err != nil {
		return nil, err
	}
	valuesContainer, ok := valuesContainer_.(*gtk.ScrolledWindow)
	if !ok {
		return nil, invalidUi("valuescontainer")
	}

	model, err := gtk.TreeStoreNew(
		glib.TYPE_STRING, // COL_PROP_NAME
		glib.TYPE_STRING, // COL_PROP_TYPE
		glib.TYPE_STRING, // COL_PROP_VALUE
	)
	if err != nil {
		return nil, err
	}
	treeView, err := gtk.TreeViewNewWithModel(model)
	if err != nil {
		return nil, err
	}
	treeView.SetHExpand(true)
	valuesContainer.Add(treeView)

	for i, colName := range []string{"Name", "Type", "Value"} {
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

		if i == COL_PROP_VALUE {
			colRenderer.Set("ellipsize", int(pango.ELLIPSIZE_END))
		}

		treeView.AppendColumn(col)
	}

	name_, err := builder.GetObject("namelabel")
	if err != nil {
		return nil, err
	}
	name, ok := name_.(*gtk.Label)
	if !ok {
		return nil, invalidUi("namelabel")
	}

	id_, err := builder.GetObject("instanceidlabel")
	if err != nil {
		return nil, err
	}
	id, ok := id_.(*gtk.Label)
	if !ok {
		return nil, invalidUi("instanceidlabel")
	}

	instancename_, err := builder.GetObject("instancenamelabel")
	if err != nil {
		return nil, err
	}
	instancename, ok := instancename_.(*gtk.Label)
	if !ok {
		return nil, invalidUi("instancenamelabel")
	}

	version_, err := builder.GetObject("versionlabel")
	if err != nil {
		return nil, err
	}
	version, ok := version_.(*gtk.Label)
	if !ok {
		return nil, invalidUi("versionlabel")
	}

	viewer.mainWidget = mainWidget
	viewer.model = model
	viewer.id = id
	viewer.name = name
	viewer.instancename = instancename
	viewer.version = version

	return viewer, nil
}
