package main

import (
	"errors"
	"strings"

	"github.com/Gskartwii/roblox-dissector/datamodel"
	"github.com/Gskartwii/roblox-dissector/peer"
	"github.com/robloxapi/rbxfile"
	"github.com/yuin/gopher-lua"
	"github.com/yuin/gopher-lua/parse"
)

var packetFields = map[string]lua.LGFunction{
	"Type": func(L *lua.LState) int {
		packet := checkPacket(L)
		typ := packet.Type()
		L.Push(lua.LNumber(float64(typ)))
		return 1
	},
}

const NotReferenced = 0

const (
	InstanceDeletion = 1 << iota
	InstanceCreation
	InstanceSetProperty
	InstanceEvent
	InstanceAck
	InstanceInValue
	InstanceAsParent
	InstanceTopReplicated
	InstancePhysicsRoot
	InstancePhysicsChild
	InstancePhysicsPlatformChild
	InstanceTouched
	InstanceTouchEnded
)

var instanceReferenceEnum = map[string]lua.LValue{
	"Deletion":             lua.LNumber(InstanceDeletion),
	"Creation":             lua.LNumber(InstanceCreation),
	"SetProperty":          lua.LNumber(InstanceSetProperty),
	"Event":                lua.LNumber(InstanceEvent),
	"Ack":                  lua.LNumber(InstanceAck),
	"InValue":              lua.LNumber(InstanceInValue),
	"AsParent":             lua.LNumber(InstanceAsParent),
	"TopReplicated":        lua.LNumber(InstanceTopReplicated),
	"PhysicsRoot":          lua.LNumber(InstancePhysicsRoot),
	"PhysicsChild":         lua.LNumber(InstancePhysicsChild),
	"PhysicsPlatformChild": lua.LNumber(InstancePhysicsPlatformChild),
	"Touched":              lua.LNumber(InstanceTouched),
	"TouchedEnded":         lua.LNumber(InstanceTouchEnded),
}

func valReferencesInstance(val rbxfile.Value, ref datamodel.Reference) bool {
	switch val.Type() {
	case datamodel.TypeReference:
		valRef := val.(datamodel.ValueReference).Reference
		return valRef.Equal(&ref)
	case datamodel.TypeArray:
		arr := val.(datamodel.ValueArray)
		for _, subval := range arr {
			if valReferencesInstance(subval, ref) {
				return true
			}
		}
	case datamodel.TypeTuple:
		arr := val.(datamodel.ValueTuple)
		for _, subval := range arr {
			if valReferencesInstance(subval, ref) {
				return true
			}
		}
	case datamodel.TypeMap:
		arr := val.(datamodel.ValueMap)
		for _, subval := range arr {
			if valReferencesInstance(subval, ref) {
				return true
			}
		}
	case datamodel.TypeDictionary:
		arr := val.(datamodel.ValueDictionary)
		for _, subval := range arr {
			if valReferencesInstance(subval, ref) {
				return true
			}
		}
	}
	return false
}

func replicationInstanceReferences(repInst *peer.ReplicationInstance, ref datamodel.Reference) uint {
	var refType uint
	creationRef := repInst.Instance.Ref
	if creationRef.Equal(&ref) {
		refType |= InstanceCreation
	}
	if repInst.Parent != nil {
		creationRef = repInst.Parent.Ref
		if creationRef.Equal(&ref) {
			refType |= InstanceAsParent
		}
	}
	for _, val := range repInst.Properties {
		if valReferencesInstance(val, ref) {
			refType |= InstanceInValue
			break // there's nothing this loop can change anymore
		}
	}
	return refType
}

func referencesInstance(packet peer.Packet83Subpacket, ref datamodel.Reference) uint {
	switch packet.Type() {
	case 0x01:
		deletion := packet.(*peer.Packet83_01)
		deletedRef := deletion.Instance.Ref
		if deletedRef.Equal(&ref) {
			return InstanceDeletion
		}
	case 0x02:
		creation := packet.(*peer.Packet83_02)
		return replicationInstanceReferences(creation.ReplicationInstance, ref)
	case 0x03:
		propSet := packet.(*peer.Packet83_03)
		propRef := propSet.Instance.Ref
		var refType uint
		if propRef.Equal(&ref) {
			refType |= InstanceSetProperty
		}
		if valReferencesInstance(propSet.Value, ref) {
			refType |= InstanceInValue
		}
		return refType
	case 0x07:
		event := packet.(*peer.Packet83_07)
		eventRef := event.Instance.Ref
		var refType uint
		if eventRef.Equal(&ref) {
			refType |= InstanceEvent
		}
		for _, arg := range event.Event.Arguments {
			if valReferencesInstance(arg, ref) {
				refType |= InstanceInValue
				break // there's nothing this loop change anymore
			}
		}
		return refType
	case 0x0A:
		ack := packet.(*peer.Packet83_0A)
		ackRef := ack.Instance.Ref
		if ackRef.Equal(&ref) {
			return InstanceAck
		}
	case 0x0B:
		joinData := packet.(*peer.Packet83_0B)
		var refType uint
		for _, inst := range joinData.Instances {
			refType |= replicationInstanceReferences(inst, ref)
		}
		return refType
	}
	// TODO: streaming stuff?
	return NotReferenced
}

func physicsDataRefs(physicsData *peer.PhysicsData, ref datamodel.Reference) uint {
	var referenceData uint
	if inst := physicsData.Instance; inst != nil {
		physicsRef := inst.Ref
		if physicsRef.Equal(&ref) {
			referenceData |= InstancePhysicsRoot
		}
	}
	if inst := physicsData.PlatformChild; inst != nil {
		physicsRef := inst.Ref
		if physicsRef.Equal(&ref) {
			referenceData |= InstancePhysicsPlatformChild
		}
	}
	return referenceData
}

func packetInstanceReferenceDataMask(packet peer.RakNetPacket, ref datamodel.Reference) uint {
	var referenceData uint

	switch packet.Type() {
	case 0x81:
		topReplic := packet.(*peer.Packet81Layer)
		for _, inst := range topReplic.Items {
			topRef := inst.Instance.Ref
			if topRef.Equal(&ref) {
				referenceData |= InstanceTopReplicated
			}
		}
	case 0x83:
		dataPacket := packet.(*peer.Packet83Layer)
		for _, subpacket := range dataPacket.SubPackets {
			if refType := referencesInstance(subpacket, ref); refType != NotReferenced {
				referenceData |= refType
			}
		}
	case 0x85:
		physics := packet.(*peer.Packet85Layer)
		for _, subpacket := range physics.SubPackets {
			referenceData |= physicsDataRefs(&subpacket.Data, ref)
			for _, child := range subpacket.Children {
				subRefData := physicsDataRefs(child, ref)
				if subRefData&InstancePhysicsRoot != 0 {
					subRefData ^= InstancePhysicsRoot
					referenceData |= InstancePhysicsChild
				}
				referenceData |= subRefData
			}
			for _, history := range subpacket.History {
				referenceData |= physicsDataRefs(history, ref)
			}
		}
	case 0x86:
		touches := packet.(*peer.Packet86Layer)
		for _, subpacket := range touches.SubPackets {
			if subpacket.Instance1.Ref.Equal(&ref) || subpacket.Instance2.Ref.Equal(&ref) {
				if subpacket.IsTouch {
					referenceData |= InstanceTouched
				} else {
					referenceData |= InstanceTouchEnded
				}
			}
		}
	}

	return referenceData
}

var packetMethods = map[string]lua.LGFunction{
	"ReferencesInstance": func(L *lua.LState) int {
		packet := checkPacket(L)

		var ref datamodel.Reference
		var maskIdx int
		if L.Get(2).Type() == lua.LTNil {
			ref.IsNull = true
			maskIdx = 3
		} else {
			ref.PeerId = uint32(L.CheckInt(2))
			ref.Id = uint32(L.CheckInt(3))
			maskIdx = 4
		}
		var mask uint
		if L.Get(maskIdx).Type() == lua.LTNil {
			mask = ^uint(0) // "any"
		} else {
			mask = uint(L.CheckInt(maskIdx))
		}

		result := packetInstanceReferenceDataMask(packet, ref)
		L.Push(lua.LBool(result&mask != 0)) // test if referenced bitset intersects with the one we asked for
		return 1
	},
	"HasSubpacket": func(L *lua.LState) int {
		packet := checkPacket(L)
		typ := uint8(L.CheckInt(2))

		if packet.Type() != 0x83 {
			L.Push(lua.LBool(false))
			return 1
		}
		for _, sub := range packet.(*peer.Packet83Layer).SubPackets {
			if sub.Type() == typ {
				L.Push(lua.LBool(true))
				return 1
			}
		}

		L.Push(lua.LBool(false))
		return 1
	},
}

var replicPacketMethods = map[string]lua.LGFunction{
	"ReferencesInstance": func(L *lua.LState) int {
		packet := checkPacket(L)

		var ref datamodel.Reference
		var maskIdx int
		if L.Get(2).Type() == lua.LTNil {
			ref.IsNull = true
			maskIdx = 3
		} else {
			ref.PeerId = uint32(L.CheckInt(2))
			ref.Id = uint32(L.CheckInt(3))
			maskIdx = 4
		}
		var mask uint
		if L.Get(maskIdx).Type() == lua.LTNil {
			mask = ^uint(0) // "any"
		} else {
			mask = uint(L.CheckInt(maskIdx))
		}

		result := referencesInstance(packet, ref)
		L.Push(lua.LBool(result&mask != 0)) // test if referenced bitset intersects with the one we asked for
		return 1
	},
	"HasSubpacket": func(L *lua.LState) int {
		packet := checkPacket(L)
		typ := uint8(L.CheckInt(2))
		L.Push(lua.LBool(packet.Type() == typ))
		return 1
	},
}

func checkPacket(L *lua.LState) peer.RakNetPacket {
	ud := L.CheckUserData(1)
	if v, ok := ud.Value.(peer.RakNetPacket); ok {
		return v
	}
	L.ArgError(1, "packet expected")
	return nil
}

func packetIndex(L *lua.LState) int {
	checkPacket(L)
	name := L.CheckString(2)
	if fieldGetter, ok := packetFields[name]; ok {
		return fieldGetter(L)
	} else if method, ok := packetMethods[name]; ok {
		L.Push(L.NewFunction(method))
		return 1
	}
	return 0
}

func replicPacketIndex(L *lua.LState) int {
	checkPacket(L)
	name := L.CheckString(2)
	if name == "Type" {
		L.Push(lua.LNumber(0x83))
		return 1
	} else if method, ok := replicPacketMethods[name]; ok {
		L.Push(L.NewFunction(method))
		return 1
	}
	return 0
}

func registerPacketType(L *lua.LState) {
	mt := L.NewTypeMetatable("Packet")
	L.SetField(mt, "__index", L.NewFunction(packetIndex))

	replicMt := L.NewTypeMetatable("ReplicPacket")
	L.SetField(replicMt, "__index", L.NewFunction(replicPacketIndex))
}

func registerInstanceRefEnum(L *lua.LState) {
	tab := L.CreateTable(0, len(instanceReferenceEnum))
	for k, v := range instanceReferenceEnum {
		tab.RawSetString(k, v)
	}
	L.SetGlobal("InstRefType", tab)
}

func BridgePacket(L *lua.LState, packet peer.RakNetPacket, mt string) lua.LValue {
	if packet == nil {
		return lua.LNil
	}
	ud := L.NewUserData()
	ud.Value = packet
	L.SetMetatable(ud, L.GetTypeMetatable(mt))
	return ud
}

func CompileFilter(filter string) (*lua.FunctionProto, error) {
	r := strings.NewReader(filter)
	parsed, err := parse.Parse(r, "<filter>")
	if err != nil {
		return nil, err
	}
	proto, err := lua.Compile(parsed, "<filter>")
	if err != nil {
		return nil, err
	}
	return proto, nil
}

func NewLuaFilterState(printFunc func(string)) *lua.LState {
	L := lua.NewState(lua.Options{
		IncludeGoStackTrace: true,
	})
	registerPacketType(L)
	registerInstanceRefEnum(L)

	L.Register("print", lua.LGFunction(func(L *lua.LState) int {
		numArgs := L.GetTop()
		if numArgs == 0 {
			printFunc("\n")
			return 0
		}

		var output strings.Builder
		output.WriteString(L.Get(1).String())
		for i := 2; i <= numArgs; i++ {
			output.WriteString("\t")
			output.WriteString(L.Get(i).String())
		}
		output.WriteString("\n")
		printFunc(output.String())
		return 0
	}))

	return L
}

func FilterAcceptsReplicPacket(L *lua.LState, filter *lua.FunctionProto, packet peer.Packet83Subpacket) (bool, error) {
	lfunc := L.NewFunctionFromProto(filter)
	L.Push(lfunc)
	var numArgs int
	L.Push(BridgePacket(L, packet, "ReplicPacket"))
	numArgs++

	err := L.PCall(numArgs, lua.MultRet, nil)
	if err != nil {
		return false, err
	}
	returnVal := L.Get(1)
	L.Pop(1)
	if b, ok := returnVal.(lua.LBool); ok {
		return bool(b), nil
	}
	return false, errors.New("invalid return value from filter")
}

func FilterAcceptsPacket(L *lua.LState, filter *lua.FunctionProto, packet peer.RakNetPacket) (bool, error) {
	lfunc := L.NewFunctionFromProto(filter)
	L.Push(lfunc)
	var numArgs int
	L.Push(BridgePacket(L, packet, "Packet"))
	numArgs++

	err := L.PCall(numArgs, lua.MultRet, nil)
	if err != nil {
		return false, err
	}
	returnVal := L.Get(1)
	L.Pop(1)
	if b, ok := returnVal.(lua.LBool); ok {
		return bool(b), nil
	}
	return false, errors.New("invalid return value from filter")
}
