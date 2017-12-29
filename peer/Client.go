package peer
import "time"
import "net"
import "fmt"
import "sort"
import "net/http"
import "encoding/json"
import "strings"

type PlaceLauncherResponse struct {
	JobId string
	Status int
	JoinScriptUrl string
	AuthenticationUrl string
	AuthenticationTicket string
}
type JoinAshxResponse struct {
	ClientTicket string
	NewClientTicket string
	SessionId string
	MachineAddress string
	ServerPort uint16
	UserId int32
}

type CustomClient struct {
	Context *CommunicationContext
	Reader *PacketReader
	Writer *PacketWriter
	Address net.UDPAddr
	ServerAddress net.UDPAddr
	Connected bool
	MustACK []int
	ClientTicket string
	SessionId string
	PlayerId int32
}

func (client *CustomClient) SendACKs() {
	if len(client.MustACK) == 0 {
		return
	}
	acks := client.MustACK
	client.MustACK = []int{}
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

func (client *CustomClient) Receive(buf []byte) {
	packet := &UDPPacket{
		Stream: BufferToStream(buf),
		Source: client.ServerAddress,
		Destination: client.Address,
	}
	client.Reader.ReadPacket(buf, packet)
}

func StartClient() (*CustomClient, error) {
	context := NewCommunicationContext()
	client := &CustomClient{}
	client.PlayerId = -306579839

	robloxCommClient := &http.Client{}
	placeLauncherRequest, err := http.NewRequest("GET", "https://assetgame.roblox.com/game/PlaceLauncher.ashx?request=RequestGame&browserTrackerId=9783257674&placeId=211851454&isPartyLeader=false&genderId=2", nil)
	if err != nil {
		return nil, err
	}
	placeLauncherRequest.Header.Set("User-Agent", "Roblox/WinInet")
	placeLauncherRequest.Header.Set("Playercount", "0")
	placeLauncherRequest.Header.Set("Requester", "Client")
	placeLauncherRequest.AddCookie(&http.Cookie{Name: "GuestData", Value: "UserId=-306579839"})
	resp, err := robloxCommClient.Do(placeLauncherRequest)
	if err != nil {
		return nil, err
	}
	var plResp PlaceLauncherResponse
	err = json.NewDecoder(resp.Body).Decode(&plResp)
	if err != nil {
		return nil, err
	}
	println("got plresp", plResp.JoinScriptUrl)

	joinScriptRequest, err := http.NewRequest("GET", plResp.JoinScriptUrl, nil)
	if err != nil {
		return nil, err
	}
	userIds := strings.Split(plResp.AuthenticationTicket, ":")
	if len(userIds) < 2 {
		println("Oh no! authtick fetch broke!", plResp.AuthenticationTicket)
	}

	joinScriptRequest.AddCookie(&http.Cookie{Name: "GuestData", Value: "UserId=" + userIds[1] + "&Gender=2"})
	resp, err = robloxCommClient.Do(joinScriptRequest)
	if err != nil {
		return nil, err
	}
	body := resp.Body

	// Discard rbxsig by reading until newline
	char := make([]byte, 1)
	_, err = body.Read(char)
	for err == nil && char[0] != 0x0A {
		_, err = body.Read(char)
	}

	if err != nil {
		return nil, err
	}

	var jsResp JoinAshxResponse
	err = json.NewDecoder(body).Decode(&jsResp)
	if err != nil {
		return nil, err
	}
	println("got jsresp", jsResp.NewClientTicket)
	client.ClientTicket = jsResp.ClientTicket
	client.SessionId = jsResp.SessionId
	client.PlayerId = jsResp.UserId
	addrp, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", jsResp.MachineAddress, jsResp.ServerPort))
	if err != nil {
		return nil, err
	}
	println("dialing addr", addrp.String())
	addr := *addrp // Yes, I'm lazy
	println("addr", addr.IP[0])

	client.Context = context
	client.ServerAddress = addr

	packetReader := NewPacketReader()
	packetReader.ACKHandler = func(packet *UDPPacket, layer *RakNetLayer) {
		println("received ack")
	}
	packetReader.ReliabilityLayerHandler = func(p *UDPPacket, re *ReliabilityLayer, ra *RakNetLayer) {
		println("receive reliabilitylayer")
	}
	packetReader.SimpleHandler = func(packetType byte, packet *UDPPacket, layers *PacketLayers) {
		if packetType == 0x6 {
			if client.Connected {
				return
			}
			println("receive 6!")
			client.Connected = true

			response := &Packet07Layer{
				GUID: 0x1122334455667788, // hahaha
				MTU: 1492,
				IPAddress: &addr,
			}
			client.Writer.WriteSimple(7, response, &addr)
		} else if packetType == 0x8 {
			println("receive 8!")

			response := &Packet09Layer{
				GUID: 0x1122334455667788,
				Timestamp: 117,
				UseSecurity: false,
				Password: []byte{0x37, 0x4F, 0x5E, 0x11, 0x6C, 0x45},
			}
			client.Writer.WriteGeneric(context, 9, response, 3, &addr)
		} else {
			println("receive simple unk", packetType)
		}
	}
	packetReader.ReliableHandler = func(packetType byte, packet *UDPPacket, layers *PacketLayers) {
		rakNetLayer := layers.RakNet
		client.MustACK = append(client.MustACK, int(rakNetLayer.DatagramNumber))
	}
	packetReader.FullReliableHandler = func(packetType byte, packet *UDPPacket, layers *PacketLayers) {
		if packetType == 0x0 {
			mainLayer := layers.Main.(*Packet00Layer)
			response := &Packet03Layer{
				SendPingTime: mainLayer.SendPingTime,
				SendPongTime: mainLayer.SendPingTime + 10,
			}

			client.Writer.WriteGeneric(context, 3, response, 3, &addr)
		} else if packetType == 0x10 {
			println("receive 10!")
			mainLayer := layers.Main.(*Packet10Layer)
			nullIP, _ := net.ResolveUDPAddr("udp", "0.0.0.0:0")
			client.Address.Port = 0
			response := &Packet13Layer{
				IPAddress: &addr,
				Addresses: [10]*net.UDPAddr{
					&client.Address,
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
				SendPongTime: 127,
			}

			client.Writer.WriteGeneric(context, 3, response, 3, &addr)
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
			client.Writer.WriteGeneric(context, 3, response90, 3, &addr)

			response92 := &Packet92Layer{
				UnknownValue: 0,
			}
			client.Writer.WriteGeneric(context, 3, response92, 3, &addr)

			response8A := &Packet8ALayer{
				PlayerId: client.PlayerId,
				ClientTicket: []byte(client.ClientTicket),
				DataModelHash: []byte("4b8387d8b57d73944b33dbe044b3707b"),
				ProtocolVersion: 36,
				SecurityKey: []byte("571cb33a3b024d7b8dafb87156909e92b7eaf86d!1ac9a51ce47836b5c1f65dfc441dfa41"),
				Platform: []byte("Win32"),
				RobloxProductName: []byte("?"),
				SessionId: []byte(client.SessionId),
				GoldenHash: 19857408,
			}
			client.Writer.WriteGeneric(context, 3, response8A, 3, &addr)

			response8F := &Packet8FLayer{
				SpawnName: "",
			}
			client.Writer.WriteGeneric(context, 3, response8F, 3, &addr)
		} else {
			println("receive generic unk", packetType)
		}
	}
	packetReader.ErrorHandler = func(err error) {
		println(err.Error())
	}
	packetReader.Context = context
	client.Reader = packetReader

	conn, err := net.DialUDP("udp", nil, &addr)
	defer conn.Close()
	if err != nil {
		return client, err
	}
	client.Address = *conn.LocalAddr().(*net.UDPAddr)

	packetWriter := NewPacketWriter()
	packetWriter.ErrorHandler = func(err error) {
		println(err.Error())
	}
	packetWriter.OutputHandler = func(payload []byte, dest *net.UDPAddr) {
		num, err := conn.Write(payload)
		if err != nil {
			fmt.Printf("Wrote %d bytes, err: %s\n", num, err.Error())
		}
	}
	client.Writer = packetWriter

	connreqpacket := &Packet05Layer{ProtocolVersion: 5}

	go func() {
		for i := 0; i < 5; i++ {
			if client.Connected {
				println("successfully dialed")
				return
			}
			client.Writer.WriteSimple(5, connreqpacket, &addr)
			time.Sleep(5)
		}
		println("dial failed after 5 attempts")
	}()
	ackTicker := time.NewTicker(17)
	go func() {
		for {
			<- ackTicker.C
			client.SendACKs()
		}
	}()

	buf := make([]byte, 1492)
	for {
		n, _, err := conn.ReadFromUDP(buf)
		if err != nil {
			println("Read err:", err.Error())
			continue
		}

		client.Receive(buf[:n])
	}
}
