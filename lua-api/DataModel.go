package api

import (
	"github.com/gskartwii/roblox-dissector/datamodel"
	lua "github.com/yuin/gopher-lua"
)

func dataModelIndex(L *lua.LState) int {
	dm := checkDataModel(L)
	name := L.CheckString(2)
	if v, ok := datamodelMethods[name]; ok {
		L.Push(L.NewFunction(v))
		return 1
	}
	return datamodelMethods["WaitForService"](L)
}

func registerDataModelType(L *lua.LState) {
	mt := L.NewTypeMetatable("DataModel")
	L.SetField(mt, "__index", L.NewFunction(dataModelIndex))
}

func checkDataModel(L *lua.LState) *datamodel.DataModel {
	ud := L.CheckUserData(1)
	if v, ok := ud.Value.(*datamodel.DataModel); ok {
		return v
	}
	L.ArgError(1, "datamodel expected")
	return nil
}

func BridgeDataModel(dataModel *datamodel.DataModel, L *lua.LState) lua.LValue {
	ud := L.NewUserData()
	ud.Value = dataModel
	L.SetMetatable(ud, L.GetTypeMetatable("DataModel"))
	return ud
}

var datamodelMethods = map[string]lua.LGFunction{
	"FindService": func(L *lua.LState) int {
		dm := checkDataModel(L)
		name := L.CheckString(2)
		// TODO: Caching?
		service := dm.FindService(name)
		if service == nil {
			L.Push(lua.LNil)
			return 1
		}
		servBridge := BridgeInstance(service, L)
		L.Push(servBridge)
		return 1
	},
	"WaitForService": func(L *lua.LState) int {
		dm := checkDataModel(L)
		name := L.CheckString(2)
		service, err := dm.WaitForService(L.Context(), name)
		if err != nil {
			L.RaiseError("WaitForService: %s", err.Error())
			return 0
		}
		servBridge := BridgeInstance(service, L)
		L.Push(servBridge)
		return 1
	},
}
