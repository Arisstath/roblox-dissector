package main

import "github.com/Gskartwii/roblox-dissector/peer"

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
