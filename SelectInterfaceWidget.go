package main
import "github.com/therecipe/qt/widgets"
import "github.com/therecipe/qt/gui"
import "github.com/therecipe/qt/core"
import "github.com/google/gopacket/pcap"

func NewSelectInterfaceWidget(parent widgets.QWidget_ITF, callback func (string, bool)) {
	window := widgets.NewQWidget(parent, core.Qt__Window)
	window.SetWindowTitle("Choose network interface")

	layout := widgets.NewQVBoxLayout()
	usePromisc := widgets.NewQCheckBox2("Use promiscuous mode", nil)
	layout.AddWidget(usePromisc, 0, 0)

	itfLabel := NewQLabelF("Interface:")
	layout.AddWidget(itfLabel, 0, 0)

	interfaces := widgets.NewQTreeView(nil)

	standardModel := NewProperSortModel(interfaces)
	standardModel.SetHorizontalHeaderLabels([]string{"Interface Name", "IP address"})
	rootNode := standardModel.InvisibleRootItem()

    interfaceList, _ := pcap.FindAllDevs()
    for i := 0; i < len(interfaceList); i++ {
        if len(interfaceList[i].Addresses) < 1 {
            println("skip", interfaceList[i].Name)
            continue
        }
        ipAddr := interfaceList[i].Addresses[0].IP.String()
		rootNode.AppendRow([]*gui.QStandardItem{
			NewQStandardItemF(interfaceList[i].Name),
			NewQStandardItemF(ipAddr),
		})
	}

	interfaces.SetModel(standardModel)
	interfaces.SetSelectionMode(1)
	layout.AddWidget(interfaces, 0, 0)

	okButton := widgets.NewQPushButton2("Capture", nil)
	layout.AddWidget(okButton, 0, 0)
	okButton.ConnectPressed(func() {
		if len(interfaces.SelectedIndexes()) < 1 {
			return
		}
		useInterface := standardModel.Item(interfaces.SelectedIndexes()[0].Row(), 0).Data(0).ToString()
		promisc := usePromisc.CheckState() == core.Qt__Checked
		window.Destroy(true, true)
		callback(useInterface, promisc)
	})

	window.SetLayout(layout)
	window.Show()
}
