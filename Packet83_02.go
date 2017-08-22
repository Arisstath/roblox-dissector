package main
import "github.com/google/gopacket"
import "github.com/therecipe/qt/widgets"
import "github.com/gskartwii/rbxfile"

type Packet83_02 struct {
	child *rbxfile.Instance
}

func (this Packet83_02) Show() widgets.QWidget_ITF {
	instanceList := widgets.NewQTreeView(nil)
	standardModel := NewProperSortModel(nil)
	standardModel.SetHorizontalHeaderLabels([]string{"Name", "Type", "Value", "Referent", "Parent"})

	rootNode := standardModel.InvisibleRootItem()
	rootNode.AppendRow(getInstanceRow(this.child))
	instanceList.SetModel(standardModel)
	instanceList.SetSelectionMode(0)
	instanceList.SetSortingEnabled(true)

	return instanceList
}

func DecodePacket83_02(thisBitstream *ExtendedReader, context *CommunicationContext, packet gopacket.Packet, instanceSchema []*InstanceSchemaItem) (interface{}, error) {
	result, err := DecodeReplicationInstance(false, thisBitstream, context, packet, instanceSchema)
	return &Packet83_02{result}, err
}
