package api

import (
	"github.com/gskartwii/roblox-dissector/datamodel"
	"github.com/gskartwii/roblox-dissector/peer"
	"github.com/robloxapi/rbxfile"
	lua "github.com/yuin/gopher-lua"
)

func getSchema(L *lua.LState) *peer.NetworkSchema {
	return L.GetGlobal("schema").(*lua.LUserData).Value.(*peer.NetworkSchema)
}

func instanceIndex(L *lua.LState) int {
	inst := checkInstance(L)
	name := L.CheckString(2)
	if v, ok := instanceMethods[name]; ok {
		L.Push(L.NewFunction(v))
		return 1
	}

	instanceClass := getSchema(L).SchemaForClass(inst.ClassName)
	for _, prop := range instanceClass.Properties {
		if prop.Name == name {
			val := inst.Get(name)
			L.Push(BridgeValue(val))
			return 1
		}
	}

	return instanceMethods["WaitForChild"](L)
}

func instanceNewIndex(L *lua.LState) int {
	inst := checkInstance(L)
	prop := L.CheckString(2)
	if prop == "Parent" {
		inst2 := checkInstanceArg(L, 3)
		err := inst2.AddChild(inst)
		if err != nil {
			L.RaiseError("Set parent: %s", err.Error())
			return 0
		}
		return 0
	}

	val := checkValue(L, 3)
	inst.Set(prop, val)
	return 0
}

func registerInstanceType(L *lua.LState) {
	mt := L.NewTypeMetatable("Instance")
	L.SetField(mt, "__index", L.NewFunction(instanceIndex))
	L.SetField(mt, "__newindex", L.NewFunction(instanceNewIndex))
}

func checkInstanceArg(L *lua.LState, index int) *datamodel.Instance {
	ud := L.CheckUserData(index)
	if v, ok := ud.Value.(*datamodel.Instance); ok {
		return v
	}
	L.ArgError(index, "instance expected")
	return nil
}

func checkInstance(L *lua.LState) *datamodel.Instance {
	return checkInstanceArg(L, 1)
}

func checkValue(L *lua.LState, index int) rbxfile.Value {
	ud := L.CheckUserData(index)
	if v, ok := ud.Value.(rbxfile.Value); ok {
		return v
	}
	L.ArgError(index, "value expected")
	return nil
}

func BridgeInstance(instance *datamodel.Instance, L *lua.LState) lua.LValue {
	ud := L.NewUserData()
	ud.Value = instance
	L.SetMetatable(ud, L.GetTypeMetatable("Instance"))
	return ud
}

var instanceMethods = map[string]lua.LGFunction{
	"FindFirstChild": func(L *lua.LState) int {
		inst := checkInstance(L)
		name := L.CheckString(2)
		child := inst.FindFirstChild(name)
		if child == nil {
			L.Push(lua.LNil)
			return 1
		}
		childBridge := BridgeInstance(child, L)
		L.Push(childBridge)
		return 1
	},
	"WaitForChild": func(L *lua.LState) int {
		inst := checkInstance(L)
		name := L.CheckString(2)
		child, err := inst.WaitForChild(L.Context(), name)
		if err != nil {
			L.RaiseError("WaitForChild: %s", err.Error())
			return 0
		}
		childBridge := BridgeInstance(child, L)
		L.Push(childBridge)
		return 1
	},
}
