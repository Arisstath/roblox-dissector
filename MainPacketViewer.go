package main

import (
	"errors"
	"fmt"

	"github.com/Gskartwii/roblox-dissector/peer"
	"github.com/gotk3/gotk3/gtk"
)

func openConnectionReq1Viewer(packet *peer.Packet05Layer) (gtk.IWidget, error) {
	return nil, errors.New("unimplemented")
}
func openConnectionResp1Viewer(packet *peer.Packet06Layer) (gtk.IWidget, error) {
	return nil, errors.New("unimplemented")
}
func openConnectionReq2Viewer(packet *peer.Packet07Layer) (gtk.IWidget, error) {
	return nil, errors.New("unimplemented")
}
func openConnectionResp2Viewer(packet *peer.Packet08Layer) (gtk.IWidget, error) {
	return nil, errors.New("unimplemented")
}
func connectionRequestViewer(packet *peer.Packet09Layer) (gtk.IWidget, error) {
	return nil, errors.New("unimplemented")
}
func connectionAcceptedViewer(packet *peer.Packet10Layer) (gtk.IWidget, error) {
	return nil, errors.New("unimplemented")
}
func newIncomingConnectionViewer(packet *peer.Packet13Layer) (gtk.IWidget, error) {
	return nil, errors.New("unimplemented")
}
func disconnectionNotificationViewer(packet *peer.Packet15Layer) (gtk.IWidget, error) {
	return nil, errors.New("unimplemented")
}
func topReplicViewer(packet *peer.Packet81Layer) (gtk.IWidget, error) {
	return nil, errors.New("unimplemented")
}
func submitTicketViewer(packet *peer.Packet8ALayer) (gtk.IWidget, error) {
	return nil, errors.New("unimplemented")
}
func clusterViewer(packet *peer.Packet8DLayer) (gtk.IWidget, error) {
	return nil, errors.New("unimplemented")
}
func protocolSyncViewer(packet *peer.Packet90Layer) (gtk.IWidget, error) {
	return nil, errors.New("unimplemented")
}
func dictionaryFormatViewer(packet *peer.Packet93Layer) (gtk.IWidget, error) {
	return nil, errors.New("unimplemented")
}
func schemaViewer(packet *peer.Packet97Layer) (gtk.IWidget, error) {
	return nil, errors.New("unimplemented")
}
func luauChallengeViewer(packet *peer.Packet9BLayer) (gtk.IWidget, error) {
	return nil, errors.New("unimplemented")
}

func viewerForMainPacket(packet peer.RakNetPacket) (gtk.IWidget, error) {
	switch packet.Type() {
	case 0x00, 0x03, 0x83, 0x84, 0x85, 0x86, 0x87, 0x8F, 0x92, 0x96, 0x98:
		return blanketViewer(packet.String())
	case 0x05:
		return openConnectionReq1Viewer(packet.(*peer.Packet05Layer))
	case 0x06:
		return openConnectionResp1Viewer(packet.(*peer.Packet06Layer))
	case 0x07:
		return openConnectionReq2Viewer(packet.(*peer.Packet07Layer))
	case 0x08:
		return openConnectionResp2Viewer(packet.(*peer.Packet08Layer))
	case 0x09:
		return connectionRequestViewer(packet.(*peer.Packet09Layer))
	case 0x10:
		return connectionAcceptedViewer(packet.(*peer.Packet10Layer))
	case 0x13:
		return newIncomingConnectionViewer(packet.(*peer.Packet13Layer))
	case 0x15:
		return disconnectionNotificationViewer(packet.(*peer.Packet15Layer))
	case 0x81:
		return topReplicViewer(packet.(*peer.Packet81Layer))
	case 0x8A:
		return submitTicketViewer(packet.(*peer.Packet8ALayer))
	case 0x8D:
		return clusterViewer(packet.(*peer.Packet8DLayer))
	case 0x90:
		return protocolSyncViewer(packet.(*peer.Packet90Layer))
	case 0x93:
		return dictionaryFormatViewer(packet.(*peer.Packet93Layer))
	case 0x97:
		return schemaViewer(packet.(*peer.Packet97Layer))
	case 0x9B:
		return luauChallengeViewer(packet.(*peer.Packet9BLayer))
	default:
		return nil, fmt.Errorf("unimplemented packet type %02X", packet.Type())
	}
}
