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
	settings := peer.Win10Settings()
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

	simpleHandler := func(layers *peer.PacketLayers) {
		packetViewer.AddFullPacket(packetType, context, layers, ActivationCallbacks[packetType])
	}
	reliableHandler := func(layers *peer.PacketLayers) {
		packetViewer.AddSplitPacket(packetType, context, layers)
	}
	fullReliableHandler := func(layers *peer.PacketLayers) {
		// special hook: we do not have a way to send specific security settings to the parser
		if packetType == 0x8A && layers.Error == nil {
			layer := layers.Main.(*peer.Packet8ALayer)
			layers.Root.Logger.Printf("hash = %8X, computed = %8X\n", settings.GenerateTicketHash(layer.ClientTicket), layer.TicketHash)
		}
		packetViewer.BindCallback(packetType, context, layers, ActivationCallbacks[packetType])
	}

	clientPacketReader := peer.NewPacketReader()
	serverPacketReader := peer.NewPacketReader()

	packetAggregate := &emitter.Group{Cap: 0}
	// even errors can be grouped to the same channel! they are handled by GUIManager
	packetAggregate.Add(
		clientPacketReader.LayerEmitter.On("*", emitter.Sync),
		clientPacketReader.ErrorEmitter.On("*", emitter.Sync),
		serverPacketReader.LayerEmitter.On("*", emitter.Sync),
		serverPacketReader.ErrorEmitter.On("*", emitter.Sync),
	)
	
	// ACK and ReliabilityLayer are nops

	clientPacketReader.SetContext(context)
	clientPacketReader.SetCaches(new(peer.Caches))
	clientPacketReader.SetIsClient(true)
	serverPacketReader.SetContext(context)
	serverPacketReader.SetCaches(new(peer.Caches))

	for true {
		select {
		case e := <- packetAggregate.On():
			switch e.OriginalTopic {
			case "simple":
				simpleHandler(e.Args[0].(*PacketLayers))
			case "full-reliable":
				fullReliableHandler(e.Args[0].(*PacketLayers))
			case "reliable":
				reliableHandler(e.Args[0].(*PacketLayers))
				// ACK and reliabilitylayer are nops
			}
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
			if packet.Layer(layers.LayerTypeIPv4) == nil {
				continue
			}
			src, dst := SrcAndDestFromGoPacket(packet)
			layers := &peer.PacketLayers{
				Root: peer.RootLayer{
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
			} else if context.Client != nil && !context.IsClient(src) && !context.IsServer(src) {
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
