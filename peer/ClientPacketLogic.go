package peer

import (
	"errors"
	"math"
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/gskartwii/rbxfile"
)

func (myClient *CustomClient) bindDefaultHandlers() {
	myClient.PacketLogicHandler.bindDefaultHandlers()

	basicHandlers := myClient.handlers
	basicHandlers.Bind(6, myClient.simple6Handler)
	basicHandlers.Bind(8, myClient.simple8Handler)
	basicHandlers.Bind(0x10, myClient.packet10Handler)
	basicHandlers.Bind(0x81, myClient.topReplicationHandler)

	dataHandlers := myClient.dataHandlers
	dataHandlers.Bind(9, myClient.idChallengeHandler)

	myClient.RegisterInstanceHandler(&InstancePath{[]string{"Players"}, nil}, myClient.handlePlayersService).Once = true
}

func (myClient *CustomClient) sendResponse7() {
	myClient.WriteSimple(&Packet07Layer{
		GUID:      myClient.GUID,
		MTU:       1492,
		IPAddress: &myClient.ServerAddress,
	})
}
func (myClient *CustomClient) simple6Handler(packetType byte, layers *PacketLayers) {
	myClient.Connected = true
	myClient.sendResponse7()
}

// transition to real RakNet communication, no more offline messaging
func (myClient *CustomClient) sendResponse9() {
	response := &Packet09Layer{
		GUID:        myClient.GUID,
		Timestamp:   uint64(time.Now().UnixNano() / int64(time.Millisecond)),
		UseSecurity: false,
		Password:    []byte{0x37, 0x4F, 0x5E, 0x11, 0x6C, 0x45},
	}
	_, err := myClient.WritePacket(response)
	if err != nil {
		println("Failed to write response9: ", err.Error())
	}
}
func (myClient *CustomClient) simple8Handler(packetType byte, layers *PacketLayers) {
	myClient.sendResponse9()
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
		SendPongTime: uint64(time.Now().UnixNano() / int64(time.Millisecond)),
	}

	_, err := myClient.WritePacket(response)
	if err != nil {
		println("Failed to write response13: ", err.Error())
	}
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
	_, err := myClient.WritePacket(response90)
	if err != nil {
		println("Failed to write response90: ", err.Error())
	}
}
func (myClient *CustomClient) sendPlaceIdVerification(placeId int64) {
	response92 := &Packet92Layer{
		PlaceId: placeId,
	}
	_, err := myClient.WritePacket(response92)
	if err != nil {
		println("Failed to write response92: ", err.Error())
	}
}
func (myClient *CustomClient) submitTicket() {
	response8A := &Packet8ALayer{
		PlayerId:          myClient.PlayerId,
		ClientTicket:      myClient.clientTicket,
		ProtocolVersion:   36,
		RobloxProductName: "?",
		SessionId:         myClient.sessionId,
	}
	myClient.SecuritySettings.PatchTicketPacket(response8A)
	_, err := myClient.WritePacket(response8A)
	if err != nil {
		println("Failed to write response8A: ", err.Error())
	}
}
func (myClient *CustomClient) sendSpawnName() {
	response8F := &Packet8FLayer{
		SpawnName: "",
	}
	_, err := myClient.WritePacket(response8F)
	if err != nil {
		println("Failed to write response8F: ", err.Error())
	}
}
func (myClient *CustomClient) packet10Handler(packetType uint8, layers *PacketLayers) {
	mainLayer := layers.Main.(*Packet10Layer)

	myClient.sendResponse13(mainLayer.SendPongTime)
	myClient.sendProtocolSync()
	myClient.sendPlaceIdVerification(0)
	myClient.submitTicket()
	myClient.sendSpawnName()
}

func (myClient *CustomClient) topReplicationHandler(packetType uint8, layers *PacketLayers) {
	mainLayer := layers.Main.(*Packet81Layer)
	for _, inst := range mainLayer.Items { // this may result in instances being announced twice!
		// be careful.
		myClient.instanceHandlers.Fire(inst.Instance)
	}

	myClient.startDataPing()
}

func (myClient *CustomClient) sendDataIdResponse(challengeInt uint32) {
	err := myClient.WriteDataPackets(&Packet83_09{
		SubpacketType: 6,
		Subpacket: &Packet83_09_06{
			Int1: challengeInt,
			Int2: myClient.SecuritySettings.GenerateIdResponse(challengeInt),
		},
	})
	if err != nil {
		println("Failed to send dataidresponse:", err.Error())
	}
}
func (myClient *CustomClient) idChallengeHandler(packetType uint8, layers *PacketLayers, item Packet83Subpacket) {
	mainPacket := item.(*Packet83_09)
	if mainPacket.SubpacketType == 5 {
		myClient.Logger.Println("recv id challenge!")
		myClient.sendDataIdResponse(mainPacket.Subpacket.(*Packet83_09_05).Int)
	}
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
			"OsPlatform":          rbxfile.ValueString(myClient.SecuritySettings.OsPlatform()),
			"userId":              rbxfile.ValueInt64(myClient.PlayerId),
			"UserId":              rbxfile.ValueInt64(myClient.PlayerId),
			"ReplicatedLocaleId":  rbxfile.ValueString("en-us"),
		},
		PropertiesMutex: &sync.RWMutex{},
	}
	players.AddChild(myPlayer)
	myClient.instanceIndex++
	myClient.Context.InstancesByReferent.AddInstance(Referent(myPlayer.Reference), myPlayer)

	err := myClient.WriteDataPackets(
		&Packet83_05{
			Timestamp:  uint64(time.Now().UnixNano() / int64(time.Millisecond)),
			IsPingBack: false,
		},
		&Packet83_0B{},
		&Packet83_02{myPlayer},
	)
	if err != nil {
		println("Failed to send localplayer:", err.Error())
	}
	myClient.LocalPlayer = myPlayer
	go myClient.instanceHandlers.Fire(myPlayer) // prevent deadlock
	return
}

func (myClient *CustomClient) InvokeRemote(instance *rbxfile.Instance, arguments []rbxfile.Value) (rbxfile.ValueTuple, error) {
	if myClient.LocalPlayer == nil {
		panic(errors.New("local player is nil!"))
	}

	myClient.remoteLock.Lock()
	myClient.remoteIndices[instance]++
	index := myClient.remoteIndices[instance]
	myClient.remoteLock.Unlock()

	// TODO: Instead of creating new event chans every time, have a global channel for each instance
	// This way the event won't be eaten in case the Arguments[0] == index check doesn't pass
	conn1, succChan := myClient.MakeEventChan(instance, "RemoteOnInvokeSuccess")
	conn2, errChan := myClient.MakeEventChan(instance, "RemoteOnInvokeError")
	// it should be ok to leave the chans open
	// they will be gc'd anyway

	err := myClient.SendEvent(instance, "RemoteOnInvokeServer",
		rbxfile.ValueInt(index),
		rbxfile.ValueReference{Instance: myClient.LocalPlayer},
		rbxfile.ValueTuple(arguments),
	)
	if err != nil {
		return nil, err
	}

	for {
		select {
		case succ := <-succChan:
			// check that this packet was sent for us specifically
			if uint32(succ.Arguments[0].(rbxfile.ValueInt)) == index {
				conn1.Disconnect()
				conn2.Disconnect()

				return succ.Arguments[1].(rbxfile.ValueTuple), nil // return any values
			}
		case err := <-errChan:
			if uint32(err.Arguments[0].(rbxfile.ValueInt)) == index {
				conn1.Disconnect()
				conn2.Disconnect()

				return nil, errors.New(string(err.Arguments[1].(rbxfile.ValueString)))
			}
		}
	}
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

func addVector(vec1, vec2 rbxfile.ValueVector3) rbxfile.ValueVector3 {
	return rbxfile.ValueVector3{
		X: vec1.X + vec2.X,
		Y: vec1.Y + vec2.Y,
		Z: vec1.Z + vec2.Z,
	}
}
func subtractVector(vec1, vec2 rbxfile.ValueVector3) rbxfile.ValueVector3 {
	return rbxfile.ValueVector3{
		X: vec1.X - vec2.X,
		Y: vec1.Y - vec2.Y,
		Z: vec1.Z - vec2.Z,
	}
}
func scaleVector(vec rbxfile.ValueVector3, scalar float32) rbxfile.ValueVector3 {
	return rbxfile.ValueVector3{
		X: vec.X * scalar,
		Y: vec.Y * scalar,
		Z: vec.Z * scalar,
	}
}
func dotProduct(vec1, vec2 rbxfile.ValueVector3) float32 {
	return vec1.X*vec2.X + vec1.Y*vec2.Y + vec1.Z*vec2.Z
}
func vectorLength(vec rbxfile.ValueVector3) float32 {
	return float32(math.Sqrt(float64(dotProduct(vec, vec))))
}
func unitVector(vec rbxfile.ValueVector3) rbxfile.ValueVector3 {
	length := vectorLength(vec)
	if length == 0 {
		return vec
	}
	return scaleVector(vec, 1/length)
}
func scaleDelta(vec1, vec2 rbxfile.ValueVector3, scale float32) rbxfile.ValueVector3 {
	delta := subtractVector(vec2, vec1)
	deltaUnit := unitVector(delta)
	scaledDelta := scaleVector(deltaUnit, scale)
	if vectorLength(scaledDelta) > vectorLength(delta) {
		return delta
	}

	return scaledDelta
}
func interpolateVector(vec1, vec2 rbxfile.ValueVector3, maxStep float32) rbxfile.ValueVector3 {
	return addVector(vec1, scaleDelta(vec1, vec2, maxStep))
}

func (myClient *CustomClient) StalkPlayer(name string) {
	myCharacter := <-myClient.WaitForRefProp(myClient.GetLocalPlayer(), "Character")
	myRootPart := <-myClient.WaitForChild(myCharacter, "HumanoidRootPart")
	myAnimator := <-myClient.WaitForChild(myCharacter, "Humanoid", "Animator")
	targetPlayer := <-myClient.WaitForInstance("Players", name)
	targetCharacter := <-myClient.WaitForRefProp(targetPlayer, "Character")
	targetRootPart := <-myClient.WaitForChild(targetCharacter, "HumanoidRootPart")
	targetAnimator := <-myClient.WaitForChild(targetCharacter, "Humanoid", "Animator")
	println("got target root part")

	var targetPosition rbxfile.ValueCFrame
	myRootPart.PropertiesMutex.RLock()
	currentPosition := myRootPart.Properties["CFrame"].(rbxfile.ValueCFrame).Position
	myRootPart.PropertiesMutex.RUnlock()
	currentHumanoidState := uint8(8)

	myClient.RegisterPacketHandler(0x85, func(packetType uint8, layers *PacketLayers) {
		mainLayer := layers.Main.(*Packet85Layer)
		for _, packet := range mainLayer.SubPackets {
			if packet.Data.Instance == targetRootPart {
				if len(packet.History) > 0 {
					latestPhysicsData := packet.History[len(packet.History)-1]

					targetPosition = latestPhysicsData.CFrame
				}
			}
		}
	})

	stalkTicker := time.NewTicker(time.Second / 30)

	reactToHealthUpdate := func(value rbxfile.Value) {
		if value.(rbxfile.ValueFloat) <= 0.0 {
			currentHumanoidState = 15
		}
	}
	var healthConnection *PacketHandlerConnection
	animConnection, animationChan := myClient.MakeEventChan(targetAnimator, "OnPlay")
	myClient.propHandlers.Bind(myClient.GetLocalPlayer(), "Character", func(value rbxfile.Value) {
		healthConnection.Disconnect()
		println("localca, updating")
		myCharacter = value.(rbxfile.ValueReference).Instance
		myRootPart = <-myClient.WaitForChild(myCharacter, "HumanoidRootPart")
		myAnimator = <-myClient.WaitForChild(myCharacter, "Humanoid", "Animator")
		myRootPart.PropertiesMutex.RLock()
		currentPosition = myRootPart.Properties["CFrame"].(rbxfile.ValueCFrame).Position
		myRootPart.PropertiesMutex.RUnlock()
		healthConnection = myClient.propHandlers.Bind(myCharacter.FindFirstChild("Humanoid", false), "Health_XML", reactToHealthUpdate)
		currentHumanoidState = 8
	})
	myClient.propHandlers.Bind(myCharacter.FindFirstChild("Humanoid", false), "Health_XML", reactToHealthUpdate)
	myClient.propHandlers.Bind(targetPlayer, "Character", func(value rbxfile.Value) {
		println("targetca, updating")
		targetCharacter = value.(rbxfile.ValueReference).Instance
		targetRootPart = <-myClient.WaitForChild(targetCharacter, "HumanoidRootPart")
		targetAnimator = <-myClient.WaitForChild(targetCharacter, "Humanoid", "Animator")
		animConnection.Disconnect()
		animConnection, animationChan = myClient.MakeEventChan(targetAnimator, "OnPlay")
	})
	for {
		select {
		case <-stalkTicker.C:
			currentPosition = interpolateVector(currentPosition, targetPosition.Position, 16.0/10.0)
			currentVelocity := scaleVector(scaleDelta(currentPosition, targetPosition.Position, 16.0), 30.0)

			myClient.stalkPart(myRootPart, rbxfile.ValueCFrame{Position: currentPosition, Rotation: targetPosition.Rotation}, currentVelocity, rbxfile.ValueVector3{}, currentHumanoidState)
		case event := <-animationChan: // Thankfully, this should receive from the updated animationChan
			err := myClient.SendEvent(myAnimator, "OnPlay", event.Arguments...)
			if err != nil {
				println("Failed to send onplay event: ", err.Error())
			}
		}
	}
}

func (myClient *CustomClient) NewTimestamp() *Packet1BLayer {
	timestamp := &Packet1BLayer{Timestamp: uint64(time.Now().UnixNano() / int64(time.Millisecond)), Timestamp2: myClient.timestamp2Index}
	myClient.timestamp2Index++
	return timestamp
}

func (myClient *CustomClient) stalkPart(movePart *rbxfile.Instance, cframe rbxfile.ValueCFrame, linearVel rbxfile.ValueVector3, rotationVel rbxfile.ValueVector3, humState uint8) {
	myCframe := rbxfile.ValueCFrame{
		Rotation: cframe.Rotation,
		Position: rbxfile.ValueVector3{
			cframe.Position.X,
			cframe.Position.Y + 10.0,
			cframe.Position.Z,
		},
	}

	physicsPacket := &Packet85Layer{
		SubPackets: []*Packet85LayerSubpacket{
			&Packet85LayerSubpacket{
				Data: PhysicsData{
					Instance:           movePart,
					CFrame:             myCframe,
					LinearVelocity:     linearVel,
					RotationalVelocity: rotationVel,
				},
				NetworkHumanoidState: humState,
			},
		},
	}

	_, err := myClient.WritePhysics(myClient.NewTimestamp(), physicsPacket)
	if err != nil {
		println("Failed to send stalking packet:", err.Error())
	}
}
