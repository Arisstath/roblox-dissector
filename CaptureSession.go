package main

import (
	"context"
	"fmt"

	"github.com/google/gopacket/pcap"

	"github.com/olebedev/emitter"
	"github.com/therecipe/qt/widgets"
)

type CaptureSession struct {
	CaptureWindow     *DissectorWindow
	CaptureContext    *CaptureContext
	IsCapturing       bool
	Context           context.Context
	CancelFunc        context.CancelFunc
	Name              string
	PacketListViewers []*PacketListViewer
	UsesInitialViewer bool
	Handle            *pcap.Handle

	SetModel bool
}

func (session *CaptureSession) AddConversation(conv *Conversation) *PacketListViewer {
	var viewer *PacketListViewer
	// The very first session window will exist before packets have been captured
	if !session.UsesInitialViewer {
		session.UsesInitialViewer = true
		viewer = session.PacketListViewers[0]
	} else {
		window := session.CaptureWindow
		viewer = NewPacketListViewer(session.Context, window, 0)
		session.PacketListViewers = append(session.PacketListViewers, viewer)

		window.TabWidget.AddTab(viewer, fmt.Sprintf("Conversation: %s#%d", session.Name, len(session.PacketListViewers)))
	}
	viewer.BindToConversation(conv)
	if session.SetModel {
		viewer.TreeView.SetModel(viewer.ProxyModel)
	}

	return viewer
}

func (session *CaptureSession) HasViewer(viewer *widgets.QWidget) bool {
	for _, v := range session.PacketListViewers {
		// TODO: Too hacky?
		if v.QWidget.Pointer() == viewer.Pointer() {
			return true
		}
	}
	return false
}

func (session *CaptureSession) StopCapture() {
	session.IsCapturing = false
	session.CancelFunc()

	if session.CaptureWindow.CurrentSession == session {
		session.CaptureWindow.StopAction.SetEnabled(false)
	}
}

func NewCaptureSession(name string, window *DissectorWindow) *CaptureSession {
	ctx, cancelFunc := context.WithCancel(context.Background())
	captureContext := NewCaptureContext()

	initialViewer := NewPacketListViewer(ctx, window, 0)

	session := &CaptureSession{
		IsCapturing:       true,
		CaptureWindow:     window,
		CaptureContext:    captureContext,
		Context:           ctx,
		CancelFunc:        cancelFunc,
		Name:              name,
		PacketListViewers: []*PacketListViewer{initialViewer},
	}
	captureContext.ConversationEmitter.On("conversation", func(e *emitter.Event) {
		conv := e.Args[0].(*Conversation)
		MainThreadRunner.RunOnMain(func() {
			session.AddConversation(conv)
		})
		<-MainThreadRunner.Wait
	}, emitter.Void)

	// TODO: Too hacky?
	captureContext.Close = func() {
		if session.Handle != nil {
			session.Handle.Close()
		}
	}

	// Can't add the tab here because the session isn't on the window yet
	return session
}

func (session *CaptureSession) CaptureFromHandle(handle *pcap.Handle, isIPv4 bool, progressChan chan int) error {
	session.Handle = handle
	return session.CaptureContext.CaptureFromHandle(session.Context, handle, isIPv4, progressChan)
}

func (session *CaptureSession) UpdateModels() {
	for _, viewer := range session.PacketListViewers {
		viewer.UpdateModel()
	}
}

func (session *CaptureSession) RemoveViewer(viewer *widgets.QWidget) bool {
	var index int
	for i, v := range session.PacketListViewers {
		// TODO: Too hacky?
		if v.QWidget.Pointer() == viewer.Pointer() {
			index = i
		}
	}
	// anti-memory leak deletion
	copy(session.PacketListViewers[index:], session.PacketListViewers[index+1:])
	session.PacketListViewers[len(session.PacketListViewers)-1] = nil
	session.PacketListViewers = session.PacketListViewers[:len(session.PacketListViewers)-1]

	return len(session.PacketListViewers) == 0
}

func (session *CaptureSession) Destroy() {
	if session.IsCapturing {
		session.StopCapture()
	}
	for _, viewer := range session.PacketListViewers {
		viewer.DestroyQWidget()
	}
	session.PacketListViewers = nil
}
