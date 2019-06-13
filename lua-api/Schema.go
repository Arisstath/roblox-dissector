package api

import (
	"github.com/Gskartwii/roblox-dissector/peer"
	"github.com/yuin/gopher-lua"
)

func BridgeSchema(schema *peer.NetworkSchema, L *lua.LState) lua.LValue {
	ud := L.NewUserData()
	ud.Value = schema
	// TODO
	return ud
}
