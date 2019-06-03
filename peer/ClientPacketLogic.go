package peer

import (
	"context"
	"errors"
	"math"
	"net"
	"time"

	"github.com/Gskartwii/roblox-dissector/datamodel"
	"github.com/olebedev/emitter"
	"github.com/robloxapi/rbxfile"
)

func (myClient *CustomClient) startDataPing() {
	// boot up dataping
	myClient.dataPingTicker = time.NewTicker(time.Duration(myClient.pingInterval) * time.Millisecond)
	go func() {
		for {
			select {
			case <-myClient.dataPingTicker.C:
				myClient.WritePacket(&Packet83Layer{
					[]Packet83Subpacket{&Packet83_05{
						Timestamp:     uint64(time.Now().UnixNano() / int64(time.Millisecond)),
						PacketVersion: 2,
						Fps1:          60,
						Fps2:          60,
						Fps3:          60,
					}},
				})
			case <-myClient.RunningContext.Done():
				return
			}
		}
	}()
}

func (myClient *CustomClient) disconnectionLogger(e *emitter.Event) {
	myClient.Logger.Println("Received disconnection:", e.Args[0].(*Packet15Layer).String())
}

func (myClient *CustomClient) bindDefaultHandlers() {
	// let the DataModel be updated properly
	myClient.DefaultPacketReader.BindDataModelHandlers()
	myClient.PacketLogicHandler.bindDefaultHandlers()

	pEmitter := myClient.PacketEmitter
	pEmitter.On("ID_OPEN_CONNECTION_REPLY_1", myClient.offline6Handler, emitter.Void)
	pEmitter.On("ID_OPEN_CONNECTION_REPLY_2", myClient.offline8Handler, emitter.Void)
	pEmitter.On("ID_CONNECTION_ACCEPTED", myClient.packet10Handler, emitter.Void)
	pEmitter.On("ID_DISCONNECTION_NOTIFICATION", myClient.disconnectionLogger, emitter.Void)
	pEmitter.On("ID_SET_GLOBALS", myClient.topReplicationHandler, emitter.Void)

	dataHandlers := myClient.DataEmitter
	dataHandlers.On("ID_REPLIC_ROCKY", myClient.idChallengeHandler, emitter.Void)

	// Even though this could be compressed into a single function call,
	// we won't do that because the chan receive MUST be executed inside
	// the go func(){}()
	go func() {
		players, err := myClient.DataModel.WaitForService(myClient.RunningContext, "Players")
		if err != nil {
			println("players serv error:", err.Error())
			return
		}
		myClient.handlePlayersService(players)
	}()
}

func (myClient *CustomClient) sendResponse7() {
	myClient.WriteOffline(&Packet07Layer{
		GUID:      myClient.GUID,
		MTU:       1492,
		IPAddress: myClient.ServerAddress,
	})
}
func (myClient *CustomClient) offline6Handler(e *emitter.Event) {
	println("receive 6")
	myClient.Connected = true
	myClient.sendResponse7()
}

// transition to real RakNet communication, no more offline messaging
func (myClient *CustomClient) sendResponse9() {
	response := &Packet09Layer{
		GUID:        myClient.GUID,
		Timestamp:   uint64(time.Now().UnixNano() / int64(time.Millisecond)),
		UseSecurity: false,
		Password:    DefaultPasswordBytes,
	}
	err := myClient.WritePacket(response)
	if err != nil {
		println("Failed to write response9: ", err.Error())
	}
}
func (myClient *CustomClient) offline8Handler(e *emitter.Event) {
	myClient.sendResponse9()
}

func (myClient *CustomClient) sendResponse13(pingTime uint64) {
	nullIP, _ := net.ResolveUDPAddr("udp", "0.0.0.0:0")
	myClient.Address.Port = 0
	response := &Packet13Layer{
		IPAddress: myClient.ServerAddress,
		Addresses: [10]*net.UDPAddr{
			myClient.Address,
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

	err := myClient.WritePacket(response)
	if err != nil {
		println("Failed to write response13: ", err.Error())
	}
}
func (myClient *CustomClient) sendProtocolSync() {
	marshalledJoin, err := myClient.joinDataObject.JSON()
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
			"PgsForAll",
			"ReplicateInterpolateRelativeHumanoidPlatformsMotion",
			"TerrainRaycastsRespectCollisionGroups",
		},
		JoinData: string(marshalledJoin),
	}
	err = myClient.WritePacket(response90)
	if err != nil {
		println("Failed to write response90: ", err.Error())
	}
}
func (myClient *CustomClient) sendPlaceIDVerification(placeID int64) {
	response92 := &Packet92Layer{
		PlaceID: placeID,
	}
	err := myClient.WritePacket(response92)
	if err != nil {
		println("Failed to write response92: ", err.Error())
	}
}
func (myClient *CustomClient) submitTicket() {
	response8A := &Packet8ALayer{
		PlayerID:          myClient.PlayerID,
		ClientTicket:      myClient.clientTicket,
		ProtocolVersion:   36,
		RobloxProductName: "?",
		SessionID:         myClient.sessionID,
	}
	myClient.SecuritySettings.PatchTicketPacket(response8A)
	err := myClient.WritePacket(response8A)
	if err != nil {
		println("Failed to write response8A: ", err.Error())
	}
}
func (myClient *CustomClient) sendSpawnName() {
	response8F := &Packet8FLayer{
		SpawnName: "",
	}
	err := myClient.WritePacket(response8F)
	if err != nil {
		println("Failed to write response8F: ", err.Error())
	}
}
func (myClient *CustomClient) packet10Handler(e *emitter.Event) {
	mainLayer := e.Args[0].(*Packet10Layer)

	myClient.sendResponse13(mainLayer.SendPongTime)
	// RakNet sends two pings when connecting
	myClient.sendPing()
	myClient.sendPing()
	myClient.sendProtocolSync()
	myClient.sendPlaceIDVerification(0)
	myClient.submitTicket()
	myClient.sendSpawnName()
}

func (myClient *CustomClient) topReplicationHandler(e *emitter.Event) {
	myClient.startDataPing()
}

func (myClient *CustomClient) sendDataIDResponse(challengeInt uint32) {
	err := myClient.WriteDataPackets(&Packet83_09{
		Subpacket: &Packet83_09_06{
			Challenge: challengeInt,
			Response:  myClient.SecuritySettings.GenerateIDResponse(challengeInt),
		},
	})
	if err != nil {
		println("Failed to send dataidresponse:", err.Error())
	}
}
func (myClient *CustomClient) idChallengeHandler(e *emitter.Event) {
	mainPacket := e.Args[0].(*Packet83_09)
	idChallenge, ok := mainPacket.Subpacket.(*Packet83_09_05)
	if ok {
		myClient.Logger.Println("recv id challenge!")
		myClient.sendDataIDResponse(idChallenge.Challenge)
	}
}

func (myClient *CustomClient) handlePlayersService(players *datamodel.Instance) {
	// this function will be called twice if both top repl and data repl contain Players
	if myClient.localPlayer != nil { // do not send localplayer twice!
		return
	}

	myPlayer, _ := datamodel.NewInstance("Player", nil)
	myPlayer.Ref = myClient.InstanceDictionary.NewReference()
	myPlayer.Properties = map[string]rbxfile.Value{
		"Name":                              rbxfile.ValueString(myClient.UserName),
		"CharacterAppearance":               rbxfile.ValueString(myClient.characterAppearance),
		"CharacterAppearanceId":             rbxfile.ValueInt64(myClient.characterAppearanceID),
		"InternalCharacterAppearanceLoaded": rbxfile.ValueBool(true),
		// TODO: Assign ID here, in case somebody wants to pass this in an event arg
		"ChatPrivacyMode":     datamodel.ValueToken{Value: 0},
		"AccountAgeReplicate": rbxfile.ValueInt(myClient.AccountAge),
		"OsPlatform":          rbxfile.ValueString(myClient.SecuritySettings.OSPlatform()),
		"userId":              rbxfile.ValueInt64(myClient.PlayerID),
		"UserId":              rbxfile.ValueInt64(myClient.PlayerID),
		"ReplicatedLocaleId":  rbxfile.ValueString("en-us"),
	}
	err := players.AddChild(myPlayer)
	if err != nil {
		println("Failed to create localpalyer:", err.Error())
	}
	myClient.Context.InstancesByReference.AddInstance(myPlayer.Ref, myPlayer)
	myClient.localPlayer = myPlayer

	err = myClient.WriteDataPackets(
		&Packet83_05{
			Timestamp:     uint64(time.Now().UnixNano() / int64(time.Millisecond)),
			PacketVersion: 2,
			Fps1:          60,
			Fps2:          60,
			Fps3:          60,
		},
		&Packet83_0B{},
	)
	if err != nil {
		println("Failed to send initial data replic:", err.Error())
		return
	}
	err = myClient.ReplicateInstance(myClient.localPlayer, true)
	if err != nil {
		println("Failed to send local player:", err.Error())
	}
	return
}

// InvokeRemote sends a RemoteFunction invocation to the server and waits for a response
func (myClient *CustomClient) InvokeRemote(instance *datamodel.Instance, arguments []rbxfile.Value) (datamodel.ValueTuple, error) {
	localPlayer, err := myClient.LocalPlayer()
	if err != nil {
		return nil, err
	}

	myClient.remoteLock.Lock()
	myClient.remoteIndices[instance]++
	index := myClient.remoteIndices[instance]
	myClient.remoteLock.Unlock()

	myClient.Logger.Printf("Sending invocation %s#%d\n", instance.GetFullName(), index)
	succEvtChan := instance.EventEmitter.Once("RemoteOnInvokeSuccess")
	errEvtChan := instance.EventEmitter.Once("RemoteOnInvokeError")
	defer instance.EventEmitter.Off("RemoteOnInvokeSuccess", succEvtChan)
	defer instance.EventEmitter.Off("RemoteOnInvokeError", errEvtChan)

	err = myClient.SendEvent(instance, "RemoteOnInvokeServer",
		rbxfile.ValueInt(index),
		datamodel.ValueReference{Instance: localPlayer},
		datamodel.ValueTuple(arguments),
	)
	if err != nil {
		return nil, err
	}

	for {
		select {
		case e, received := <-succEvtChan:
			if !received {
				continue
			}
			succ := e.Args[0].([]rbxfile.Value)
			myClient.Logger.Printf("Listener #%d received success on %s#%d", index, instance.GetFullName(), uint32(succ[0].(rbxfile.ValueInt)))
			// check that this packet was sent for us specifically
			if uint32(succ[0].(rbxfile.ValueInt)) == index {
				return succ[1].(datamodel.ValueTuple), nil // return any values
			}
		case e, received := <-errEvtChan:
			if !received {
				continue
			}
			err := e.Args[0].([]rbxfile.Value)
			myClient.Logger.Printf("Listener #%d received error on %s#%d", index, instance.GetFullName(), uint32(err[0].(rbxfile.ValueInt)))
			if uint32(err[0].(rbxfile.ValueInt)) == index {
				return nil, errors.New(string(err[1].(rbxfile.ValueString)))
			}
		}
	}
}

// FireRemote fires a RemoteEvent
func (myClient *CustomClient) FireRemote(instance *datamodel.Instance, arguments ...rbxfile.Value) error {
	localPlayer, err := myClient.LocalPlayer()
	if err != nil {
		return err
	}
	return myClient.SendEvent(instance, "OnServerEvent",
		datamodel.ValueReference{Instance: localPlayer},
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

func (myClient *CustomClient) stalkPlayer(name string) error {
	// We must be able to cancel the process from within this function
	// while not halting the entire client
	ctx, stalkEnd := context.WithCancel(myClient.RunningContext)
	defer stalkEnd()

	localPlayer, err := myClient.LocalPlayer()
	if err != nil {
		return err
	}
	myCharacter, err := localPlayer.WaitForRefProp(ctx, "Character")
	if err != nil {
		return err
	}
	myRootPart, err := myCharacter.WaitForChild(ctx, "HumanoidRootPart")
	if err != nil {
		return err
	}
	myAnimator, err := myCharacter.WaitForChild(ctx, "Humanoid", "Animator")
	if err != nil {
		return err
	}
	targetPlayer, err := myClient.DataModel.WaitForChild(ctx, "Players", name)
	if err != nil {
		return err
	}
	targetCharacter, err := targetPlayer.WaitForRefProp(ctx, "Character")
	if err != nil {
		return err
	}
	targetRootPart, err := targetCharacter.WaitForChild(ctx, "HumanoidRootPart")
	if err != nil {
		return err
	}
	targetAnimator, err := targetCharacter.WaitForChild(ctx, "Humanoid", "Animator")
	if err != nil {
		return err
	}
	println("got target animator")

	var targetPosition rbxfile.ValueCFrame
	currentPosition := myRootPart.Get("CFrame").(rbxfile.ValueCFrame).Position
	currentHumanoidState := uint8(8)

	physicsEvtChan := myClient.PacketEmitter.On("ID_PHYSICS", func(e *emitter.Event) {
		mainLayer := e.Args[0].(*Packet85Layer)
		for _, packet := range mainLayer.SubPackets {
			if packet.Data.Instance == targetRootPart {
				if len(packet.History) > 0 {
					latestPhysicsData := packet.History[len(packet.History)-1]

					targetPosition = latestPhysicsData.CFrame
				}
			}
		}
	}, emitter.Void)
	defer myClient.PacketEmitter.Off("ID_PHYSICS", physicsEvtChan)

	stalkTicker := time.NewTicker(time.Second / 30)

	animationEvtChan := targetAnimator.EventEmitter.On("OnCombinedUpdate")
	myHumanoid, err := myCharacter.WaitForChild(ctx, "Humanoid")
	if err != nil {
		targetAnimator.EventEmitter.Off("OnCombinedUpdate", animationEvtChan)
		return err
	}
	healthHandler := func(e *emitter.Event) {
		if e.Args[0].(rbxfile.ValueFloat) <= 0.0 {
			currentHumanoidState = 15
		}
	}
	healthCh := myHumanoid.PropertyEmitter.On("Health_XML", healthHandler, emitter.Void)
	targetHandler := func(e *emitter.Event) {
		targetAnimator.EventEmitter.Off("OnPlay", animationEvtChan)
		animationEvtChan = nil
		println("targetca, updating")
		targetCharacter = e.Args[0].(datamodel.ValueReference).Instance
		if targetCharacter == nil {
			println("Stalk process finished!")
			stalkEnd()
			return
		}
		// this "middleware is called async, hence we don't want to block"
		go func() {
			targetRootPart, err = targetCharacter.WaitForChild(ctx, "HumanoidRootPart")
			if err != nil {
				println(err.Error())
				stalkEnd()
			}
			targetAnimator, err = targetCharacter.WaitForChild(ctx, "Humanoid", "Animator")
			if err != nil {
				println(err.Error())
				stalkEnd()
			}
			animationEvtChan = targetAnimator.EventEmitter.On("OnCombinedUpdate")
		}()
	}
	charHandler := func(e *emitter.Event) {
		myHumanoid.PropertyEmitter.Off("Health_XML", healthCh)
		println("localca, updating")
		myCharacter = e.Args[0].(datamodel.ValueReference).Instance
		// this "middleware is called async, hence we don't want to block"
		go func() {
			myRootPart, err = myCharacter.WaitForChild(ctx, "HumanoidRootPart")
			if err != nil {
				println(err.Error())
				stalkEnd()
			}
			myHumanoid, err = myCharacter.WaitForChild(ctx, "Humanoid")
			if err != nil {
				println(err.Error())
				stalkEnd()
			}
			myAnimator, err = myHumanoid.WaitForChild(ctx, "Animator")
			if err != nil {
				println(err.Error())
				stalkEnd()
			}
			currentPosition = myRootPart.Get("CFrame").(rbxfile.ValueCFrame).Position
			healthCh = myHumanoid.PropertyEmitter.On("Health_XML")
			currentHumanoidState = 8
		}()
	}
	localCharEvtChan := localPlayer.PropertyEmitter.On("Character", charHandler, emitter.Void)
	targetCharEvtChan := targetPlayer.PropertyEmitter.On("Character", targetHandler, emitter.Void)
	defer localPlayer.PropertyEmitter.Off("Character", localCharEvtChan)
	defer targetPlayer.PropertyEmitter.Off("Character", targetCharEvtChan)
	for {
		// TODO: Break out of this loop?
		select {
		case <-stalkTicker.C:
			currentPosition = interpolateVector(currentPosition, targetPosition.Position, 16.0/10.0)
			currentVelocity := scaleVector(scaleDelta(currentPosition, targetPosition.Position, 16.0), 30.0)

			myClient.stalkPart(myRootPart, rbxfile.ValueCFrame{Position: currentPosition, Rotation: targetPosition.Rotation}, currentVelocity, rbxfile.ValueVector3{}, currentHumanoidState)
		case e, received := <-animationEvtChan: // Thankfully, this should receive from the updated animationChan
			if !received {
				continue
			}
			event := e.Args[0].([]rbxfile.Value)
			err := myClient.SendEvent(myAnimator, "OnCombinedUpdate", event...)
			if err != nil {
				println("Failed to send onplay event: ", err.Error())
			}
		case <-myClient.RunningContext.Done():
			targetAnimator.EventEmitter.Off("OnCombinedUpdate", animationEvtChan)
			return myClient.RunningContext.Err()
		}
	}
}

// NewTimestamp generates a new Packet1BLayer
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

	err := myClient.WriteTimestamped(myClient.NewTimestamp(), physicsPacket)
	if err != nil {
		println("Failed to send stalking packet:", err.Error())
	}
}
