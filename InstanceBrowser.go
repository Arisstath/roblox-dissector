package main

import (
	"github.com/Gskartwii/roblox-dissector/datamodel"
	"github.com/Gskartwii/roblox-dissector/peer"
	"github.com/gotk3/gotk3/cairo"
	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
	"github.com/gotk3/gotk3/pango"
	"github.com/robloxapi/rbxfile"

	"encoding/hex"
	"fmt"
	"strconv"
)

const (
	COL_PROP_NAME = iota
	COL_PROP_TYPE
	COL_PROP_VALUE
	COL_SHOW_PIXBUF
	COL_PIXBUF
	COL_PROP_ADDITIONAL_VALUE
)

type InstanceViewer struct {
	mainWidget *gtk.Box
	model      *gtk.TreeStore
	view       *gtk.TreeView

	class  *gtk.Label
	id     *gtk.Label
	parent *gtk.Label
}

func bindValueCopy(model *gtk.TreeStore, view *gtk.TreeView) error {
	_, err := view.Connect("button-press-event", func(_ gtk.IWidget, evt *gdk.Event) {
		trueEvt := gdk.EventButtonNewFromEvent(evt)
		// only care about right clicks
		if trueEvt.Button() != gdk.BUTTON_SECONDARY {
			return
		}
		x := trueEvt.X()
		y := trueEvt.Y()
		path, _, _, _, _ := view.GetPathAtPos(int(x), int(y))
		if path == nil {
			// There's no row here
			return
		}
		iter, err := model.GetIter(path)
		if err != nil {
			println("Failed to get iter:", err.Error())
			return
		}
		popupMenu, err := gtk.MenuNew()
		if err != nil {
			println("Failed to make menu:", err.Error())
			return
		}
		copyAction, err := gtk.MenuItemNewWithLabel("Copy value")
		if err != nil {
			println("Failed to make menu:", err.Error())
			return
		}
		copyAction.Connect("activate", func() {
			copied, err := model.GetValue(iter, COL_PROP_ADDITIONAL_VALUE)
			if err != nil {
				println("Failed to copy:", err.Error())
				return
			}
			val, err := copied.GetString()
			if err != nil {
				println("Failed to copy:", err.Error())
				return
			}
			clipboard, err := gtk.ClipboardGet(gdk.GdkAtomIntern("CLIPBOARD", true))
			if err != nil {
				println("Failed to get clipboard:", err.Error())
				return
			}
			clipboard.SetText(val)
		})
		popupMenu.Append(copyAction)
		popupMenu.ShowAll()
		popupMenu.PopupAtPointer(evt)
	})
	return err
}

func appendValueRow(model *gtk.TreeStore, parent *gtk.TreeIter, name string, value rbxfile.Value, treeView  *gtk.TreeView) {
	newRow := model.Append(parent)
	model.SetValue(newRow, COL_PROP_NAME, name)
	model.SetValue(newRow, COL_PROP_TYPE, datamodel.TypeString(value))
	model.SetValue(newRow, COL_PROP_VALUE, value.String())
	model.SetValue(newRow, COL_SHOW_PIXBUF, false)
	model.SetValue(newRow, COL_PROP_ADDITIONAL_VALUE, value.String())

	switch value.Type() {
	case rbxfile.TypeCFrame:
		cf := value.(rbxfile.ValueCFrame)
		appendValueRow(model, newRow, "X", rbxfile.ValueFloat(cf.Position.X), treeView)
		appendValueRow(model, newRow, "Y", rbxfile.ValueFloat(cf.Position.Y), treeView)
		appendValueRow(model, newRow, "Z", rbxfile.ValueFloat(cf.Position.Z), treeView)

		for i := 0; i < 3; i++ {
			for j := 0; j < 3; j++ {
				appendValueRow(model, newRow, fmt.Sprintf("R%d%d", i, j), rbxfile.ValueFloat(cf.Rotation[3*i+j]), treeView)
			}
		}
	case rbxfile.TypeColor3uint8:
		c3 := value.(rbxfile.ValueColor3uint8)

		styles, err := treeView.GetStyleContext()
		if err != nil {
    		println("failed to get style context")
    		return
		}
		borderColor := styles.GetColor(gtk.STATE_FLAG_NORMAL).Floats()
		
		surface := cairo.CreateImageSurface(cairo.FORMAT_ARGB32, 16, 16)
		context := cairo.Create(surface)

		context.SetSourceRGBA(float64(c3.R)/255.0, float64(c3.G)/255.0, float64(c3.B)/255.0, 1.0)
		context.Rectangle(0, 0, 16, 16)
		context.Fill()

		context.SetSourceRGBA(borderColor[0], borderColor[1], borderColor[2], borderColor[3])
		context.Rectangle(0, 0, 16, 16)
		context.Stroke()

		surface.Flush()
		pixbuf, err := gdk.PixbufGetFromSurface(surface, 0, 0, 16, 16)
		if err != nil {
    		println("failed to get pixbuf")
    		return
		}
		model.SetValue(newRow, COL_PIXBUF, pixbuf)
		model.SetValue(newRow, COL_SHOW_PIXBUF, true)
	case rbxfile.TypeColorSequence, datamodel.TypeColorSequence:
		cs := value.(datamodel.ValueColorSequence)
		for i, keypoint := range cs {
			appendValueRow(model, newRow, fmt.Sprintf("Keypoint %d", i), keypoint, treeView)
		}
	case datamodel.TypeColorSequenceKeypoint:
		kp := value.(datamodel.ValueColorSequenceKeypoint)
		appendValueRow(model, newRow, "Color", kp.Value, treeView)
		appendValueRow(model, newRow, "Time", rbxfile.ValueFloat(kp.Time), treeView)
		appendValueRow(model, newRow, "Envelope", rbxfile.ValueFloat(kp.Envelope), treeView)
	case rbxfile.TypeNumberRange:
		ra := value.(rbxfile.ValueNumberRange)
		appendValueRow(model, newRow, "Min", rbxfile.ValueFloat(ra.Min), treeView)
		appendValueRow(model, newRow, "Max", rbxfile.ValueFloat(ra.Max), treeView)
	case rbxfile.TypeNumberSequence, datamodel.TypeNumberSequence:
		ns := value.(datamodel.ValueNumberSequence)
		for i, keypoint := range ns {
			appendValueRow(model, newRow, fmt.Sprintf("Keypoint %d", i), keypoint, treeView)
		}
	case datamodel.TypeNumberSequenceKeypoint:
		kp := value.(datamodel.ValueNumberSequenceKeypoint)
		appendValueRow(model, newRow, "Value", rbxfile.ValueFloat(kp.Value), treeView)
		appendValueRow(model, newRow, "Time", rbxfile.ValueFloat(kp.Time), treeView)
		appendValueRow(model, newRow, "Envelope", rbxfile.ValueFloat(kp.Envelope), treeView)
	case rbxfile.TypePhysicalProperties:
		pp := value.(rbxfile.ValuePhysicalProperties)
		if pp.CustomPhysics {
			appendValueRow(model, newRow, "Density", rbxfile.ValueFloat(pp.Density), treeView)
			appendValueRow(model, newRow, "Friction", rbxfile.ValueFloat(pp.Friction), treeView)
			appendValueRow(model, newRow, "Elasticity", rbxfile.ValueFloat(pp.Elasticity), treeView)
			appendValueRow(model, newRow, "Friction weight", rbxfile.ValueFloat(pp.FrictionWeight), treeView)
			appendValueRow(model, newRow, "Elasticity weight", rbxfile.ValueFloat(pp.ElasticityWeight), treeView)
		}
	case rbxfile.TypeRay:
		ray := value.(rbxfile.ValueRay)
		appendValueRow(model, newRow, "Origin", ray.Origin, treeView)
		appendValueRow(model, newRow, "Direction", ray.Direction, treeView)
	case rbxfile.TypeRect2D:
		rect := value.(rbxfile.ValueRect2D)
		appendValueRow(model, newRow, "Min", rect.Min, treeView)
		appendValueRow(model, newRow, "Max", rect.Max, treeView)
	case rbxfile.TypeUDim:
		ud := value.(rbxfile.ValueUDim)
		appendValueRow(model, newRow, "Scale", rbxfile.ValueFloat(ud.Scale), treeView)
		appendValueRow(model, newRow, "Offset", rbxfile.ValueInt(ud.Offset), treeView)
	case rbxfile.TypeUDim2:
		ud2 := value.(rbxfile.ValueUDim2)
		appendValueRow(model, newRow, "X", ud2.X, treeView)
		appendValueRow(model, newRow, "Y", ud2.Y, treeView)
	case rbxfile.TypeVector2:
		v2 := value.(rbxfile.ValueVector2)
		appendValueRow(model, newRow, "X", rbxfile.ValueFloat(v2.X), treeView)
		appendValueRow(model, newRow, "Y", rbxfile.ValueFloat(v2.Y), treeView)
	case rbxfile.TypeVector2int16:
		v2 := value.(rbxfile.ValueVector2int16)
		appendValueRow(model, newRow, "X", rbxfile.ValueInt(v2.X), treeView)
		appendValueRow(model, newRow, "Y", rbxfile.ValueInt(v2.Y), treeView)
	case rbxfile.TypeVector3:
		v3 := value.(rbxfile.ValueVector3)
		appendValueRow(model, newRow, "X", rbxfile.ValueFloat(v3.X), treeView)
		appendValueRow(model, newRow, "Y", rbxfile.ValueFloat(v3.Y), treeView)
		appendValueRow(model, newRow, "Z", rbxfile.ValueFloat(v3.Z), treeView)
	case rbxfile.TypeVector3int16:
		v3 := value.(rbxfile.ValueVector3int16)
		appendValueRow(model, newRow, "X", rbxfile.ValueInt(v3.X), treeView)
		appendValueRow(model, newRow, "Y", rbxfile.ValueInt(v3.Y), treeView)
		appendValueRow(model, newRow, "Y", rbxfile.ValueInt(v3.Z), treeView)

	case datamodel.TypeArray:
		arr := value.(datamodel.ValueArray)
		for i, v := range arr {
			appendValueRow(model, newRow, strconv.Itoa(i), v, treeView)
		}
	case datamodel.TypeDictionary:
		dict := value.(datamodel.ValueDictionary)
		for i, v := range dict {
			appendValueRow(model, newRow, i, v, treeView)
		}
	case datamodel.TypeMap:
		dict := value.(datamodel.ValueMap)
		for i, v := range dict {
			appendValueRow(model, newRow, i, v, treeView)
		}
	case datamodel.TypeRegion3:
		r3 := value.(datamodel.ValueRegion3)
		appendValueRow(model, newRow, "Start", r3.Start, treeView)
		appendValueRow(model, newRow, "End", r3.End, treeView)
	case datamodel.TypeRegion3int16:
		r3 := value.(datamodel.ValueRegion3int16)
		appendValueRow(model, newRow, "Start", r3.Start, treeView)
		appendValueRow(model, newRow, "End", r3.End, treeView)
	case datamodel.TypeTuple:
		arr := value.(datamodel.ValueTuple)
		for i, v := range arr {
			appendValueRow(model, newRow, strconv.Itoa(i), v, treeView)
		}
	case datamodel.TypeSignedProtectedString:
    	str := value.(datamodel.ValueSignedProtectedString)
    	model.SetValue(newRow, COL_PROP_ADDITIONAL_VALUE, hex.EncodeToString(str.Value))
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
		appendValueRow(viewer.model, nil, name, value, viewer.view)
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

	emptyPixbuf, err := gdk.PixbufNew(0, false, 8, 1, 1)
	if err != nil {
		return nil, err
	}
	model, err := gtk.TreeStoreNew(
		glib.TYPE_STRING,               // COL_PROP_NAME
		glib.TYPE_STRING,               // COL_PROP_TYPE
		glib.TYPE_STRING,               // COL_PROP_VALUE
		glib.TYPE_BOOLEAN,              // COL_SHOW_PIXBUF
		emptyPixbuf.TypeFromInstance(), // COL_PIXBUF
		glib.TYPE_STRING, // COL_PROP_ADDITIONAL_VALUE
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

	for i, colName := range []string{"Name", "Type"} {
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

	colorRenderer, err := gtk.CellRendererPixbufNew()
	if err != nil {
		return nil, err
	}
	colRenderer, err := gtk.CellRendererTextNew()
	if err != nil {
		return nil, err
	}
	col, err := gtk.TreeViewColumnNew()
	if err != nil {
		return nil, err
	}
	colRenderer.Set("ellipsize", int(pango.ELLIPSIZE_END))
	col.PackStart(colorRenderer, false)
	col.PackStart(colRenderer, true)
	col.SetSpacing(4)
	col.SetTitle("Value")
	col.AddAttribute(colRenderer, "text", COL_PROP_VALUE)
	col.AddAttribute(colorRenderer, "visible", COL_SHOW_PIXBUF)
	col.AddAttribute(colorRenderer, "pixbuf", COL_PIXBUF)
	col.SetSortColumnID(COL_PROP_VALUE)
	treeView.AppendColumn(col)

	err = bindValueCopy(model, treeView)
	if err != nil {
		return nil, err
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
	viewer.view = treeView
	viewer.model = model
	viewer.id = id
	viewer.class = class
	viewer.parent = parent

	return viewer, nil
}

type PropEventViewer struct {
	mainWidget *gtk.Box
	model      *gtk.TreeStore
	view *gtk.TreeView

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
	appendValueRow(viewer.model, nil, name, newValue, viewer.view)
}
func (viewer *PropEventViewer) ViewEvent(instance *datamodel.Instance, name string, arguments []rbxfile.Value) {
	viewer.name.SetText(name)
	viewer.id.SetText("ID: " + instance.Ref.String())
	viewer.version.SetVisible(false)
	viewer.instancename.SetText("Instance name: " + instance.Name())

	viewer.model.Clear()
	for i, val := range arguments {
		appendValueRow(viewer.model, nil, "Argument "+strconv.Itoa(i), val, viewer.view)
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

	emptyPixbuf, err := gdk.PixbufNew(0, false, 8, 1, 1)
	if err != nil {
		return nil, err
	}

	model, err := gtk.TreeStoreNew(
		glib.TYPE_STRING, // COL_PROP_NAME
		glib.TYPE_STRING, // COL_PROP_TYPE
		glib.TYPE_STRING, // COL_PROP_VALUE
    	glib.TYPE_BOOLEAN, // COL_SHOW_PIXBUF
		emptyPixbuf.TypeFromInstance(), // COL_PIXBUF
    	glib.TYPE_STRING, // COL_PROP_ADDITIONAL_VALUE
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

	for i, colName := range []string{"Name", "Type"} {
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

	colorRenderer, err := gtk.CellRendererPixbufNew()
	if err != nil {
		return nil, err
	}
	colRenderer, err := gtk.CellRendererTextNew()
	if err != nil {
		return nil, err
	}
	col, err := gtk.TreeViewColumnNew()
	if err != nil {
		return nil, err
	}
	colRenderer.Set("ellipsize", int(pango.ELLIPSIZE_END))
	col.PackStart(colorRenderer, false)
	col.PackStart(colRenderer, true)
	col.SetSpacing(4)
	col.SetTitle("Value")
	col.AddAttribute(colRenderer, "text", COL_PROP_VALUE)
	col.AddAttribute(colorRenderer, "visible", COL_SHOW_PIXBUF)
	col.AddAttribute(colorRenderer, "pixbuf", COL_PIXBUF)
	col.SetSortColumnID(COL_PROP_VALUE)
	treeView.AppendColumn(col)

	err = bindValueCopy(model, treeView)
	if err != nil {
		return nil, err
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
	viewer.view = treeView
	viewer.model = model
	viewer.id = id
	viewer.name = name
	viewer.instancename = instancename
	viewer.version = version

	return viewer, nil
}
