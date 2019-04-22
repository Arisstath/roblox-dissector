package main

import (
	"github.com/Gskartwii/roblox-dissector/peer"
	"github.com/olebedev/emitter"
)

func (captureContext *CaptureContext) HookClient(client *peer.CustomClient) {
	conversation := NewProviderConversation(client.DefaultPacketWriter, client.DefaultPacketReader)
	// The client address assigned here may be wrong. It doesn't really matter
	conversation.ClientAddress = &client.Address
	conversation.ServerAddress = &client.ServerAddress
	captureContext.AddConversation(conversation)
}

func (session *CaptureSession) CaptureFromClient(client *peer.CustomClient, placeID uint32, authTicket string) {
	session.CaptureContext.HookClient(client)

	client.ConnectWithAuthTicket(placeID, authTicket)
}

func (captureContext *CaptureContext) HookServer(server *peer.CustomServer) {
	server.ClientEmitter.On("client", func(e *emitter.Event) {
		client := e.Args[0].(*peer.ServerClient)
		conversation := NewProviderConversation(client.DefaultPacketReader, client.DefaultPacketWriter)
		conversation.ClientAddress = client.Address
		conversation.ServerAddress = client.Server.Address

		captureContext.AddConversation(conversation)
	}, emitter.Void)
}

func (session *CaptureSession) CaptureFromServer(server *peer.CustomServer) {
	session.CaptureContext.HookServer(server)

	server.Start()
}
