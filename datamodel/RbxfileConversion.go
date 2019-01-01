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
	if instance == nil {
		return nil
	}
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
	if instance == nil {
		return nil
	}
	return pool.MakeWithRef(instance.Reference)
}
func (pool *SelfReferencePool) MakeWithRef(ref string) *Instance {
	key := ref
	inst, ok := pool.pool[key]
	if ok {
		return inst
	}
	inst, _ = NewInstance("", nil)

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
	inst.Properties = make(map[string]rbxfile.Value, len(instance.Properties))
	for name, value := range instance.Properties {
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
		case ValueNumberSequence:
			oldSeq := value.(ValueNumberSequence)
			newSeq := make([]rbxfile.ValueNumberSequenceKeypoint, len(oldSeq))
			for i, keypoint := range oldSeq {
				newSeq[i] = rbxfile.ValueNumberSequenceKeypoint(keypoint)
			}
			inst.Properties[name] = rbxfile.ValueNumberSequence(newSeq)
		case ValueColorSequence:
			oldSeq := value.(ValueColorSequence)
			newSeq := make([]rbxfile.ValueColorSequenceKeypoint, len(oldSeq))
			for i, keypoint := range oldSeq {
				newSeq[i] = rbxfile.ValueColorSequenceKeypoint(keypoint)
			}
			inst.Properties[name] = rbxfile.ValueColorSequence(newSeq)
		case ValueToken:
			inst.Properties[name] = rbxfile.ValueToken(value.(ValueToken).Value)
		case ValueReference:
			inst.Properties[name] = rbxfile.ValueReference{
				Instance: pool.Make(value.(ValueReference).Instance),
			}
		case nil:
			// Strip this property
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
		switch value.Type() {
		case rbxfile.TypeToken:
			newToken := ValueToken{Value: uint32(value.(rbxfile.ValueToken))}

			instance.Properties[name] = newToken
		case rbxfile.TypeReference:
			refInst := pool.Make(value.(rbxfile.ValueReference).Instance)
			if refInst == nil {
				instance.Properties[name] = ValueReference{Reference: NullReference}
			} else {
				instance.Properties[name] = ValueReference{
					Reference: refInst.Ref,
					Instance:  refInst,
				}
			}
		case rbxfile.TypeNumberSequence:
			oldSeq := value.(rbxfile.ValueNumberSequence)
			newSeq := make([]ValueNumberSequenceKeypoint, len(oldSeq))

			for i, keypoint := range oldSeq {
				newSeq[i] = ValueNumberSequenceKeypoint(keypoint)
			}
			instance.Properties[name] = ValueNumberSequence(newSeq)
		case rbxfile.TypeColorSequence:
			oldSeq := value.(rbxfile.ValueColorSequence)
			newSeq := make([]ValueColorSequenceKeypoint, len(oldSeq))
			for i, keypoint := range oldSeq {
				newSeq[i] = ValueColorSequenceKeypoint(keypoint)
			}

			instance.Properties[name] = ValueColorSequence(newSeq)
		}
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

	// We do not need to register the dummy root in the pool
	// Nothing should ever refer to it
	dummyRoot, _ := NewInstance("DataModel", nil)
	dummyRoot.IsService = true
	dummyRoot.Ref = dictionary.NewReference()

	for _, serv := range root.Instances {
		thisServ := InstanceFromRbxfile(serv, pool, dictionary)
		thisServ.SetParent(dummyRoot)
		model.AddService(thisServ)
	}

	return model
}
