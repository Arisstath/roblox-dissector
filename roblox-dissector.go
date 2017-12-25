package main
import "github.com/google/gopacket"
import "github.com/google/gopacket/pcap"
import "github.com/google/gopacket/layers"
import "github.com/fatih/color"
import "fmt"
import "github.com/gskartwii/roblox-dissector/peer"
import "time"
import "net"
import "net/http"
import "io"
import "io/ioutil"
import "crypto/tls"
import "compress/gzip"
import "regexp"
import "bytes"
import "strconv"

const DEBUG bool = false

var PacketNames map[byte]string = map[byte]string{
	0xFF: "???",
	0x00: "ID_CONNECTED_PING",
	0x01: "ID_UNCONNECTED_PING",
	0x03: "ID_CONNECTED_PONG",
	0x04: "ID_DETECT_LOST_CONNECTIONS",
	0x05: "ID_OPEN_CONNECTION_REQUEST_1",
	0x06: "ID_OPEN_CONNECTION_REPLY_1",
	0x07: "ID_OPEN_CONNECTION_REQUEST_2",
	0x08: "ID_OPEN_CONNECTION_REPLY_2",
	0x09: "ID_CONNECTION_REQUEST",
	0x10: "ID_CONNECTION_REQUEST_ACCEPTED",
	0x11: "ID_CONNECTION_ATTEMPT_FAILED",
	0x13: "ID_NEW_INCOMING_CONNECTION",
	0x15: "ID_DISCONNECTION_NOTIFICATION",
	0x18: "ID_INVALID_PASSWORD",
	0x1B: "ID_TIMESTAMP",
	0x1C: "ID_UNCONNECTED_PONG",
	0x81: "ID_ROBLOX_INIT_INSTANCES",
	0x82: "ID_ROBLOX_DICTIONARIES",
	0x83: "ID_ROBLOX_REPLICATION",
	0x85: "ID_ROBLOX_PHYSICS",
	0x86: "ID_ROBLOX_TOUCH",
	0x89: "ID_ROBLOX_REPORT_ABUSE",
	0x8A: "ID_ROBLOX_AUTH",
	0x8E: "ID_ROBLOX_PROTOCOL_MISMATCH",
	0x8F: "ID_ROBLOX_INITIAL_SPAWN_NAME",
	0x90: "ID_ROBLOX_SCHEMA_VERSION",
	0x91: "ID_ROBLOX_NETWORK_SCHEMA",
	0x92: "ID_ROBLOX_START_AUTH_THREAD",
	0x93: "ID_ROBLOX_NETWORK_PARAMS",
	0x94: "ID_ROBLOX_HASH_REJECTED",
	0x95: "ID_ROBLOX_SECURITY_KEY_REJECTED",
	0x97: "ID_ROBLOX_NEW_SCHEMA",
}

type ActivationCallback func(byte, *peer.UDPPacket, *peer.CommunicationContext, *peer.PacketLayers)
var ActivationCallbacks map[byte]ActivationCallback = map[byte]ActivationCallback{
	0x05: ShowPacket05,
	0x06: ShowPacket06,
	0x07: ShowPacket07,
	0x08: ShowPacket08,
	0x09: ShowPacket09,
	0x10: ShowPacket10,
	0x13: ShowPacket13,
	0x00: ShowPacket00,
	0x03: ShowPacket03,

	0x93: ShowPacket93,
	//0x8A: ShowPacket8A,
	0x82: ShowPacket82,
	0x92: ShowPacket92,
	0x90: ShowPacket90,
	0x8F: ShowPacket8F,
	0x81: ShowPacket81,
	0x83: ShowPacket83,
    0x97: ShowPacket97,
	0x85: ShowPacket85,
	0x86: ShowPacket86,
}

func captureJob(handle *pcap.Handle, useIPv4 bool, stopCaptureJob chan struct{}, packetViewer *MyPacketListView, context *peer.CommunicationContext) {
	handle.SetBPFFilter("udp")
	var packetSource *gopacket.PacketSource
	if useIPv4 {
		packetSource = gopacket.NewPacketSource(handle, layers.LayerTypeIPv4)
	} else {
		packetSource = gopacket.NewPacketSource(handle, handle.LinkType())
	}
	packetChannel := make(chan gopacket.Packet, 0x100) // Buffer the packets to avoid dropping them

	go func() {
		for packet := range packetSource.Packets() {
			packetChannel <- packet
		}
	}()

	packetReader := peer.NewPacketReader()
	packetReader.SimpleHandler = func(packetType byte, packet *peer.UDPPacket, layers *peer.PacketLayers) {
		if context.IsValid {
			packetViewer.AddFullPacket(packetType, packet, context, layers, ActivationCallbacks[packetType])
		}
	}
	packetReader.ReliableHandler = func(packetType byte, packet *peer.UDPPacket, layers *peer.PacketLayers) {
		if context.IsValid {
			packetViewer.AddSplitPacket(packetType, packet, context, layers)
		}
	}
	packetReader.FullReliableHandler = func(packetType byte, packet *peer.UDPPacket, layers *peer.PacketLayers) {
		if context.IsValid {
			packetViewer.BindCallback(packetType, packet, context, layers, ActivationCallbacks[packetType])
		}
	}
	packetReader.ReliabilityLayerHandler = func(p *peer.UDPPacket, re *peer.ReliabilityLayer, ra *peer.RakNetLayer) {
		// nop
	}
	packetReader.ACKHandler = func(p *peer.UDPPacket, ra *peer.RakNetLayer) {
		// nop
	}
	packetReader.ErrorHandler = func(err error) {
		println(err.Error())
	}
	packetReader.Context = context

	for true {
		select {
		case _ = <- stopCaptureJob:
			context.IsValid = false
			return
		case packet := <- packetChannel:
			if packet.ApplicationLayer() == nil {
				color.Red("Ignoring packet because ApplicationLayer can't be decoded")
				continue
			}
			payload := packet.ApplicationLayer().Payload()
			if len(payload) == 0 {
				println("Ignoring 0 payload")
				continue
			}
			if context.Client == "" && payload[0] != 5 {
				println("Ignoring non5")
				continue
			}
			newPacket := peer.UDPPacketFromGoPacket(packet)
			if newPacket == nil {
				continue
			}
			if context.Client != "" && !context.IsClient(newPacket.Source) && !context.IsServer(newPacket.Source) {
				continue
			}

			if newPacket != nil {
				packetReader.ReadPacket(payload, newPacket)
			}
		}
	}
	return
}

func captureFromFile(filename string, useIPv4 bool, stopCaptureJob chan struct{}, packetViewer *MyPacketListView, context *peer.CommunicationContext) {
	fmt.Printf("Will capture from file %s\n", filename)
	handle, err := pcap.OpenOffline(filename)
	if err != nil {
		println(err.Error())
		return
	}
	captureJob(handle, useIPv4, stopCaptureJob, packetViewer, context)
}

func captureFromLive(livename string, useIPv4 bool, usePromisc bool, stopCaptureJob chan struct{}, packetViewer *MyPacketListView, context *peer.CommunicationContext) {
	fmt.Printf("Will capture from live device %s\n", livename)
	handle, err := pcap.OpenLive(livename, 2000, usePromisc, 10 * time.Second)
	if err != nil {
		println(err.Error())
		return
	}
	captureJob(handle, useIPv4, stopCaptureJob, packetViewer, context)
}

type ProxiedPacket struct {
	Packet *peer.UDPPacket
	Payload []byte
}
func captureFromProxy(src string, dst string, stopCaptureJob chan struct{}, packetViewer *MyPacketListView, context *peer.CommunicationContext) {
	fmt.Printf("Will capture from proxy %s -> %s\n", src, dst)

	srcAddr, _ := net.ResolveUDPAddr("udp", src)
	dstAddr, _ := net.ResolveUDPAddr("udp", dst)
	conn, err := net.ListenUDP("udp", srcAddr)
	if err != nil {
		fmt.Printf("Failed to start proxy: %s", err.Error())
		return
	}
	dstConn, err := net.DialUDP("udp", nil, dstAddr)
	if err != nil {
		fmt.Printf("Failed to start proxy: %s", err.Error())
		return
	}

	packetReader := peer.NewPacketReader()
	packetReader.SimpleHandler = func(packetType byte, packet *peer.UDPPacket, layers *peer.PacketLayers) {
		if context.IsValid {
			packetViewer.AddFullPacket(packetType, packet, context, layers, ActivationCallbacks[packetType])
		}
	}
	packetReader.ReliableHandler = func(packetType byte, packet *peer.UDPPacket, layers *peer.PacketLayers) {
		if context.IsValid {
			packetViewer.AddSplitPacket(packetType, packet, context, layers)
		}
	}
	packetReader.FullReliableHandler = func(packetType byte, packet *peer.UDPPacket, layers *peer.PacketLayers) {
		if context.IsValid {
			packetViewer.BindCallback(packetType, packet, context, layers, ActivationCallbacks[packetType])
		}
	}
	packetReader.ReliabilityLayerHandler = func(p *peer.UDPPacket, re *peer.ReliabilityLayer, ra *peer.RakNetLayer) {
		// nop
	}
	packetReader.ACKHandler = func(p *peer.UDPPacket, ra *peer.RakNetLayer) {
		// nop
	}
	packetReader.ErrorHandler = func(err error) {
		println(err.Error())
	}
	packetReader.Context = context

	var clientAddr *net.UDPAddr
	var n int
	packetChan := make(chan ProxiedPacket, 100)

	go func() {
		for {
			payload := make([]byte, 1500)
			n, clientAddr, err = conn.ReadFromUDP(payload)
			if err != nil {
				fmt.Println("readfromudp fail: %s", err.Error())
				continue
			}
			_, err = dstConn.Write(payload[:n])
			if err != nil {
				fmt.Println("write fail: %s", err.Error())
				continue
			}
			newPacket := peer.UDPPacket{
				peer.BufferToStream(payload[:n]),
				*srcAddr,
				*dstAddr,
			}
			if payload[0] > 0x8 {
				packetChan <- ProxiedPacket{Packet: &newPacket, Payload: payload[:n]}
			} else { // Need priority for join packets
				packetReader.ReadPacket(payload[:n], &newPacket)
			}
		}
	}()
	go func() {
		for {
			payload := make([]byte, 1500)
			n, _, err := dstConn.ReadFromUDP(payload)
			if err != nil {
				fmt.Println("readfromudp fail: %s", err.Error())
				continue
			}
			_, err = conn.WriteToUDP(payload[:n], clientAddr)
			if err != nil {
				fmt.Println("write fail: %s", err.Error())
				continue
			}
			newPacket := peer.UDPPacket{
				peer.BufferToStream(payload[:n]),
				*dstAddr,
				*srcAddr,
			}
			if payload[0] > 0x8 {
				packetChan <- ProxiedPacket{Packet: &newPacket, Payload: payload[:n]}
			} else { // Need priority for join packets
				packetReader.ReadPacket(payload[:n], &newPacket)
			}
		}
	}()

	for {
		select {
		case newPacket := <- packetChan:
			packetReader.ReadPacket(newPacket.Payload, newPacket.Packet)
		case _ = <- stopCaptureJob:
			return
		}
	}
	return
}

func captureFromInjectionProxy(src string, dst string, stopCaptureJob chan struct{}, injectPacket chan peer.RakNetPacket, packetViewer *MyPacketListView, context *peer.CommunicationContext) {
	fmt.Printf("Will capture from injproxy %s -> %s\n", src, dst)

	srcAddr, _ := net.ResolveUDPAddr("udp", src)
	dstAddr, _ := net.ResolveUDPAddr("udp", dst)
	conn, err := net.ListenUDP("udp", srcAddr)
	if err != nil {
		fmt.Printf("Failed to start proxy: %s", err.Error())
		return
	}
	dstConn, err := net.DialUDP("udp", nil, dstAddr)
	if err != nil {
		fmt.Printf("Failed to start proxy: %s", err.Error())
		return
	}

	context.Client = srcAddr.String()
	context.Server = dstAddr.String()
	proxyWriter := peer.NewProxyWriter(context)
	proxyWriter.ServerAddr = dstAddr

	proxyWriter.ClientHalf.OutputHandler = func(p []byte, d *net.UDPAddr) {
		_, err := conn.WriteToUDP(p, d)
		if err != nil {
			fmt.Println("write fail: %s", err.Error())
			return
		}
	}
	proxyWriter.ServerHalf.OutputHandler = func(p []byte, d *net.UDPAddr) {
		_, err := dstConn.Write(p)
		if err != nil {
			fmt.Println("write fail: %s", err.Error())
			return
		}
	}

	var n int
	packetChan := make(chan ProxiedPacket, 100)

	go func() {
		for {
			payload := make([]byte, 1500)
			n, proxyWriter.ClientAddr, err = conn.ReadFromUDP(payload)
			if err != nil {
				fmt.Println("readfromudp fail: %s", err.Error())
				continue
			}
			newPacket := peer.UDPPacket{
				peer.BufferToStream(payload[:n]),
				*srcAddr,
				*dstAddr,
			}
			if payload[0] > 0x8 {
				packetChan <- ProxiedPacket{Packet: &newPacket, Payload: payload[:n]}
			} else { // Need priority for join packets
				proxyWriter.ProxyClient(payload[:n], &newPacket)
			}
		}
	}()
	go func() {
		for {
			payload := make([]byte, 1500)
			n, _, err := dstConn.ReadFromUDP(payload)
			if err != nil {
				fmt.Println("readfromudp fail: %s", err.Error())
				continue
			}
			newPacket := peer.UDPPacket{
				peer.BufferToStream(payload[:n]),
				*dstAddr,
				*srcAddr,
			}
			if payload[0] > 0x8 {
				packetChan <- ProxiedPacket{Packet: &newPacket, Payload: payload[:n]}
			} else { // Need priority for join packets
				proxyWriter.ProxyServer(payload[:n], &newPacket)
			}
		}
	}()
	for {
		select {
		case newPacket := <- packetChan:
			if newPacket.Packet.Source.String() == srcAddr.String() {
				proxyWriter.ProxyClient(newPacket.Payload, newPacket.Packet)
			} else {
				proxyWriter.ProxyServer(newPacket.Payload, newPacket.Packet)
			}
		case injectedPacket := <- injectPacket:
			proxyWriter.InjectServer(injectedPacket)
		case _ = <- stopCaptureJob:
			return
		}
	}
	return
}

// Requires you to patch the player with memcheck bypass and rbxsig ignoring! But it could work...
func captureFromPlayerProxy(settings *PlayerProxySettings, stopCaptureJob chan struct{}, injectPacket chan peer.RakNetPacket, packetViewer *MyPacketListView, context *peer.CommunicationContext) {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		fmt.Printf("Request: %s/%s %s %v\n", req.Host, req.URL.String(), req.Method, req.Header)
		req.URL.Host = "8.42.96.30"
		req.URL.Scheme = "https"

		if req.URL.Path == "/Game/Join.ashx" {
			println("patching join.ashx gzip")
			req.Header.Set("Accept-Encoding", "none")
		}

		resp, err := transport.RoundTrip(req)
		if err != nil {
			println("error:", err.Error())
			return
		}
		defer resp.Body.Close()

		for k, vv := range resp.Header {
			for _, v := range vv {
				w.Header().Add(k, v)
			}
		}

		if req.URL.Path == "/Game/Join.ashx" {
			w.Header().Set("Content-Encoding", "gzip")
			response, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				println("joinashx err:", err.Error())
				return
			}

			newBuffer := bytes.NewBuffer(make([]byte, 0, len(response)))
			//result := regexp.MustCompile(`MachineAddress":"\d+.\d+.\d+.\d+","ServerPort":\d+`).ReplaceAll(response, []byte(`MachineAddress":"127.0.0.1","ServerPort":53640`))
			result := []byte(`--rbxsig%QtNI8SiXF7yjeJg6d4S2J8Jo5PeMtgp7xwVM/VMkoMLRb49dWV/8O7NEU8xrFidalt/bQ3/3FLsBupknl2fJLm1MPRSTDK5NCYjzwqn6xOrU/yt7ZUsOA7UHUHRqc4Fq/8wBleW/sbc07evcnP0F/ukFjh/NgYq5u0LV58jjSzs=%
{"ClientPort":0,"MachineAddress":"localhost","ServerPort":53640,"PingUrl":"","PingInterval":300,"UserName":"Player","SeleniumTestMode":false,"UserId":0,"SuperSafeChat":true,"CharacterAppearance":"https://api.sitetest3.robloxlabs.com/v1.1/avatar-fetch/?placeId=0&userId=0","ClientTicket":"11/12/2017 3:47:30 AM;c3MUdaYefbSrJEd51QITWKVPMea12cMN2vwWFahiBKaiTTvM280xDTabnOsFsmQLgvhqLeVAgcvAzAtGgVTLWl64BVzvQA7FOaXqjgaFWwtt1I8Zu37nnsp00PNBwtKuUeZ5uTe2e2U/S6Z2xy9zSCxMuLcC78Yua/5Y6ppAkE4=;bfMULtcz6pXut6DQyGm2K7IWamsW+pBFlWwfQXVwgqJWvd+c85wRQIG0VOhuznN0Oqjgalt47qjAxJbflnPqyb02Jrh+hbmKpRe/VoQWJxPC6i1atq/fpiO5lO8ysLOgxR42I4BxHuBDN59zS9GAkzhqckLnSqvty4bmQ7tIN48=","NewClientTicket":"11/12/2017 3:47:30 AM;j0wgKOqDExCg/mlccmtbzEFS4GayxlGb3w3b9liZhTbPt07YDkhqda3+hcVHileH5tu3U6V6+e7/vsv992lTtRjWz9n+HqT7aECecfyMtmC7dEitpmgjgDChMX+TS43Kp2aEfLizRWkGRQxBeDH21x8OfaLiqDCRBbgP29Fl0bU=;bfMULtcz6pXut6DQyGm2K7IWamsW+pBFlWwfQXVwgqJWvd+c85wRQIG0VOhuznN0Oqjgalt47qjAxJbflnPqyb02Jrh+hbmKpRe/VoQWJxPC6i1atq/fpiO5lO8ysLOgxR42I4BxHuBDN59zS9GAkzhqckLnSqvty4bmQ7tIN48=","GameId":"00000000-0000-0000-0000-000000000000","PlaceId":0,"MeasurementUrl":"","WaitingForCharacterGuid":"774fc427-1665-4cb8-b0e5-50618ead81ce","BaseUrl":"http://assetgame.sitetest3.robloxlabs.com/","ChatStyle":"Classic","VendorId":0,"ScreenShotInfo":"","VideoInfo":"<?xml version=\"1.0\"?><entry xmlns=\"http://www.w3.org/2005/Atom\" xmlns:media=\"http://search.yahoo.com/mrss/\" xmlns:yt=\"http://gdata.youtube.com/schemas/2007\"><media:group><media:title type=\"plain\"><![CDATA[ROBLOX Place]]></media:title><media:description type=\"plain\"><![CDATA[ For more games visit http://www.roblox.com]]></media:description><media:category scheme=\"http://gdata.youtube.com/schemas/2007/categories.cat\">Games</media:category><media:keywords>ROBLOX, video, free game, online virtual world</media:keywords></media:group></entry>","CreatorId":0,"CreatorTypeEnum":"User","MembershipType":"None","AccountAge":0,"CookieStoreFirstTimePlayKey":"rbx_evt_ftp","CookieStoreFiveMinutePlayKey":"rbx_evt_fmp","CookieStoreEnabled":true,"IsRobloxPlace":false,"GenerateTeleportJoin":false,"IsUnknownOrUnder13":true,"GameChatType":"NoOne","SessionId":"87aa47bb-5eee-4599-982b-da0b07c913ba|00000000-0000-0000-0000-000000000000|0|109.240.79.235|5|2017-11-12T09:47:31.1863008Z|0|null|null|null|null|null|null","DataCenterId":0,"UniverseId":0,"BrowserTrackerId":0,"UsePortraitMode":false,"FollowUserId":0,"characterAppearanceId":0}`)

			args := regexp.MustCompile(`MachineAddress":"(\d+.\d+.\d+.\d+)","ServerPort":(\d+)`).FindSubmatch(response)

			serverAddr := string(args[1]) + ":" + string(args[2])
			go captureFromInjectionProxy("127.0.0.1:53640", serverAddr, stopCaptureJob, injectPacket, packetViewer, context)

			compressStream := gzip.NewWriter(newBuffer)

			_, err = compressStream.Write(result)
			if err != nil {
				println("joinashx gz w err:", err.Error())
				return
			}
			err = compressStream.Close()
			if err != nil {
				println("joinashx gz close err:", err.Error())
				return
			}
			w.Header().Set("Content-Length", strconv.Itoa(newBuffer.Len()))
			w.WriteHeader(resp.StatusCode)

			w.Write(newBuffer.Bytes())
		} else {
			io.Copy(w, resp.Body)
		}
	})
	err := http.ListenAndServeTLS(":443", settings.Certfile, settings.Keyfile, nil)
	if err != nil {
		println("listen err:", err.Error())
		return
	}
}

func main() {
	GUIMain()
}
