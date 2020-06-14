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

const (
    NotReferenced = iota
    InstanceDeletion
    InstanceCreation
	InstanceSetProperty
	InstanceEvent
	InstanceAck
	InstanceInValue
	InstanceAsParent
)

func valReferencesInstance(val rbxfile.Value, peerId uint32, id uint32) bool {
	switch val.Type() {
    case datamodel.TypeReference:
		ref := val.(datamodel.ValueReference).Reference
    	return ref.PeerId == peerId && ref.Id == id
    case datamodel.TypeArray:
        arr := val.(datamodel.ValueArray)
        for _, subval := range arr {
            if valReferencesInstance(subval, peerId, id) {
                return true
            }
        }
    case datamodel.TypeTuple:
        arr := val.(datamodel.ValueTuple)
        for _, subval := range arr {
            if valReferencesInstance(subval, peerId, id) {
                return true
            }
        }
    case datamodel.TypeMap:
        arr := val.(datamodel.ValueMap)
        for _, subval := range arr {
            if valReferencesInstance(subval, peerId, id) {
                return true
            }
        }
    case datamodel.TypeDictionary:
        arr := val.(datamodel.ValueDictionary)
        for _, subval := range arr {
            if valReferencesInstance(subval, peerId, id) {
                return true
            }
        }
	}
	return false
}

func replicationInstanceReferences(repInst *peer.ReplicationInstance, peerId uint32, id uint32) uint {
	ref := repInst.Instance.Ref
	if ref.PeerId == peerId && ref.Id == id {
		return InstanceCreation
	}
	if repInst.Parent != nil {
    	ref = repInst.Parent.Ref
    	if ref.PeerId == peerId && ref.Id == id {
    		return InstanceAsParent
    	}
	}
	for _, val := range repInst.Properties {
    	if valReferencesInstance(val, peerId, id) {
        	return InstanceInValue
    	}
	}
	return NotReferenced
}

func referencesInstance(packet peer.Packet83Subpacket, peerId uint32, id uint32) uint {
	switch packet.Type() {
	case 0x01:
		deletion := packet.(*peer.Packet83_01)
		ref := deletion.Instance.Ref
		if ref.PeerId == peerId && ref.Id == id {
    		return InstanceDeletion
		}
	case 0x02:
    	creation := packet.(*peer.Packet83_02)
    	return replicationInstanceReferences(creation.ReplicationInstance, peerId, id)
    case 0x03:
        propSet := packet.(*peer.Packet83_03)
        ref := propSet.Instance.Ref
    	if ref.PeerId == peerId && ref.Id == id {
			return InstanceSetProperty
    	}
    	if valReferencesInstance(propSet.Value, peerId, id) {
        	return InstanceInValue
    	}
    case 0x07:
        event := packet.(*peer.Packet83_07)
        ref := event.Instance.Ref
        if ref.PeerId == peerId && ref.Id == id {
            return InstanceEvent
        }
        for _, arg := range event.Event.Arguments {
			if valReferencesInstance(arg, peerId, id) {
    			return InstanceInValue
			}
        }
    case 0x0A:
        ack := packet.(*peer.Packet83_0A)
        ref := ack.Instance.Ref
        if ref.PeerId == peerId && ref.Id == id {
            return InstanceAck
        }
    case 0x0B:
		joinData := packet.(*peer.Packet83_0B)
		for _, inst := range joinData.Instances {
    		if ref := replicationInstanceReferences(inst, peerId, id); ref != NotReferenced {
        		return ref
    		}
		}
	}
	// TODO: streaming stuff?
	return NotReferenced
}

var packetMethods = map[string]lua.LGFunction{
    "ReferencesInstance": func(L *lua.LState) int {
		packet := checkPacket(L)
		peerId := uint32(L.CheckInt(2))
		id := uint32(L.CheckInt(3))

		switch packet.Type() {
		case 0x81:
			topReplic := packet.(*peer.Packet81Layer)
			for _, inst := range topReplic.Items {
    			ref := inst.Instance.Ref
    			if ref.PeerId == peerId && ref.Id == id {
        			L.Push(lua.LBool(true))
        			return 1
    			}
			}
			L.Push(lua.LBool(false))
			return 1
		case 0x83:
    		dataPacket := packet.(*peer.Packet83Layer)
			for _, subpacket := range dataPacket.SubPackets {
    			if referencesInstance(subpacket, peerId, id) != NotReferenced {
        			L.Push(lua.LBool(true))
        			return 1
    			}
			}
			L.Push(lua.LBool(false))
			return 1
		default:
    		L.Push(lua.LBool(false))
    		return 1
		}
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

func registerPacketType(L *lua.LState) {
    mt := L.NewTypeMetatable("Packet")
    L.SetField(mt, "__index", L.NewFunction(packetIndex))
}

func BridgePacket(L *lua.LState, packet peer.RakNetPacket) lua.LValue {
	ud := L.NewUserData()
	ud.Value = packet
	L.SetMetatable(ud, L.GetTypeMetatable("Packet"))
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

func NewLuaFilterState() *lua.LState {
    L := lua.NewState(lua.Options{
    	IncludeGoStackTrace: true,
    })
    registerPacketType(L)

    return L
}
func FilterAcceptsPacket(L *lua.LState, filter *lua.FunctionProto, packet peer.RakNetPacket, id int) (bool, error) {
	lfunc := L.NewFunctionFromProto(filter)
	L.Push(lfunc)
    L.Push(BridgePacket(L, packet))
    L.Push(lua.LNumber(float64(id)))

	err := L.PCall(2, lua.MultRet, nil)
	if err != nil {
    	return false, err
	}
	returnVal := L.CheckAny(1)
	if b, ok := returnVal.(lua.LBool); ok {
    	return bool(b), nil
	}
	return false, errors.New("invalid return value from filter")
}
