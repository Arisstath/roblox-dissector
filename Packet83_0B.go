package main
import "github.com/google/gopacket"
import "github.com/therecipe/qt/widgets"
import "errors"

type Packet83_0B struct {
	Instances []*ReplicationInstance
}

func NewPacket83_0BLayer(length int) *Packet83_0B {
	return &Packet83_0B{make([]*ReplicationInstance, length)}
}

func (this Packet83_0B) Show() widgets.QWidget_ITF {
	instanceList := widgets.NewQTreeView(nil)
	standardModel := NewProperSortModel(nil)
	standardModel.SetHorizontalHeaderLabels([]string{"Name", "Type", "Value", "Referent", "Unknown bool", "Parent"})

	rootNode := standardModel.InvisibleRootItem()
	for _, instance := range(this.Instances) {
		rootNode.AppendRow(instance.Show())
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
