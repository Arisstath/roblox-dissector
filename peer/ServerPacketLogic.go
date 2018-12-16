package peer

import "net"
import "time"

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
	client.WritePacket(&Packet97Layer{
		Schema: *client.Context.StaticSchema,
	})

	topReplicationItems := make([]*Packet81LayerItem, len(client.Context.DataModel.Instances))
	for i, instance := range client.Context.DataModel.Instances {
		topReplicationItems[i] = &Packet81LayerItem{
			ClassID:  uint16(client.Context.StaticSchema.ClassesByName[instance.ClassName]),
			Instance: instance,
			Bool1:    false,
			Bool2:    false,
		}
	}

	client.WritePacket(&Packet81Layer{
		StreamJob:            false,
		FilteringEnabled:     true,
		AllowThirdPartySales: false,
		CharacterAutoSpawn:   true,
		ReferentString:       client.Server.Scope,
		// TODO: VM ints
		Int1:  1,
		Int2:  1,
		Items: topReplicationItems,
	})
}

func (client *ServerClient) bindDefaultHandlers() {
	client.RegisterPacketHandler(0x05, client.simple5Handler)
	client.RegisterPacketHandler(0x07, client.simple7Handler)
	client.RegisterPacketHandler(0x09, client.connectionRequestHandler)
	client.RegisterPacketHandler(0x90, client.requestParamsHandler)
	client.RegisterPacketHandler(0x8A, client.authHandler)
}
