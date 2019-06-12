package api

import (
	"github.com/gskartwii/roblox-dissector/datamodel"
	"github.com/robloxapi/rbxfile"
	lua "github.com/yuin/gopher-lua"
)

func BridgeValue(value rbxfile.Value) *lua.LValue {
	if value == nil {
		return lua.LNil
	}
	ud := L.NewUserData()
	ud.Value = value
	L.SetMetatable(ud, L.GetTypeMetatable(datamodel.TypeString(value)))
	return ud
}
