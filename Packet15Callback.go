package main

import (
	"fmt"
	"net/http"

	"github.com/therecipe/qt/widgets"

	"github.com/Gskartwii/roblox-dissector/peer"
	"github.com/robloxapi/rbxapi/rbxapijson"
	"github.com/robloxapi/rbxapiref/fetch"
)

// Can't use https:// because the site is broken
const CDNURL = "http://setup.roblox.com/"

var latestRobloxAPI *rbxapijson.Root
var latestRobloxAPIChan chan struct{} // Closed when retrievement is done

func init() {
	latestRobloxAPIChan = make(chan struct{})
	go func() {
		defer func() {
			close(latestRobloxAPIChan)
		}()
		robloxApiClient := &fetch.Client{
			Client: &http.Client{},
			Config: fetch.Config{
				Builds:             []fetch.Location{fetch.NewLocation(CDNURL + "DeployHistory.txt")},
				Latest:             []fetch.Location{fetch.NewLocation(CDNURL + "versionQTStudio")},
				APIDump:            []fetch.Location{fetch.NewLocation(CDNURL + "$HASH-API-Dump.json")},
				ReflectionMetadata: []fetch.Location{fetch.NewLocation(CDNURL + "$HASH-RobloxStudio.zip#ReflectionMetadata.xml")},
				ExplorerIcons:      []fetch.Location{fetch.NewLocation(CDNURL + "$HASH-RobloxStudio.zip#RobloxStudioBeta.exe")},
			},
		}
		latestBuild, err := robloxApiClient.Latest()
		if err != nil {
			fmt.Println("Error retrieving API:", err.Error())
			return
		}
		apiDump, err := robloxApiClient.APIDump(latestBuild.Hash)
		if err != nil {
			fmt.Println("Error retrieving API:", err.Error())
			return
		}
		latestRobloxAPI = apiDump
	}()
}

func ShowPacket15(layerLayout *widgets.QVBoxLayout, context *peer.CommunicationContext, layers *peer.PacketLayers) {
	MainLayer := layers.Main.(*peer.Packet15Layer)

	var reasonLabel *widgets.QLabel
	reasonLabel = NewLabel("")
	reasonSuccess := func(r string) {
		reasonLabel.SetText("Disconnection reason: " + r)
	}
	reasonFail := func(r string) {
		reasonLabel.SetText("Failed to fetch reason: " + r)
	}
	if MainLayer.Reason == -1 {
		reasonSuccess("Generic disconnection -1")
	} else {
		reasonLabel.SetText("Fetching API...")
		go func() {
			wait := true
			for wait {
				_, wait = <-latestRobloxAPIChan
			}
			if latestRobloxAPI == nil {
				reasonFail("Roblox API not available")
				return
			}
			disconnectionEnum := latestRobloxAPI.GetEnum("ConnectionError")
			if disconnectionEnum == nil {
				reasonFail("ConnectionError Enum not available")
				return
			}
			items := disconnectionEnum.GetEnumItems()
			for _, item := range items {
				if item.GetValue() == int(MainLayer.Reason) {
					reasonSuccess(item.GetName())
					return
				}
			}
			reasonFail(fmt.Sprintf("Unknown disconnection %d", MainLayer.Reason))
		}()
	}
	layerLayout.AddWidget(reasonLabel, 0, 0)
}
