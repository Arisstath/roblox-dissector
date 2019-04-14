package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"

	"github.com/robloxapi/rbxfile"
	"github.com/therecipe/qt/core"
	"github.com/therecipe/qt/gui"
	"github.com/therecipe/qt/widgets"
)

// Cheating
type mainThreadHelper struct {
	core.QObject

	_    func()       `constructor:"init"`
	_    func(func()) `signal:"runOnMain"`
	Wait chan struct{}
}

func (helper *mainThreadHelper) init() {
	// send doesn't need to block, only receive
	// that's why we can make the cap 1
	// this also helps prevent deadlocks
	helper.Wait = make(chan struct{}, 1)
}

func (helper *mainThreadHelper) runOnMain(f func()) {
	f()
	helper.Wait <- struct{}{}
}

var MainThreadRunner = NewMainThreadHelper(nil)

func init() {
	MainThreadRunner.ConnectRunOnMain(MainThreadRunner.runOnMain)
}

type DefaultValues map[string](map[string]rbxfile.Value)

func NewTopAlignLayout() *widgets.QVBoxLayout {
	layout := widgets.NewQVBoxLayout()
	layout.SetAlign(core.Qt__AlignTop)
	return layout
}
func NewLabel(content string) *widgets.QLabel {
	return widgets.NewQLabel2(content, nil, 0)
}
func NewQLabelF(format string, args ...interface{}) *widgets.QLabel {
	return widgets.NewQLabel2(fmt.Sprintf(format, args...), nil, 0)
}

func NewStringItem(content string) *gui.QStandardItem {
	ret := gui.NewQStandardItem2(content)
	ret.SetEditable(false)
	return ret
}
func NewIntItem(content interface{}) *gui.QStandardItem {
	var normalized int
	switch content.(type) {
	case int:
		normalized = content.(int)
	case int8:
		normalized = int(content.(int8))
	case int16:
		normalized = int(content.(int16))
	case int32:
		normalized = int(content.(int32))
	case int64:
		normalized = int(content.(int64))
	case uint:
		normalized = int(content.(uint))
	case uint8:
		normalized = int(content.(uint8))
	case uint16:
		normalized = int(content.(uint16))
	case uint32:
		normalized = int(content.(uint32))
	case uint64:
		normalized = int(content.(uint64))
	}

	ret := gui.NewQStandardItem()
	ret.SetData(core.NewQVariant7(normalized), 0)
	ret.SetEditable(false)
	return ret
}
func NewUintItem(content interface{}) *gui.QStandardItem {
	var normalized uint
	switch content.(type) {
	case int:
		normalized = uint(content.(int))
	case int8:
		normalized = uint(content.(int8))
	case int16:
		normalized = uint(content.(int16))
	case int32:
		normalized = uint(content.(int32))
	case int64:
		normalized = uint(content.(int64))
	case uint:
		normalized = content.(uint)
	case uint8:
		normalized = uint(content.(uint8))
	case uint16:
		normalized = uint(content.(uint16))
	case uint32:
		normalized = uint(content.(uint32))
	case uint64:
		normalized = uint(content.(uint64))
	}

	ret := gui.NewQStandardItem()
	ret.SetData(core.NewQVariant8(normalized), 0)
	ret.SetEditable(false)
	return ret
}
func NewQStandardItemF(format string, args ...interface{}) *gui.QStandardItem {
	ret := gui.NewQStandardItem2(fmt.Sprintf(format, args...))
	ret.SetEditable(false)
	return ret
}

func paintItems(row []*gui.QStandardItem, color *gui.QColor) {
	for i := 0; i < len(row); i++ {
		row[i].SetBackground(gui.NewQBrush3(color, core.Qt__SolidPattern))
	}
}

func GUIMain() {
	widgets.NewQApplication(len(os.Args), os.Args)
	window := NewDissectorWindow(nil, 0)
	window.ShowMaximized()

	joinFlag := flag.String("join", "", "roblox-dissector:<authTicket>:<placeID>:<browserTrackerID>")
	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()
	flag.Parse()
	if *joinFlag != "" {
		println("Received protocol invocation?")
		protocolRegex := regexp.MustCompile(`roblox-dissector:([0-9A-Fa-f]+):(\d+):(\d+)`)
		uri := *joinFlag
		parts := protocolRegex.FindStringSubmatch(uri)
		if len(parts) < 4 {
			widgets.QMessageBox_Critical(window, "Invalid protocol invocation", "Invalid protocol invocation: "+os.Args[1], widgets.QMessageBox__Ok, widgets.QMessageBox__NoButton)
		} else {
			authTicket := parts[1]
			placeID, _ := strconv.Atoi(parts[2])
			browserTrackerId, _ := strconv.Atoi(parts[3])

			window.StartClient(uint32(placeID), uint64(browserTrackerId), authTicket)
		}
	}
	openFile := flag.Arg(0)
	// React to command line arg
	if openFile != "" {
		window.CaptureFromFile(openFile, false)
	}

	/*toolsBar := window.MenuBar().AddMenu2("&Tools")

	scriptDumperAction := toolsBar.AddAction("Dump &scripts")
	scriptDumperAction.ConnectTriggered(func(checked bool) {
		dumpScripts(packetViewer.Context.DataModel.ToRbxfile().Instances, 0)
		scriptData, err := os.OpenFile("dumps/scriptKeys", os.O_RDWR|os.O_CREATE, 0666)
		defer scriptData.Close()
		if err != nil {
			println("while dumping script keys:", err.Error())
			return
		}

		_, err = fmt.Fprintf(scriptData, "Int 1: %d\nInt 2: %d", packetViewer.Context.Int1, packetViewer.Context.Int2)
		if err != nil {
			println("while dumping script keys:", err.Error())
			return
		}
	})
	dumperAction := toolsBar.AddAction("&DataModel dumper lite...")
	dumperAction.ConnectTriggered(func(checked bool) {
		location := widgets.QFileDialog_GetSaveFileName(packetViewer, "Save as RBXL...", "", "Roblox place files (*.rbxl)", "", 0)
		writer, err := os.OpenFile(location, os.O_RDWR|os.O_CREATE, 0666)
		if err != nil {
			println("while opening file:", err.Error())
			return
		}

		writableClone := packetViewer.Context.DataModel.ToRbxfile()

		dumpScripts(writableClone.Instances, 0)

		err = xml.Serialize(writer, nil, writableClone)
		if err != nil {
			println("while serializing place:", err.Error())
			return
		}
	})
	browseAction := toolsBar.AddAction("&Browse DataModel...")
	browseAction.ConnectTriggered(func(checked bool) {
		if packetViewer.Context != nil {
			NewDataModelBrowser(packetViewer.Context, packetViewer.Context.DataModel, packetViewer.DefaultValues)
		}
	})

	readDefaults := toolsBar.AddAction("Parse &default values...")
	readDefaults.ConnectTriggered(func(checked bool) {
		NewFindDefaultsWidget(window, packetViewer.DefaultsSettings, func(settings *DefaultsSettings) {
			packetViewer.DefaultValues = ParseDefaultValues(settings.Files)
		})
	})

	/*viewCache := toolsBar.AddAction("&View string cache...")
	viewCache.ConnectTriggered(func(checked bool) {
		NewViewCacheWidget(packetViewer, packetViewer.Context)
	})*

	injectChat := toolsBar.AddAction("Inject &chat message...")
	injectChat.ConnectTriggered(func(checked bool) {
		if packetViewer.Context == nil {
			println("context is nil!")
			return
		} else if packetViewer.Context.DataModel == nil {
			println("datamodel instances is nil!")
			return
		}

		dataModel := packetViewer.Context.DataModel.Instances
		var players, replicatedStorage *datamodel.Instance
		for i := 0; i < len(dataModel); i++ {
			if dataModel[i].ClassName == "Players" {
				players = dataModel[i]
			} else if dataModel[i].ClassName == "ReplicatedStorage" {
				replicatedStorage = dataModel[i]
			}
		}
		player := players.Children[0]
		println("chose player", player.Name())
		chatEvent := replicatedStorage.FindFirstChild("DefaultChatSystemChatEvents").FindFirstChild("SayMessageRequest")
		subpacket := &peer.Packet83_07{
			Instance: chatEvent,
			Schema:   packetViewer.Context.StaticSchema.SchemaForClass("RemoteEvent").SchemaForEvent("OnServerEvent"),
			Event: &peer.ReplicationEvent{
				Arguments: []rbxfile.Value{
					datamodel.ValueReference{Instance: player, Reference: player.Ref},
					datamodel.ValueTuple{
						rbxfile.ValueString("Hello, this is a hacked message"),
						rbxfile.ValueString("All"),
					},
				},
			},
		}

		packetViewer.InjectPacket <- &peer.Packet83Layer{
			SubPackets: []peer.Packet83Subpacket{subpacket},
		}
	})

	peersBar := window.MenuBar().AddMenu2("&Peers...")
	startSelfServer := peersBar.AddAction("Start self &server...")
	startSelfClient := peersBar.AddAction("Start self &client...")
	startSelfServer.ConnectTriggered(func(checked bool) {
		NewServerStartWidget(window, packetViewer.ServerSettings, func(settings *ServerSettings) {
			port, _ := strconv.Atoi(settings.Port)
			enums, err := os.Open(settings.EnumSchemaLocation)
			if err != nil {
				println("while parsing schema:", err.Error())
				return
			}
			instances, err := os.Open(settings.InstanceSchemaLocation)
			if err != nil {
				println("while parsing schema:", err.Error())
				return
			}
			schema, err := peer.ParseSchema(instances, enums)
			if err != nil {
				println("while parsing schema:", err.Error())
				return
			}
			dataModelReader, err := os.Open(settings.RBXLLocation)
			if err != nil {
				println("while reading instances:", err.Error())
				return
			}
			dataModelRoot, err := xml.Deserialize(dataModelReader, nil)
			if err != nil {
				println("while reading instances:", err.Error())
				return
			}

			instanceDictionary := datamodel.NewInstanceDictionary()
			thisRoot := datamodel.FromRbxfile(instanceDictionary, dataModelRoot)
			normalizeTypes(thisRoot.Instances, &schema)

			server, err := peer.NewCustomServer(uint16(port), &schema, thisRoot)
			if err != nil {
				println("while creating server", err.Error())
				return
			}
			server.InstanceDictionary = instanceDictionary
			server.Context.InstancesByReference.Populate(thisRoot.Instances)

			NewServerConsole(window, server)

			go server.Start()
		})
	})
	startSelfClient.ConnectTriggered(func(checked bool) {
		customClient := peer.NewCustomClient()
		NewClientStartWidget(window, customClient, func(placeId uint32, username string, password string) {
			NewClientConsole(window, customClient)
			customClient.SecuritySettings = peer.Win10Settings()
			// No more guests! Roblox won't let us connect as one.
			go func() {
				ticket, err := GetAuthTicket(username, password)
				if err != nil {
					widgets.QMessageBox_Critical(window, "Failed to start client", "While getting authticket: "+err.Error(), widgets.QMessageBox__Ok, widgets.QMessageBox__NoButton)
				} else {
					customClient.ConnectWithAuthTicket(placeId, ticket)
				}
			}()
		})
	})*/

	widgets.QApplication_Exec()
}
