package peer

import (
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
	JoinDataConfig{"ReplicatedFirst", false, false},
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

func (client *ServerClient) authHandler(e *emitter.Event) {
	err := client.WritePacket(&Packet97Layer{
		Schema: client.Context.StaticSchema,
	})
	if err != nil {
		println("schema error: ", err.Error())
		return
	}

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
			thisContainer.UpdateBinding(client)
		}
	}

	err = client.WritePacket(&Packet81Layer{
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
	if err != nil {
		println("topreplic error: ", err.Error())
		return
	}

	err = client.WritePacket(&Packet84Layer{
		MarkerId: 1,
	})
	if err != nil {
		println("topreplic error: ", err.Error())
		return
	}

	player, _ := datamodel.NewInstance("Player", nil)
	player.Set("Name", rbxfile.ValueString("Player1"))
	player.Set("AccountAgeReplicate", rbxfile.ValueInt(-1712138672))
	player.Set("CharacterAppearanceId", rbxfile.ValueInt64(36537369)) // gskw
	player.Set("ChatPrivacyMode", datamodel.ValueToken{Value: 0})
	player.Set("ReplicatedLocaleId", rbxfile.ValueString("en-us"))
	player.Set("UserId", rbxfile.ValueInt64(-1))
	player.Set("userId", rbxfile.ValueInt64(-1))
	player.Ref = client.Server.InstanceDictionary.NewReference()

	playerGui, _ := datamodel.NewInstance("PlayerGui", nil)
	playerGui.Ref = client.Server.InstanceDictionary.NewReference()
	err = player.AddChild(playerGui)
	if err != nil {
		println("playergui error: ", err.Error())
		return
	}

	err = client.DataModel.FindService("Players").AddChild(player)
	if err != nil {
		println("player error: ", err.Error())
		return
	}

	// REPLICATION BEGIN
	// Do not include ReplicatedFirst itself in replication
	replicatedFirstStreamer := NewJoinDataStreamer(client.DefaultPacketWriter)
	replicatedFirstStreamer.BufferEmitter.On("join-data", func(e *emitter.Event) {
		err := client.WriteDataPackets(e.Args[0].(Packet83Subpacket))
		if err != nil {
			println("replicatedfirst error: ", err.Error())
		}
	}, emitter.Void)
	replicatedFirst := client.constructInstanceList(nil, client.DataModel.FindService("ReplicatedFirst"))
	for _, repFirstInstance := range replicatedFirst {
		err := replicatedFirstStreamer.AddInstance(repFirstInstance)
		if err != nil {
			println("repfirst join data error: ", err.Error())
			return
		}
	}
	err = replicatedFirstStreamer.Close()
	if err != nil {
		println("repfirst join data error: ", err.Error())
		return
	}
	// Tag: ReplicatedFirst finished!
	err = client.WriteDataPackets(&Packet83_10{
		TagId: 12,
	})
	if err != nil {
		println("tag error: ", err.Error())
		return
	}

	joinDataStreamer := NewJoinDataStreamer(client.DefaultPacketWriter)
	joinDataStreamer.BufferEmitter.On("join-data", func(e *emitter.Event) {
		err := client.WriteDataPackets(e.Args[0].(Packet83Subpacket))
		if err != nil {
			println("joindata error: ", err.Error())
		}
	}, emitter.Void)
	for _, dataConfig := range JoinDataConfiguration {
		service := client.DataModel.FindService(dataConfig.ClassName)
		if service != nil {
			println("Replicating service ", dataConfig.ClassName)
			err = client.ReplicateJoinData(service, dataConfig.ReplicateProperties, dataConfig.ReplicateChildren, joinDataStreamer, player)
			if err != nil {
				println("repfirst join data error: ", err.Error())
				return
			}
		}
	}
	err = joinDataStreamer.Close()
	if err != nil {
		println("join data error: ", err.Error())
		return
	}

	err = client.WriteDataPackets(&Packet83_10{
		TagId: 13,
	})
	if err != nil {
		println("tag error: ", err.Error())
		return
	}
	// REPLICATION END

	println("setting playername")
	player.Set("Name", rbxfile.ValueString("Player15"))
	println("set playername")
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
}
