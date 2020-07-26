package main

import (
	"strings"

	"github.com/google/gopacket/pcap"
	"github.com/gotk3/gotk3/gtk"
)

func PromptInterfaceName(callback func(string)) error {
	builder, err := gtk.BuilderNewFromFile("res/interfaceselector.ui")
	if err != nil {
		return err
	}
	mainWidget_, err := builder.GetObject("maindialog")
	if err != nil {
		return err
	}
	mainWidget, ok := mainWidget_.(*gtk.Dialog)
	if !ok {
		return invalidUi("maindialog")
	}

	model_, err := builder.GetObject("interfaceliststore")
	if err != nil {
		return err
	}
	model, ok := model_.(*gtk.ListStore)
	if !ok {
		return invalidUi("liststore")
	}

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
			addrs = addrStringBuilder.String()[:addrStringBuilder.Len()-2]
		}

		model.InsertWithValues(nil, -1, []int{0, 1, 2}, []interface{}{dev.Name, dev.Description, addrs})
	}

	cancelButton_, err := builder.GetObject("cancelbutton")
	if err != nil {
		return err
	}
	cancelButton, ok := cancelButton_.(*gtk.Button)
	if !ok {
		return invalidUi("cancelbutton")
	}
	cancelButton.Connect("clicked", func() {
		mainWidget.Hide()
		mainWidget.Destroy()
	})
	okButton_, err := builder.GetObject("okbutton")
	if err != nil {
		return err
	}
	okButton, ok := okButton_.(*gtk.Button)
	if !ok {
		return invalidUi("okbutton")
	}

	sel_, err := builder.GetObject("itfselection")
	if err != nil {
		return err
	}
	sel, ok := sel_.(*gtk.TreeSelection)
	if !ok {
		return invalidUi("sel")
	}
	okButton.Connect("clicked", func() {
		mainWidget.Hide()
		_, iter, ok := sel.GetSelected()
		if !ok {
			mainWidget.Destroy()
			return
		}
		name, err := model.GetValue(iter, 0)
		if err != nil {
			println("selection failed:", err.Error())
			return
		}
		nameG, err := name.GoValue()
		if err != nil {
			println("selection failed:", err.Error())
			return
		}
		callback(nameG.(string))
	})

	mainWidget.ShowAll()
	return nil
}
