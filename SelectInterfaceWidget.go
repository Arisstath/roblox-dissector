package main

import (
	"strings"

	"github.com/google/gopacket/pcap"
	"github.com/therecipe/qt/core"
	"github.com/therecipe/qt/gui"
	"github.com/therecipe/qt/widgets"
)

func NewSelectInterfaceWidget(parent widgets.QWidget_ITF, callback func(string, bool)) error {
	window := widgets.NewQWidget(parent, core.Qt__Window)
	window.SetWindowTitle("Choose network interface")

	layout := NewTopAlignLayout()
	usePromisc := widgets.NewQCheckBox2("Use promiscuous mode", nil)
	layout.AddWidget(usePromisc, 0, 0)

	itfLabel := NewLabel("Interface:")
	layout.AddWidget(itfLabel, 0, 0)

	interfaces := widgets.NewQTreeView(nil)

	standardModel := NewProperSortModel(interfaces)
	standardModel.SetHorizontalHeaderLabels([]string{"Interface Name", "Description", "IP addresses"})
	rootNode := standardModel.InvisibleRootItem()

	devs, err := pcap.FindAllDevs()
	if err != nil {
		return err
	}

	for _, dev := range devs {
		var addrStringBuilder strings.Builder
		for _, addr := range dev.Addresses {
			addrStringBuilder.WriteString(addr.IP.String())
			addrStringBuilder.WriteString(", ")
		}
		var addrs string
		if addrStringBuilder.Len() > 2 {
			// Remove trailing comma
			addrs = addrStringBuilder.String()[:addrStringBuilder.Len() - 2]
		}
		rootNode.AppendRow([]*gui.QStandardItem{
			NewStringItem(dev.Name),
			NewStringItem(dev.Description),
			// Remove trailing comma
			NewStringItem(addrs),
		})

	}

	interfaces.SetModel(standardModel)
	interfaces.SetSelectionMode(1)
	layout.AddWidget(interfaces, 0, 0)

	okButton := widgets.NewQPushButton2("Capture", nil)
	layout.AddWidget(okButton, 0, 0)
	okButton.ConnectReleased(func() {
		if len(interfaces.SelectedIndexes()) < 1 {
			window.Close()
			return
		}
		useInterface := standardModel.Item(interfaces.SelectedIndexes()[0].Row(), 0).Data(0).ToString()
		promisc := usePromisc.CheckState() == core.Qt__Checked
		window.Close()
		callback(useInterface, promisc)
	})

	window.SetLayout(layout)
	window.Show()

	return nil
}
