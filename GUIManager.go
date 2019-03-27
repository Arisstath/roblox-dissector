package main

import "github.com/therecipe/qt/widgets"
import "github.com/therecipe/qt/gui"
import "github.com/therecipe/qt/core"
import "github.com/robloxapi/rbxfile"

import "os"
import "fmt"
import "strconv"

// Cheating
type mainThreadHelper struct {
	core.QObject

	_    func()       `constructor:"init`
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
func NewQLabelF(format string, args ...interface{}) *widgets.QLabel {
	return widgets.NewQLabel2(fmt.Sprintf(format, args...), nil, 0)
}

func NewQStandardItemF(format string, args ...interface{}) *gui.QStandardItem {
	if format == "%d" {
		ret := gui.NewQStandardItem()
		i, _ := strconv.Atoi(fmt.Sprintf(format, args...)) // hack
		ret.SetData(core.NewQVariant7(i), 0)
		ret.SetEditable(false)
		return ret
	}
	ret := gui.NewQStandardItem2(fmt.Sprintf(format, args...))
	ret.SetEditable(false)
	return ret
}

func paintItems(row []*gui.QStandardItem, color *gui.QColor) {
	for i := 0; i < len(row); i++ {
		row[i].SetBackground(gui.NewQBrush3(color, core.Qt__SolidPattern))
	}
}

func GUIMain(openFile string) {
	widgets.NewQApplication(len(os.Args), os.Args)
	window := NewDissectorWindow(nil, 0)

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
			server.Context.InstancesByReferent.Populate(thisRoot.Instances)

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

	window.ShowMaximized()

	widgets.QApplication_Exec()
}
