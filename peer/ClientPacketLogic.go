package peer

import "github.com/gskartwii/rbxfile"
import "time"
import "net"
import "strconv"
import "errors"

func (myClient *CustomClient) FindService(name string) *rbxfile.Instance {
	if myClient.Context == nil || myClient.Context.DataModel == nil {
		return nil
	}
	for _, service := range myClient.Context.DataModel.Instances {
		if service.ClassName == name {
			return service
		}
	}
	return nil
}

func (myClient *CustomClient) WaitForInstance(path ...string) <-chan *rbxfile.Instance { // returned channels are output only
	// chan needs to be buffered if we need to immediately write to it
	retChannel := make(chan *rbxfile.Instance, 1)
	currInstance := myClient.FindService(path[0])
	if currInstance != nil {
		for i := 1; i < len(path); i++ {
			currInstance = currInstance.FindFirstChild(path[i], false)
			if currInstance == nil {
				break
			}
		}
	}
	if currInstance == nil {
		myClient.instanceHandlers.Bind(&InstancePath{path}, func(inst *rbxfile.Instance) {
			retChannel <- inst
		})
	} else { // immediately write if it exists
		retChannel <- currInstance
	}
	return retChannel
}

func (myClient *CustomClient) MakeEventChan(instance *rbxfile.Instance, name string) (int, chan *ReplicationEvent) {
	newChan := make(chan *ReplicationEvent)
	unbind := myClient.eventHandlers.Bind(instance, name, func(evt *ReplicationEvent) {
		newChan <- evt
	})
	return unbind, newChan
}

func (myClient *CustomClient) bindDefaultHandlers() {
	basicHandlers := myClient.handlers
	basicHandlers.Bind(6, myClient.simple6Handler)
	basicHandlers.Bind(8, myClient.simple8Handler)
	basicHandlers.Bind(0x10, myClient.packet10Handler)
	basicHandlers.Bind(0x15, myClient.disconnectHandler)
	basicHandlers.Bind(0x81, myClient.topReplicationHandler)
	basicHandlers.Bind(0x83, myClient.dataHandler)

	dataHandlers := myClient.dataHandlers
	dataHandlers.Bind(1, myClient.deleteHandler)
	dataHandlers.Bind(2, myClient.newInstanceHandler)
	dataHandlers.Bind(5, myClient.dataPingHandler)
	dataHandlers.Bind(7, myClient.eventHandler)
	dataHandlers.Bind(9, myClient.idChallengeHandler)
	dataHandlers.Bind(0xB, myClient.joinDataHandler)

	instHandlers := myClient.instanceHandlers
	instHandlers.Bind(&InstancePath{[]string{"Players"}}, myClient.handlePlayersService)
}

func (myClient *CustomClient) sendResponse7() {
	myClient.WriteSimple(&Packet07Layer{
		GUID:      myClient.GUID,
		MTU:       1492,
		IPAddress: &myClient.ServerAddress,
	})
}
func (myClient *CustomClient) simple6Handler(packetType byte, packet *UDPPacket, layers *PacketLayers) {
	myClient.Connected = true
	myClient.sendResponse7()
}

// transition to teal RakNet communication, no more offline messaging
func (myClient *CustomClient) sendResponse9() {
	response := &Packet09Layer{
		GUID:        myClient.GUID,
		Timestamp:   uint64(time.Now().Unix()),
		UseSecurity: false,
		Password:    []byte{0x37, 0x4F, 0x5E, 0x11, 0x6C, 0x45},
	}
	myClient.WritePacket(response)
}
func (myClient *CustomClient) simple8Handler(packetType byte, packet *UDPPacket, layers *PacketLayers) {
	myClient.sendResponse9()
}

func (myClient *CustomClient) sendPong(pingTime uint64) {
	response := &Packet03Layer{
		SendPingTime: pingTime,
		SendPongTime: uint64(time.Now().Unix()),
	}

	myClient.WritePacket(response)
}
func (myClient *CustomClient) pingHandler(packetType byte, packet *UDPPacket, layers *PacketLayers) {
	mainLayer := layers.Main.(*Packet00Layer)

	myClient.sendPong(mainLayer.SendPingTime)
}

func (myClient *CustomClient) sendResponse13(pingTime uint64) {
	nullIP, _ := net.ResolveUDPAddr("udp", "0.0.0.0:0")
	myClient.Address.Port = 0
	response := &Packet13Layer{
		IPAddress: &myClient.ServerAddress,
		Addresses: [10]*net.UDPAddr{
			&myClient.Address,
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
		SendPingTime: pingTime,
		SendPongTime: uint64(time.Now().Unix()),
	}

	myClient.WritePacket(response)
}
func (myClient *CustomClient) sendProtocolSync() {
	response90 := &Packet90Layer{
		SchemaVersion: 36,
		RequestedFlags: []string{
			"AllowMoreAngles",
			"BodyColorsColor3PropertyReplicationEnabled",
			"FixWeldConstraintReplicationCLI19374",
			"NetworkClusterByte2",
			"NetworkCompressorNewRotation",
			"NetworkCompressorNewTranslation",
			"NetworkCompressorNewVelocity",
			"NetworkNewInstanceNoDefault",
			"NetworkS2NewFraming",
			"SendAdditionalNonAdjustedTimeStamp",
			"SendPlayerGuiEarly2",
			"UseNativePathWaypoint",
			"UseNewPhysicsSender7",
			"UseNewProtocolForStreaming",
		},
	}
	myClient.WritePacket(response90)
}
func (myClient *CustomClient) sendPlaceIdVerification(placeId int64) {
	response92 := &Packet92Layer{
		PlaceId: placeId,
	}
	myClient.WritePacket(response92)
}
func (myClient *CustomClient) submitTicket() {
	response8A := &Packet8ALayer{
		PlayerId:          myClient.PlayerId,
		ClientTicket:      myClient.clientTicket,
		DataModelHash:     myClient.SecuritySettings.DataModelHash,
		ProtocolVersion:   36,
		SecurityKey:       myClient.SecuritySettings.SecurityKey,
		Platform:          myClient.SecuritySettings.OsPlatform,
		RobloxProductName: "?",
		SessionId:         myClient.sessionId,
		GoldenHash:        myClient.SecuritySettings.GoldenHash,
	}
	myClient.WritePacket(response8A)
}
func (myClient *CustomClient) sendSpawnName() {
	response8F := &Packet8FLayer{
		SpawnName: "",
	}
	myClient.WritePacket(response8F)
}
func (myClient *CustomClient) packet10Handler(packetType uint8, packet *UDPPacket, layers *PacketLayers) {
	mainLayer := layers.Main.(*Packet10Layer)

	myClient.sendResponse13(mainLayer.SendPongTime)
	myClient.sendProtocolSync()
	myClient.sendPlaceIdVerification(0)
	myClient.submitTicket()
	myClient.sendSpawnName()
}

func (myClient *CustomClient) topReplicationHandler(packetType uint8, packet *UDPPacket, layers *PacketLayers) {
	mainLayer := layers.Main.(*Packet81Layer)
	for _, inst := range mainLayer.Items { // this may result in instances being announced twice!
		// be careful.
		myClient.instanceHandlers.Fire(inst.Instance)
	}

	myClient.startDataPing()
}

func (myClient *CustomClient) dataHandler(packetType uint8, packet *UDPPacket, layers *PacketLayers) {
	mainLayer := layers.Main.(*Packet83Layer)
	for _, item := range mainLayer.SubPackets {
		myClient.dataHandlers.Fire(Packet83ToType(item), packet, layers, item)
	}
}

func (myClient *CustomClient) WriteDataPackets(packets ...Packet83Subpacket) {
	myClient.WritePacket(&Packet83Layer{
		SubPackets: packets,
	})
}

func (myClient *CustomClient) sendDataPingBack() {
	response := &Packet83_06{
		SendStats:  8,
		Timestamp:  uint64(time.Now().Unix()),
		IsPingBack: true,
	}

	myClient.WriteDataPackets(response)
}
func (myClient *CustomClient) dataPingHandler(packetType uint8, packet *UDPPacket, layers *PacketLayers, item Packet83Subpacket) {
	myClient.sendDataPingBack()
}

func (myClient *CustomClient) sendDataIdResponse(challengeInt uint32) {
	myClient.WriteDataPackets(&Packet83_09{
		Type: 6,
		Subpacket: &Packet83_09_06{
			Int1: challengeInt,
			Int2: myClient.SecuritySettings.IdChallengeResponse - challengeInt,
		},
	})
}
func (myClient *CustomClient) idChallengeHandler(packetType uint8, packet *UDPPacket, layers *PacketLayers, item Packet83Subpacket) {
	mainPacket := item.(*Packet83_09)
	if mainPacket.Type == 5 {
		myClient.Logger.Println("recv id challenge!")
		myClient.sendDataIdResponse(mainPacket.Subpacket.(*Packet83_09_05).Int)
	}
}

func (myClient *CustomClient) eventHandler(packetType uint8, packet *UDPPacket, layers *PacketLayers, item Packet83Subpacket) {
	mainPacket := item.(*Packet83_07)

	myClient.eventHandlers.Fire(mainPacket.Instance, mainPacket.EventName, mainPacket.Event)
}

func (myClient *CustomClient) handlePlayersService(players *rbxfile.Instance) {
	// this function will be called twice if both top repl and data repl contain Players
	if myClient.LocalPlayer != nil { // do not send localplayer twice!
		return
	}

	myPlayer := &rbxfile.Instance{
		ClassName: "Player",
		Reference: myClient.scope + "_" + strconv.Itoa(int(myClient.instanceIndex)),
		IsService: false,
		Properties: map[string]rbxfile.Value{
			"Name":                  rbxfile.ValueString(myClient.UserName),
			"CharacterAppearance":   rbxfile.ValueString(myClient.characterAppearance),
			"CharacterAppearanceId": rbxfile.ValueInt64(myClient.characterAppearanceId),
			"ChatPrivacyMode": rbxfile.ValueToken{
				Value: 0,
				ID:    uint16(myClient.Context.StaticSchema.EnumsByName["ChatPrivacyMode"]),
				Name:  "ChatPrivacyMode",
			},
			"AccountAgeReplicate": rbxfile.ValueInt(myClient.AccountAge),
			"OsPlatform":          rbxfile.ValueString(myClient.SecuritySettings.OsPlatform),
			"userId":              rbxfile.ValueInt64(myClient.PlayerId),
			"UserId":              rbxfile.ValueInt64(myClient.PlayerId),
			"ReplicatedLocaleId":  rbxfile.ValueString("en-us"),
		},
	}
	players.AddChild(myPlayer)
	myClient.instanceIndex++

	myClient.WriteDataPackets(
		&Packet83_05{
			SendStats:  8,
			Timestamp:  uint64(time.Now().Unix()),
			IsPingBack: false,
		},
		&Packet83_0B{},
		&Packet83_02{myPlayer},
	)
	myClient.LocalPlayer = myPlayer
	myClient.instanceHandlers.Fire(myPlayer)
}

func (myClient *CustomClient) newInstanceHandler(packetType uint8, packet *UDPPacket, layers *PacketLayers, subpacket Packet83Subpacket) {
	mainpacket := subpacket.(*Packet83_02)

	myClient.instanceHandlers.Fire(mainpacket.Child)
}

func (myClient *CustomClient) joinDataHandler(packetType uint8, packet *UDPPacket, layers *PacketLayers, subpacket Packet83Subpacket) {
	mainpacket := subpacket.(*Packet83_0B)

	for _, inst := range mainpacket.Instances {
		myClient.instanceHandlers.Fire(inst)
	}
}

func (myClient *CustomClient) SendEvent(instance *rbxfile.Instance, name string, arguments ...rbxfile.Value) {
	myClient.WriteDataPackets(
		&Packet83_07{
			Instance:  instance,
			EventName: name,
			Event:     &ReplicationEvent{arguments},
		},
	)
}

func (myClient *CustomClient) InvokeRemote(instance *rbxfile.Instance, arguments []rbxfile.Value) (rbxfile.ValueTuple, string) {
	if myClient.LocalPlayer == nil {
		panic(errors.New("local player is nil!"))
	}

	myClient.remoteLock.Lock()
	myClient.remoteIndices[instance]++
	index := myClient.remoteIndices[instance]
	myClient.remoteLock.Unlock()

	unbind1, succChan := myClient.MakeEventChan(instance, "RemoteOnInvokeSuccess")
	unbind2, errChan := myClient.MakeEventChan(instance, "RemoteOnInvokeError")
	// it should be ok to leave the chans open
	// they will be gc'd anyway

	myClient.SendEvent(instance, "RemoteOnInvokeServer",
		rbxfile.ValueInt(index),
		rbxfile.ValueReference{Instance: myClient.LocalPlayer},
		rbxfile.ValueTuple(arguments),
	)

	for true {
		select {
		case succ := <-succChan:
			// check that this packet was sent for us specifically
			if uint32(succ.Arguments[0].(rbxfile.ValueInt)) == index {
				myClient.eventHandlers.Unbind(unbind1)
				myClient.eventHandlers.Unbind(unbind2)

				return succ.Arguments[1].(rbxfile.ValueTuple), "" // return any values
			}
		case err := <-errChan:
			if uint32(err.Arguments[0].(rbxfile.ValueInt)) == index {
				myClient.eventHandlers.Unbind(unbind1)
				myClient.eventHandlers.Unbind(unbind2)

				return nil, string(err.Arguments[1].(rbxfile.ValueString))
			}
		}
	}

	return nil, ""
}

func (myClient *CustomClient) FireRemote(instance *rbxfile.Instance, arguments ...rbxfile.Value) {
	if myClient.LocalPlayer == nil {
		panic(errors.New("local player is nil!"))
	}
	myClient.SendEvent(instance, "OnServerEvent",
		rbxfile.ValueReference{Instance: myClient.LocalPlayer},
		rbxfile.ValueTuple(arguments),
	)
}

func (myClient *CustomClient) disconnectHandler(packetType uint8, packet *UDPPacket, layers *PacketLayers) {
	mainLayer := layers.Main.(*Packet15Layer)
	myClient.Logger.Printf("Disconnected because of reason %d\n", mainLayer.Reason)

	myClient.disconnectInternal()
}

func (myClient *CustomClient) MakeChildChan(instance *rbxfile.Instance) chan *rbxfile.Instance {
	newChan := make(chan *rbxfile.Instance)

	go func() { // we don't want this to block
		for _, child := range instance.Children {
			newChan <- child
		}
	}()

	path := NewInstancePath(instance)
	path.p = append(path.p, "*") // wildcard

	myClient.instanceHandlers.Bind(path, func(inst *rbxfile.Instance) {
		newChan <- inst
	})

	return newChan
}

func (myClient *CustomClient) deleteHandler(packetType uint8, packet *UDPPacket, layers *PacketLayers, subpacket Packet83Subpacket) {
	mainpacket := subpacket.(*Packet83_01)

	myClient.deleteHandlers.Fire(mainpacket.Instance)
}

type GroupDeleteChan struct {
	C         chan *rbxfile.Instance
	binding   int
	client    *CustomClient
	referents []Referent
}

func (channel *GroupDeleteChan) AddInstances(instances ...*rbxfile.Instance) {
	for _, inst := range instances {
		channel.referents = append(channel.referents, Referent(inst.Reference))
	}
}
func (channel *GroupDeleteChan) Destroy() {
	channel.client.dataHandlers.Unbind(1, channel.binding)
}
func (myClient *CustomClient) MakeGroupDeleteChan(instances []*rbxfile.Instance) *GroupDeleteChan {
	channel := &GroupDeleteChan{
		C:         make(chan *rbxfile.Instance),
		referents: make([]Referent, len(instances)),
		client:    myClient,
	}

	for i, inst := range instances {
		channel.referents[i] = Referent(inst.Reference)
	}

	channel.binding = myClient.dataHandlers.Bind(1, func(packetType uint8, packet *UDPPacket, layers *PacketLayers, subpacket Packet83Subpacket) {
		mainpacket := subpacket.(*Packet83_01)
		for _, inst := range channel.referents {
			if string(inst) == mainpacket.Instance.Reference {
				channel.C <- mainpacket.Instance
				break
			}
		}
	})

	return channel
}
