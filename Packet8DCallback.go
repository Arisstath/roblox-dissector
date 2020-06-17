package main

import (
	"github.com/Gskartwii/roblox-dissector/peer"
	"github.com/therecipe/qt/gui"
	"github.com/therecipe/qt/widgets"
)

func ShowPacket8D(layerLayout *widgets.QVBoxLayout, context *peer.CommunicationContext, layers *peer.PacketLayers) {
	MainLayer := layers.Main.(*peer.Packet8DLayer)

	labelForSubpackets := NewQLabelF("Terrain cluster (%d chunks):", len(MainLayer.Chunks))
	layerLayout.AddWidget(labelForSubpackets, 0, 0)

	chunkList := widgets.NewQTreeView(nil)
	standardModel := NewProperSortModel(nil)
	standardModel.SetHorizontalHeaderLabels([]string{"Index", "Dimensions", "Int 1"})

	rootNode := standardModel.InvisibleRootItem()
	for _, chunk := range MainLayer.Chunks {
		indexItem := NewStringItem("Chunk: " + chunk.ChunkIndex.String())
		countItem := NewQStandardItemF("%d x %d x %d", chunk.SideLength, chunk.SideLength, chunk.SideLength)
		int1Item := NewUintItem(chunk.Int1)

		rootNode.AppendRow([]*gui.QStandardItem{indexItem, countItem, int1Item})
	}

	chunkList.SetModel(standardModel)
	chunkList.SetSelectionMode(0)
	chunkList.SetSortingEnabled(true)
	layerLayout.AddWidget(chunkList, 0, 0)
}
