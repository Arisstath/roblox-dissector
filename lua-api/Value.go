package api

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/Gskartwii/roblox-dissector/datamodel"
	"github.com/robloxapi/rbxfile"
	lua "github.com/yuin/gopher-lua"
)

func valueAs(L *lua.LState) int {
	value := L.Get(1)
	name := L.CheckString(2)

	typeMt := L.NewTable()
	typeMt.RawSetString("__type", lua.LString(name))

	if tableVal, ok := value.(*lua.LTable); ok {
		L.SetMetatable(tableVal, typeMt)
		L.Push(tableVal)
		return 1
	}
	proxyTable := L.NewTable()
	proxyTable.RawSetString("Value", value)
	L.SetMetatable(proxyTable, typeMt)
	L.Push(proxyTable)
	return 1
}

func registerValueAs(L *lua.LState) {
	L.SetGlobal("as", L.NewFunction(valueAs))
}

func checkValue(L *lua.LState, index int, typ rbxfile.Type) rbxfile.Value {
	val := L.Get(index)
	rbxfileValue, err := coerceValue(L, val, typ)
	if err != nil {
		L.ArgError(index, err.Error())
		return nil
	}
	return rbxfileValue
}

func getValue(L *lua.LState, val lua.LValue) rbxfile.Value {
	switch val.(type) {
	case *lua.LNilType:
		return nil
	case lua.LBool:
		return rbxfile.ValueBool(val == lua.LTrue)
	case lua.LString:
		return rbxfile.ValueString(val.(lua.LString))
	case lua.LNumber:
		return rbxfile.ValueDouble(val.(lua.LNumber))
	case *lua.LUserData:
		ud, ok := val.(*lua.LUserData)
		if !ok {
			L.RaiseError("not instance")
		}
		inst, ok := ud.Value.(*datamodel.Instance)
		if !ok {
			L.RaiseError("not instance")
		}
		return datamodel.ValueReference{Instance: inst, Reference: inst.Ref}
	case *lua.LTable:
		valTable := val.(*lua.LTable)
		reportedType := L.GetMetaField(valTable, "__type").(lua.LString)
		rbxfileType := datamodel.TypeFromString(string(reportedType))
		if rbxfileType == rbxfile.TypeInvalid {
			L.RaiseError("invalid type %s", reportedType)
		}

		createdValue := datamodel.NewValue(rbxfileType)
		reflectVal := reflect.ValueOf(createdValue)
		switch reflectVal.Kind() {
		case reflect.Bool:
			stored := valTable.RawGetString("Value").(lua.LBool)
			reflectVal.SetBool(stored == lua.LTrue)
		case reflect.String:
			stored := valTable.RawGetString("Value").(lua.LString)
			reflectVal.SetString(string(stored))
		case reflect.Float32, reflect.Float64:
			stored := valTable.RawGetString("Value").(lua.LNumber)
			reflectVal.SetFloat(float64(stored))
		case reflect.Int, reflect.Int64:
			stored := valTable.RawGetString("Value").(lua.LNumber)
			reflectVal.SetInt(int64(stored))
		case reflect.Slice:
			if reflectVal.Elem().Kind() == reflect.Uint8 {
				stored := valTable.RawGetString("Value").(lua.LString)
				reflectVal.SetBytes([]byte(stored))
				return createdValue
			}
			// Don't use Len(), use MaxN() to find the real length
			for i := 0; i < valTable.MaxN(); i++ {
				subValue := getValue(L, valTable.RawGetInt(i))
				reflectVal = reflect.Append(reflectVal, reflect.ValueOf(subValue))
			}

			// createdVal may not contain the slice anymore
			return reflectVal.Interface().(rbxfile.Value)
		case reflect.Map:
			if reflectVal.Type().Key().Kind() != reflect.String {
				L.RaiseError("can't handle map key type %s", reflectVal.Type().Key().Kind())
			}
			k := lua.LNil
			var v lua.LValue
			for k, v = valTable.Next(k); v != nil; k, v = valTable.Next(k) {
				keyStr, ok := k.(lua.LString)
				if !ok {
					L.RaiseError("invalid map key %T", k)
				}

				subValue := getValue(L, v)
				reflectVal.SetMapIndex(reflect.ValueOf(keyStr), reflect.ValueOf(subValue))
			}
		case reflect.Struct:
			for i := 0; i < reflectVal.NumField(); i++ {
				keyName := reflectVal.Type().Field(i).Name
				keyValue := valTable.RawGetString(keyName)
				field := reflectVal.Field(i)

				switch reflectVal.Type().Field(i).Type.Kind() {
				// Don't need to support everything here
				case reflect.Bool:
					field.SetBool(keyValue == lua.LTrue)
				case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
					field.SetInt(int64(keyValue.(lua.LNumber)))
				case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
					field.SetUint(uint64(keyValue.(lua.LNumber)))
				case reflect.Float32, reflect.Float64:
					field.SetFloat(float64(keyValue.(lua.LNumber)))
				case reflect.String:
					field.SetString(string(keyValue.(lua.LString)))
				}
			}
		}
		return createdValue
	default:
		L.RaiseError("can't dereflect type %T", val)
		return nil
	}
}

func coerceValue(L *lua.LState, val lua.LValue, typ rbxfile.Type) (rbxfile.Value, error) {
	if val == lua.LNil {
		return nil, nil
	}
	switch typ {
	case rbxfile.TypeBool:
		valBool, ok := val.(lua.LBool)
		if !ok {
			return nil, errors.New("not bool")
		}
		return rbxfile.ValueBool(valBool), nil
	case rbxfile.TypeString:
		valStr, ok := val.(lua.LString)
		if !ok {
			return nil, errors.New("not string")
		}
		return rbxfile.ValueString(valStr), nil
	case rbxfile.TypeContent:
		valStr, ok := val.(lua.LString)
		if !ok {
			return nil, errors.New("not string")
		}
		return rbxfile.ValueContent(valStr), nil
	case rbxfile.TypeBinaryString:
		valStr, ok := val.(lua.LString)
		if !ok {
			return nil, errors.New("not string")
		}
		return rbxfile.ValueBinaryString(valStr), nil
	case rbxfile.TypeProtectedString:
		valStr, ok := val.(lua.LString)
		if !ok {
			return nil, errors.New("not string")
		}
		return rbxfile.ValueProtectedString(valStr), nil
	case rbxfile.TypeFloat:
		valNum, ok := val.(lua.LNumber)
		if !ok {
			return nil, errors.New("not number")
		}
		return rbxfile.ValueFloat(valNum), nil
	case rbxfile.TypeDouble:
		valNum, ok := val.(lua.LNumber)
		if !ok {
			return nil, errors.New("not number")
		}
		return rbxfile.ValueDouble(valNum), nil
	case rbxfile.TypeInt:
		valNum, ok := val.(lua.LNumber)
		if !ok {
			return nil, errors.New("not number")
		}
		return rbxfile.ValueInt(valNum), nil
	case rbxfile.TypeInt64:
		valNum, ok := val.(lua.LNumber)
		if !ok {
			return nil, errors.New("not number")
		}
		return rbxfile.ValueInt64(valNum), nil
	default:
		createdVal := getValue(L, val)

		if createdVal.Type() != typ {
			return nil, fmt.Errorf("expected %s", datamodel.TypeString(datamodel.NewValue(typ)))
		}

		return createdVal, nil
	}
}

func BridgeValue(L *lua.LState, value interface{}) lua.LValue {
	if ref, ok := value.(datamodel.ValueReference); ok {
		return BridgeInstance(ref.Instance, L)
	}

	reflectVal := reflect.ValueOf(value)
	switch reflectVal.Kind() {
	case reflect.Invalid:
		return lua.LNil
	case reflect.Bool:
		return lua.LBool(reflectVal.Bool())
	case reflect.Float32, reflect.Float64:
		return lua.LNumber(reflectVal.Float())
	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return lua.LNumber(reflectVal.Int())
	case reflect.String:
		return lua.LString(reflectVal.String())
	case reflect.Slice:
		elemType := reflectVal.Type().Elem()
		if elemType.Kind() == reflect.Uint8 {
			return lua.LString(reflectVal.Bytes())
		}
		expectedType := reflect.TypeOf((*rbxfile.Value)(nil)).Elem()
		if !elemType.Implements(expectedType) {
			L.RaiseError("invalid slice type %s", elemType.String())
		}

		out := L.NewTable()
		for i := 0; i < reflectVal.Len(); i++ {
			out.Append(BridgeValue(L, reflectVal.Index(i).Interface().(rbxfile.Value)))
		}

		typeMt := L.NewTable()
		typeMt.RawSetString("__type", lua.LString(datamodel.TypeString(value.(rbxfile.Value))))
		L.SetMetatable(out, typeMt)

		return out
	case reflect.Map:
		keyKind := reflectVal.Type().Key().Kind()
		if keyKind != reflect.String {
			L.RaiseError("invalid key kind %s", keyKind.String())
		}
		// map values should always implement rbxfile.Value
		elemType := reflectVal.Type().Elem()
		expectedType := reflect.TypeOf((*rbxfile.Value)(nil)).Elem()
		if !elemType.Implements(expectedType) {
			L.RaiseError("invalid elem type %s", elemType.String())
		}

		out := L.NewTable()
		iter := reflectVal.MapRange()
		for iter.Next() {
			k := iter.Key().String()
			v := iter.Value().Interface().(rbxfile.Value)
			out.RawSetString(k, BridgeValue(L, v))
		}

		typeMt := L.NewTable()
		typeMt.RawSetString("__type", lua.LString(datamodel.TypeString(value.(rbxfile.Value))))
		L.SetMetatable(out, typeMt)

		return out
	case reflect.Struct:
		out := L.NewTable()
		for i := 0; i < reflectVal.NumField(); i++ {
			fieldName := reflectVal.Type().Field(i).Name
			fieldVal := reflectVal.Field(i)
			// val doesn't necessarily have to be rbxfile value

			out.RawSetString(fieldName, BridgeValue(L, fieldVal.Interface()))
		}

		typeMt := L.NewTable()
		typeMt.RawSetString("__type", lua.LString(datamodel.TypeString(value.(rbxfile.Value))))
		L.SetMetatable(out, typeMt)

		return out
	default:
		L.RaiseError("invalid value kind %s", reflectVal.Kind().String())
		return lua.LNil
	}
}
