package datamodel

import (
	"context"

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

func (model *DataModel) WaitForService(ctx context.Context, name string) (*Instance, error) {
	service := model.FindService(name)
	if service != nil {
		return service, nil
	}
	emitterChan := model.ServiceEmitter.Once(name)
	// If the service was added while the emitter was being created
	service = model.FindService(name)
	if service != nil {
		model.ServiceEmitter.Off(name, emitterChan)
		return service, nil
	}

	select {
	case e := <-emitterChan:
		servMiddle := e.Args[0].(*Instance)
		return servMiddle, nil
	case <-ctx.Done():
		model.ServiceEmitter.Off(name, emitterChan)
		return nil, ctx.Err()
	}
}

func (model *DataModel) WaitForChild(ctx context.Context, names ...string) (*Instance, error) {
	instance, err := model.WaitForService(ctx, names[0])
	// If we only have one name, don't need to call WaitForChild()
	if err != nil || len(names) == 1 {
		return instance, err
	}
	return instance.WaitForChild(ctx, names[1:]...)
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
