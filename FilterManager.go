package main

/*
import "github.com/therecipe/qt/widgets"

// TODO: Perhaps subclass QListWidgetItem to set its data?
// And also do that for the main packet view and other stuff?
// What would that imply?
func NewFilterManager(parent widgets.QWidget_ITF, filters FilterSettings) {
	window := widgets.NewQWidget(parent, 1)
	window.SetWindowTitle("Manage packet view filters")
	layout := widgets.NewQVBoxLayout()

	ruleList := widgets.NewQListWidget(window)
	ruleList.SetSortingEnabled(false)
	for _, filter := range filters {
		ruleList.AddItem(filter.String())
	}

	layout.AddWidget(ruleList, 0, 0)

	newTypeFilter := widgets.NewQPushButton2("New type filter...", window)
	newDataFilter := widgets.NewQPushButton2("New data filter...", window)
	newDirectionFilter := widgets.NewQPushButton2("New direction filter...", window)
	removeFilterButton := widgets.NewQPushButton2("Remove", window)
	moveUpButton := widgets.NewQPushButton2("Up", window)
	moveDownButton := widgets.NewQPushButton2("Down", window)

	newTypeFilter.ConnectClicked(func(checked bool) {
		PickPacketType(PacketNames, func(picked uint8) {
			newItem := &BasicPacketTypeFilter{picked}
			newRule := &FilterRule{false, newItem}
			ruleList.AddItem(newRule.String())
			filters = append(filters, newRule)
		})
	})
}
*/
