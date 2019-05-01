package peer

import (
	"fmt"
	"net"
	"time"

	"github.com/gskartwii/roblox-dissector/datamodel"

	"github.com/olebedev/emitter"
	"github.com/robloxapi/rbxfile"
)

type JoinDataConfig struct {
	ClassName           string
	ReplicateProperties bool
	ReplicateChildren   bool
}

var JoinDataConfiguration = []JoinDataConfig{
	JoinDataConfig{"ReplicatedFirst", true, true},
	JoinDataConfig{"Lighting", true, true},
	JoinDataConfig{"SoundService", true, true},
	JoinDataConfig{"TeleportService", true, false},
	JoinDataConfig{"StarterPack", false, true},
	JoinDataConfig{"StarterGui", true, true},
	JoinDataConfig{"StarterPlayer", true, true},
	JoinDataConfig{"CSGDictionaryService", false, true},
	JoinDataConfig{"Workspace", true, true},
	JoinDataConfig{"JointsService", false, true},
	JoinDataConfig{"Players", true, true},
	JoinDataConfig{"Teams", false, true},
	JoinDataConfig{"InsertService", true, true},
	JoinDataConfig{"Chat", true, true},
	JoinDataConfig{"LocalizationService", true, true},
	JoinDataConfig{"FriendService", true, true},
	JoinDataConfig{"MarketplaceService", true, true},
	JoinDataConfig{"BadgeService", true, false},
	JoinDataConfig{"ReplicatedStorage", true, true},
	JoinDataConfig{"RobloxReplicatedStorage", true, true},
	JoinDataConfig{"TestService", true, true},
	JoinDataConfig{"LogService", true, false},
	JoinDataConfig{"PointsService", true, false},
	JoinDataConfig{"AdService", true, false},
	JoinDataConfig{"SocialService", true, false},
}

func (client *ServerClient) offline5Handler(e *emitter.Event) {
	println("Received connection!", client.Address.String())
	client.WriteOffline(&Packet06Layer{
		GUID:        client.Server.GUID,
		UseSecurity: false,
		MTU:         1492,
	})
}
func (client *ServerClient) offline7Handler(e *emitter.Event) {
	println("Received reply 7!", client.Address.String())
	client.WriteOffline(&Packet08Layer{
		GUID:      client.Server.GUID,
		IPAddress: client.Address,
		MTU:       1492,
	})
}
func (client *ServerClient) connectionRequestHandler(e *emitter.Event) {
	nullIP, _ := net.ResolveUDPAddr("udp", "0.0.0.0:0")
	println("received connection request!", client.Address.String())
	client.WritePacket(&Packet10Layer{
		IPAddress:   client.Address,
		SystemIndex: uint16(len(client.Server.Clients) - 1),
		Addresses: [10]*net.UDPAddr{
			client.Address,
			nullIP,
			nullIP,
			nullIP,
			nullIP,
			nullIP,
			nullIP,
			nullIP,
			nullIP,
			nullIP,
		},
		SendPingTime: e.Args[0].(*Packet09Layer).Timestamp,
		SendPongTime: uint64(time.Now().UnixNano() / int64(time.Millisecond)),
	})
}

func (client *ServerClient) requestParamsHandler(e *emitter.Event) {
	params := make(map[string]bool)

	for _, flag := range e.Args[0].(*Packet90Layer).RequestedFlags {
		value := false
		switch flag {
		case "FixDictionaryScopePlatformsReplication",
			"ReplicateInterpolateRelativeHumanoidPlatformsMotion",
			"FixRaysInWedges",
			"FixBallRaycasts",
			"UseNativePathWaypoint",
			"PgsForAll":
			value = true
		}
		params[flag] = value
	}

	client.WritePacket(&Packet93Layer{
		ProtocolSchemaSync:       false,
		ApiDictionaryCompression: false,
		Params:                   params,
	})
}

func (client *ServerClient) topReplicate() error {
	topReplicationItems := make([]*Packet81LayerItem, 0, len(JoinDataConfiguration))
	for _, instance := range JoinDataConfiguration {
		service := client.Context.DataModel.FindService(instance.ClassName)
		if service != nil {
			topReplicationItems = append(topReplicationItems, &Packet81LayerItem{
				Schema:        client.Context.StaticSchema.SchemaForClass(instance.ClassName),
				Instance:      service,
				WatchChanges:  instance.ReplicateProperties,
				WatchChildren: instance.ReplicateChildren,
			})

			thisContainer := &ReplicationContainer{
				Instance:            service,
				ReplicateProperties: instance.ReplicateProperties,
				ReplicateChildren:   instance.ReplicateChildren,

				// service parent should never change, but this will be
				// inherited by children
				ReplicateParent: true,
			}
			client.replicatedInstances = append(client.replicatedInstances, thisContainer)
			thisContainer.UpdateBinding(client, false)
		}
	}

	return client.WritePacket(&Packet81Layer{
		StreamJob:            false,
		FilteringEnabled:     true,
		AllowThirdPartySales: false,
		CharacterAutoSpawn:   true,
		ReferenceString:      client.Server.InstanceDictionary.Scope,
		// TODO: VM ints
		ScriptKey:     1,
		CoreScriptKey: 1,
		Items:         topReplicationItems,
	})
}

func (client *ServerClient) createCameraScript(parent *datamodel.Instance) error {
	cameraScript, _ := datamodel.NewInstance("LocalScript", nil)
	cameraScript.Set("Name", rbxfile.ValueString("SalaCamera"))
	cameraScript.Set("Source", rbxfile.ValueProtectedString(CameraScript))
	cameraScript.Ref = client.Server.InstanceDictionary.NewReference()

	return parent.AddChild(cameraScript)
}

func (client *ServerClient) createPlayer() error {
	player, _ := datamodel.NewInstance("Player", nil)
	player.Set("Name", rbxfile.ValueString(fmt.Sprintf("Player%d", client.Index)))
	player.Set("AccountAgeReplicate", rbxfile.ValueInt(117))
	player.Set("CharacterAppearanceId", rbxfile.ValueInt64(1))
	player.Set("ChatPrivacyMode", datamodel.ValueToken{Value: 0})
	player.Set("ReplicatedLocaleId", rbxfile.ValueString("en-us"))
	player.Set("UserId", rbxfile.ValueInt64(-client.Index))
	player.Set("userId", rbxfile.ValueInt64(-client.Index))
	player.Ref = client.Server.InstanceDictionary.NewReference()

	client.Player = player

	hum, _ := datamodel.NewInstance("Model", nil)
	hum.Set("Name", player.Get("Name"))
	hum.Ref = client.Server.InstanceDictionary.NewReference()
	err := client.DataModel.FindService("Workspace").AddChild(hum)
	if err != nil {
		return err
	}

	player.Set("Character", datamodel.ValueReference{
		Reference: hum.Ref,
		Instance:  hum,
	})

	playerGui, _ := datamodel.NewInstance("PlayerGui", nil)
	playerGui.Ref = client.Server.InstanceDictionary.NewReference()
	err = player.AddChild(playerGui)
	if err != nil {
		return err
	}

	return client.DataModel.FindService("Players").AddChild(player)
}

func (client *ServerClient) authHandler(e *emitter.Event) {
	err := client.WritePacket(&Packet97Layer{
		Schema: client.Context.StaticSchema,
	})
	if err != nil {
		println("schema error: ", err.Error())
		return
	}

	err = client.topReplicate()
	if err != nil {
		println("topreplic error: ", err.Error())
		return
	}

	err = client.WritePacket(&Packet84Layer{
		MarkerId: 1,
	})
	if err != nil {
		println("marker error: ", err.Error())
		return
	}

	err = client.createPlayer()
	if err != nil {
		println("player error: ", err.Error())
		return
	}

	err = client.sendReplicatedFirst()
	if err != nil {
		println("replicatedfirst error: ", err.Error())
		return
	}
	err = client.sendContainers()
	if err != nil {
		println("joindata error: ", err.Error())
		return
	}

	err = client.createCameraScript(client.Player.FindFirstChild("PlayerGui"))
	if err != nil {
		println("camera error: ", err.Error())
		return
	}
}

func (client *ServerClient) bindDefaultHandlers() {
	client.DefaultPacketReader.LayerEmitter.On("reliability", client.defaultReliabilityLayerHandler, emitter.Void)
	// TODO: Error handling?
	pEmitter := client.PacketEmitter
	pEmitter.On("ID_OPEN_CONNECTION_REQUEST_1", client.offline5Handler, emitter.Void)
	pEmitter.On("ID_OPEN_CONNECTION_REQUEST_2", client.offline7Handler, emitter.Void)
	pEmitter.On("ID_CONNECTION_REQUEST", client.connectionRequestHandler, emitter.Void)
	pEmitter.On("ID_PROTOCOL_SYNC", client.requestParamsHandler, emitter.Void)
	pEmitter.On("ID_SUBMIT_TICKET", client.authHandler, emitter.Void)

	client.PacketLogicHandler.bindDefaultHandlers()
}
