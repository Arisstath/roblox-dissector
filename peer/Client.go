package peer

import "time"
import "net"
import "fmt"
import "net/http"
import "encoding/json"
import "math/rand"
import "errors"
import "encoding/hex"
import "log"
import "github.com/gskartwii/rbxfile"
import "sync"
import "github.com/gskartwii/roblox-dissector/packets"

var LauncherStatuses = [...]string{
	"Wait",
	"Wait (2)",
	"Success",
	"Maintenance",
	"Error",
	"Game ended",
	"Game is full",
	"Roblox is updating",
	"Requesting a server",
	"Unknown 9",
	"User left",
	"Game blocked on this platform",
	"Unauthorized",
}

type placeLauncherResponse struct {
	JobId                string
	Status               int
	JoinScriptUrl        string
	AuthenticationUrl    string
	AuthenticationTicket string
}
type JoinAshxResponse struct {
	ClientTicket          string
	NewClientTicket       string
	SessionId             string
	MachineAddress        string
	ServerPort            uint16
	UserId                int64
	UserName              string
	CharacterAppearance   string
	CharacterAppearanceId int64
	PingInterval          int
	AccountAge            int
}

type SecuritySettings struct {
	RakPassword         []byte
	GoldenHash          uint32
	SecurityKey         string
	DataModelHash       string
	OsPlatform          string
	IdChallengeResponse uint32
	UserAgent           string
}

type CustomClient struct {
	*ConnectedPeer
	Context               *CommunicationContext
	Address               net.UDPAddr
	ServerAddress         net.UDPAddr
	Connected             bool
	clientTicket          string
	sessionId             string
	PlayerId              int64
	UserName              string
	characterAppearance   string
	characterAppearanceId int64
	pingInterval          int
	PlaceId               uint32
	httpClient            *http.Client
	GUID                  uint64
	BrowserTrackerId      uint64
	GenderId              uint8
	IsPartyLeader         bool
	AccountAge            int

	SecuritySettings SecuritySettings

	instanceIndex uint32
	scope         string

	ackTicker      *time.Ticker
	dataPingTicker *time.Ticker

	Connection *net.UDPConn
	Logger     *log.Logger

	handlers         *RawPacketHandlerMap
	dataHandlers     *DataPacketHandlerMap
	instanceHandlers *NewInstanceHandlerMap
	deleteHandlers   *DeleteInstanceHandlerMap
	eventHandlers    *EventHandlerMap

	remoteIndices map[*rbxfile.Instance]uint32
	remoteLock    *sync.Mutex

	LocalPlayer *rbxfile.Instance
}

func (myClient *CustomClient) RegisterPacketHandler(packetType uint8, handler ReceiveHandler) {
	myClient.handlers.Bind(packetType, handler)
}
func (myClient *CustomClient) RegisterDataHandler(packetType uint8, handler DataReceiveHandler) {
	myClient.dataHandlers.Bind(packetType, handler)
}
func (myClient *CustomClient) RegisterInstanceHandler(path *InstancePath, handler NewInstanceHandler) {
	myClient.instanceHandlers.Bind(path, handler)
}

func (myClient *CustomClient) ReadPacket(buf []byte) {
	layers := &PacketLayers{
		Root: RootLayer{
			Source:      &myClient.ServerAddress,
			Destination: &myClient.Address,
			FromServer:  true,
		},
	}
	myClient.ConnectedPeer.ReadPacket(buf, layers)
}

func NewCustomClient() *CustomClient {
	rand.Seed(time.Now().UnixNano())
	context := NewCommunicationContext()

	scope := make([]byte, 0x10)
	n, err := rand.Read(scope)
	if n < 0x10 && err != nil {
		panic(err)
	}

	return &CustomClient{
		ConnectedPeer: NewConnectedPeer(context),
		httpClient:    &http.Client{},
		Context:       context,
		GUID:          rand.Uint64(),
		instanceIndex: 1000,
		scope:         "RBX" + hex.EncodeToString(scope),

		handlers:         NewRawPacketHandlerMap(),
		dataHandlers:     NewDataHandlerMap(),
		instanceHandlers: NewNewInstanceHandlerMap(),
		deleteHandlers:   NewDeleteInstanceHandlerMap(),
		eventHandlers:    NewEventHandlerMap(),

		remoteIndices: make(map[*rbxfile.Instance]uint32),
		remoteLock:    &sync.Mutex{},
	}
}

func (myClient *CustomClient) GetLocalPlayer() *rbxfile.Instance { // may yield! do not call from main thread
	return <-myClient.WaitForInstance("Players", myClient.UserName)
}

// call this asynchronously! it will wait a lot
func (myClient *CustomClient) setupChat() error {
	getInitDataRequest := <-myClient.WaitForInstance("ReplicatedStorage", "DefaultChatSystemChatEvents", "GetInitDataRequest")
	println("got req")

	_, err := myClient.InvokeRemote(getInitDataRequest, []rbxfile.Value{})
	if err != "" {
		return errors.New(err)
	}
	// unimportant
	//myClient.Logger.Printf("chat init data 0: %s\n", initData.String())

	/*_, newMessageChan := myClient.MakeEventChan( // never unbind
		<- myClient.WaitForInstance("ReplicatedStorage", "DefaultChatSystemChatEvents", "OnNewMessage"),
		"OnClientEvent",
	)*/

	_, newFilteredMessageChan := myClient.MakeEventChan(
		<-myClient.WaitForInstance("ReplicatedStorage", "DefaultChatSystemChatEvents", "OnMessageDoneFiltering"),
		"OnClientEvent",
	)

	playerJoinChan := myClient.MakeChildChan(myClient.FindService("Players"))
	playerLeaveChan := myClient.MakeGroupDeleteChan(myClient.FindService("Players").Children)

	for true {
		select {
		case message := <-newFilteredMessageChan:
			dict := message.Arguments[0].(rbxfile.ValueTuple)[0].(rbxfile.ValueDictionary)
			myClient.Logger.Printf("<%s (%s)> %s\n", dict["FromSpeaker"].(rbxfile.ValueString), dict["MessageType"].(rbxfile.ValueString), dict["Message"].(rbxfile.ValueString))
		case player := <-playerJoinChan:
			myClient.Logger.Printf("SYSTEM: %s has joined the game.\n", player.Name())
			playerLeaveChan.AddInstances(player)
		case player := <-playerLeaveChan.C:
			myClient.Logger.Printf("SYSTEM: %s has left the game.\n", player.Name())
		}
	}
	return nil
}

func (myClient *CustomClient) SendChat(message string, toPlayer string, channel string) {
	if channel == "" {
		channel = "All" // assume default channel
	}

	remote := <-myClient.WaitForInstance("ReplicatedStorage", "DefaultChatSystemChatEvents", "SayMessageRequest")
	if toPlayer != "" {
		message = "/w " + toPlayer + " " + message
	}

	myClient.FireRemote(remote, rbxfile.ValueString(message), rbxfile.ValueString(channel))
}

func (myClient *CustomClient) joinWithJoinScript(url string, cookies []*http.Cookie) error {
	joinScriptRequest, err := http.NewRequest("GET", url, nil)
	robloxCommClient := myClient.httpClient
	if err != nil {
		return err
	}

	for _, cook := range cookies {
		if x, _ := joinScriptRequest.Cookie(cook.Name); x == nil {
			joinScriptRequest.AddCookie(cook)
		}
	}

	resp, err := robloxCommClient.Do(joinScriptRequest)
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

	var jsResp JoinAshxResponse
	err = json.NewDecoder(body).Decode(&jsResp)
	if err != nil {
		return err
	}
	myClient.characterAppearance = jsResp.CharacterAppearance
	myClient.characterAppearanceId = jsResp.CharacterAppearanceId
	myClient.clientTicket = jsResp.ClientTicket
	myClient.sessionId = jsResp.SessionId
	myClient.PlayerId = jsResp.UserId
	myClient.UserName = jsResp.UserName
	myClient.pingInterval = jsResp.PingInterval
	myClient.AccountAge = jsResp.AccountAge

	addrp, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", jsResp.MachineAddress, jsResp.ServerPort))
	if err != nil {
		return err
	}

	myClient.Logger.Println("Connecting to", jsResp.MachineAddress)

	myClient.ServerAddress = *addrp
	return myClient.rakConnect()
}

func (myClient *CustomClient) joinWithPlaceLauncher(url string, cookies []*http.Cookie) error {
	var plResp placeLauncherResponse
	var resp *http.Response
	robloxCommClient := myClient.httpClient
	for i := 0; i < 5; i++ {
		myClient.Logger.Println("requesting placelauncher", url, "attempt", i)
		placeLauncherRequest, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return err
		}
		placeLauncherRequest.Header.Set("User-Agent", myClient.SecuritySettings.UserAgent)
		for _, cookie := range cookies {
			placeLauncherRequest.AddCookie(cookie)
		}

		resp, err = robloxCommClient.Do(placeLauncherRequest)
		if err != nil {
			return err
		}
		err = json.NewDecoder(resp.Body).Decode(&plResp)
		if err != nil {
			return err
		}
		if plResp.Status == 0 || plResp.Status == 1 { // status: wait
			myClient.Logger.Println("status=0 --> retrying after 5s")
			time.Sleep(time.Second * 5)
			continue
		} else if plResp.Status != 2 { // status: success
			myClient.Logger.Println("failed to connect, reason: ", LauncherStatuses[plResp.Status])
			return errors.New("PlaceLauncher returned fatal status")
		}

		if plResp.JoinScriptUrl == "" {
			myClient.Logger.Println("joinscript failure, status", plResp.Status)
			return errors.New("couldn't get joinscripturl")
		} else {
			break
		}
	}

	for _, cook := range resp.Cookies() {
		cookies = append(cookies, cook)
	}

	return myClient.joinWithJoinScript(plResp.JoinScriptUrl, cookies)
}

func (myClient *CustomClient) ConnectWithAuthTicket(placeId uint32, ticket string) error {
	myClient.PlaceId = placeId
	robloxCommClient := myClient.httpClient
	negotiationRequest, err := http.NewRequest("POST", "https://www.roblox.com/Login/Negotiate.ashx?suggest="+ticket, nil)
	if err != nil {
		return err
	}

	negotiationRequest.Header.Set("Playercount", "0")
	negotiationRequest.Header.Set("Requester", "Client")
	negotiationRequest.Header.Set("User-Agent", myClient.SecuritySettings.UserAgent)
	negotiationRequest.Header.Set("Content-Length", "0")
	negotiationRequest.Header.Set("X-Csrf-Token", "")
	negotiationRequest.Header.Set("Rbxauthenticationnegotiation", "www.roblox.com")
	negotiationRequest.Header.Set("Host", "www.roblox.com")

	resp, err := robloxCommClient.Do(negotiationRequest)
	if err != nil {
		return err
	}
	if resp.StatusCode == 403 { // token verification failed
		negotiationRequest.Header.Set("X-Csrf-Token", resp.Header.Get("X-Csrf-Token"))
		myClient.Logger.Println("Set csrftoken:", resp.Header.Get("X-Csrf-Token"))

		resp, err = robloxCommClient.Do(negotiationRequest)
		if err != nil {
			return err
		}
		if resp.StatusCode == 403 {
			return errors.New("couldn't negotiate ticket: " + resp.Status)
		}
	}
	cookies := resp.Cookies()

	return myClient.joinWithPlaceLauncher(fmt.Sprintf("https://www.roblox.com/game/PlaceLauncher.ashx?request=RequestGame&browserTrackerId=%d&placeId=%d&isPartyLeader=false&genderId=%d", myClient.BrowserTrackerId, myClient.PlaceId, myClient.GenderId), cookies)
}

// Automatically fills in any needed hashes/key for Windows 10 clients
func (settings *SecuritySettings) InitWin10() {
	settings.SecurityKey = "2e427f51c4dab762fe9e3471c6cfa1650841723b!c41c6af331e84e7d96dca37f759078fa\002"
	settings.OsPlatform = "Windows_Universal"
	settings.GoldenHash = 0xC001CAFE
	settings.DataModelHash = "ios,ios"
	settings.UserAgent = "Roblox/WinINet"
	settings.IdChallengeResponse = 0x512265E4
}

// Automatically fills in any needed hashes/key for Android clients
func (settings *SecuritySettings) InitAndroid() {
	// good job on your security :-p
	settings.SecurityKey = "2e427f51c4dab762fe9e3471c6cfa1650841723b!10ddf3176164dab2c7b4ba9c0e986001"
	settings.OsPlatform = "Android"
	settings.GoldenHash = 0xC001CAFE
	settings.DataModelHash = "ios,ios"
	settings.UserAgent = "Mozilla/5.0 (512MB; 576x480; 300x300; 300x300; Samsung Galaxy S8; 6.0.1 Marshmallow) AppleWebKit/537.36 (KHTML, like Gecko) Roblox Android App 0.334.0.195932 Phone Hybrid()"
	settings.IdChallengeResponse = 0xBA4CE5C4
}

// genderId 1 ==> default genderless
// genderId 2 ==> Billy
// genderId 3 ==> Betty
func (myClient *CustomClient) ConnectGuest(placeId uint32, genderId uint8) error {
	myClient.PlaceId = placeId
	myClient.GenderId = genderId
	return myClient.joinWithPlaceLauncher(fmt.Sprintf("https://assetgame.roblox.com/game/PlaceLauncher.ashx?request=RequestGame&browserTrackerId=%d&placeId=%d&isPartyLeader=false&genderId=%d", myClient.BrowserTrackerId, myClient.PlaceId, myClient.GenderId), []*http.Cookie{})
}

func (myClient *CustomClient) defaultAckHandler(layers *PacketLayers) {
	if myClient.ACKHandler != nil {
		myClient.ACKHandler(layers)
	}
}
func (myClient *CustomClient) defaultReliabilityLayerHandler(layers *PacketLayers) {
	myClient.mustACK = append(myClient.mustACK, int(layers.RakNet.DatagramNumber))
	if myClient.ReliabilityLayerHandler != nil {
		myClient.ReliabilityLayerHandler(layers)
	}
}
func (myClient *CustomClient) defaultSimpleHandler(packetType byte, layers *PacketLayers) {
	if myClient.SimpleHandler != nil {
		myClient.SimpleHandler(packetType, layers)
	}
	myClient.handlers.Fire(packetType, layers)
}
func (myClient *CustomClient) defaultReliableHandler(packetType byte, layers *PacketLayers) {
	if myClient.ReliableHandler != nil {
		myClient.ReliableHandler(packetType, layers)
	}
}
func (myClient *CustomClient) defaultFullReliableHandler(packetType byte, layers *PacketLayers) {
	if myClient.FullReliableHandler != nil {
		myClient.FullReliableHandler(packetType, layers)
	}
	if layers.Main != nil {
		myClient.handlers.Fire(packetType, layers)
	}
}

func (myClient *CustomClient) createReader() {
	myClient.ACKHandler = myClient.defaultAckHandler
	myClient.ReliabilityLayerHandler = myClient.defaultReliabilityLayerHandler
	myClient.SimpleHandler = myClient.defaultSimpleHandler
	myClient.ReliableHandler = myClient.defaultReliableHandler
	myClient.FullReliableHandler = myClient.defaultFullReliableHandler

	myClient.DefaultPacketReader.SetContext(myClient.Context)
}

func (myClient *CustomClient) createWriter() {
	myClient.OutputHandler = func(payload []byte) {
		num, err := myClient.Connection.Write(payload)
		if err != nil {
			fmt.Printf("Wrote %d bytes, err: %s\n", num, err.Error())
		}
	}
}

func (myClient *CustomClient) dial() {
	connreqpacket := &ConnectionRequest1{ProtocolVersion: 5, maxLength: 1492}
	go func() {
		for i := 0; i < 5; i++ {
			if myClient.Connected {
				myClient.Logger.Println("successfully dialed")
				return
			}
			myClient.WriteSimple(connreqpacket)
			time.Sleep(5 * time.Second)
			if i > 2 {
				connreqpacket.maxLength = 576 // try smaller mtu, is this why our packets are getting lost?
			}
		}
		myClient.Logger.Println("dial failed after 5 attempts")
	}()
}

func (myClient *CustomClient) startAcker() {
	myClient.ackTicker = time.NewTicker(500 * time.Millisecond)
	go func() {
		for {
			<-myClient.ackTicker.C
			myClient.sendACKs()
		}
	}()

}
func (myClient *CustomClient) startDataPing() {
	// boot up dataping
	myClient.dataPingTicker = time.NewTicker(time.Duration(myClient.pingInterval) * time.Millisecond)
	go func() {
		for {
			<-myClient.dataPingTicker.C

			myClient.WritePacket(&Packet83Layer{
				[]Packet83Subpacket{&Packet83_05{
					SendStats:  8,
					Timestamp:  uint64(time.Now().Unix()),
					IsPingBack: false,
				}},
			})
		}
	}()
}

func (myClient *CustomClient) mainReadLoop() error {
	buf := make([]byte, 1492)
	for {
		n, _, err := myClient.Connection.ReadFromUDP(buf)
		if err != nil {
			myClient.Logger.Println("fatal read err:", err.Error(), "read", n, "bytes")
			return err // a read error may be a sign that the connection was closed
			// hence we can't run this loop anymore; we would get infinitely many errors
		}

		myClient.ReadPacket(buf[:n])
	}
}

func (myClient *CustomClient) disconnectInternal() error {
	myClient.ackTicker.Stop()
	myClient.dataPingTicker.Stop()
	return myClient.Connection.Close()
}

func (myClient *CustomClient) Disconnect() {
	myClient.WritePacket(&Packet15Layer{
		Reason: 0xFFFFFFFF,
	})

	myClient.disconnectInternal()
}

func (myClient *CustomClient) rakConnect() error {
	var err error
	addr := myClient.ServerAddress

	myClient.createReader()
	myClient.bindDefaultHandlers()
	myClient.Connection, err = net.DialUDP("udp", nil, &addr)
	defer myClient.Connection.Close()
	if err != nil {
		return err
	}
	myClient.Address = *myClient.Connection.LocalAddr().(*net.UDPAddr)
	myClient.createWriter()

	myClient.dial()
	myClient.startAcker()

	go myClient.setupChat() // needs to run async; does a lot of waiting for instances

	return myClient.mainReadLoop()
}
