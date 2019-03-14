package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/Gskartwii/roblox-dissector/peer"

	windivert "github.com/gskartwii/windivert-go"
)

type ProxiedPacket struct {
	Payload []byte
	Layers  *peer.PacketLayers
}

func captureFromWinDivertProxy(realServerAddr string, captureJobContext context.Context, injectPacket chan peer.RakNetPacket, packetViewer *MyPacketListView, commContext *peer.CommunicationContext) {
	fmt.Printf("Will capture from windivert proxy ??? -> %s\n", realServerAddr)

	dstAddr, err := net.ResolveUDPAddr("udp", realServerAddr)
	if err != nil {
		fmt.Printf("Failed to start proxy: %s", err.Error())
		return
	}

	filter := fmt.Sprintf("(ip.SrcAddr == %s and udp.SrcPort == %d) or (ip.DstAddr == %s and udp.DstPort == %d)",
		dstAddr.IP.String(), dstAddr.Port,
		dstAddr.IP.String(), dstAddr.Port)

	divertConnection, err := windivert.Open(filter, windivert.LayerNetwork, 405, 0)
	if err != nil {
		fmt.Printf("failed to open windivert context: %s\n", err.Error())
		return
	}
	var ifIdx, subIfIdx uint32

	commContext.Server = dstAddr
	proxyWriter := peer.NewProxyWriter(commContext)
	proxyWriter.ServerAddr = dstAddr
	proxyWriter.SecuritySettings = peer.Win10Settings()
	proxyWriter.RuntimeContext, proxyWriter.CancelFunc = context.WithCancel(captureJobContext)

	proxyWriter.ClientHalf.OutputHandler = func(p []byte) { // writes TO client
		err := divertConnection.SendUDP(p, proxyWriter.ServerAddr, proxyWriter.ClientAddr, false, ifIdx, subIfIdx)
		if err != nil {
			fmt.Println("write fail to client %s/%d/%d: %s", proxyWriter.ClientAddr.String(), ifIdx, subIfIdx, err.Error())
			return
		}
	}
	proxyWriter.ServerHalf.OutputHandler = func(p []byte) { // writes TO server
		err := divertConnection.SendUDP(p, proxyWriter.ClientAddr, proxyWriter.ServerAddr, true, ifIdx, subIfIdx)
		if err != nil {
			fmt.Println("write fail to server %d/%d: %s", ifIdx, subIfIdx, err.Error())
			return
		}
	}

	packetChan := make(chan ProxiedPacket, 100)

	go func() {
		var pktSrcAddr, pktDstAddr *net.UDPAddr
		var winDivertAddr *windivert.Address
		var err error
		var udpPayload []byte
		for {
			payload := make([]byte, 1500)
			winDivertAddr, _, err = divertConnection.Recv(payload)
			if err != nil {
				fmt.Printf("divert recv fail: %s\n", err.Error())
				return
			}
			ifIdx = winDivertAddr.InterfaceIndex
			subIfIdx = winDivertAddr.SubInterfaceIndex

			pktSrcAddr, pktDstAddr, udpPayload, err = windivert.ExtractUDP(payload)
			if err != nil {
				fmt.Printf("parse udp fail: %s\n", err.Error())
				return
			}
			// Is this packet for the server?
			// TODO: Better address comparison?
			if proxyWriter.ClientAddr == nil && pktDstAddr.String() == dstAddr.String() {
				// If so, assign the client to the src address
				proxyWriter.ClientAddr = pktSrcAddr
				commContext.Client = pktSrcAddr
			} else if proxyWriter.ClientAddr == nil {
				proxyWriter.ClientAddr = pktDstAddr
				commContext.Client = pktDstAddr
			}

			layers := &peer.PacketLayers{
				Root: peer.RootLayer{
					Source:      pktSrcAddr,
					Destination: pktDstAddr,
					FromClient:  commContext.IsClient(pktSrcAddr),
					FromServer:  commContext.IsServer(pktSrcAddr),
				},
			}
			if payload[0] > 0x8 {
				packetChan <- ProxiedPacket{Layers: layers, Payload: udpPayload}
			} else { // Need priority for join packets
				proxyWriter.ProxyClient(udpPayload, layers)
			}
		}
	}()
	for {
		select {
		case newPacket := <-packetChan:
			if newPacket.Layers.Root.FromClient { // from client? handled by client side
				proxyWriter.ProxyClient(newPacket.Payload, newPacket.Layers)
			} else {
				proxyWriter.ProxyServer(newPacket.Payload, newPacket.Layers)
			}
		case inject := <-injectPacket:
			println("Attempting injection to client")
			injResult, err := proxyWriter.ClientHalf.WritePacket(inject)
			if err != nil {
				println("Error while injecting to client: ", err.Error())
			}
			fmt.Printf("ProxyWriter injection finished: %X\n", injResult)
		case <-captureJobContext.Done():
			proxyWriter.CancelFunc()
			return
		}
	}
	return
}

func autoDetectWinDivertProxy(settings *PlayerProxySettings, captureJobContext context.Context, injectPacket chan peer.RakNetPacket, packetViewer *MyPacketListView, commContext *peer.CommunicationContext) {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	mux := http.NewServeMux()
	acceptNewJoinAshx := true
	proxyContext, proxyContextCancel := context.WithCancel(captureJobContext)

	mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		req.URL.Host = "209.206.41.230"
		req.URL.Scheme = "https"

		if req.URL.Path == "/Game/Join.ashx" {
			//println("patching join.ashx for gzip encoding")
			req.Header.Set("Accept-Encoding", "none")
			// We must use a raw setter, because req.AddCookie() "sanitizes" the values in the wrong way
			req.Header.Set("Cookie", req.Header.Get("Cookie")+"; RBXAppDeviceIdentifier=AppDeviceIdentifier=ROBLOX UWP")
		} else if req.URL.Path == "/Game/PlaceLauncher.ashx" {
			// We must use a raw setter, because req.AddCookie() "sanitizes" the values in the wrong way
			req.Header.Set("Cookie", req.Header.Get("Cookie")+"; RBXAppDeviceIdentifier=AppDeviceIdentifier=ROBLOX UWP")
		}

		resp, err := transport.RoundTrip(req)
		//fmt.Printf("Request: %s/%s %s %v\n%s %v\n", req.Host, req.URL.String(), req.Method, req.Header, resp.Status, resp.Header)
		if err != nil {
			println("error:", err.Error())
			return
		}
		if resp.StatusCode == 403 { // CSRF check fail?
			req.Header.Set("X-Csrf-Token", resp.Header.Get("X-Csrf-Token"))
			//println("Set csrftoken:", resp.Header.Get("X-Csrf-Token"))
			//println("retrying")
			resp, err = transport.RoundTrip(req)
			//fmt.Printf("Request: %s/%s %s %v\n%s %v\n", req.Host, req.URL.String(), req.Method, req.Header, resp.Status, resp.Header)
			if err != nil {
				println("error:", err.Error())
				return
			}
		}
		defer resp.Body.Close()

		for k, vv := range resp.Header {
			for _, v := range vv {
				w.Header().Add(k, v)
			}
		}
		w.WriteHeader(resp.StatusCode)

		if req.URL.Path == "/Game/Join.ashx" && acceptNewJoinAshx { // TODO: Stop using global variables for this
			acceptNewJoinAshx = false
			response, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				println("joinashx err:", err.Error())
				return
			}
			//println("joinashx:", string(response))

			args := regexp.MustCompile(`MachineAddress":"(\d+.\d+.\d+.\d+)","ServerPort":(\d+)`).FindSubmatch(response)

			serverAddr := string(args[1]) + ":" + string(args[2])
			//println("joinashx response:", serverAddr)
			go captureFromWinDivertProxy(serverAddr, proxyContext, injectPacket, packetViewer, commContext)

			w.Write(response)
		} else {
			//println("dumping to", "dumps/" + strings.Replace(req.URL.Path, "/", "_", -1))
			dumpfile, err := os.Create("dumps/" + strings.Replace(req.URL.Path, "/", "_", -1))
			if err != nil {
				println("fail:", err.Error())
				return
			}
			defer dumpfile.Close()
			tee := io.TeeReader(resp.Body, dumpfile)

			io.Copy(w, tee)
		}
	})
	server := &http.Server{Addr: ":443", Handler: mux}

	// HTTP listener must run on its own thread!
	go func() {
		err := server.ListenAndServeTLS(settings.Certfile, settings.Keyfile)
		if err != nil {
			println("listen err:", err.Error())
			return
		}
	}()
	go func() {
		<-captureJobContext.Done()
		println("closing proxy server")
		server.Close()
		proxyContextCancel()
	}()

}
