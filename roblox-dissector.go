package main

import "github.com/Gskartwii/roblox-dissector/peer"

const DEBUG bool = false

var PacketNames map[byte]string = map[byte]string{
	0xFF: "???",
	0x00: "ID_CONNECTED_PING",
	0x01: "ID_UNCONNECTED_PING",
	0x03: "ID_CONNECTED_PONG",
	0x04: "ID_DETECT_LOST_CONNECTIONS",
	0x05: "ID_OPEN_CONNECTION_REQUEST_1",
	0x06: "ID_OPEN_CONNECTION_REPLY_1",
	0x07: "ID_OPEN_CONNECTION_REQUEST_2",
	0x08: "ID_OPEN_CONNECTION_REPLY_2",
	0x09: "ID_CONNECTION_REQUEST",
	0x10: "ID_CONNECTION_REQUEST_ACCEPTED",
	0x11: "ID_CONNECTION_ATTEMPT_FAILED",
	0x13: "ID_NEW_INCOMING_CONNECTION",
	0x15: "ID_DISCONNECTION_NOTIFICATION",
	0x18: "ID_INVALID_PASSWORD",
	0x1B: "ID_TIMESTAMP",
	0x1C: "ID_UNCONNECTED_PONG",
	0x81: "ID_ROBLOX_INIT_INSTANCES", // ID_SET_GLOBALS
	0x82: "ID_ROBLOX_DICTIONARIES",   // ID_TEACH_DESCRIPTOR_DICTIONARIES
	0x83: "ID_ROBLOX_REPLICATION",    // ID_DATA
	0x85: "ID_ROBLOX_PHYSICS",        // ID_PHYSICS
	0x86: "ID_ROBLOX_TOUCH",          // ID_TOUCHES
	0x87: "ID_ROBLOX_CHAT",           // unused
	0x88: "ID_ROBLOX_CHAT_TEAM",      // unused
	0x89: "ID_ROBLOX_REPORT_ABUSE",
	0x8A: "ID_ROBLOX_AUTH",        // ID_SUBMIT_TICKET
	0x8B: "ID_ROBLOX_CHAT_GAME",   // unused
	0x8C: "ID_ROBLOX_CHAT_PLAYER", // unused
	0x8D: "ID_ROBLOX_CLUSTER",
	0x8E: "ID_ROBLOX_PROTOCOL_MISMATCH",
	0x8F: "ID_ROBLOX_INITIAL_SPAWN_NAME",
	0x90: "ID_ROBLOX_REQUEST_PARAMS",        // ID_PROTOCOL_SYNC
	0x91: "ID_ROBLOX_NETWORK_SCHEMA",        // ID_SCHEMA_SYNC
	0x92: "ID_ROBLOX_VERIFY_PLACEID",        // ID_PLACEID_VERIFICATION
	0x93: "ID_ROBLOX_NETWORK_PARAMS",        // ID_DICTIONARY_FORMAT
	0x94: "ID_ROBLOX_HASH_REJECTED",         // ID_HASH_MISMATCH
	0x95: "ID_ROBLOX_SECURITY_KEY_REJECTED", // ID_SECURITYKEY_MISMATCH
	0x96: "ID_ROBLOX_REQUEST_STATS",
	0x97: "ID_ROBLOX_NEW_SCHEMA",
}

type ActivationCallback func(byte, *peer.UDPPacket, *peer.CommunicationContext, *peer.PacketLayers)

var ActivationCallbacks map[byte]ActivationCallback = map[byte]ActivationCallback{
	0x05: ShowPacket05,
	0x06: ShowPacket06,
	0x07: ShowPacket07,
	0x08: ShowPacket08,
	0x09: ShowPacket09,
	0x10: ShowPacket10,
	0x13: ShowPacket13,
	0x00: ShowPacket00,
	0x03: ShowPacket03,
	0x15: ShowPacket15,

	0x93: ShowPacket93,
	//0x82: ShowPacket82,
	0x92: ShowPacket92,
	0x90: ShowPacket90,
	0x8F: ShowPacket8F,
	0x81: ShowPacket81,
	0x83: ShowPacket83,
	0x97: ShowPacket97,
	0x85: ShowPacket85,
	0x86: ShowPacket86,
	0x8A: ShowPacket8A,
}

func main() {
	GUIMain()
}
