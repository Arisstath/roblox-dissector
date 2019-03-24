package main

import (
	"context"
	"fmt"
	"time"

	"github.com/Gskartwii/roblox-dissector/peer"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"github.com/olebedev/emitter"
)

// TODO: Can this use ConnectedPeer?
func captureJob(handle *pcap.Handle, useIPv4 bool, captureJobContext context.Context, packetViewer *MyPacketListView, context *peer.CommunicationContext) {
	clientPacketReader := peer.NewPacketReader()
	serverPacketReader := peer.NewPacketReader()

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

	simpleHandler := func(e *emitter.Event) {
		layers := e.Args[0].(*peer.PacketLayers)
		//println("simple: ", layers.PacketType)
		packetViewer.AddFullPacket(layers.PacketType, context, layers, ActivationCallbacks[layers.PacketType])
	}
	reliableHandler := func(e *emitter.Event) {
		layers := e.Args[0].(*peer.PacketLayers)
		//println("reliable: ", layers.Reliability.SplitBuffer.UniqueID)
		packetViewer.AddSplitPacket(layers.PacketType, context, layers)
	}
	fullReliableHandler := func(e *emitter.Event) {
		layers := e.Args[0].(*peer.PacketLayers)
		//println("full-reliable: ", layers.Reliability.SplitBuffer.UniqueID)
		// special hook: we do not have a way to send specific security settings to the parser
		if layers.PacketType == 0x8A && layers.Error == nil {
			layer := layers.Main.(*peer.Packet8ALayer)
			layers.Root.Logger.Printf("hash = %8X, computed = %8X\n", settings.GenerateTicketHash(layer.ClientTicket), layer.TicketHash)
		}
		packetViewer.BindCallback(layers.PacketType, context, layers, ActivationCallbacks[layers.PacketType])
	}
	// ACK and ReliabilityLayer are nops

	clientPacketReader.SetContext(context)
	clientPacketReader.SetCaches(new(peer.Caches))
	clientPacketReader.SetIsClient(true)
	clientPacketReader.BindDataModelHandlers()
	serverPacketReader.SetContext(context)
	serverPacketReader.SetCaches(new(peer.Caches))
	serverPacketReader.BindDataModelHandlers()

	clientPacketReader.LayerEmitter.On("simple", simpleHandler, emitter.Void)
	clientPacketReader.LayerEmitter.On("reliable", reliableHandler, emitter.Void)
	clientPacketReader.LayerEmitter.On("full-reliable", fullReliableHandler, emitter.Void)
	clientPacketReader.ErrorEmitter.On("simple", simpleHandler, emitter.Void)
	clientPacketReader.ErrorEmitter.On("reliable", reliableHandler, emitter.Void)
	clientPacketReader.ErrorEmitter.On("full-reliable", fullReliableHandler, emitter.Void)

	serverPacketReader.LayerEmitter.On("simple", simpleHandler, emitter.Void)
	serverPacketReader.LayerEmitter.On("reliable", reliableHandler, emitter.Void)
	serverPacketReader.LayerEmitter.On("full-reliable", fullReliableHandler, emitter.Void)
	serverPacketReader.ErrorEmitter.On("simple", simpleHandler, emitter.Void)
	serverPacketReader.ErrorEmitter.On("reliable", reliableHandler, emitter.Void)
	serverPacketReader.ErrorEmitter.On("full-reliable", fullReliableHandler, emitter.Void)

	for {
		select {
		case <-captureJobContext.Done():
			return
		case packet := <-packetChannel:
			if packet.ApplicationLayer() == nil {
				println("Ignoring packet because ApplicationLayer can't be decoded")
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
