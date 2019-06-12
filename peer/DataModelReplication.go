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
// is specific to a network peer.
// ReplicationContainer keeps track of DataModel events for the instance it's
// bound to.
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

// Replicator is a system that listens to DataModel updates, keeps track of
// which instances have been replicated and sends the appropriate packets to
// the peer it is bound to.
type Replicator struct {
	ReplicatedInstances map[*datamodel.Instance]*ReplicationContainer
	handlingChild       *datamodel.Instance
	handlingProp        handledChange
	handlingEvent       handledChange
	handlingRemoval     *datamodel.Instance

	reader *DefaultPacketReader
	writer *DefaultPacketWriter
}

// ContainerFor returns the ReplicationContainer corresponding to an instance.
// If no such container exists, it will be created according to the options:
// (1) replicate children
// (2) replicate properties and events
// (3) replicate parent changes
// If options are given but the instance has been replicated, the container will be updated.
func (replicator *Replicator) ContainerFor(instance *datamodel.Instance, options ...bool) *ReplicationContainer {
	cont, ok := replicator.ReplicatedInstances[instance]
	if !ok {
		cont = &ReplicationContainer{
			Instance:            instance,
			ReplicateProperties: true,
			ReplicateChildren:   true,
			ReplicateParent:     true,
		}

		replicator.ReplicatedInstances[instance] = cont
	}
	if !ok || len(options) != 0 {
		if len(options) != 0 {
			cont.ReplicateChildren = options[0]
			cont.ReplicateProperties = options[1]
			cont.ReplicateParent = options[2]
		}
		cont.update(replicator)
	}

	return cont
}

type eventHandler func(*emitter.Event)

func (replicator *Replicator) isHandlingProperty(instance *datamodel.Instance, name string) bool {
	return replicator.handlingProp.Instance == instance && replicator.handlingProp.Name == name
}
func (replicator *Replicator) isHandlingEvent(instance *datamodel.Instance, name string) bool {
	return replicator.handlingEvent.Instance == instance && replicator.handlingEvent.Name == name
}

func (replicator *Replicator) writeDataPacket(packet Packet83Subpacket) error {
	return replicator.writer.WritePacket(&Packet83Layer{SubPackets: []Packet83Subpacket{packet}})
}

func (replicator *Replicator) parentHandler(instance *datamodel.Instance) eventHandler {
	return func(e *emitter.Event) {
		newParent := e.Args[0].(*datamodel.Instance)
		if !instance.IsService && newParent == nil {
			// instance was destroyed
			// so it must have been replicated previously
			// however, service parents are locked so parent updates for them
			// are ignored
			// sometimes the server will report their parent as nil
			// unbind listeners
			replicator.ContainerFor(instance, false, false, false)

			if replicator.handlingRemoval != instance && !replicator.isHandlingProperty(instance, "Parent") {
				err := replicator.writeDataPacket(&Packet83_01{Instance: instance})
				if err != nil {
					println("Parent handler error:", instance.GetFullName(), err.Error())
				}
			}
		}
	}
}

// ReplicationInstance creates a new ReplicationInstance for
// the given DataModel instance
func (replicator *Replicator) ReplicationInstance(inst *datamodel.Instance, deleteOnDisconnect bool) *ReplicationInstance {
	repInstance := &ReplicationInstance{}
	repInstance.DeleteOnDisconnect = deleteOnDisconnect
	repInstance.Instance = inst
	repInstance.Parent = inst.Parent()
	repInstance.Schema = replicator.writer.context.NetworkSchema.SchemaForClass(inst.ClassName)
	inst.PropertiesMutex.RLock()
	repInstance.Properties = make(map[string]rbxfile.Value, len(inst.Properties))
	for name, value := range inst.Properties {
		repInstance.Properties[name] = value
	}
	inst.PropertiesMutex.RUnlock()

	return repInstance
}

func (replicator *Replicator) replicateChildren(parent *datamodel.Instance) error {
	for _, subChild := range parent.Children {
		err := replicator.replicateNewChild(parent, subChild)
		if err != nil {
			return err
		}
	}
	return nil
}

func (replicator *Replicator) replicateNewChild(parent *datamodel.Instance, newChild *datamodel.Instance) error {
	childConfig := replicator.ContainerFor(newChild)
	if !childConfig.hasReplicated && replicator.handlingChild != newChild {
		// send the child as a new instance
		err := replicator.writeDataPacket(&Packet83_02{replicator.ReplicationInstance(newChild, false)})
		if err != nil {
			return err
		}

		// this instance's children probably haven't been replicated
		// and if they have, their parent update hasn't been replicated
		// because this instance (as a parent) didn't have childAdded bound
		// when the children's parents were updated
		err = replicator.replicateChildren(newChild)
		if err != nil {
			return err
		}
	} else if childConfig.hasReplicated && !replicator.isHandlingProperty(newChild, "Parent") && childConfig.ReplicateParent {
		// the child exists on the client; its parent has been updated
		err := replicator.writeDataPacket(&Packet83_03{
			Instance: newChild,
			Schema:   nil, // Parent property
			Value:    datamodel.ValueReference{Instance: parent, Reference: parent.Ref},
		})
		if err != nil {
			return err
		}

		// because this is a parent update, we don't need to cascade it to children
	}
	childConfig.hasReplicated = true

	return nil
}

func (replicator *Replicator) childHandler(instance *datamodel.Instance) eventHandler {
	return func(e *emitter.Event) {
		newChild := e.Args[0].(*datamodel.Instance)
		err := replicator.replicateNewChild(instance, newChild)
		if err != nil {
			println("Child handler error:", instance.GetFullName(), err.Error())
		}
	}
}

func (replicator *Replicator) propertyHandler(instance *datamodel.Instance) eventHandler {
	return func(e *emitter.Event) {
		name := e.OriginalTopic
		value := e.Args[0].(rbxfile.Value)
		if !replicator.isHandlingProperty(instance, name) {
			replicator.writeDataPacket(&Packet83_03{
				Instance: instance,
				Schema:   replicator.writer.context.NetworkSchema.SchemaForClass(instance.ClassName).SchemaForProp(name),
				Value:    value,
			})
		}
	}
}

func (replicator *Replicator) eventHandler(instance *datamodel.Instance) eventHandler {
	return func(e *emitter.Event) {
		name := e.OriginalTopic
		args := e.Args[0].([]rbxfile.Value)
		if !replicator.isHandlingEvent(instance, name) {
			replicator.writeDataPacket(&Packet83_07{
				Instance: instance,
				Schema:   replicator.writer.context.NetworkSchema.SchemaForClass(instance.ClassName).SchemaForEvent(name),
				Event:    &ReplicationEvent{args},
			})
		}
	}
}

func (replicator *Replicator) addTopInstanceChildren(children []*datamodel.Instance, streamer *JoinDataStreamer) error {
	for _, child := range children {
		config := replicator.ContainerFor(child)
		if config.hasReplicated {
			// skip if this instance has been replicated
			continue
		}
		config.hasReplicated = true
		err := streamer.AddInstance(replicator.ReplicationInstance(child, false))
		if err != nil {
			return err
		}
		err = replicator.addTopInstanceChildren(child.Children, streamer)
		if err != nil {
			return err
		}
	}
	return nil
}

// AddTopInstance manually adds an instance for replication.
// Typically it is used by the server to add services for replication.
func (replicator *Replicator) AddTopInstance(rootInstance *datamodel.Instance, replicateChildren, replicateProperties bool, streamer *JoinDataStreamer) error {
	var err error

	rootInstance.PropertiesMutex.RLock()
	if replicateProperties && len(rootInstance.Properties) != 0 {
		err = streamer.AddInstance(replicator.ReplicationInstance(rootInstance, false))
		if err != nil {
			return err
		}
	} else if replicateProperties {
		switch rootInstance.ClassName {
		case "AdService",
			"Workspace",
			"JointsService",
			"Players",
			"StarterGui",
			"StarterPack":

			fmt.Printf("Warning: skipping replication of bad instance %s (no properties and no defaults), replicateProperties: %v\n", rootInstance.ClassName, replicateProperties)
		default:
			err = streamer.AddInstance(replicator.ReplicationInstance(rootInstance, false))
			if err != nil {
				return err
			}
		}
	}
	rootInstance.PropertiesMutex.RUnlock()

	if replicateChildren {
		err = replicator.addTopInstanceChildren(rootInstance.Children, streamer)
		if err != nil {
			return err
		}
	}

	// Don't pass bool arguments here; the config should have already been done by topReplicate() or a similar system
	cont := replicator.ContainerFor(rootInstance)
	cont.hasReplicated = true

	return nil
}

// Bind makes the Replicator listen to reader's packets and write packets to the given writer.
func (replicator *Replicator) Bind(reader *DefaultPacketReader, writer *DefaultPacketWriter) {
	replicator.reader = reader
	replicator.writer = writer

	reader.PacketEmitter.On("ID_SET_GLOBALS", func(e *emitter.Event) {
		packet := e.Args[0].(*Packet81Layer)

		for _, item := range packet.Items {
			replicator.ContainerFor(item.Instance, item.WatchChildren, item.WatchChanges, true)

			reader.context.DataModel.AddService(item.Instance)
		}
	}, emitter.Void)

	dataEmitter := reader.DataEmitter
	dataEmitter.On("ID_REPLIC_DELETE_INSTANCE", func(e *emitter.Event) {
		inst := e.Args[0].(*Packet83_01).Instance
		replicator.handlingRemoval = inst
		reader.HandlePacket01(e)
		replicator.handlingRemoval = nil
	}, emitter.Void)

	dataEmitter.On("ID_REPLIC_NEW_INSTANCE", func(e *emitter.Event) {
		inst := e.Args[0].(*Packet83_02).ReplicationInstance.Instance
		replicator.handlingChild = inst
		replicator.handlingProp = handledChange{
			Instance: inst,
			Name:     "Parent",
		}

		reader.HandlePacket02(e)

		replicator.handlingChild = nil
		replicator.handlingProp = handledChange{}
	}, emitter.Void)

	dataEmitter.On("ID_REPLIC_PROP", func(e *emitter.Event) {
		packet := e.Args[0].(*Packet83_03)
		propName := "Parent"
		if packet.Schema != nil {
			propName = packet.Schema.Name
		}
		replicator.handlingProp = handledChange{
			Instance: packet.Instance,
			Name:     propName,
		}
		reader.HandlePacket03(e)
		replicator.handlingProp = handledChange{}
	}, emitter.Void)

	dataEmitter.On("ID_REPLIC_EVENT", func(e *emitter.Event) {
		packet := e.Args[0].(*Packet83_07)
		replicator.handlingEvent = handledChange{
			Instance: packet.Instance,
			Name:     packet.Schema.Name,
		}
		reader.HandlePacket07(e)
		replicator.handlingEvent = handledChange{}
	}, emitter.Void)

	dataEmitter.On("ID_REPLIC_JOIN_DATA", func(e *emitter.Event) {
		instanceList := e.Args[0].(*Packet83_0B).Instances
		for _, inst := range instanceList {
			replicator.handlingChild = inst.Instance
			replicator.handlingProp = handledChange{
				Instance: inst.Instance,
				Name:     "Parent",
			}

			err := reader.handleReplicationInstance(inst)
			if err != nil {
				e.Args[1].(*PacketLayers).Error = err
				return
			}
		}

		replicator.handlingChild = nil
		replicator.handlingProp = handledChange{}
	}, emitter.Void)

	dataEmitter.On("ID_REPLIC_ATOMIC", func(e *emitter.Event) {
		packet := e.Args[0].(*Packet83_13)
		replicator.handlingChild = packet.Instance
		replicator.handlingProp = handledChange{
			Instance: packet.Instance,
			Name:     "Parent",
		}

		packet.Instance.SetParent(packet.Parent)

		replicator.handlingChild = nil
		replicator.handlingProp = handledChange{}
	}, emitter.Void)
}

// NewReplicator creates a new Replicator object.
func NewReplicator() *Replicator {
	return &Replicator{
		ReplicatedInstances: make(map[*datamodel.Instance]*ReplicationContainer),
	}
}

func (cont *ReplicationContainer) update(replicator *Replicator) {
	inst := cont.Instance

	if cont.parentBinding == nil && cont.ReplicateParent {
		cont.parentBinding = inst.ParentEmitter.On("*", replicator.parentHandler(inst), emitter.Void)
	} else if !cont.ReplicateParent && cont.parentBinding != nil {
		// Use a goroutine here!
		// The emitter's mutex may be locked, because ReplicateParent
		// is typically updated by the container's own parentBinding
		go func() {
			inst.ParentEmitter.Off("*", cont.parentBinding)
		}()
	}

	if cont.childBinding == nil && cont.ReplicateChildren {
		cont.childBinding = inst.ChildEmitter.On("*", replicator.childHandler(inst), emitter.Void)
	} else if !cont.ReplicateChildren && cont.childBinding != nil {
		// ChildEmitter's binding should never be locked here
		inst.ChildEmitter.Off("*", cont.childBinding)
	}

	if cont.propBinding == nil && cont.ReplicateProperties {
		cont.propBinding = inst.PropertyEmitter.On("*", replicator.propertyHandler(inst), emitter.Void)
		cont.eventBinding = inst.EventEmitter.On("*", replicator.eventHandler(inst), emitter.Void)
	} else if !cont.ReplicateProperties && cont.propBinding != nil {
		// These bindings should never be lccked here
		inst.PropertyEmitter.Off("*", cont.propBinding)
		inst.EventEmitter.Off("*", cont.eventBinding)
	}

	if cont.ReplicateChildren {
		// Cascade update
		for _, child := range inst.Children {
			replicator.ContainerFor(child)
		}
	}
}
