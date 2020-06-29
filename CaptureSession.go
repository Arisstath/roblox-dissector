package main

import (
	"bytes"
	"context"
	"fmt"
	"github.com/Gskartwii/roblox-dissector/peer"
	"net"
)

type Conversation struct {
	Client       *net.UDPAddr
	Server       *net.UDPAddr
	ClientReader PacketProvider
	ServerReader PacketProvider
	Context      *peer.CommunicationContext
}

type CaptureSession struct {
	Name                  string
	ViewerCounter         uint
	IsCapturing           bool
	Conversations         []*Conversation
	CancelFunc            context.CancelFunc
	InitialViewerOccupied bool
	ListViewers           []*PacketListViewer
	ListViewerCallback    func(*PacketListViewer, error)
}

func AddressEq(a *net.UDPAddr, b *net.UDPAddr) bool {
	return a.Port == b.Port && bytes.Equal(a.IP, b.IP)
}

func NewCaptureSession(name string, cancelFunc context.CancelFunc, listViewerCallback func(*PacketListViewer, error)) (*CaptureSession, error) {
	initialViewer, err := NewPacketListViewer(fmt.Sprintf("%s#%d", name, 1))
	if err != nil {
		return nil, err
	}
	listViewerCallback(initialViewer, nil)

	return &CaptureSession{
		Name:                  name,
		ViewerCounter:         2,
		IsCapturing:           true,
		Conversations:         nil,
		CancelFunc:            cancelFunc,
		InitialViewerOccupied: false,
		ListViewers:           []*PacketListViewer{initialViewer},
	}, nil
}

func (session *CaptureSession) ConversationFor(source *net.UDPAddr, dest *net.UDPAddr, payload []byte) *Conversation {
	for _, conv := range session.Conversations {
		if AddressEq(source, conv.Client) && AddressEq(dest, conv.Server) {
			return conv
		}
		if AddressEq(source, conv.Server) && AddressEq(dest, conv.Client) {
			return conv
		}
	}

	if len(payload) < 1 || payload[0] != 0x05 {
		return nil
	}
	isHandshake := peer.IsOfflineMessage(payload)
	if !isHandshake {
		return nil
	}

	newContext := peer.NewCommunicationContext()
	clientR := peer.NewPacketReader()
	serverR := peer.NewPacketReader()
	clientR.SetContext(newContext)
	serverR.SetContext(newContext)
	clientR.SetIsClient(true)
	clientR.BindDataModelHandlers()
	serverR.BindDataModelHandlers()
	newConv := &Conversation{
		Client:       source,
		Server:       dest,
		ClientReader: clientR,
		ServerReader: serverR,
		Context:      newContext,
	}
	session.Conversations = append(session.Conversations, newConv)

	return newConv
}

func (session *CaptureSession) AddConversation(conv *Conversation) (*PacketListViewer, error) {
	var err error
	var viewer *PacketListViewer
	if !session.InitialViewerOccupied {
		session.InitialViewerOccupied = true
		viewer = session.ListViewers[0]
	} else {
		title := fmt.Sprintf("%s#d", session.Name, session.ViewerCounter)
		session.ViewerCounter++
		viewer, err = NewPacketListViewer(title)
		session.ListViewerCallback(viewer, err)
	}

	return viewer, err
}

func (session *CaptureSession) StopCapture() {
	if session.IsCapturing {
		session.IsCapturing = false
		session.CancelFunc()
	}
}
