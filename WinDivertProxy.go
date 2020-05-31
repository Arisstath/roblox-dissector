// +build divert

package main

// WinDivertProxy code will only be included if the "divert" tags is set
// This is because the windivert dependency causes problems on many build platforms
import (
	"context"
	"fmt"
	"net"
	"strings"

	"github.com/Gskartwii/roblox-dissector/peer"

	windivert "github.com/Gskartwii/windivert-go"
	"github.com/olebedev/emitter"
)

const WinDivertEnabled = true

type ProxiedPacket struct {
	Payload []byte
	Layers  *peer.PacketLayers
}

func (captureContext *CaptureContext) CaptureWithDivertedPacket(ctx context.Context, clientAddr *net.UDPAddr, serverAddr *net.UDPAddr, payload []byte, ifIdx uint32, subIfIdx uint32, opened chan struct{}) error {
	filter := fmt.Sprintf("(ip.SrcAddr == %s and udp.SrcPort == %d) or (ip.DstAddr == %s and udp.DstPort == %d)",
		clientAddr.IP.String(), clientAddr.Port,
		clientAddr.IP.String(), clientAddr.Port)

	divertConnection, err := windivert.Open(filter, windivert.LayerNetwork, 405, 0)
	if err != nil {
		return err
	}
	// inform other threads that it is now safe to pass the join.ashx
	if opened != nil {
		close(opened)
	}
	// this must ALWAYS be executed
	// if not, the WinDivert kernel driver may remain loaded
	// even after the application is closed, resulting in WinDivert??.sys
	// being locked
	defer divertConnection.Close()

	proxyWriter := peer.NewProxyWriter(ctx)
	proxyWriter.ClientAddr = clientAddr
	proxyWriter.ServerAddr = serverAddr

	proxyWriter.ClientHalf.Output.On("udp", func(e *emitter.Event) { // writes TO client
		p := e.Args[0].([]byte)
		err := divertConnection.SendUDP(p, proxyWriter.ServerAddr, proxyWriter.ClientAddr, false, ifIdx, subIfIdx)
		if err != nil {
			fmt.Println("write fail to client %s/%d/%d: %s", proxyWriter.ClientAddr.String(), ifIdx, subIfIdx, err.Error())
			return
		}
	}, emitter.Void)
	proxyWriter.ServerHalf.Output.On("udp", func(e *emitter.Event) { // writes TO server
		p := e.Args[0].([]byte)
		err := divertConnection.SendUDP(p, proxyWriter.ClientAddr, proxyWriter.ServerAddr, true, ifIdx, subIfIdx)
		if err != nil {
			fmt.Println("write fail to server %d/%d: %s", ifIdx, subIfIdx, err.Error())
			return
		}
	}, emitter.Void)

	clientConversation := NewProviderConversation(proxyWriter.ClientHalf.DefaultPacketWriter, proxyWriter.ClientHalf.DefaultPacketReader)
	serverConversation := NewProviderConversation(proxyWriter.ServerHalf.DefaultPacketReader, proxyWriter.ServerHalf.DefaultPacketWriter)
	clientConversation.ClientAddress = clientAddr
	serverConversation.ClientAddress = clientAddr
	captureContext.AddConversation(clientConversation)
	captureContext.AddConversation(serverConversation)

	packetChan := make(chan ProxiedPacket, 100)

	divertedLayers := &peer.PacketLayers{
		Root: peer.RootLayer{
			Source:      clientAddr,
			Destination: serverAddr,
			FromClient:  true,
			FromServer:  false,
		},
	}
	proxyWriter.ProxyClient(payload, divertedLayers)
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

			layers := &peer.PacketLayers{
				Root: peer.RootLayer{
					Source:      pktSrcAddr,
					Destination: pktDstAddr,
					// TODO: Can this be improved?
					FromClient: proxyWriter.ClientAddr.String() == pktSrcAddr.String(),
					FromServer: proxyWriter.ServerAddr.String() == pktSrcAddr.String(),
				},
			}
			if payload[0] > 0x8 {
				select {
				case packetChan <- ProxiedPacket{Layers: layers, Payload: udpPayload}:
				case <-ctx.Done():
					return
				}
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
		case <-ctx.Done():
			return nil
		}
	}
	return nil
}

func genPayloadFilter(offset int, bytes []byte) string {
	var build strings.Builder
	build.WriteString(fmt.Sprintf("udp.PayloadLength >= %d", offset+len(bytes)))
	for i := 0; i < len(bytes); i++ {
		build.WriteString(fmt.Sprintf(" and udp.Payload[%d] == 0x%02X", i+offset, bytes[i]))
	}
	return build.String()
}

func (session *CaptureSession) CaptureFromWinDivert() error {
	ctx := session.Context
	captureCtx := session.CaptureContext
	//filter := genPayloadFilter(0, append([]byte{5}, peer.OfflineMessageID...))
	filter := genPayloadFilter(0, []byte{5})

	divertConnection, err := windivert.Open(filter, windivert.LayerNetwork, 405, 0)
	if err != nil {
		return err
	}

	go func() {
		for {
			payload := make([]byte, 1500)
			winDivertAddr, _, err := divertConnection.Recv(payload)
			if err != nil {
				fmt.Printf("divert recv fail: %s\n", err.Error())
				return
			}
			divertConnection.Close()
			ifIdx := winDivertAddr.InterfaceIndex
			subIfIdx := winDivertAddr.SubInterfaceIndex

			pktSrcAddr, pktDstAddr, udpPayload, err := windivert.ExtractUDP(payload)
			if err != nil {
				fmt.Printf("parse udp fail: %s\n", err.Error())
				return
			}
			err = captureCtx.CaptureWithDivertedPacket(ctx, pktSrcAddr, pktDstAddr, udpPayload, ifIdx, subIfIdx, nil)
			if err != nil {
				fmt.Printf("open divert connection fail: %s\n", err.Error())
				return
			}
		}
	}()
	go func() {
		<-ctx.Done()
		divertConnection.Close()
	}()

	return nil
}
