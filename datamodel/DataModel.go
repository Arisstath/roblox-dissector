package datamodel

import (
	"github.com/olebedev/emitter"
)

type DataModel struct {
	Instances      []*Instance
	ServiceEmitter *emitter.Emitter
}

func New() *DataModel {
	return &DataModel{
		Instances:      make([]*Instance, 0),
		ServiceEmitter: emitter.New(1),
	}
}

func (model *DataModel) AddService(service *Instance) {
	model.Instances = append(model.Instances, service)
	<-model.ServiceEmitter.Emit(service.ClassName, service)
}
func (model *DataModel) FindService(name string) *Instance {
	for _, service := range model.Instances {
		if service.ClassName == name {
			return service
		}
	}
	return nil
}

func (model *DataModel) WaitForService(name string) <-chan *Instance {
	servChan := make(chan *Instance, 1)
	service := model.FindService(name)
	if service != nil {
		servChan <- service
		return servChan
	}
	emitterChan := model.ServiceEmitter.Once(name)
	go func(emitterChan <-chan emitter.Event) {
		// If the service was added while the emitter was being created
		service := model.FindService(name)
		if service != nil {
			model.ServiceEmitter.Off(name, emitterChan)
			servChan <- service
			return
		}

		servMiddle := (<-emitterChan).Args[0].(*Instance)
		servChan <- servMiddle
	}(emitterChan)
	return servChan
}

func (model *DataModel) WaitForChild(names ...string) <-chan *Instance {
	retChan := make(chan *Instance, 1)
	go func() {
		instance := <-model.WaitForService(names[0])
		if len(names) > 0 {
			retChan <- (<-instance.WaitForChild(names[1:]...))
		} else {
			retChan <- instance
		}
	}()
	return retChan
}

func (model *DataModel) Copy() *DataModel {
	newModel := New()
	pool := NewSelfReferencePool()
	newModel.Instances = make([]*Instance, len(model.Instances))

	for i, inst := range model.Instances {
		newModel.Instances[i] = inst.Copy(pool)
	}

	return newModel
}
