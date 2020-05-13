package datamodel

import (
	"errors"
)

// TODO: Should work with PeerID
type instanceScope struct {
	Instances map[uint32]*Instance
}

type InstanceList struct {
	scopes map[string]*instanceScope
}

var ErrNullInstance = errors.New("instance is null")
var ErrInstanceDoesntExist = errors.New("instance doesn't exist")

func newInstanceScope() *instanceScope {
	return &instanceScope{
		Instances: make(map[uint32]*Instance),
	}
}

func NewInstanceList() *InstanceList {
	return &InstanceList{scopes: make(map[string]*instanceScope)}
}

func (s *instanceScope) remove(id uint32) {
	delete(s.Instances, id)
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
}

func (l *InstanceList) Populate(instances []*Instance) {
	for _, inst := range instances {
		l.AddInstance(inst.Ref, inst)
		l.Populate(inst.Children)
	}
}

func (l *InstanceList) RemoveTree(instance *Instance) {
	l.getScope(instance.Ref).remove(instance.Ref.Id)

	for _, child := range instance.Children {
		l.RemoveTree(child)
	}
}
