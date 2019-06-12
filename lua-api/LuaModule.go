package api

import lua "github.com/yuin/gopher-lua"

var exports = map[string]lua.LGFunction{}

func Loader(L *lua.LState) int {
	registerDataModelType(L)
	registerInstanceType(L)

	module := L.SetFuncs(L.NewTable(), exports)
	L.Push(mod)

	return 1
}
