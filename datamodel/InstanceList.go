package datamodel

import (
	"errors"
)

type instanceScope struct {
	Instances    map[uint32]*Instance
	AddCallbacks map[uint32][]func(*Instance)
}

type InstanceList struct {
	scopes map[string]*instanceScope
}

var ErrNullInstance = errors.New("instance is null")
var ErrInstanceDoesntExist = errors.New("instance doesn't exist")

func newInstanceScope() *instanceScope {
	return &instanceScope{
		Instances:    make(map[uint32]*Instance),
		AddCallbacks: make(map[uint32][]func(*Instance)),
	}
}

func NewInstanceList() *InstanceList {
	return &InstanceList{scopes: make(map[string]*instanceScope)}
}

func (l *InstanceList) getScope(ref Reference) *instanceScope {
	scope, ok := l.scopes[ref.Scope]
	if !ok {
		scope = newInstanceScope()
		l.scopes[ref.Scope] = scope
	}
	return scope
}

func (l *InstanceList) CreateInstance(ref Reference) (*Instance, error) {
	if ref.IsNull {
		return nil, ErrNullInstance
	}
	instance := l.getScope(ref).Instances[ref.Id]
	if instance == nil {
		instance, _ = NewInstance("", nil)
		instance.Ref = ref
		l.AddInstance(ref, instance)
		return instance, nil
	}
	// Allow rebinds. I don't know if this is right, but it can't hurt, right?
	return instance, nil
}

func (l *InstanceList) TryGetInstance(ref Reference) (*Instance, error) {
	if ref.IsNull {
		return nil, nil
	}

	instance := l.getScope(ref).Instances[ref.Id]
	if instance == nil {
		return nil, ErrInstanceDoesntExist
	}
	return instance, nil
}

func (l *InstanceList) AddInstance(ref Reference, instance *Instance) {
	l.getScope(ref).Instances[ref.Id] = instance
	for _, callback := range l.getScope(ref).AddCallbacks[ref.Id] {
		callback(instance)
	}
}

func (l *InstanceList) OnAddInstance(ref Reference, callback func(*Instance)) error {
	if ref.IsNull {
		return ErrNullInstance
	}

	scope := l.getScope(ref)
	instance := scope.Instances[ref.Id]
	if instance == nil {
		scope.AddCallbacks[ref.Id] = append(scope.AddCallbacks[ref.Id], callback)
	} else {
		callback(instance)
	}

	return nil
}

func (l *InstanceList) Populate(instances []*Instance) {
	for _, inst := range instances {
		l.AddInstance(inst.Ref, inst)
		l.Populate(inst.Children)
	}
}
