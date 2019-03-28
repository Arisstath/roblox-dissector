package main

import (
	"context"
	"net"

	"github.com/Gskartwii/roblox-dissector/peer"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"github.com/olebedev/emitter"
)

type Conversation struct {
	ClientAddress *net.UDPAddr
	ServerAddress *net.UDPAddr
	ClientReader  *peer.DefaultPacketReader
	ServerReader  *peer.DefaultPacketReader
	Context       *peer.CommunicationContext
}
type CaptureContext struct {
	Conversations       []*Conversation
	ConversationEmitter *emitter.Emitter
}

func NewCaptureContext() *CaptureContext {
	return &CaptureContext{
		Conversations:       make([]*Conversation, 0, 1),
		ConversationEmitter: emitter.New(8),
	}
}

func (ctx *CaptureContext) FindClient(client *net.UDPAddr) *Conversation {
	for _, conv := range ctx.Conversations {
		if conv.ClientAddress.IP.Equal(client.IP) && conv.ClientAddress.Port == client.Port {
			return conv
		}
	}
	return nil
}
func (ctx *CaptureContext) FindServer(server *net.UDPAddr) *Conversation {
	for _, conv := range ctx.Conversations {
		if conv.ServerAddress.IP.Equal(server.IP) && conv.ServerAddress.Port == server.Port {
			return conv
		}
	}
	return nil
}
func (ctx *CaptureContext) Find(src *net.UDPAddr, dst *net.UDPAddr) (*Conversation, bool) {
	conv := ctx.FindClient(src)
	if conv != nil {
		return conv, true
	}
	conv = ctx.FindClient(dst)
	if conv != nil {
		return conv, false
	}

	return nil, false
}
func (ctx *CaptureContext) AddConversation(conv *Conversation) {
	ctx.Conversations = append(ctx.Conversations, conv)
	<-ctx.ConversationEmitter.Emit("conversation", conv)
}

func NewConversation(client *net.UDPAddr, server *net.UDPAddr) *Conversation {
	context := peer.NewCommunicationContext()
	clientReader := peer.NewPacketReader()
	clientReader.SetIsClient(true)
	clientReader.SetContext(context)
	serverReader := peer.NewPacketReader()
	serverReader.SetContext(context)
	// We do not automatically bind DataModel handlers here
	// the user can do that manually

	conv := &Conversation{
		ClientAddress: client,
		ServerAddress: server,
		ClientReader:  clientReader,
		ServerReader:  serverReader,
		Context:       context,
	}

	return conv
}

func (captureContext *CaptureContext) Capture(ctx context.Context, packetSource *gopacket.PacketSource, progressChan chan int) error {
	var progress int
	for packet := range packetSource.Packets() {
		select {
		case <-ctx.Done():
			return nil
		case progressChan <- progress:
		default:
		}
		progress++
		if packet.ApplicationLayer() == nil || packet.Layer(layers.LayerTypeIPv4) == nil {
			continue
		}
		payload := packet.ApplicationLayer().Payload()
		if len(payload) == 0 {
			continue
		}

		src, dst := SrcAndDestFromGoPacket(packet)
		conv, fromClient := captureContext.Find(src, dst)
		layers := &peer.PacketLayers{
			Root: peer.RootLayer{
				Source:      src,
				Destination: dst,
				// Do not set FromClient or FromServer yet
				// fromClient may be false if conv is nil!
			},
		}

		if conv == nil {
			if !peer.IsOfflineMessage(payload) {
				// Conversation not recognized and not an offline message:
				// skip this packet
				continue
			}
			if payload[0] != 5 {
				println("Warning: receiving unknown offline message: ", payload[0])
				continue
			}
			fromClient = true

			conv = NewConversation(src, dst)
			captureContext.AddConversation(conv)
		}
		layers.Root.FromClient = fromClient
		layers.Root.FromServer = !fromClient

		if fromClient {
			conv.ClientReader.ReadPacket(payload, layers)
		} else {
			conv.ServerReader.ReadPacket(payload, layers)
		}
	}

	return nil
}

func (captureContext *CaptureContext) CaptureFromHandle(ctx context.Context, handle *pcap.Handle, isIPv4 bool, progressChan chan int) error {
	err := handle.SetBPFFilter("udp")
	if err != nil {
		return err
	}

	var packetSource *gopacket.PacketSource
	if isIPv4 {
		packetSource = gopacket.NewPacketSource(handle, layers.LayerTypeIPv4)
	} else {
		packetSource = gopacket.NewPacketSource(handle, handle.LinkType())
	}

	return captureContext.Capture(ctx, packetSource, progressChan)
}
