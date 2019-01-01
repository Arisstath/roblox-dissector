package datamodel

import (
	"errors"
	"strings"
	"sync"

	"github.com/olebedev/emitter"
	"github.com/robloxapi/rbxfile"
)

type Instance struct {
	ClassName          string
	PropertiesMutex    *sync.RWMutex
	Properties         map[string]rbxfile.Value
	Children           []*Instance
	IsService          bool
	Ref                Reference
	ChildEmitter       *emitter.Emitter
	PropertyEmitter    *emitter.Emitter
	EventEmitter       *emitter.Emitter
	parent             *Instance
	DeleteOnDisconnect bool
}

func NewInstance(className string, parent *Instance) (*Instance, error) {
	inst := &Instance{
		ClassName:       className,
		Properties:      make(map[string]rbxfile.Value),
		PropertiesMutex: &sync.RWMutex{},
		ChildEmitter:    emitter.New(1),
		PropertyEmitter: emitter.New(1),
		EventEmitter:    emitter.New(1),
	}

	return inst, inst.SetParent(parent)
}

func (instance *Instance) HasAncestor(ancestor *Instance) bool {
	if instance == ancestor {
		return true
	}
	for instance != nil {
		if instance.parent == ancestor {
			return true
		}
		instance = instance.parent
	}
	return false
}

func (instance *Instance) AddChild(child *Instance) error {
	if instance.HasAncestor(child) {
		return errors.New("instance references can't be cyclic")
	}
	child.parent = instance
	if instance != nil {
		instance.Children = append(instance.Children, child)
		// Do NOT wait for the emitter to finish!
		instance.ChildEmitter.Emit(child.Name(), child)
	}
	// Do NOT wait for the emitter to finish!
	child.PropertyEmitter.Emit("Parent", instance)
	return nil
}

func (instance *Instance) Get(name string) rbxfile.Value {
	instance.PropertiesMutex.RLock()
	defer instance.PropertiesMutex.RUnlock()
	return instance.Properties[name]
}
func (instance *Instance) Set(name string, value rbxfile.Value) {
	instance.PropertiesMutex.Lock()
	instance.Properties[name] = value
	instance.PropertiesMutex.Unlock()
	// Do NOT wait for the emitter to finish!
	instance.PropertyEmitter.Emit(name, value)
}

func (instance *Instance) SetParent(parent *Instance) error {
	return parent.AddChild(instance)
}

func (instance *Instance) FindFirstChild(name string) *Instance {
	for _, child := range instance.Children {
		if child.Name() == name {
			return child
		}
	}
	return nil
}
func (instance *Instance) waitForChild(name string) <-chan *Instance {
	childChan := make(chan *Instance, 1)
	if child := instance.FindFirstChild(name); child != nil {
		childChan <- child
		return childChan
	}
	emitterChan := instance.ChildEmitter.Once(name)
	go func() {
		defer instance.ChildEmitter.Off(name, emitterChan)
		// If the child is added while we created the emitter
		if child := instance.FindFirstChild(name); child != nil {
			childChan <- child
		}
		childChan <- (<-emitterChan).Args[0].(*Instance)
	}()
	return childChan
}

func (instance *Instance) WaitForChild(names ...string) <-chan *Instance {
	retChan := make(chan *Instance, 1)
	go func() {
		for _, name := range names {
			instance = <-instance.waitForChild(name)
		}
		retChan <- instance
	}()
	return retChan
}

func (instance *Instance) WaitForProp(name string) <-chan rbxfile.Value {
	instance.PropertiesMutex.RLock()
	propChan := make(chan rbxfile.Value, 1)
	emitterChan := instance.PropertyEmitter.Once(name)

	go func() {
		propChan <- (<-emitterChan).Args[0].(rbxfile.Value)
	}()

	instance.PropertiesMutex.RUnlock()
	return propChan
}
func (instance *Instance) WaitForRefProp(name string) <-chan *Instance {
	propChan := make(chan *Instance, 1)
	currProp := instance.Get(name)
	if currProp != nil && currProp.(ValueReference).Instance != nil {
		propChan <- currProp.(ValueReference).Instance
		return propChan
	}
	go func() {
		for propEvt := range instance.PropertyEmitter.On(name) {
			if propEvt.Args[0] == nil {
				println("skipping nil prop")
				continue
			}
			currProp := propEvt.Args[0].(ValueReference)
			if currProp.Instance != nil {
				println("got arg")
				propChan <- currProp.Instance
				return
			}
			println("skipping nil ref")
		}
	}()
	return propChan
}

func (instance *Instance) Name() string {
	name := instance.Get("Name")
	if name == nil {
		return instance.ClassName
	}
	var nameStr rbxfile.ValueString
	var ok bool
	if nameStr, ok = name.(rbxfile.ValueString); !ok {
		return instance.ClassName
	}
	return string(nameStr)
}
func (instance *Instance) GetFullName() string {
	if instance == nil {
		return "nil"
	}
	parts := make([]string, 0, 8)
	for instance != nil {
		parts = append([]string{instance.Name()}, parts...)
		instance = instance.parent
	}
	var builder strings.Builder
	for _, part := range parts {
		builder.WriteByte('.')
		builder.WriteString(part)
	}
	return builder.String()
}
func (instance *Instance) MakeEventChan(name string, once bool) (<-chan emitter.Event, <-chan []rbxfile.Value) {
	evChan := make(chan []rbxfile.Value, 1)
	var emitterChan <-chan emitter.Event
	if once {
		emitterChan = instance.EventEmitter.Once(name)
	} else {
		emitterChan = instance.EventEmitter.On(name)
	}

	go func() {
		for ev := range emitterChan {
			evChan <- ev.Args[0].([]rbxfile.Value)
		}
		close(evChan)
	}()
	return emitterChan, evChan
}
func (instance *Instance) FireEvent(name string, args ...rbxfile.Value) {
	// Should not wait for the emitter to finish
	instance.EventEmitter.Emit(name, []rbxfile.Value(args))
}

func (instance *Instance) Parent() *Instance {
	return instance.parent
}

func (instance *Instance) Copy(pool *SelfReferencePool) *Instance {
	newInst := pool.MakeWithRef(instance.Ref.String())
	newInst.ClassName = instance.ClassName
	newInst.Ref = instance.Ref
	newInst.Children = make([]*Instance, len(instance.Children))
	newInst.Properties = make(map[string]rbxfile.Value, len(instance.Properties))
	// We intentionally do NOT set the parent here!
	// The parent may not be a copied instance

	newInst.PropertiesMutex.Lock()
	instance.PropertiesMutex.RLock()
	for name, value := range instance.Properties {
		newInst.Properties[name] = value.Copy()
		// Copy() clears the Instance field
		// hence we need to set it again here
		if value.Type() == TypeReference {
			newInst.Properties[name] = ValueReference{
				Instance:  pool.MakeWithRef(value.(ValueReference).Reference.String()),
				Reference: value.(ValueReference).Reference,
			}
		}
	}
	newInst.PropertiesMutex.Unlock()
	instance.PropertiesMutex.RUnlock()

	for i, child := range instance.Children {
		newChild := child.Copy(pool)
		newChild.parent = newInst
		newInst.Children[i] = newChild
	}

	return newInst
}
