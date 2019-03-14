package peer

import (
	"errors"
	"math"
	"net"
	"time"

	"github.com/gskartwii/roblox-dissector/datamodel"
	"github.com/robloxapi/rbxfile"
)

func (myClient *CustomClient) startDataPing() {
	// boot up dataping
	myClient.dataPingTicker = time.NewTicker(time.Duration(myClient.pingInterval) * time.Millisecond)
	go func() {
		for {
			<-myClient.dataPingTicker.C

			myClient.WritePacket(&Packet83Layer{
				[]Packet83Subpacket{&Packet83_05{
					Timestamp:     uint64(time.Now().UnixNano() / int64(time.Millisecond)),
					PacketVersion: 2,
					Fps1:          60,
					Fps2:          60,
					Fps3:          60,
				}},
			})
		}
	}()
}

func (myClient *CustomClient) disconnectionLogger(packetType uint8, layers *PacketLayers) {
	myClient.Logger.Println("Received disconnection:", layers.Main.(*Packet15Layer).String())
}

func (myClient *CustomClient) bindDefaultHandlers() {
	myClient.PacketLogicHandler.bindDefaultHandlers()

	basicHandlers := myClient.handlers
	basicHandlers.Bind(6, myClient.simple6Handler)
	basicHandlers.Bind(8, myClient.simple8Handler)
	basicHandlers.Bind(0x10, myClient.packet10Handler)
	basicHandlers.Bind(0x15, myClient.disconnectionLogger)
	basicHandlers.Bind(0x81, myClient.topReplicationHandler)

	dataHandlers := myClient.dataHandlers
	dataHandlers.Bind(9, myClient.idChallengeHandler)

	// Even though this could be compressed into a single function call,
	// we won't do that because the chan receive MUST be executed inside
	// the go func(){}()
	go func() {
		players := <-myClient.DataModel.WaitForService("Players")
		myClient.handlePlayersService(players)
	}()
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
	marshalledJoin, err := myClient.JoinDataObject.JSON()
	if err != nil {
		myClient.Logger.Println("Failed to marshal join JSON: ", err.Error())
		return
	}
	response90 := &Packet90Layer{
		SchemaVersion: 36,
		RequestedFlags: []string{
			"FixDictionaryScopePlatformsReplication",
			"UseNativePathWaypoint",
			"BodyColorsColor3PropertyReplicationEnabled",
			"FixBallRaycasts",
			"FixRaysInWedges",
			"FixHats",
			"EnableRootPriority",
			"PartMasslessEnabled",
			"KeepRedundantWeldsAlways",
			"KeepRedundantWeldsExplicit",
			"PgsForAll",
			"WeldedToAnchoredIsntSpecial",
			"ReplicateInterpolateRelativeHumanoidPlatformsMotion",
			"TerrainRaycastsRespectCollisionGroups",
		},
		JoinData: string(marshalledJoin),
	}
	_, err = myClient.WritePacket(response90)
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
	// RakNet sends two pings when connecting
	myClient.sendPing()
	myClient.sendPing()
	myClient.sendProtocolSync() // This is broken and I have no clue why
	myClient.sendPlaceIdVerification(0)
	myClient.submitTicket()
	myClient.sendSpawnName()
}

func (myClient *CustomClient) topReplicationHandler(packetType uint8, layers *PacketLayers) {
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

func (myClient *CustomClient) handlePlayersService(players *datamodel.Instance) {
	// this function will be called twice if both top repl and data repl contain Players
	if myClient.LocalPlayer != nil { // do not send localplayer twice!
		return
	}

	myPlayer, _ := datamodel.NewInstance("Player", nil)
	myPlayer.Ref = myClient.InstanceDictionary.NewReference()
	myPlayer.Properties = map[string]rbxfile.Value{
		"Name":                              rbxfile.ValueString(myClient.UserName),
		"CharacterAppearance":               rbxfile.ValueString(myClient.characterAppearance),
		"CharacterAppearanceId":             rbxfile.ValueInt64(myClient.characterAppearanceId),
		"InternalCharacterAppearanceLoaded": rbxfile.ValueBool(true),
		// TODO: Assign ID here, in case somebody wants to pass this in an event arg
		"ChatPrivacyMode":     datamodel.ValueToken{Value: 0},
		"AccountAgeReplicate": rbxfile.ValueInt(myClient.AccountAge),
		"OsPlatform":          rbxfile.ValueString(myClient.SecuritySettings.OsPlatform()),
		"userId":              rbxfile.ValueInt64(myClient.PlayerId),
		"UserId":              rbxfile.ValueInt64(myClient.PlayerId),
		"ReplicatedLocaleId":  rbxfile.ValueString("en-us"),
	}
	err := players.AddChild(myPlayer)
	if err != nil {
		println("Failed to create localpalyer:", err.Error())
	}
	myClient.Context.InstancesByReferent.AddInstance(myPlayer.Ref, myPlayer)
	myClient.LocalPlayer = myPlayer

	err = myClient.WriteDataPackets(
		&Packet83_05{
			Timestamp:     uint64(time.Now().UnixNano() / int64(time.Millisecond)),
			PacketVersion: 2,
			Fps1:          60,
			Fps2:          60,
			Fps3:          60,
		},
		&Packet83_0B{},
		&Packet83_02{myPlayer},
	)
	if err != nil {
		println("Failed to send localplayer:", err.Error())
	}
	return
}

func (myClient *CustomClient) InvokeRemote(instance *datamodel.Instance, arguments []rbxfile.Value) (datamodel.ValueTuple, error) {
	if myClient.LocalPlayer == nil {
		panic(errors.New("local player is nil"))
	}

	myClient.remoteLock.Lock()
	myClient.remoteIndices[instance]++
	index := myClient.remoteIndices[instance]
	myClient.remoteLock.Unlock()

	succEmitter, succChan := instance.MakeEventChan("RemoteOnInvokeSuccess", true)
	errEmitter, errChan := instance.MakeEventChan("RemoteOnInvokeError", true)

	err := myClient.SendEvent(instance, "RemoteOnInvokeServer",
		rbxfile.ValueInt(index),
		datamodel.ValueReference{Instance: myClient.LocalPlayer},
		datamodel.ValueTuple(arguments),
	)
	if err != nil {
		return nil, err
	}

	for {
		select {
		case succ := <-succChan:
			// check that this packet was sent for us specifically
			if uint32(succ[0].(rbxfile.ValueInt)) == index {
				instance.EventEmitter.Off("RemoteOnInvokeError", errEmitter)

				return succ[1].(datamodel.ValueTuple), nil // return any values
			}
		case err := <-errChan:
			if uint32(err[0].(rbxfile.ValueInt)) == index {
				instance.EventEmitter.Off("RemoteOnInvokeSuccess", succEmitter)

				return nil, errors.New(string(err[1].(rbxfile.ValueString)))
			}
		}
	}
}

func (myClient *CustomClient) FireRemote(instance *datamodel.Instance, arguments ...rbxfile.Value) {
	if myClient.LocalPlayer == nil {
		panic(errors.New("local player is nil"))
	}
	myClient.SendEvent(instance, "OnServerEvent",
		datamodel.ValueReference{Instance: myClient.LocalPlayer},
		datamodel.ValueTuple(arguments),
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
	myCharacter := <-myClient.GetLocalPlayer().WaitForRefProp("Character")
	myRootPart := <-myCharacter.WaitForChild("HumanoidRootPart")
	myAnimator := <-myCharacter.WaitForChild("Humanoid", "Animator")
	targetPlayer := <-myClient.DataModel.WaitForChild("Players", name)
	targetCharacter := <-targetPlayer.WaitForRefProp("Character")
	targetRootPart := <-targetCharacter.WaitForChild("HumanoidRootPart")
	targetAnimator := <-targetCharacter.WaitForChild("Humanoid", "Animator")
	println("got target animator")

	var targetPosition rbxfile.ValueCFrame
	currentPosition := myRootPart.Get("CFrame").(rbxfile.ValueCFrame).Position
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

	animationEmitter, animationChan := targetAnimator.MakeEventChan("OnPlay", false)
	charEmitter := myClient.GetLocalPlayer().PropertyEmitter.On("Character")
	myHumanoid := <-myCharacter.WaitForChild("Humanoid")
	healthEmitter := myHumanoid.PropertyEmitter.On("Health_XML")
	targetEmitter := targetPlayer.PropertyEmitter.On("Character")
	for {
		select {
		case value := <-healthEmitter:
			if value.Args[0].(rbxfile.ValueFloat) <= 0.0 {
				currentHumanoidState = 15
			}
		case newChar := <-charEmitter:
			myHumanoid.PropertyEmitter.Off("Health_XML", healthEmitter)
			println("localca, updating")
			myCharacter = newChar.Args[0].(datamodel.ValueReference).Instance
			myRootPart = <-myCharacter.WaitForChild("HumanoidRootPart")
			myHumanoid = <-myCharacter.WaitForChild("Humanoid")
			myAnimator = <-myHumanoid.WaitForChild("Animator")
			currentPosition = myRootPart.Get("CFrame").(rbxfile.ValueCFrame).Position
			healthEmitter = myHumanoid.PropertyEmitter.On("Health_XML")
			currentHumanoidState = 8
		case newTarget := <-targetEmitter:
			targetAnimator.EventEmitter.Off("OnPlay", animationEmitter)
			println("targetca, updating")
			targetCharacter = newTarget.Args[0].(datamodel.ValueReference).Instance
			if targetCharacter == nil {
				println("Stalk process finished!")
				return
			}
			targetRootPart = <-targetCharacter.WaitForChild("HumanoidRootPart")
			targetAnimator = <-targetCharacter.WaitForChild("Humanoid", "Animator")
			animationEmitter, animationChan = targetAnimator.MakeEventChan("OnPlay", false)
		case <-stalkTicker.C:
			currentPosition = interpolateVector(currentPosition, targetPosition.Position, 16.0/10.0)
			currentVelocity := scaleVector(scaleDelta(currentPosition, targetPosition.Position, 16.0), 30.0)

			myClient.stalkPart(myRootPart, rbxfile.ValueCFrame{Position: currentPosition, Rotation: targetPosition.Rotation}, currentVelocity, rbxfile.ValueVector3{}, currentHumanoidState)
		case event := <-animationChan: // Thankfully, this should receive from the updated animationChan
			err := myClient.SendEvent(myAnimator, "OnPlay", event...)
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

func (myClient *CustomClient) stalkPart(movePart *datamodel.Instance, cframe rbxfile.ValueCFrame, linearVel rbxfile.ValueVector3, rotationVel rbxfile.ValueVector3, humState uint8) {
	myCframe := rbxfile.ValueCFrame{
		Rotation: cframe.Rotation,
		Position: rbxfile.ValueVector3{
			X: cframe.Position.X,
			Y: cframe.Position.Y + 10.0,
			Z: cframe.Position.Z,
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
