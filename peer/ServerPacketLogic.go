package peer

import (
	"net"
	"time"

	"github.com/gskartwii/roblox-dissector/datamodel"

	"github.com/robloxapi/rbxfile"
)

type JoinDataConfig struct {
	ClassName           string
	ReplicateProperties bool
	ReplicateChildren   bool
}

var joinDataConfiguration = []JoinDataConfig{
	JoinDataConfig{"ReplicatedFirst", false, false}, // Replicated separately
	JoinDataConfig{"Lighting", true, true},
	JoinDataConfig{"SoundService", true, true},
	JoinDataConfig{"StarterPack", false, true},
	JoinDataConfig{"StarterGui", true, true},
	//JoinDataConfig{"StarterPlayer", true, true},
	JoinDataConfig{"CSGDictionaryService", false, true},
	JoinDataConfig{"Workspace", true, true},
	JoinDataConfig{"JointsService", false, true},
	JoinDataConfig{"Teams", false, true},
	JoinDataConfig{"InsertService", true, true},
	JoinDataConfig{"Chat", true, true},
	JoinDataConfig{"FriendService", true, true},
	JoinDataConfig{"MarketplaceService", true, true},
	JoinDataConfig{"BadgeService", true, false},
	//JoinDataConfig{"ReplicatedStorage", true, true},
	JoinDataConfig{"RobloxReplicatedStorage", true, true},
	JoinDataConfig{"TestService", true, true},
	JoinDataConfig{"LogService", true, false},
	JoinDataConfig{"PointsService", true, false},
	JoinDataConfig{"AdService", true, false},
	JoinDataConfig{"TeleportService", true, false},
	JoinDataConfig{"LocalizationService", true, true},
}

func (client *ServerClient) simple5Handler(packetType byte, layers *PacketLayers) {
	println("Received connection!", client.Address.String())
	client.WriteSimple(&Packet06Layer{
		GUID:        client.Server.GUID,
		UseSecurity: false,
		MTU:         1492,
	})
}
func (client *ServerClient) simple7Handler(packetType byte, layers *PacketLayers) {
	println("Received reply 7!", client.Address.String())
	client.WriteSimple(&Packet08Layer{
		GUID:      client.Server.GUID,
		IPAddress: client.Address,
		MTU:       1492,
	})
}
func (client *ServerClient) connectionRequestHandler(packetType byte, layers *PacketLayers) {
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
		SendPingTime: layers.Main.(*Packet09Layer).Timestamp,
		SendPongTime: uint64(time.Now().UnixNano() / int64(time.Millisecond)),
	})
}

func (client *ServerClient) requestParamsHandler(packetType byte, layers *PacketLayers) {
	params := make(map[string]bool)

	for _, flag := range layers.Main.(*Packet90Layer).RequestedFlags {
		value := false
		switch flag {
		case "UseNativePathWaypoint",
			"ReplicationInterpolateRelativeHumanoidPlatformsMotion",
			"FixRaysInWedges",
			"PartMasslessEnabled",
			"KeepRedundantWeldsExplicit",
			"FixBallRaycasts",
			"EnableRootPriority",
			"FixHats":
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

func (client *ServerClient) authHandler(packetType byte, layers *PacketLayers) {
	_, err := client.WritePacket(&Packet97Layer{
		Schema: *client.Context.StaticSchema,
	})
	if err != nil {
		println("schema error: ", err.Error())
		return
	}

	topReplicationItems := make([]*Packet81LayerItem, len(client.Context.DataModel.Instances))
	for i, instance := range client.Context.DataModel.Instances {
		topReplicationItems[i] = &Packet81LayerItem{
			ClassID:  uint16(client.Context.StaticSchema.ClassesByName[instance.ClassName]),
			Instance: instance,
			Bool1:    false,
			Bool2:    false,
		}
	}

	_, err = client.WritePacket(&Packet81Layer{
		StreamJob:            false,
		FilteringEnabled:     true,
		AllowThirdPartySales: false,
		CharacterAutoSpawn:   true,
		ReferentString:       client.Server.InstanceDictionary.Scope,
		// TODO: VM ints
		Int1:  1,
		Int2:  1,
		Items: topReplicationItems,
	})
	if err != nil {
		println("topreplic error: ", err.Error())
		return
	}

	partTest, _ := datamodel.NewInstance("NumberValue", nil)
	partTest.Set("Value", rbxfile.ValueDouble(3.0))
	partTest.Ref = client.Server.InstanceDictionary.NewReference()
	err = client.DataModel.FindService("Workspace").AddChild(partTest)
	if err != nil {
		println("parttest error: ", err.Error())
		return
	}

	err = client.WriteDataPackets(&Packet83_02{partTest})
	if err != nil {
		println("parttest error: ", err.Error())
		return
	}

	// REPLICATION BEGIN
	// Do not include ReplicatedFirst itself in replication
	replicatedFirst := constructInstanceList(nil, client.DataModel.FindService("ReplicatedFirst"))
	newInstanceList := make([]Packet83Subpacket, 0, len(replicatedFirst))
	for _, repFirstInstance := range replicatedFirst {
		newInstanceList = append(newInstanceList, &Packet83_02{Child: repFirstInstance})
	}
	err = client.WriteDataPackets(newInstanceList...)
	if err != nil {
		println("replicfirst error: ", err.Error())
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
	// For now, we will not limit the size of JoinData
	// This may become a problem later on with larger places
	for _, dataConfig := range joinDataConfiguration {
		service := client.DataModel.FindService(dataConfig.ClassName)
		if service != nil {
			println("Replicating service ", dataConfig.ClassName)
			err = client.ReplicateJoinData(service, dataConfig.ReplicateProperties, dataConfig.ReplicateChildren)
			if err != nil {
				println("replicate join data error: ", err.Error())
				return
			}
		}
	}
	err = client.WriteDataPackets(&Packet83_10{
		TagId: 13,
	})
	if err != nil {
		println("tag error: ", err.Error())
		return
	}
	// REPLICATION END
}

func (client *ServerClient) bindDefaultHandlers() {
	client.RegisterPacketHandler(0x05, client.simple5Handler)
	client.RegisterPacketHandler(0x07, client.simple7Handler)
	client.RegisterPacketHandler(0x09, client.connectionRequestHandler)
	client.RegisterPacketHandler(0x90, client.requestParamsHandler)
	client.RegisterPacketHandler(0x8A, client.authHandler)
}
