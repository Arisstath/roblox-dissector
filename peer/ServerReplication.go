package peer

import (
	"github.com/gskartwii/roblox-dissector/datamodel"
	"github.com/olebedev/emitter"
	"github.com/robloxapi/rbxfile"
)

type handledChange struct {
	Instance *datamodel.Instance
	Name     string
}

// ReplicationContainer represents replication config for an instance that
// is specific to a server client
type ReplicationContainer struct {
	Instance            *datamodel.Instance
	ReplicateProperties bool
	ReplicateChildren   bool
	ReplicateParent     bool

	childBinding  <-chan emitter.Event
	propBinding   <-chan emitter.Event
	eventBinding  <-chan emitter.Event
	parentBinding <-chan emitter.Event
}

func (cont *ReplicationContainer) UpdateBinding(client *ServerClient, isNew bool) {
	inst := cont.Instance

	if cont.parentBinding == nil && cont.ReplicateParent {
		cont.parentBinding = inst.ParentEmitter.On("*", func(e *emitter.Event) {
			client.ParentChangedHandler(inst, e)
		}, emitter.Void)
	} else if !cont.ReplicateParent && cont.parentBinding != nil {
		inst.ParentEmitter.Off("*", cont.parentBinding)
	}

	if cont.childBinding == nil && cont.ReplicateChildren {
		cont.childBinding = inst.ChildEmitter.On("*", func(e *emitter.Event) {
			client.ChildAddedHandler(inst, e)
		}, emitter.Void)
	} else if !cont.ReplicateChildren && cont.childBinding != nil {
		inst.ChildEmitter.Off("*", cont.childBinding)
	}

	if cont.propBinding == nil && cont.ReplicateProperties {
		cont.propBinding = inst.PropertyEmitter.On("*", func(e *emitter.Event) {
			client.PropChangedHandler(inst, e)
		}, emitter.Void)
		cont.eventBinding = inst.EventEmitter.On("*", func(e *emitter.Event) {
			client.EventHandler(inst, e)
		}, emitter.Void)
	} else if !cont.ReplicateProperties && cont.propBinding != nil {
		inst.PropertyEmitter.Off("*", cont.propBinding)
		inst.EventEmitter.Off("*", cont.eventBinding)
	}

	// Cascade update
	for _, child := range inst.Children {
		client.UpdateBinding(child, isNew)
	}
}

func joinDataConfigForInstance(inst *datamodel.Instance) *JoinDataConfig {
	for inst.Parent() != nil && inst.Parent().ClassName != "DataModel" {
		inst = inst.Parent()
	}

	for _, config := range JoinDataConfiguration {
		if config.ClassName == inst.ClassName {
			return &config
		}
	}

	return nil
}

func (client *ServerClient) BindDefaultDatamodelHandlers() {
	dataEmitter := client.DataEmitter
	dataEmitter.On("ID_REPLIC_DELETE_INSTANCE", func(e *emitter.Event) {
		inst := e.Args[0].(*Packet83_01).Instance
		client.handlingRemoval = inst
		client.DefaultPacketReader.HandlePacket01(e)
		client.handlingRemoval = nil
	}, emitter.Void)

	dataEmitter.On("ID_REPLIC_NEW_INSTANCE", func(e *emitter.Event) {
		inst := e.Args[0].(*Packet83_02).ReplicationInstance.Instance
		client.handlingChild = inst
		client.handlingProp = handledChange{
			Instance: inst,
			Name:     "Parent",
		}

		client.DefaultPacketReader.HandlePacket02(e)

		client.handlingChild = nil
		client.handlingProp = handledChange{}
	}, emitter.Void)

	dataEmitter.On("ID_REPLIC_PROP", func(e *emitter.Event) {
		packet := e.Args[0].(*Packet83_03)
		propName := "Parent"
		if packet.Schema != nil {
			propName = packet.Schema.Name
		}
		client.handlingProp = handledChange{
			Instance: packet.Instance,
			Name:     propName,
		}
		client.DefaultPacketReader.HandlePacket03(e)
		client.handlingProp = handledChange{}
	}, emitter.Void)

	dataEmitter.On("ID_REPLIC_EVENT", func(e *emitter.Event) {
		packet := e.Args[0].(*Packet83_07)
		client.handlingEvent = handledChange{
			Instance: packet.Instance,
			Name:     packet.Schema.Name,
		}
		client.DefaultPacketReader.HandlePacket07(e)
		client.handlingEvent = handledChange{}
	}, emitter.Void)

	dataEmitter.On("ID_REPLIC_JOIN_DATA", func(e *emitter.Event) {
		instanceList := e.Args[0].(*Packet83_0B).Instances
		for _, inst := range instanceList {
			client.handlingChild = inst.Instance
			client.handlingProp = handledChange{
				Instance: inst.Instance,
				Name:     "Parent",
			}

			err := client.handleReplicationInstance(inst)
			if err != nil {
				e.Args[1].(*PacketLayers).Error = err
				return
			}
		}

		client.handlingChild = nil
		client.handlingProp = handledChange{}
	}, emitter.Void)

	dataEmitter.On("ID_REPLIC_ATOMIC", func(e *emitter.Event) {
		packet := e.Args[0].(*Packet83_13)
		client.handlingChild = packet.Instance
		client.handlingProp = handledChange{
			Instance: packet.Instance,
			Name:     "Parent",
		}

		packet.Instance.SetParent(packet.Parent)

		client.handlingChild = nil
		client.handlingProp = handledChange{}
	}, emitter.Void)
}

func (client *ServerClient) ReplicationConfig(inst *datamodel.Instance) *ReplicationContainer {
	for _, conf := range client.replicatedInstances {
		if conf.Instance == inst {
			return conf
		}
	}

	return nil
}
func (client *ServerClient) IsHandlingChild(child *datamodel.Instance) bool {
	return client.handlingChild == child
}
func (client *ServerClient) IsHandlingProp(inst *datamodel.Instance, name string) bool {
	return client.handlingProp.Instance == inst && client.handlingProp.Name == name
}
func (client *ServerClient) IsHandlingEvent(inst *datamodel.Instance, name string) bool {
	return client.handlingEvent.Instance == inst && client.handlingEvent.Name == name
}
func (client *ServerClient) IsHandlingRemoval(inst *datamodel.Instance) bool {
	return client.handlingRemoval == inst
}

func (client *ServerClient) ParentChangedHandler(inst *datamodel.Instance, e *emitter.Event) {
	if client.IsHandlingRemoval(inst) || client.IsHandlingProp(inst, "Parent") {
		// avoid circular replication: if this parent change
		// comes from the client, we ignore it
		return
	}

	newParent := e.Args[0].(*datamodel.Instance)
	if newParent == nil {
		// instance has been removed, :Destroy() it
		client.WriteDataPackets(&Packet83_01{
			Instance: inst,
		})
	}

	parentConfig := client.ReplicationConfig(newParent)
	if parentConfig == nil {
		// if the new parent hasn't been replicated to the client,
		// the scenario should be handled by parenting this instance to nil
		client.WriteDataPackets(&Packet83_03{
			Instance: inst,
			Schema:   nil, // Parent
			Value:    datamodel.ValueReference{Reference: datamodel.NullReference},
		})
	}

	// If the parent has been replicated, ChildAddedHandler will replicate the appropriate change
}
func (client *ServerClient) ChildAddedHandler(parent *datamodel.Instance, e *emitter.Event) {
	child := e.Args[0].(*datamodel.Instance)

	childConfig := client.ReplicationConfig(child)
	if childConfig != nil {
		client.UpdateBinding(child, false)

		if client.IsHandlingProp(child, "Parent") {
			// this client caused the parent change, won't replicate
			return
		}
		// Instance has already been replicated
		// don't call ReplicateInstance(), instead update the parent

		client.WriteDataPackets(&Packet83_03{
			Instance: child,
			Schema:   nil, // Parent property
			Value:    datamodel.ValueReference{Instance: parent, Reference: parent.Ref},
		})
		return
	}

	// Bind to instance before replicating it to the client
	client.UpdateBinding(child, true)
}

func (client *ServerClient) PropChangedHandler(inst *datamodel.Instance, e *emitter.Event) {
	name := e.OriginalTopic
	value := e.Args[0].(rbxfile.Value)

	if !client.IsHandlingProp(inst, name) {
		client.WriteDataPackets(&Packet83_03{
			Instance: inst,
			Schema:   client.Context.StaticSchema.SchemaForClass(inst.ClassName).SchemaForProp(name),
			Value:    value,
		})
	}
}

func (client *ServerClient) EventHandler(inst *datamodel.Instance, e *emitter.Event) {
	name := e.OriginalTopic
	args := e.Args[0].([]rbxfile.Value)

	if !client.IsHandlingEvent(inst, name) {
		client.WriteDataPackets(&Packet83_07{
			Instance: inst,
			Schema:   client.Context.StaticSchema.SchemaForClass(inst.ClassName).SchemaForEvent(name),
			Event:    &ReplicationEvent{args},
		})
	}
}

func (client *ServerClient) UpdateBinding(inst *datamodel.Instance, isNew bool) {
	found := client.ReplicationConfig(inst)
	var parentConfig *ReplicationContainer
	if inst.Parent() == nil {
		// Consider the instance destroyed
		parentConfig = &ReplicationContainer{
			Instance:            inst,
			ReplicateProperties: false,
			ReplicateChildren:   false,
			ReplicateParent:     false,
		}
	} else {
		parentConfig = client.ReplicationConfig(inst.Parent())
	}
	if found == nil {
		// This instance has never been replicated to the client
		// inherit
		newBinding := &ReplicationContainer{
			Instance:            inst,
			ReplicateProperties: parentConfig.ReplicateProperties,
			ReplicateChildren:   parentConfig.ReplicateChildren,
			ReplicateParent:     parentConfig.ReplicateParent,
		}
		client.replicatedInstances = append(client.replicatedInstances, newBinding)

		if !client.IsHandlingChild(inst) && isNew {
			client.PacketLogicHandler.ReplicateInstance(inst, false)
		}

		// Because this instance hasn't been replicated, neither are its children
		newBinding.UpdateBinding(client, isNew)
	} else {
		found.ReplicateProperties = parentConfig.ReplicateProperties
		found.ReplicateChildren = parentConfig.ReplicateChildren
		found.ReplicateParent = parentConfig.ReplicateParent

		found.UpdateBinding(client, isNew)
	}
}

func (client *ServerClient) sendReplicatedFirst() error {
	replicatedFirstStreamer := NewJoinDataStreamer(client.DefaultPacketWriter)
	replicatedFirstStreamer.BufferEmitter.On("join-data", func(e *emitter.Event) {
		err := client.WriteDataPackets(e.Args[0].(Packet83Subpacket))
		if err != nil {
			println("replicatedfirst error: ", err.Error())
		}
	}, emitter.Void)
	replicatedFirst := client.constructInstanceList(nil, client.DataModel.FindService("ReplicatedFirst"))
	for _, repFirstInstance := range replicatedFirst {
		err := replicatedFirstStreamer.AddInstance(repFirstInstance)
		if err != nil {
			return err
		}
	}
	err := replicatedFirstStreamer.Close()
	if err != nil {
		return err
	}
	// Tag: ReplicatedFirst finished!
	return client.WriteDataPackets(&Packet83_10{
		TagId: 12,
	})
}

func (client *ServerClient) sendContainer(streamer *JoinDataStreamer, config JoinDataConfig) error {
	service := client.DataModel.FindService(config.ClassName)
	if service != nil {
		return client.ReplicateJoinData(service, config.ReplicateProperties, config.ReplicateChildren, streamer, client.Player)
	}
	return nil
}

func (client *ServerClient) sendContainers() error {
	var err error

	joinDataStreamer := NewJoinDataStreamer(client.DefaultPacketWriter)
	joinDataStreamer.BufferEmitter.On("join-data", func(e *emitter.Event) {
		err := client.WriteDataPackets(e.Args[0].(Packet83Subpacket))
		if err != nil {
			println("joindata error: ", err.Error())
		}
	}, emitter.Void)
	for _, dataConfig := range JoinDataConfiguration {
		// Previously replicated for priority, don't duplicate
		if dataConfig.ClassName != "ReplicatedFirst" {
			err = client.sendContainer(joinDataStreamer, dataConfig)
			if err != nil {
				return err
			}
		}
	}
	err = joinDataStreamer.Close()
	if err != nil {
		return err
	}

	return client.WriteDataPackets(&Packet83_10{
		TagId: 13,
	})
}
