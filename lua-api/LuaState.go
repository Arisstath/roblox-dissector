package api

import (
	"github.com/Gskartwii/roblox-dissector/datamodel"
	"github.com/Gskartwii/roblox-dissector/peer"
	lua "github.com/yuin/gopher-lua"
)

type State struct {
	*lua.LState
}

func NewState(opts ...lua.Options) *State {
	L := lua.NewState(opts...)
	registerDataModelType(L)
	registerInstanceType(L)
	registerValueAs(L)

	return &State{LState: L}
}

func (state *State) RegisterSchema(schema *peer.NetworkSchema) {
	state.SetGlobal("schema", BridgeSchema(schema, state.LState))
}

func (state *State) RegisterDataModel(datamodel *datamodel.DataModel) {
	state.SetGlobal("game", BridgeDataModel(datamodel, state.LState))
}
