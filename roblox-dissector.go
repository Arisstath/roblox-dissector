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
import "crypto/tls"
import "strings"
import "os"

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
	0x81: "ID_ROBLOX_INIT_INSTANCES", // ID_SET_GLOBALS
	0x82: "ID_ROBLOX_DICTIONARIES", // ID_TEACH_DESCRIPTOR_DICTIONARIES
	0x83: "ID_ROBLOX_REPLICATION", // ID_DATA
	0x85: "ID_ROBLOX_PHYSICS", // ID_PHYSICS
	0x86: "ID_ROBLOX_TOUCH", // ID_TOUCHES
	0x87: "ID_ROBLOX_CHAT_ALL", // unused
	0x88: "ID_ROBLOX_CHAT_TEAM", // unused
	0x89: "ID_ROBLOX_REPORT_ABUSE",
	0x8A: "ID_ROBLOX_AUTH", // ID_SUBMIT_TICKET
	0x8B: "ID_ROBLOX_CHAT_GAME", // unused
	0x8C: "ID_ROBLOX_CHAT_PLAYER", // unused
	0x8D: "ID_ROBLOX_CLUSTER", // ???
	0x8E: "ID_ROBLOX_PROTOCOL_MISMATCH",
	0x8F: "ID_ROBLOX_INITIAL_SPAWN_NAME",
	0x90: "ID_ROBLOX_REQUEST_PARAMS", // ID_PROTOCOL_SYNC
	0x91: "ID_ROBLOX_NETWORK_SCHEMA", // ID_SCHEMA_SYNC
	0x92: "ID_ROBLOX_VERIFY_PLACEID", // ID_PLACEID_VERIFICATION
	0x93: "ID_ROBLOX_NETWORK_PARAMS", // ID_DICTIONARY_FORMAT
	0x94: "ID_ROBLOX_HASH_REJECTED", // ID_HASH_MISMATCH
	0x95: "ID_ROBLOX_SECURITY_KEY_REJECTED", // ID_SECURITYKEY_MISMATCH
	0x96: "ID_ROBLOX_REQUEST_STATS",
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
	0x8A: ShowPacket8A,
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
			newPacket := peer.UDPPacketFromBytes(payload[:n])
			newPacket.Source = *srcAddr
			newPacket.Destination = *dstAddr
			if payload[0] > 0x8 {
				packetChan <- ProxiedPacket{Packet: newPacket, Payload: payload[:n]}
			} else { // Need priority for join packets
				packetReader.ReadPacket(payload[:n], newPacket)
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
			newPacket := peer.UDPPacketFromBytes(payload[:n])
			newPacket.Source = *srcAddr
			newPacket.Destination = *dstAddr
			if payload[0] > 0x8 {
				packetChan <- ProxiedPacket{Packet: newPacket, Payload: payload[:n]}
			} else { // Need priority for join packets
				packetReader.ReadPacket(payload[:n], newPacket)
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
			newPacket := peer.UDPPacketFromBytes(payload[:n])
			newPacket.Source = *srcAddr
			newPacket.Destination = *dstAddr
			if payload[0] > 0x8 {
				packetChan <- ProxiedPacket{Packet: newPacket, Payload: payload[:n]}
			} else { // Need priority for join packets
				proxyWriter.ProxyClient(payload[:n], newPacket)
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
			newPacket := peer.UDPPacketFromBytes(payload[:n])
			newPacket.Source = *srcAddr
			newPacket.Destination = *dstAddr
			if payload[0] > 0x8 {
				packetChan <- ProxiedPacket{Packet: newPacket, Payload: payload[:n]}
			} else { // Need priority for join packets
				proxyWriter.ProxyServer(payload[:n], newPacket)
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
		req.URL.Host = "8.42.96.30"
		req.URL.Scheme = "https"
		resp, err := transport.RoundTrip(req)
		fmt.Printf("Request: %s/%s %s %v %v\n", req.Host, req.URL.String(), req.Method, req.Header, resp.Header)
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
		println("dumping to", "dumps/" + strings.Replace(req.URL.Path, "/", "_", -1))
		dumpfile, err := os.Create("dumps/" + strings.Replace(req.URL.Path, "/", "_", -1))
		if err != nil {
			println("fail:", err.Error())
			return
		}
		defer dumpfile.Close()
		tee := io.TeeReader(resp.Body, dumpfile)

		io.Copy(w, tee)
	})
	err := http.ListenAndServeTLS(":443", settings.Certfile, settings.Keyfile, nil)
	if err != nil {
		println("listen err:", err.Error())
		return
	}
}

func main() {
	go func() {
		client := peer.NewCustomClient()
		client.SecurityKey = "571cb33a3b024d7b8dafb87156909e92b7eaf86d!1ac9a51ce47836b5c1f65dfc441dfa41"
		client.OsPlatform = "Win32"
		client.GoldenHash = 19857408
		client.DataModelHash = "4b8387d8b57d73944b33dbe044b3707b"
		client.BrowserTrackerId = 9783257674
		err := client.ConnectGuest(12109643, 2)
		if err != nil {
			panic(err)
		}
	}()
	GUIMain()
}
