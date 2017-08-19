package main
import "golang.org/x/sys/windows/registry"
import "github.com/therecipe/qt/widgets"
import "github.com/therecipe/qt/gui"
import "github.com/therecipe/qt/core"

const BaseKey = `SYSTEM\CurrentControlSet\Services\Tcpip\Parameters\Interfaces`

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
	standardModel.SetHorizontalHeaderLabels([]string{"Interface GUID", "IP address"})
	rootNode := standardModel.InvisibleRootItem()

	itfList, err := registry.OpenKey(registry.LOCAL_MACHINE, BaseKey, registry.READ)
	if err != nil {
		println("trying to open base key: " + err.Error())
		return
	}

	keyNames, err := itfList.ReadSubKeyNames(-1) 
	if err != nil {
		println("trying to read sub key names: " + err.Error())
		return
	}
	for _, SubKeyName := range keyNames {
		thisItf, err := registry.OpenKey(registry.LOCAL_MACHINE, BaseKey + `\` + SubKeyName, registry.QUERY_VALUE)
		if err != nil {
			println("trying to open sub key: " + err.Error())
			return
		}
		ipAddr, _, err := thisItf.GetStringValue("DhcpIPAddress")
		if err != nil {
			if err.Error() == "The system cannot find the file specified." {
				continue
			}
			println("trying to get sub key content: " + err.Error())
			return
		}
		rootNode.AppendRow([]*gui.QStandardItem{
			NewQStandardItemF(SubKeyName),
			NewQStandardItemF(ipAddr),
		})

		thisItf.Close()
	}
	itfList.Close()

	interfaces.SetModel(standardModel)
	interfaces.SetSelectionMode(1)
	layout.AddWidget(interfaces, 0, 0)

	okButton := widgets.NewQPushButton2("Capture", nil)
	layout.AddWidget(okButton, 0, 0)
	okButton.ConnectPressed(func() {
		useInterface := standardModel.Item(interfaces.SelectedIndexes()[0].Row(), 0).Data(0).ToString()
		promisc := usePromisc.CheckState() == core.Qt__Checked
		window.Destroy(true, true)
		callback(`\Devices\NPF_` + useInterface, promisc)
	})

	window.SetLayout(layout)
	window.Show()
}
