package peer

import (
	"fmt"

	"github.com/Gskartwii/roblox-dissector/datamodel"
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

	hasReplicated bool
}

func (cont *ReplicationContainer) updateBinding(client *ServerClient, isNew bool) {
	inst := cont.Instance

	if cont.parentBinding == nil && cont.ReplicateParent {
		cont.parentBinding = inst.ParentEmitter.On("*", func(e *emitter.Event) {
			client.parentChangedHandler(inst, e)
		}, emitter.Void)
	} else if !cont.ReplicateParent && cont.parentBinding != nil {
		inst.ParentEmitter.Off("*", cont.parentBinding)
	}

	if cont.childBinding == nil && cont.ReplicateChildren {
		cont.childBinding = inst.ChildEmitter.On("*", func(e *emitter.Event) {
			client.childAddedHandler(inst, e)
		}, emitter.Void)
	} else if !cont.ReplicateChildren && cont.childBinding != nil {
		inst.ChildEmitter.Off("*", cont.childBinding)
	}

	if cont.propBinding == nil && cont.ReplicateProperties {
		cont.propBinding = inst.PropertyEmitter.On("*", func(e *emitter.Event) {
			client.propChangedHandler(inst, e)
		}, emitter.Void)
		cont.eventBinding = inst.EventEmitter.On("*", func(e *emitter.Event) {
			client.eventHandler(inst, e)
		}, emitter.Void)
	} else if !cont.ReplicateProperties && cont.propBinding != nil {
		inst.PropertyEmitter.Off("*", cont.propBinding)
		inst.EventEmitter.Off("*", cont.eventBinding)
	}

	// Cascade update
	for _, child := range inst.Children {
		client.updateBinding(child, isNew)
	}
}

func joinDataConfigForInstance(inst *datamodel.Instance) *joinDataConfig {
	for inst.Parent() != nil && inst.Parent().ClassName != "DataModel" {
		inst = inst.Parent()
	}

	for _, config := range joinDataConfiguration {
		if config.ClassName == inst.ClassName {
			return &config
		}
	}

	return nil
}

// BindDefaultDataModelHandlers binds the client's DataModel
// handlers so that the client's changes will be reflected in
// the DataModel
func (client *ServerClient) BindDefaultDataModelHandlers() {
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

// ReplicationConfig returns the replication configuration for
// an instance
func (client *ServerClient) ReplicationConfig(inst *datamodel.Instance) *ReplicationContainer {
	for _, conf := range client.replicatedInstances {
		if conf.Instance == inst {
			return conf
		}
	}

	return nil
}
func (client *ServerClient) isHandlingChild(child *datamodel.Instance) bool {
	return client.handlingChild == child
}
func (client *ServerClient) isHandlingProp(inst *datamodel.Instance, name string) bool {
	return client.handlingProp.Instance == inst && client.handlingProp.Name == name
}
func (client *ServerClient) isHandlingEvent(inst *datamodel.Instance, name string) bool {
	return client.handlingEvent.Instance == inst && client.handlingEvent.Name == name
}
func (client *ServerClient) isHandlingRemoval(inst *datamodel.Instance) bool {
	return client.handlingRemoval == inst
}

func (client *ServerClient) parentChangedHandler(inst *datamodel.Instance, e *emitter.Event) {
	if client.isHandlingRemoval(inst) || client.isHandlingProp(inst, "Parent") {
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
		// Note: we don't check hasReplicated here!
		// hasReplicated has the sense "has replicated _yet_"
		client.WriteDataPackets(&Packet83_03{
			Instance: inst,
			Schema:   nil, // Parent
			Value:    datamodel.ValueReference{Reference: datamodel.NullReference},
		})
	}

	// If the parent has been replicated, ChildAddedHandler will replicate the appropriate change
}
func (client *ServerClient) childAddedHandler(parent *datamodel.Instance, e *emitter.Event) {
	child := e.Args[0].(*datamodel.Instance)

	childConfig := client.ReplicationConfig(child)
	if childConfig != nil && childConfig.hasReplicated {
		client.updateBinding(child, false)

		if client.isHandlingProp(child, "Parent") {
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

	client.updateBinding(child, true)
}

func (client *ServerClient) propChangedHandler(inst *datamodel.Instance, e *emitter.Event) {
	name := e.OriginalTopic
	value := e.Args[0].(rbxfile.Value)

	if !client.isHandlingProp(inst, name) {
		client.WriteDataPackets(&Packet83_03{
			Instance: inst,
			Schema:   client.Context.NetworkSchema.SchemaForClass(inst.ClassName).SchemaForProp(name),
			Value:    value,
		})
	}
}

func (client *ServerClient) eventHandler(inst *datamodel.Instance, e *emitter.Event) {
	name := e.OriginalTopic
	args := e.Args[0].([]rbxfile.Value)

	if !client.isHandlingEvent(inst, name) {
		switch name {
		case "RemoteOnInvokeClient", "OnClientEvent":
			client.WriteDataPackets(&Packet83_07{
				Instance: inst,
				Schema:   client.Context.NetworkSchema.SchemaForClass(inst.ClassName).SchemaForEvent(name),
				Event:    &ReplicationEvent{args},
			})
		default:
			println("Warning: not replicating non-whitelisted event", name)
		}
	}
}

func (client *ServerClient) updateBinding(inst *datamodel.Instance, canReplicate bool) {
	found := client.ReplicationConfig(inst)
	var parentConfig *ReplicationContainer
	if inst.Parent() == nil {
		// Consider the instance destroyed
		parentConfig = &ReplicationContainer{
			Instance:            nil,
			ReplicateProperties: false,
			ReplicateChildren:   false,
			ReplicateParent:     false,
		}
	} else {
		parentConfig = client.ReplicationConfig(inst.Parent())
	}
	if found == nil {
		// This instance has never been replicated to the client
		// found == nil && !isNew means that this is joinData replication
		// hence we don't create a binding yet
		// inherit
		newBinding := &ReplicationContainer{
			Instance:            inst,
			ReplicateProperties: parentConfig.ReplicateProperties,
			ReplicateChildren:   parentConfig.ReplicateChildren,
			ReplicateParent:     parentConfig.ReplicateParent,
		}

		client.replicatedInstances = append(client.replicatedInstances, newBinding)
		if canReplicate && !client.isHandlingChild(inst) && !newBinding.hasReplicated {
			newBinding.hasReplicated = true
			client.PacketLogicHandler.ReplicateInstance(inst, false)
		} else if client.isHandlingChild(inst) {
			newBinding.hasReplicated = true
		}

		// Cascade to children
		newBinding.updateBinding(client, canReplicate)
	} else if found != nil {
		found.ReplicateProperties = parentConfig.ReplicateProperties
		found.ReplicateChildren = parentConfig.ReplicateChildren
		found.ReplicateParent = parentConfig.ReplicateParent

		found.updateBinding(client, canReplicate)
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

	service := client.DataModel.FindService("ReplicatedFirst")
	config := client.ReplicationConfig(service)
	err := client.replicateJoinData(service, config, replicatedFirstStreamer)
	if err != nil {
		return err
	}
	err = replicatedFirstStreamer.Close()
	if err != nil {
		return err
	}
	// Tag: ReplicatedFirst finished!
	return client.WriteDataPackets(&Packet83_10{
		TagID: 12,
	})
}

func (client *ServerClient) sendContainer(streamer *JoinDataStreamer, config joinDataConfig) error {
	service := client.DataModel.FindService(config.ClassName)
	if service != nil {
		repConfig := client.ReplicationConfig(service)
		return client.replicateJoinData(service, repConfig, streamer)
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
	for _, dataConfig := range joinDataConfiguration {
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
		TagID: 13,
	})
}

func (client *ServerClient) replicateJoinDataChildren(children []*datamodel.Instance, streamer *JoinDataStreamer) error {
	for _, child := range children {
		config := client.ReplicationConfig(child)
		if config == nil {
			println("warning: nil config for instance", child.GetFullName(), "skipping")
			continue
		}
		if config.hasReplicated {
			// Skip instances that have already been replicated
			continue
		}
		config.hasReplicated = true

		err := streamer.AddInstance(client.ReplicationInstance(child, false))
		println("added child", child.GetFullName())
		if err != nil {
			return err
		}
		err = client.replicateJoinDataChildren(child.Children, streamer)
		if err != nil {
			return err
		}
	}
	return nil
}

func (client *ServerClient) replicateJoinData(rootInstance *datamodel.Instance, rootConfig *ReplicationContainer, streamer *JoinDataStreamer) error {
	var err error
	// HACK: Replicating some instances to the client without including properties
	// may result in an error and a disconnection.
	// Here's a bad workaround
	rootInstance.PropertiesMutex.RLock()
	if rootConfig.ReplicateProperties && len(rootInstance.Properties) != 0 {
		println("Writing instance with properties", rootInstance.ClassName)
		rootConfig.hasReplicated = true
		err = streamer.AddInstance(client.ReplicationInstance(rootInstance, false))
		if err != nil {
			return err
		}
	} else if rootConfig.ReplicateProperties {
		switch rootInstance.ClassName {
		case "AdService",
			"Workspace",
			"JointsService",
			"Players",
			"StarterGui",
			"StarterPack":
			fmt.Printf("Warning: skipping replication of bad instance %s (no properties and no defaults), replicateProperties: %v\n", rootInstance.ClassName, rootConfig.ReplicateProperties)
		default:
			println("Writing instance without properties", rootInstance.ClassName)
			rootConfig.hasReplicated = true
			err = streamer.AddInstance(client.ReplicationInstance(rootInstance, false))
			if err != nil {
				return err
			}
		}
	}
	rootInstance.PropertiesMutex.RUnlock()

	if rootConfig.ReplicateChildren {
		return client.replicateJoinDataChildren(rootInstance.Children, streamer)
	}
	return nil
}
