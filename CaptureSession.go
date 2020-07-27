package main

import (
	"bytes"
	"context"
	"fmt"
	"github.com/Gskartwii/roblox-dissector/peer"
	"github.com/gotk3/gotk3/glib"
	"github.com/olebedev/emitter"
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
	ProgressCallback      func(int)
	progress              int
	ForgetAcks            bool
}

func AddressEq(a *net.UDPAddr, b *net.UDPAddr) bool {
	return a.Port == b.Port && bytes.Equal(a.IP, b.IP)
}

func NewCaptureSession(name string, cancelFunc context.CancelFunc, listViewerCallback func(*CaptureSession, *PacketListViewer, error)) (*CaptureSession, error) {
	initialViewer, err := NewPacketListViewer(fmt.Sprintf("%s#%d", name, 1))
	if err != nil {
		return nil, err
	}
	session := &CaptureSession{
		Name:                  name,
		ViewerCounter:         2,
		IsCapturing:           true,
		Conversations:         nil,
		CancelFunc:            cancelFunc,
		InitialViewerOccupied: false,
		ListViewers:           []*PacketListViewer{initialViewer},
	}
	listViewerCallback(session, initialViewer, nil)

	return session, nil
}

func (session *CaptureSession) SetProgress(prog int) {
	session.progress = prog
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

	if len(payload) < 1 || payload[0] != 0x7B {
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
	session.AddConversation(newConv)

	return newConv
}

func (session *CaptureSession) AddConversation(conv *Conversation) (*PacketListViewer, error) {
	var err error
	var viewer *PacketListViewer
	if !session.InitialViewerOccupied {
		session.InitialViewerOccupied = true
		viewer = session.ListViewers[0]
	} else {
		title := fmt.Sprintf("%s#%d", session.Name, session.ViewerCounter)
		session.ViewerCounter++
		viewer, err = NewPacketListViewer(title)
		session.ListViewerCallback(viewer, err)
	}
	handler := func(e *emitter.Event) {
		topic := e.OriginalTopic
		layers := e.Args[0].(*peer.PacketLayers)

		associatedProgress := session.progress
		_, err := glib.IdleAdd(func() bool {
			viewer.NotifyPacket(topic, layers, session.ForgetAcks)
			session.ProgressCallback(associatedProgress)
			return false
		})
		if err != nil {
			println("idleadd failed:", err.Error())
		}
	}
	conv.ClientReader.Layers().On("*", handler, emitter.Void)
	conv.ClientReader.Errors().On("*", handler, emitter.Void)
	conv.ServerReader.Layers().On("*", handler, emitter.Void)
	conv.ServerReader.Errors().On("*", handler, emitter.Void)

	return viewer, err
}

func (session *CaptureSession) StopCapture() {
	if session.IsCapturing {
		session.IsCapturing = false
		session.CancelFunc()
	}
}

func (session *CaptureSession) ReportDone() {
	glib.IdleAdd(func() bool {
		session.IsCapturing = false
		if session.ProgressCallback != nil {
			session.ProgressCallback(-1)
		}
		return false
	})
}
