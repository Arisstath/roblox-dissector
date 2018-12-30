package datamodel

import "github.com/robloxapi/rbxfile"

// RbxfileReferencePool holds a collection of visited instances
// This prevents the conversion from creating a duplicate
// rbxfile representation of an instance if it is visited twice
// (e.g.) normally and via ValueReference
type RbxfileReferencePool struct {
	pool map[string]*rbxfile.Instance
}

func NewRbxfileReferencePool() *RbxfileReferencePool {
	return &RbxfileReferencePool{pool: make(map[string]*rbxfile.Instance)}
}

func (pool *RbxfileReferencePool) Make(instance *Instance) *rbxfile.Instance {
	key := instance.Ref.String()
	inst, ok := pool.pool[key]
	if ok {
		return inst
	}
	inst = rbxfile.NewInstance("", nil)

	pool.pool[key] = inst
	return inst
}

// SelfReferencePool holds a collection of visited instances
// This prevents the conversion from creating a duplicate
// rbxfile representation of an instance if it is visited twice
// (e.g.) normally and via ValueReference
type SelfReferencePool struct {
	pool map[string]*Instance
}

func NewSelfReferencePool() *SelfReferencePool {
	return &SelfReferencePool{pool: make(map[string]*Instance)}
}

func (pool *SelfReferencePool) Make(instance *rbxfile.Instance) *Instance {
	return pool.MakeWithRef(instance.Reference)
}
func (pool *SelfReferencePool) MakeWithRef(ref string) *Instance {
	key := ref
	inst, ok := pool.pool[key]
	if ok {
		return inst
	}
	inst = NewInstance("", nil)

	pool.pool[key] = inst
	return inst
}

// ToRbxfile converts an Instance to the rbxfile format
// Note: the parent of this instance must be assigned manually, however
// its children will have their respective parents set correctly
//
// The conv argument is required so that
// the ValueReferences can be converted without creating
// duplicate referents
func (instance *Instance) ToRbxfile(pool *RbxfileReferencePool) *rbxfile.Instance {
	inst := pool.Make(instance)
	inst.ClassName = instance.ClassName
	inst.IsService = instance.IsService
	inst.Reference = instance.Ref.String()
	instance.PropertiesMutex.RLock()
	inst.Properties = make(map[string]rbxfile.Value, len(inst.Properties))
	for name, value := range inst.Properties {
		switch value.(type) {
		case ValueNumberSequenceKeypoint,
			ValueColorSequenceKeypoint,
			ValueSystemAddress,
			ValueMap,
			ValueDictionary,
			ValueArray,
			ValueTuple,
			ValueRegion3,
			ValueRegion3int16:
			println("Dropping property", name)
		case ValueToken:
			inst.Properties[name] = rbxfile.ValueToken(value.(ValueToken).Value)
		case ValueReference:
			inst.Properties[name] = rbxfile.ValueReference{
				Instance: pool.Make(value.(ValueReference).Instance),
			}
		default:
			inst.Properties[name] = value
		}
	}
	instance.PropertiesMutex.RUnlock()

	inst.Children = make([]*rbxfile.Instance, 0, len(instance.Children))
	for _, child := range instance.Children {
		inst.AddChild(child.ToRbxfile(pool))
	}

	return inst
}

func (model *DataModel) ToRbxfile() *rbxfile.Root {
	pool := NewRbxfileReferencePool()
	root := &rbxfile.Root{Instances: make([]*rbxfile.Instance, len(model.Instances))}
	for i, inst := range model.Instances {
		root.Instances[i] = inst.ToRbxfile(pool)
	}

	return root
}

func InstanceFromRbxfile(inst *rbxfile.Instance, pool *SelfReferencePool, dictionary *InstanceDictionary) *Instance {
	instance := pool.Make(inst)
	instance.ClassName = inst.ClassName
	instance.IsService = inst.IsService
	instance.Ref = dictionary.NewReference()

	instance.PropertiesMutex.Lock()
	instance.Properties = make(map[string]rbxfile.Value, len(inst.Properties))
	for name, value := range inst.Properties {
		instance.Properties[name] = value
	}
	instance.PropertiesMutex.Unlock()

	instance.Children = make([]*Instance, 0, len(inst.Children))
	for _, child := range inst.Children {
		instance.AddChild(InstanceFromRbxfile(child, pool, dictionary))
	}

	return instance
}

func FromRbxfile(dictionary *InstanceDictionary, root *rbxfile.Root) *DataModel {
	pool := NewSelfReferencePool()
	model := New()

	dummyRoot := NewInstance("DataModel", nil)
	dummyRoot.IsService = true

	for _, serv := range root.Instances {
		thisServ := InstanceFromRbxfile(serv, pool, dictionary)
		thisServ.SetParent(dummyRoot)
		model.AddService(thisServ)
	}

	return model
}
