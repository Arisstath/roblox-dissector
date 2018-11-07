package main

import (
	"context"
	"fmt"
	"time"

	"github.com/Gskartwii/roblox-dissector/peer"

	"github.com/fatih/color"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
)

// TODO: Can this use ConnectedPeer?
func captureJob(handle *pcap.Handle, useIPv4 bool, captureJobContext context.Context, packetViewer *MyPacketListView, context *peer.CommunicationContext) {
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

	clientPacketReader := peer.NewPacketReader()
	clientPacketReader.SimpleHandler = func(packetType byte, packet *peer.UDPPacket, layers *peer.PacketLayers) {
		packetViewer.AddFullPacket(packetType, packet, context, layers, ActivationCallbacks[packetType])
	}
	clientPacketReader.ReliableHandler = func(packetType byte, packet *peer.UDPPacket, layers *peer.PacketLayers) {
		packetViewer.AddSplitPacket(packetType, packet, context, layers)
	}
	clientPacketReader.FullReliableHandler = func(packetType byte, packet *peer.UDPPacket, layers *peer.PacketLayers) {
		packetViewer.BindCallback(packetType, packet, context, layers, ActivationCallbacks[packetType])
	}
	clientPacketReader.ReliabilityLayerHandler = func(p *peer.UDPPacket, re *peer.ReliabilityLayer, ra *peer.RakNetLayer) {
		// nop
	}
	clientPacketReader.ACKHandler = func(p *peer.UDPPacket, ra *peer.RakNetLayer) {
		// nop
	}
	clientPacketReader.ErrorHandler = func(err error, packet *peer.UDPPacket) {
		println(err.Error())
	}
	clientPacketReader.ValContext = context
	clientPacketReader.ValCaches = new(peer.Caches)
	clientPacketReader.ValIsClient = true

	serverPacketReader := peer.NewPacketReader()
	serverPacketReader.SimpleHandler = clientPacketReader.SimpleHandler
	serverPacketReader.ReliableHandler = clientPacketReader.ReliableHandler
	serverPacketReader.FullReliableHandler = clientPacketReader.FullReliableHandler
	serverPacketReader.ReliabilityLayerHandler = clientPacketReader.ReliabilityLayerHandler
	serverPacketReader.ACKHandler = clientPacketReader.ACKHandler
	serverPacketReader.ErrorHandler = clientPacketReader.ErrorHandler
	serverPacketReader.ValContext = clientPacketReader.ValContext
	serverPacketReader.ValCaches = new(peer.Caches)

	for true {
		select {
		case <-captureJobContext.Done():
			return
		case packet := <-packetChannel:
			if packet.ApplicationLayer() == nil {
				color.Red("Ignoring packet because ApplicationLayer can't be decoded")
				continue
			}
			payload := packet.ApplicationLayer().Payload()
			if len(payload) == 0 {
				//println("Ignoring 0 payload")
				continue
			}
			src, dst := SrcAndDestFromGoPacket(packet)
			layers := &peer.PacketLayers{
				Root: RootLayer{
					Source:      src,
					Destination: dst,
				},
			}
			if context.Client == nil && !peer.IsOfflineMessage(payload) {
				//println("Ignoring non5")
				continue
			} else if context.Client == nil {
				if payload[0]%2 == 1 { // hack: detect packets 5 and 7
					context.Client, context.Server = src, dst
				} else {
					context.Client, context.Server = dst, src
				}
			} else if context.Client != "" && !context.IsClient(src) && !context.IsServer(src) {
				continue
			}
			layers.Root.FromClient = context.IsClient(src)
			layers.Root.FromServer = context.IsServer(src)

			if layers.Root.FromClient {
				clientPacketReader.ReadPacket(payload, layers)
			} else {
				serverPacketReader.ReadPacket(payload, layers)
			}
		}
	}
	return
}

func captureFromFile(filename string, useIPv4 bool, captureJobContext context.Context, packetViewer *MyPacketListView, context *peer.CommunicationContext) {
	fmt.Printf("Will capture from file %s\n", filename)
	handle, err := pcap.OpenOffline(filename)
	if err != nil {
		println(err.Error())
		return
	}
	captureJob(handle, useIPv4, captureJobContext, packetViewer, context)
}

func captureFromLive(livename string, useIPv4 bool, usePromisc bool, captureJobContext context.Context, packetViewer *MyPacketListView, context *peer.CommunicationContext) {
	fmt.Printf("Will capture from live device %s\n", livename)
	handle, err := pcap.OpenLive(livename, 2000, usePromisc, 10*time.Second)
	if err != nil {
		println(err.Error())
		return
	}
	captureJob(handle, useIPv4, captureJobContext, packetViewer, context)
}
