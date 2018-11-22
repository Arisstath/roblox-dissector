package peer

import (
	"errors"
	"sync"

	"github.com/gskartwii/rbxfile"
)

type InstanceList struct {
	Instances    map[string]*rbxfile.Instance
	AddCallbacks map[string][]func(*rbxfile.Instance)
}

func (l *InstanceList) CreateInstance(ref Referent) (*rbxfile.Instance, error) {
	if ref.IsNull() {
		return nil, errors.New("Instance to create is nil!")
	}
	instance := l.Instances[string(ref)]
	if instance == nil {
		instance = &rbxfile.Instance{Reference: string(ref), Properties: make(map[string]rbxfile.Value), PropertiesMutex: &sync.RWMutex{}}
		l.AddInstance(ref, instance)
		return instance, nil
	}
	// Allow rebinds. I don't know if this is right, but it can't hurt, right?
	return instance, nil
}

func (l *InstanceList) TryGetInstance(ref Referent) (*rbxfile.Instance, error) {
	if ref.IsNull() {
		return nil, nil
	}

	instance := l.Instances[string(ref)]
	if instance == nil {
		return nil, errors.New("Instance doesn't exist!")
	}
	return instance, nil
}

func (l *InstanceList) WaitForInstance(ref Referent) *rbxfile.Instance {
	instance := l.Instances[string(ref)]
	return instance
}
func (l *InstanceList) AddInstance(ref Referent, instance *rbxfile.Instance) {
	l.Instances[string(ref)] = instance
	for _, callback := range l.AddCallbacks[string(ref)] {
		callback(instance)
	}
}

func (l *InstanceList) OnAddInstance(ref Referent, callback func(*rbxfile.Instance)) error {
	if ref == "null" {
		return errors.New("Instance to wait on can't be nil!")
	}

	instance := l.Instances[string(ref)]
	if instance == nil {
		thisPend := l.AddCallbacks[string(ref)]
		thisPend = append(thisPend, callback)
	} else {
		callback(instance)
	}

	return nil
}
