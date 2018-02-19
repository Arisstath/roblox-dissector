package peer
import "time"
import "net"
import "fmt"
import "sort"
import "net/http"
import "encoding/json"
import "github.com/gskartwii/rbxfile"
import "math/rand"
import "strconv"
import "errors"

type placeLauncherResponse struct {
	JobId string
	Status int
	JoinScriptUrl string
	AuthenticationUrl string
	AuthenticationTicket string
}
type joinAshxResponse struct {
	ClientTicket string
	NewClientTicket string
	SessionId string
	MachineAddress string
	ServerPort uint16
	UserId int32
	UserName string
	CharacterAppearance string
}

type CustomClient struct {
	*ConnectedPeer
	Context *CommunicationContext
	Address net.UDPAddr
	ServerAddress net.UDPAddr
	Connected bool
	mustACK []int
	clientTicket string
	sessionId string
	PlayerId int32
	UserName string
	characterAppearance string
	PlaceId uint32
	httpClient *http.Client
	GUID uint64
	BrowserTrackerId uint64
	GenderId uint8
	IsPartyLeader bool

	RakPassword []byte
	GoldenHash uint32
	SecurityKey string
	DataModelHash string
	OsPlatform string
	instanceIndex uint32
	scope string
}

func (client *CustomClient) sendACKs() {
	if len(client.mustACK) == 0 {
		return
	}
	acks := client.mustACK
	client.mustACK = []int{}
	var ackStructure []ACKRange
	sort.Ints(acks)

	for _, ack := range acks {
		if len(ackStructure) == 0 {
			ackStructure = append(ackStructure, ACKRange{uint32(ack), uint32(ack)})
			continue
		}

		inserted := false
		for i, ackRange := range ackStructure {
			if int(ackRange.Max) == ack {
				inserted = true
                break
			}
			if int(ackRange.Max + 1) == ack {
				ackStructure[i].Max++
				inserted = true
                break
			}
		}
		if inserted {
			continue
		}

		ackStructure = append(ackStructure, ACKRange{uint32(ack), uint32(ack)})
	}

	result := &RakNetLayer{
		IsValid: true,
		IsACK: true,
		ACKs: ackStructure,
	}

	client.Writer.WriteRakNet(result, &client.ServerAddress)
}

func (client *CustomClient) receive(buf []byte) {
	packet := UDPPacketFromBytes(buf)
	packet.Source = client.ServerAddress
	packet.Destination = client.Address
	client.Reader.ReadPacket(buf, packet)
}

func NewCustomClient() *CustomClient {
	rand.Seed(time.Now().UnixNano())
	return &CustomClient{
		httpClient: &http.Client{},
		Context: NewCommunicationContext(),
		GUID: rand.Uint64(),
		instanceIndex: 1000,
		scope: "RBX224117",
	}
}

func (myClient *CustomClient) joinWithPlaceLauncher(url string, cookies []*http.Cookie) error {
	println("requesting placelauncher", url)
	robloxCommClient := myClient.httpClient
	placeLauncherRequest, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	placeLauncherRequest.Header.Set("User-Agent", "Roblox/WinInet")
	for _, cookie := range cookies {
		placeLauncherRequest.AddCookie(cookie)
	}

	resp, err := robloxCommClient.Do(placeLauncherRequest)
	if err != nil {
		return err
	}
	var plResp placeLauncherResponse
	err = json.NewDecoder(resp.Body).Decode(&plResp)
	if err != nil {
		return err
	}
	if plResp.JoinScriptUrl == "" {
		println("joinscript failure, status", plResp.Status)
		return errors.New("couldn't get joinscripturl")
	}

	joinScriptRequest, err := http.NewRequest("GET", plResp.JoinScriptUrl, nil)
	if err != nil {
		return err
	}

	for _, cook := range resp.Cookies() {
		joinScriptRequest.AddCookie(cook)
	}
	for _, cook := range cookies {
		if x, _ := joinScriptRequest.Cookie(cook.Name); x == nil {
			joinScriptRequest.AddCookie(cook)
		}
	}

	resp, err = robloxCommClient.Do(joinScriptRequest)
	if err != nil {
		return err
	}
	body := resp.Body

	// Discard rbxsig by reading until newline
	char := make([]byte, 1)
	_, err = body.Read(char)
	for err == nil && char[0] != 0x0A {
		_, err = body.Read(char)
	}

	if err != nil {
		return err
	}

	var jsResp joinAshxResponse
	err = json.NewDecoder(body).Decode(&jsResp)
	if err != nil {
		return err
	}
	myClient.clientTicket = jsResp.ClientTicket
	myClient.sessionId = jsResp.SessionId
	myClient.PlayerId = jsResp.UserId
	myClient.UserName = jsResp.UserName
	addrp, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", jsResp.MachineAddress, jsResp.ServerPort))
	if err != nil {
		return err
	}

	myClient.ServerAddress = *addrp
	return myClient.rakConnect()
}

func (myClient *CustomClient) ConnectGuest(placeId uint32, genderId uint8) error {
	myClient.PlaceId = placeId
	myClient.GenderId = genderId
	return myClient.joinWithPlaceLauncher(fmt.Sprintf("https://assetgame.roblox.com/game/PlaceLauncher.ashx?request=RequestGame&browserTrackerId=%d&placeId=%d&isPartyLeader=false&genderId=%d", myClient.BrowserTrackerId, myClient.PlaceId, myClient.GenderId),[]*http.Cookie{})
}

func (myClient *CustomClient) rakConnect() error {
	context := myClient.Context
	addr := myClient.ServerAddress

	packetReader := NewPacketReader()
	packetReader.ACKHandler = func(packet *UDPPacket, layer *RakNetLayer) {
		println("received ack")
		if myClient.ACKHandler != nil {
			myClient.ACKHandler(packet, layer)
		}
	}
	packetReader.ReliabilityLayerHandler = func(p *UDPPacket, re *ReliabilityLayer, ra *RakNetLayer) {
		println("receive reliabilitylayer")
		if myClient.ReliabilityLayerHandler != nil {
			myClient.ReliabilityLayerHandler(p, re, ra)
		}
	}
	packetReader.SimpleHandler = func(packetType byte, packet *UDPPacket, layers *PacketLayers) {
		if packetType == 0x6 {
			if myClient.Connected {
				return
			}
			println("receive 6!")
			myClient.Connected = true

			response := &Packet07Layer{
				GUID: myClient.GUID,
				MTU: 1492,
				IPAddress: &addr,
			}
			myClient.Writer.WriteSimple(7, response, &addr)
		} else if packetType == 0x8 {
			println("receive 8!")

			response := &Packet09Layer{
				GUID: myClient.GUID,
				Timestamp: uint64(time.Now().Unix()),
				UseSecurity: false,
				Password: []byte{0x37, 0x4F, 0x5E, 0x11, 0x6C, 0x45},
			}
			myClient.Writer.WriteGeneric(context, 9, response, 3, &addr)
		} else {
			println("receive simple unk", packetType)
		}
		if myClient.SimpleHandler != nil {
			myClient.SimpleHandler(packetType, packet, layers)
		}
	}
	packetReader.ReliableHandler = func(packetType byte, packet *UDPPacket, layers *PacketLayers) {
		rakNetLayer := layers.RakNet
		myClient.mustACK = append(myClient.mustACK, int(rakNetLayer.DatagramNumber))
		if myClient.ReliableHandler != nil {
			myClient.ReliableHandler(packetType, packet, layers)
		}
	}
	packetReader.FullReliableHandler = func(packetType byte, packet *UDPPacket, layers *PacketLayers) {
		if packetType == 0x0 {
			mainLayer := layers.Main.(*Packet00Layer)
			response := &Packet03Layer{
				SendPingTime: mainLayer.SendPingTime,
				SendPongTime: mainLayer.SendPingTime + 10,
			}

			myClient.Writer.WriteGeneric(context, 3, response, 3, &addr)
		} else if packetType == 0x10 {
			println("receive 10!")
			mainLayer := layers.Main.(*Packet10Layer)
			nullIP, _ := net.ResolveUDPAddr("udp", "0.0.0.0:0")
			myClient.Address.Port = 0
			response := &Packet13Layer{
				IPAddress: &addr,
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
				SendPingTime: mainLayer.SendPongTime,
				SendPongTime: uint64(time.Now().Unix()),
			}

			myClient.Writer.WriteGeneric(context, 3, response, 3, &addr)
			response90 := &Packet90Layer{
				SchemaVersion: 36,
				RequestedFlags: []string{
					"AllowMoreAngles",
					"UseNewProtocolForStreaming",
					"ReplicatorSupportRegion3Types",
					"SendAdditionalNonAdjustedTimeStamp",
					"SendPlayerGuiEarly2",
					"UseNewPhysicsSender6",
					"FixWeldedHumanoidsDeath",
					"PartColor3Uint8Enabled",
					"ReplicatorUseZstd",
					"BodyColorsColor3PropertyReplicationEnabled",
					"ReplicatorSupportInt64Type",
				},
			}
			myClient.Writer.WriteGeneric(context, 3, response90, 3, &addr)

			response92 := &Packet92Layer{
				PlaceId: 0,
			}
			myClient.Writer.WriteGeneric(context, 3, response92, 3, &addr)

			response8A := &Packet8ALayer{
				PlayerId: myClient.PlayerId,
				ClientTicket: []byte(myClient.clientTicket),
				DataModelHash: []byte(myClient.DataModelHash),
				ProtocolVersion: 36,
				SecurityKey: []byte(myClient.SecurityKey),
				Platform: []byte(myClient.OsPlatform),
				RobloxProductName: []byte("?"),
				SessionId: []byte(myClient.sessionId),
				GoldenHash: myClient.GoldenHash,
			}
			myClient.Writer.WriteGeneric(context, 3, response8A, 3, &addr)

			response8F := &Packet8FLayer{
				SpawnName: "",
			}
			myClient.Writer.WriteGeneric(context, 3, response8F, 3, &addr)
		} else if packetType == 0x81 {
			var players *rbxfile.Instance
			for i := 0; i < len(context.DataModel.Instances); i++ {
				instance := context.DataModel.Instances[i]
				if instance.Name() == "Players" {
					players = instance
					break
				}
			}

			myPlayer := &rbxfile.Instance{
				ClassName: "Player",
				Reference: myClient.scope + "_" + strconv.Itoa(int(myClient.instanceIndex)),
				IsService: false,
				Properties: map[string]rbxfile.Value{
					"Name": rbxfile.ValueString(myClient.UserName),
					"CharacterAppearance": rbxfile.ValueString(myClient.characterAppearance),
					"CharacterAppearanceId": rbxfile.ValueInt(15437777),
					"ChatPrivacyMode": rbxfile.ValueToken{
						Value: 0,
						ID: uint16(context.StaticSchema.EnumsByName["ChatPrivacyMode"]),
						Name: "ChatPrivacyMode",
					},
					"AccountAgeReplicate": rbxfile.ValueInt(0),
					"OsPlatform": rbxfile.ValueString("Win32"),
					"userId": rbxfile.ValueInt(myClient.PlayerId),
					"UserId": rbxfile.ValueInt(myClient.PlayerId),
				},
			}
			players.AddChild(myPlayer)
			myClient.instanceIndex++

			response83 := &Packet83Layer{
				SubPackets: []Packet83Subpacket{
					&Packet83_0B{[]*rbxfile.Instance{myPlayer}},
				},
			}
			myClient.Writer.WriteGeneric(context, 0x83, response83, 3, &addr)
		} else if packetType == 0x83 {
			response := &Packet83Layer{make([]Packet83Subpacket, 0)}
			mainLayer := layers.Main.(*Packet83Layer)
			for _, packet := range mainLayer.SubPackets {
				if Packet83ToType(packet) == 5 {
					response.SubPackets = append(response.SubPackets, &Packet83_06{
						Timestamp: uint64(time.Now().Unix()),
						IsPingBack: true,
					})
				}
			}
			if len(response.SubPackets) > 0 {
				myClient.Writer.WriteGeneric(context, 0x83, response, 3, &addr)
			}
		} else {
			println("receive generic unk", packetType)
		}
		if myClient.FullReliableHandler != nil {
			myClient.FullReliableHandler(packetType, packet, layers)
		}
	}
	packetReader.ErrorHandler = func(err error) {
		println(err.Error())
		if myClient.ErrorHandler != nil {
			myClient.ErrorHandler(err)
		}
	}
	packetReader.Context = context
	println("will set reader", myClient, packetReader)
	myClient.Reader = packetReader

	conn, err := net.DialUDP("udp", nil, &addr)
	defer conn.Close()
	if err != nil {
		return err
	}
	myClient.Address = *conn.LocalAddr().(*net.UDPAddr)

	packetWriter := NewPacketWriter()
	packetWriter.ErrorHandler = func(err error) {
		println(err.Error())
		if myClient.ErrorHandler != nil {
			myClient.ErrorHandler(err)
		}
	}
	packetWriter.OutputHandler = func(payload []byte, dest *net.UDPAddr) {
		num, err := conn.Write(payload)
		if err != nil {
			fmt.Printf("Wrote %d bytes, err: %s\n", num, err.Error())
		}
		if myClient.OutputHandler != nil {
			myClient.OutputHandler(payload, dest)
		}
	}
	myClient.Writer = packetWriter

	connreqpacket := &Packet05Layer{ProtocolVersion: 5}

	go func() {
		for i := 0; i < 5; i++ {
			if myClient.Connected {
				println("successfully dialed")
				return
			}
			myClient.Writer.WriteSimple(5, connreqpacket, &addr)
			time.Sleep(5)
		}
		println("dial failed after 5 attempts")
	}()
	ackTicker := time.NewTicker(17)
	go func() {
		for {
			<- ackTicker.C
			myClient.sendACKs()
		}
	}()

	buf := make([]byte, 1492)
	for {
		n, _, err := conn.ReadFromUDP(buf)
		if err != nil {
			println("read err:", err.Error())
			continue
		}

		myClient.receive(buf[:n])
	}
}
