package main

import (
	"strings"

	"github.com/google/gopacket/pcap"
	"github.com/therecipe/qt/core"
	"github.com/therecipe/qt/gui"
	"github.com/therecipe/qt/widgets"
)

func NewSelectInterfaceWidget(parent widgets.QWidget_ITF, callback func(string, bool)) {
	window := widgets.NewQWidget(parent, core.Qt__Window)
	window.SetWindowTitle("Choose network interface")

	layout := NewTopAlignLayout()
	usePromisc := widgets.NewQCheckBox2("Use promiscuous mode", nil)
	layout.AddWidget(usePromisc, 0, 0)

	itfLabel := NewLabel("Interface:")
	layout.AddWidget(itfLabel, 0, 0)

	interfaces := widgets.NewQTreeView(nil)

	standardModel := NewProperSortModel(interfaces)
	standardModel.SetHorizontalHeaderLabels([]string{"Interface Name", "IP addresses"})
	rootNode := standardModel.InvisibleRootItem()

	devs, err := pcap.FindAllDevs()
	if err != nil {
		println("trying to get devs: " + err.Error())
		return
	}

	for _, dev := range devs {
		if len(dev.Addresses) < 1 {
			println("skip", dev.Name)
			continue
		}
		var addrStringBuilder strings.Builder
		for _, addr := range dev.Addresses {
			addrStringBuilder.WriteString(addr.IP.String())
			addrStringBuilder.WriteString(", ")
		}
		rootNode.AppendRow([]*gui.QStandardItem{
			NewStringItem(dev.Name),
			// Remove trailing comma
			NewStringItem(addrStringBuilder.String()[:addrStringBuilder.Len()-2]),
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
		window.Close()
		callback(useInterface, promisc)
	})

	window.SetLayout(layout)
	window.Show()
}
