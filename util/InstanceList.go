package util

import "github.com/gskartwii/rbxfile"
import "github.com/gskartwii/roblox-dissector/bitstreams"
import "errors"

type InstanceListScope struct {
    Instances map[uint32]*rbxfile.Instance
    AddCallbacks map[uint32][]func(*rbxfile.Instance)

type InstanceList struct {
    Scopes map[string]InstanceListScope
}

func (l *InstanceList) GetScope(ref Reference) map[uint32]*rbxfile.Instance {
    scope, ok := l.Scopes[ref.Scope]
    if !ok {
        scope = make(map[uint32]*rbxfile.Instance)
        l.Scopes[ref.Scope] = scope
    }
    return scope
}

func (l *InstanceList) CreateInstance(ref Reference) (*rbxfile.Instance, error) {
	if ref.IsNull {
		return nil, errors.New("Instance to create is nil!")
	}
	instance := l.GetScope(ref).Instance[ref.Id]
	if instance == nil {
		instance = &rbxfile.Instance{Reference: ref.String(), Properties: make(map[string]rbxfile.Value)}
		l.AddInstance(ref, instance)
		return instance, nil
	}
	// Allow rebinds. I don't know if this is right, but it can't hurt, right?
	return instance, nil
}

func (l *InstanceList) TryGetInstance(ref Reference) (*rbxfile.Instance, error) {
	if ref.IsNull {
		return nil, nil
	}

	instance := l.GetScope(ref).Instances[ref.Id]
	if instance == nil {
		return nil, errors.New("Instance doesn't exist!")
	}
	return instance, nil
}

func (l *InstanceList) AddInstance(ref Reference, instance *rbxfile.Instance) {
    scope := l.GetScope(ref)
	scope.Instances[ref.Id] = instance
	for _, callback := range scope.AddCallbacks[ref.Id] {
		callback(instance)
	}
}

func (l *InstanceList) OnAddInstance(ref Reference, callback func(*rbxfile.Instance)) error {
	if ref.IsNull {
		return errors.New("Instance to wait on can't be nil!")
	}

    scope := l.GetScope(ref)
	instance := scope.Instances[ref.Id]
	if instance == nil {
		thisPend := scope.AddCallbacks[ref.Id]
		thisPend = append(thisPend, callback)
	} else {
		callback(instance)
	}

	return nil
}
