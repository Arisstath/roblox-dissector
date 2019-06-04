package main

import (
	"fmt"

	"github.com/Gskartwii/roblox-dissector/peer"
	"github.com/therecipe/qt/gui"
	"github.com/therecipe/qt/widgets"
)

// TODO: Dynamically fetch?
var TerrainMaterials = [...]string{
	"Air",
}

func MaterialToString(material uint8) string {
	if len(TerrainMaterials) > int(material) {
		return TerrainMaterials[material]
	}

	return fmt.Sprintf("0x%02X", material)
}

func ShowPacket8D(layerLayout *widgets.QVBoxLayout, context *peer.CommunicationContext, layers *peer.PacketLayers) {
	MainLayer := layers.Main.(*peer.Packet8DLayer)

	labelForSubpackets := NewQLabelF("Terrain cluster (%d chunks):", len(MainLayer.Chunks))
	layerLayout.AddWidget(labelForSubpackets, 0, 0)

	chunkList := widgets.NewQTreeView(nil)
	standardModel := NewProperSortModel(nil)
	standardModel.SetHorizontalHeaderLabels([]string{"Index", "Dimensions"})

	rootNode := standardModel.InvisibleRootItem()
	for _, chunk := range MainLayer.Chunks {
		indexItem := NewStringItem("Chunk: " + chunk.ChunkIndex.String())
		countItem := NewQStandardItemF("%d x %d x %d", chunk.SideLength, chunk.SideLength, chunk.SideLength)

		rootNode.AppendRow([]*gui.QStandardItem{indexItem, countItem})
	}

	chunkList.SetModel(standardModel)
	chunkList.SetSelectionMode(0)
	chunkList.SetSortingEnabled(true)
	layerLayout.AddWidget(chunkList, 0, 0)
}
