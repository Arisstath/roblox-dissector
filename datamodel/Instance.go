package datamodel

import (
	"strings"
	"sync"

	"github.com/olebedev/emitter"
	"github.com/robloxapi/rbxfile"
)

type Instance struct {
	ClassName       string
	PropertiesMutex *sync.RWMutex
	Properties      map[string]rbxfile.Value
	Children        []*Instance
	Ref             Reference
	ChildEmitter    *emitter.Emitter
	PropertyEmitter *emitter.Emitter
	EventEmitter    *emitter.Emitter
	parent          *Instance
}

func NewInstance(className string, parent *Instance) *Instance {
	inst := &Instance{
		ClassName:       className,
		Properties:      make(map[string]rbxfile.Value),
		PropertiesMutex: &sync.RWMutex{},
		ChildEmitter:    emitter.New(1),
		PropertyEmitter: emitter.New(1),
	}
	inst.SetParent(inst)
	return inst
}

func (instance *Instance) AddChild(child *Instance) {
	if instance != nil {
		instance.Children = append(instance.Children, child)
		instance.ChildEmitter.Emit(child.Name(), child)
	}
	child.PropertyEmitter.Emit("Parent", instance)
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
	instance.PropertyEmitter.Emit(name, value)
}

func (instance *Instance) SetParent(parent *Instance) {
	parent.AddChild(instance)
}

func (instance *Instance) FindFirstChild(name string) *Instance {
	for _, child := range instance.Children {
		if child.Name() == name {
			return child
		}
	}
	return nil
}
func (instance *Instance) WaitForChild(name string) <-chan *Instance {
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
	if currProp != nil && currProp.(*ValueReference).Instance != nil {
		propChan <- currProp.(*ValueReference).Instance
		return propChan
	}
	go func() {
		for currProp := range instance.WaitForProp(name) {
			if currProp != nil && currProp.(*ValueReference).Instance != nil {
				propChan <- currProp.(*ValueReference).Instance
				return
			}
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
	var builder strings.Builder
	for instance != nil {
		builder.WriteByte('.')
		builder.WriteString(string(instance.Get("Name").(rbxfile.ValueString)))
		instance = instance.parent
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
	}()
	return emitterChan, evChan
}
func (instance *Instance) FireEvent(name string, args ...rbxfile.Value) {
	instance.EventEmitter.Emit(name, []rbxfile.Value(args))
}

func (instance *Instance) Parent() *Instance {
	return instance.parent
}
