package main
import "roblox-dissector/peer"

var disconnectionReasons = [...]string{
	"Disconnected due to a bad hash",
	"Disconnected due to a Security Key Mismatch",
	"Protocol mismatch, please reconnect",
	"Error while receiving data, please reconnect",
	"Error while streaming data, please reconnect",
	"Error while sending data, please reconnect",
	"Place ID verification failed",
	"You are already joined to a game. Please shutdown other game and try again",
	"Error processing ticket, please reconnect",
	"Lost connection to server due to timeout",
	"You have been kicked from the game",
	"Kicked by server. Please close and rejoin another game",
	"Disconnected due to timeout, pleas reconnect",
	"You have been disconnected from Team Create, please reconnect",
	"Server was shutdown due to no active players",
	"Disconnected due to Security Key Mismatch (15)",
	"Disconnected from game, possibly due to game joined from another device",
}

func ShowPacket15(packetType byte, packet *peer.UDPPacket, context *peer.CommunicationContext, layers *peer.PacketLayers) {
	MainLayer := layers.Main.(*peer.Packet15Layer)

	var reason string
	if MainLayer.Reason == 0xFFFFFFFF {
		reason = "Developer has shut down all game servers or game has shut down for other reasons, please reconnect"
	} else {
		reason = disconnectionReasons[MainLayer.Reason]
	}

	layerLayout := NewBasicPacketViewer(packetType, packet, context, layers)
	layerLayout.AddWidget(NewQLabelF("Disconnection reason: %s", reason), 0, 0)
}
