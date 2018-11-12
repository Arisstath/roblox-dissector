package util

import "github.com/gskartwii/rbxfile"
import "errors"

type DeserializedInstance struct {
    *rbxfile.Instance
    Reference Reference
    Parent Reference
}

type InstanceListScope struct {
    Instances map[uint32]DeserializedInstance
    AddCallbacks map[uint32][]func(DeserializedInstance)
}

type InstanceList struct {
    Scopes map[string]InstanceListScope
}

func NewInstanceList() *InstanceList {
    return &InstanceList{Scopes: make(map[string]InstanceListScope)}
}

func (l *InstanceList) GetScope(ref Reference) InstanceListScope {
    scope, ok := l.Scopes[ref.Scope]
    if !ok {
        scope = InstanceListScope{Instances: make(map[uint32]DeserializedInstance), AddCallbacks: make(map[uint32][]func(DeserializedInstance))}
        l.Scopes[ref.Scope] = scope
    }
    return scope
}

func (l *InstanceList) CreateInstance(ref Reference) (DeserializedInstance, error) {
	if ref.IsNull {
		return DeserializedInstance{}, errors.New("Instance to create is nil!")
	}
	instance, ok := l.GetScope(ref).Instances[ref.Id]
	if !ok {
        instance = DeserializedInstance{Instance: &rbxfile.Instance{Reference: ref.String(), Properties: make(map[string]rbxfile.Value)}, Reference: ref}
		l.AddInstance(ref, instance)
		return instance, nil
	}
	// Allow rebinds. I don't know if this is right, but it can't hurt, right?
	return instance, nil
}

func (l *InstanceList) TryGetInstance(ref Reference) (DeserializedInstance, error) {
	if ref.IsNull {
		return DeserializedInstance{}, nil
	}

	instance, ok := l.GetScope(ref).Instances[ref.Id]
	if !ok {
		return DeserializedInstance{}, errors.New("Instance doesn't exist!")
	}
	return instance, nil
}

func (l *InstanceList) AddInstance(ref Reference, instance DeserializedInstance) {
    scope := l.GetScope(ref)
	scope.Instances[ref.Id] = instance
	for _, callback := range scope.AddCallbacks[ref.Id] {
		callback(instance)
	}
}

func (l *InstanceList) OnAddInstance(ref Reference, callback func(DeserializedInstance)) error {
	if ref.IsNull {
		return errors.New("Instance to wait on can't be nil!")
	}

    scope := l.GetScope(ref)
	instance, ok := scope.Instances[ref.Id]
	if !ok {
		thisPend := scope.AddCallbacks[ref.Id]
		thisPend = append(thisPend, callback)
	} else {
		callback(instance)
	}

	return nil
}
