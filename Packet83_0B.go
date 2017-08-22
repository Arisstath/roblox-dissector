package main
import "github.com/google/gopacket"
import "github.com/therecipe/qt/widgets"
import "github.com/therecipe/qt/gui"
import "errors"
import "github.com/gskartwii/rbxfile"

type Packet83_0B struct {
	Instances []*rbxfile.Instance
}

func NewPacket83_0BLayer(length int) *Packet83_0B {
	return &Packet83_0B{make([]*rbxfile.Instance, length)}
}

func getInstanceRow(this *rbxfile.Instance) []*gui.QStandardItem {
    rootNameItem := NewQStandardItemF("Name: %s", this.Properties["Name"].String())
	typeItem := NewQStandardItemF(this.ClassName)
	referentItem := NewQStandardItemF(this.Reference)
	parentItem := NewQStandardItemF(this.Parent().Reference)

	for name, property := range this.Properties {
		nameItem := NewQStandardItemF(name)
		typeItem := NewQStandardItemF(property.Type().String())
		valueItem := NewQStandardItemF(property.String())

		rootNameItem.AppendRow([]*gui.QStandardItem{
			nameItem,
			typeItem,
			valueItem,
			nil,
			nil,
			nil,
		})
	}

	return []*gui.QStandardItem{
		rootNameItem,
		typeItem,
		nil,
		referentItem,
		parentItem,
	}
}

func (this Packet83_0B) Show() widgets.QWidget_ITF {
	instanceList := widgets.NewQTreeView(nil)
	standardModel := NewProperSortModel(nil)
	standardModel.SetHorizontalHeaderLabels([]string{"Name", "Type", "Value", "Referent", "Parent"})

	rootNode := standardModel.InvisibleRootItem()
	for _, instance := range(this.Instances) {
		rootNode.AppendRow(getInstanceRow(instance))
	}
	instanceList.SetModel(standardModel)
	instanceList.SetSelectionMode(0)
	instanceList.SetSortingEnabled(true)

	return instanceList
}

func DecodePacket83_0B(thisBitstream *ExtendedReader, context *CommunicationContext, packet gopacket.Packet, instanceSchema []*InstanceSchemaItem) (interface{}, error) {
	var layer *Packet83_0B
	thisBitstream.Align()
	arrayLen, err := thisBitstream.ReadUint32BE()
	if err != nil {
		return layer, err
	}
	if arrayLen > 0x10000 {
		return layer, errors.New("sanity check: array len too long")
	}

	layer = NewPacket83_0BLayer(int(arrayLen))

	gzipStream, err := thisBitstream.RegionToGZipStream()
	if err != nil {
		return layer, err
	}

	var i uint32
	for i = 0; i < arrayLen; i++ {
		layer.Instances[i], err = DecodeReplicationInstance(true, gzipStream, context, packet, instanceSchema)
		if err != nil {
			return layer, err
		}

		gzipStream.Align()
	}
	return layer, nil
}
