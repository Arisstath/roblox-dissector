package datamodel

import (
	"context"
	"errors"
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
	IsService       bool
	Ref             Reference
	ChildEmitter    *emitter.Emitter
	PropertyEmitter *emitter.Emitter
	EventEmitter    *emitter.Emitter
	ParentEmitter   *emitter.Emitter
	parent          *Instance
}

func NewInstance(className string, parent *Instance) (*Instance, error) {
	inst := &Instance{
		ClassName:       className,
		Properties:      make(map[string]rbxfile.Value),
		PropertiesMutex: &sync.RWMutex{},
		ChildEmitter:    emitter.New(1),
		PropertyEmitter: emitter.New(1),
		EventEmitter:    emitter.New(1),
		ParentEmitter:   emitter.New(1),
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
	oldParent := child.parent
	if oldParent != nil {
		for i, c := range oldParent.Children {
			if c == child {
				copy(oldParent.Children[i:], oldParent.Children[i+1:])
				oldParent.Children[len(oldParent.Children)-1] = nil
				oldParent.Children = oldParent.Children[:len(oldParent.Children)-1]
			}
		}
	}

	child.parent = instance
	if instance != nil {
		instance.Children = append(instance.Children, child)
		<-instance.ChildEmitter.Emit(child.Name(), child)
	}

	parentName := ""
	if instance != nil {
		parentName = instance.Name()
	}
	<-child.ParentEmitter.Emit(parentName, instance)
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
	<-instance.PropertyEmitter.Emit(name, value)
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
func (instance *Instance) waitForChild(ctx context.Context, name string) (*Instance, error) {
	if child := instance.FindFirstChild(name); child != nil {
		return child, nil
	}
	emitterChan := instance.ChildEmitter.Once(name)
	// If the child is added while we created the emitter
	if child := instance.FindFirstChild(name); child != nil {
		instance.ChildEmitter.Off(name, emitterChan)
		return child, nil
	}
	select {
	case e := <-emitterChan:
		// No need to unbind because it's Once()
		child := e.Args[0].(*Instance)
		return child, nil
	case <-ctx.Done():
		instance.ChildEmitter.Off(name, emitterChan)
		return nil, ctx.Err()
	}
}

func (instance *Instance) WaitForChild(ctx context.Context, names ...string) (*Instance, error) {
	var err error
	for _, name := range names {
		instance, err = instance.waitForChild(ctx, name)
		if err != nil {
			return instance, err
		}
	}
	return instance, nil
}

func (instance *Instance) WaitForProp(ctx context.Context, name string) (rbxfile.Value, error) {
	instance.PropertiesMutex.RLock()
	emitterChan := instance.PropertyEmitter.Once(name)
	instance.PropertiesMutex.RUnlock()

	select {
	case e := <-emitterChan:
		return e.Args[0].(rbxfile.Value), nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}
func (instance *Instance) WaitForRefProp(ctx context.Context, name string) (*Instance, error) {
	instance.PropertiesMutex.RLock()
	currProp := instance.Properties[name]
	if currProp != nil && currProp.(ValueReference).Instance != nil {
		instance.PropertiesMutex.RUnlock()
		return currProp.(ValueReference).Instance, nil
	}
	propEvtChan := instance.PropertyEmitter.On(name)
	instance.PropertiesMutex.RUnlock()
	for {
		select {
		case propEvt := <-propEvtChan:
			if propEvt.Args[0] == nil {
				continue
			}
			currProp := propEvt.Args[0].(ValueReference)
			if currProp.Instance != nil {
				instance.PropertyEmitter.Off(name, propEvtChan)
				return currProp.Instance, nil
			}
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
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
	return builder.String()[1:]
}
func (instance *Instance) FireEvent(name string, args ...rbxfile.Value) {
	<-instance.EventEmitter.Emit(name, []rbxfile.Value(args))
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
