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
			if context.Client == "" && payload[0] != 5 {
				//println("Ignoring non5")
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
				if context.IsClient(newPacket.Source) {
					clientPacketReader.ReadPacket(payload, newPacket)
				} else {
					serverPacketReader.ReadPacket(payload, newPacket)
				}
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
