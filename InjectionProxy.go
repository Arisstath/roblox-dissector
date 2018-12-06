package main

import (
	"context"
	"fmt"
	"net"

	"github.com/Gskartwii/roblox-dissector/peer"
)

func captureFromInjectionProxy(src string, dst string, captureJobContext context.Context, injectPacket chan peer.RakNetPacket, packetViewer *MyPacketListView, commContext *peer.CommunicationContext) {
	fmt.Printf("Will capture from injproxy %s -> %s\n", src, dst)

	srcAddr, err := net.ResolveUDPAddr("udp", src)
	if err != nil {
		fmt.Printf("Failed to start proxy: %s", err.Error())
		return
	}
	dstAddr, err := net.ResolveUDPAddr("udp", dst)
	if err != nil {
		fmt.Printf("Failed to start proxy: %s", err.Error())
		return
	}
	conn, err := net.ListenUDP("udp", srcAddr)
	if err != nil {
		fmt.Printf("Failed to start proxy: %s", err.Error())
		return
	}
	defer conn.Close()
	dstConn, err := net.DialUDP("udp", nil, dstAddr)
	if err != nil {
		fmt.Printf("Failed to start proxy: %s", err.Error())
		return
	}
	defer dstConn.Close()

	// srcAddr = client listen address
	// dstAddr = server connection address

	commContext.Client = srcAddr
	commContext.Server = dstAddr
	proxyWriter := peer.NewProxyWriter(commContext)
	proxyWriter.ServerAddr = dstAddr
	proxyWriter.SecuritySettings = peer.Win10Settings()
	proxyWriter.RuntimeContext, proxyWriter.CancelFunc = context.WithCancel(captureJobContext)

	proxyWriter.ClientHalf.OutputHandler = func(p []byte) {
		_, err := conn.WriteToUDP(p, proxyWriter.ClientAddr)
		if err != nil {
			fmt.Println("write fail to client %s: %s", proxyWriter.ClientAddr.String(), err.Error())
			return
		}
	}
	proxyWriter.ServerHalf.OutputHandler = func(p []byte) {
		_, err := dstConn.Write(p)
		if err != nil {
			fmt.Println("write fail to server: %s", err.Error())
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
				return
			}
			layers := &peer.PacketLayers{
				Root: peer.RootLayer{
					Source:      srcAddr,
					Destination: dstAddr,
					FromClient:  true,
				},
			}
			if payload[0] > 0x8 {
				packetChan <- ProxiedPacket{Layers: layers, Payload: payload[:n]}
			} else { // Need priority for join packets
				proxyWriter.ProxyClient(payload[:n], layers)
			}
		}
	}()
	go func() {
		for {
			payload := make([]byte, 1500)
			n, addr, err := dstConn.ReadFromUDP(payload)
			if err != nil {
				fmt.Println("readfromudp fail %s: %s", addr.String(), err.Error())
				return
			}
			layers := &peer.PacketLayers{
				Root: peer.RootLayer{
					Source:      dstAddr,
					Destination: srcAddr,
					FromServer:  true,
				},
			}
			if payload[0] > 0x8 {
				packetChan <- ProxiedPacket{Layers: layers, Payload: payload[:n]}
			} else { // Need priority for join packets
				proxyWriter.ProxyServer(payload[:n], layers)
			}
		}
	}()
	for {
		select {
		case newPacket := <-packetChan:
			if newPacket.Layers.Root.FromClient {
				proxyWriter.ProxyClient(newPacket.Payload, newPacket.Layers)
			} else {
				proxyWriter.ProxyServer(newPacket.Payload, newPacket.Layers)
			}
		case _ = <-injectPacket:
			//proxyWriter.InjectServer(injectedPacket)
			println("Attempt to inject packet not implemented")
		case <-captureJobContext.Done():
			proxyWriter.CancelFunc()
			return
		}
	}
	return
}
