package main

import (
	"context"
	"fmt"
	"net"

	"github.com/Gskartwii/roblox-dissector/peer"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

type Conversation struct {
	ClientAddress *net.UDPAddr
	ServerAddress *net.UDPAddr
	ClientReader  *peer.DefaultPacketReader
	ServerReader  *peer.DefaultPacketReader
	Context       *peer.CommunicationContext
}
type CaptureContext []*Conversation

func (ctx CaptureContext) FindClient(client *net.UDPAddr) *Conversation {
	for _, conv := range ctx {
		if conv.ClientAddress.IP.Equal(client.IP) && conv.ClientAddress.Port == client.Port {
			return conv
		}
	}
	return nil
}
func (ctx CaptureContext) FindServer(server *net.UDPAddr) *Conversation {
	for _, conv := range ctx {
		if conv.ServerAddress.IP.Equal(server.IP) && conv.ServerAddress.Port == server.Port {
			return conv
		}
	}
	return nil
}
func (ctx CaptureContext) Find(src *net.UDPAddr, dst *net.UDPAddr) (*Conversation, bool) {
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

func NewConversation(client *net.UDPAddr, server *net.UDPAddr) *Conversation {
	context := peer.NewCommunicationContext()
	// TODO: Remove?
	context.Client = client
	context.Server = server

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

func captureJob(captureJobContext context.Context, name string, packetSource *gopacket.PacketSource, window *DissectorWindow, progressChan chan int) error {
	var progress int
	defer close(progressChan)
	conversations := CaptureContext(make([]*Conversation, 0, 1))
	listViewers := make([]*PacketListViewer, 0, 1)
	for packet := range packetSource.Packets() {
		select {
		case <-captureJobContext.Done():
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
		conv, fromClient := conversations.Find(src, dst)
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
			conversations = append(conversations, conv)
			newListViewer := window.AddConversation(fmt.Sprintf("%s#%d", name, len(conversations)), conv)
			listViewers = append(listViewers, newListViewer)
		}
		layers.Root.FromClient = fromClient
		layers.Root.FromServer = !fromClient

		if fromClient {
			conv.ClientReader.ReadPacket(payload, layers)
		} else {
			conv.ServerReader.ReadPacket(payload, layers)
		}
	}

	for _, viewer := range listViewers {
		viewer.UpdateModel()
	}
	return nil
}
