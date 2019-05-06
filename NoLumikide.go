// +build !lumikide

package main

import (
	"github.com/Gskartwii/roblox-dissector/peer"
	"github.com/therecipe/qt/widgets"
)

const LumikideEnabled = false

func LumikideProcessContext(parent widgets.QWidget_ITF, context *peer.CommunicationContext, rbxfileDataModel *rbxfile.Root) error {
	widgets.QMessageBox_Critical(parent, "Lumikide disabled", "Lumikide disabled at compile time, sorry!", widgets.QMessageBox__Ok, widgets.QMessageBox__Ok)
}
