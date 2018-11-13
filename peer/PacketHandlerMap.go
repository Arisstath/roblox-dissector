package peer

import "github.com/gskartwii/rbxfile"
import "strings"

type RawPacketHandlerMap struct {
	m map[uint8][]ReceiveHandler
}

func (m *RawPacketHandlerMap) Bind(packetType uint8, handler ReceiveHandler) int {
	m.m[packetType] = append(m.m[packetType], handler)
	return len(m.m[packetType]) - 1
}
func (m *RawPacketHandlerMap) Unbind(packetType uint8, index int) {
	m.m[packetType] = append(m.m[packetType][:index], m.m[packetType][index+1:]...)
}
func (m *RawPacketHandlerMap) Fire(packetType uint8, layers *PacketLayers) {
	for _, handler := range m.m[packetType] {
		handler(packetType, layers)
	}
}
func NewRawPacketHandlerMap() *RawPacketHandlerMap {
	return &RawPacketHandlerMap{map[uint8][]ReceiveHandler{}}
}

type DeleteInstanceHandler func(*rbxfile.Instance)
type DeleteInstanceHandlerMap struct {
	m map[Referent][]DeleteInstanceHandler
}

func (m *DeleteInstanceHandlerMap) Bind(packetType Referent, handler DeleteInstanceHandler) int {
	m.m[packetType] = append(m.m[packetType], handler)
	return len(m.m[packetType]) - 1
}
func (m *DeleteInstanceHandlerMap) Unbind(packetType Referent, index int) {
	m.m[packetType] = append(m.m[packetType][:index], m.m[packetType][index+1:]...)
}
func (m *DeleteInstanceHandlerMap) Fire(packetType *rbxfile.Instance) {
	for _, handler := range m.m[Referent(packetType.Reference)] {
		handler(packetType)
	}
}
func NewDeleteInstanceHandlerMap() *DeleteInstanceHandlerMap {
	return &DeleteInstanceHandlerMap{map[Referent][]DeleteInstanceHandler{}}
}

type DataPacketHandlerMap struct {
	m map[uint8][]DataReceiveHandler
}

type DataReceiveHandler func(uint8, *PacketLayers, Packet83Subpacket)

func (m *DataPacketHandlerMap) Bind(packetType uint8, handler DataReceiveHandler) int {
	m.m[packetType] = append(m.m[packetType], handler)
	return len(m.m[packetType]) - 1
}
func (m *DataPacketHandlerMap) Unbind(packetType uint8, index int) {
	m.m[packetType] = append(m.m[packetType][:index], m.m[packetType][index+1:]...)
}
func (m *DataPacketHandlerMap) Fire(packetType uint8, layers *PacketLayers, subpacket Packet83Subpacket) {
	for _, handler := range m.m[packetType] {
		handler(packetType, layers, subpacket)
	}
}
func NewDataHandlerMap() *DataPacketHandlerMap {
	return &DataPacketHandlerMap{map[uint8][]DataReceiveHandler{}}
}

type NewInstanceHandler func(*rbxfile.Instance)
type newInstanceHandlerMapKey struct {
	key     *InstancePath
	handler NewInstanceHandler
}

type NewInstanceHandlerMap struct {
	handlers []newInstanceHandlerMapKey
}

func (m *NewInstanceHandlerMap) Bind(packetType *InstancePath, handler NewInstanceHandler) int {
	m.handlers = append(m.handlers, newInstanceHandlerMapKey{
		key:     packetType,
		handler: handler,
	})

	addedIndex := len(m.handlers) - 1

	return addedIndex
}
func (m *NewInstanceHandlerMap) Unbind(index int) {
	m.handlers = append(m.handlers[index:], m.handlers[:index+1]...)
}
func (m *NewInstanceHandlerMap) Fire(instance *rbxfile.Instance) {
	//println("fire", NewInstancePath(instance).String())
	for _, handler := range m.handlers {
		if handler.key.Matches(instance) {
			handler.handler(instance)
		}
	}
}
func NewNewInstanceHandlerMap() *NewInstanceHandlerMap {
	return &NewInstanceHandlerMap{}
}

// InstancePath describes a single instance's path
type InstancePath struct {
	p []string
}

// Creates an instance path from an instance
func NewInstancePath(instance *rbxfile.Instance) *InstancePath {
	path := []string{instance.Name()}
	for instance.Parent() != nil && !instance.IsService { // HACK: see InstancePath::Matches()
		instance = instance.Parent()
		if !instance.IsService {
			path = append([]string{instance.Name()}, path...)
		} else {
			path = append([]string{instance.ClassName}, path...)
		}
	}

	return &InstancePath{path}
}

// Returns a string representation where the names are joined with dots
func (path *InstancePath) String() string {
	return strings.Join(path.p, ".")
}

// Checks if the path is the same as an Instance's
func (path *InstancePath) Matches(instance *rbxfile.Instance) bool {
	index := len(path.p) - 1
	if instance.Name() != path.p[index] && path.p[index] != "*" {
		return false
	}

	// HACK: when the instance is a service, we pretend it has no parent
	// this is because the DataModel is never properly replicated/created, it is only
	// used as a parent for services in JoinData packets
	for instance.Parent() != nil && !instance.IsService {
		instance = instance.Parent()
		index -= 1
		if index < 0 {
			return false
		}
		if !instance.IsService {
			if instance.Name() != path.p[index] && path.p[index] != "*" { // wildcard handling
				return false
			}
		} else {
			if instance.ClassName != path.p[index] && path.p[index] != "*" {
				// HACK: use classname detection for services
				// this is to circumvent the anti-hack system found in Jailbreak
				return false
			}
		}
	}
	if index != 0 {
		return false
	}
	return true
}

// Handler that receives events
type EventHandler func(event *ReplicationEvent)
type eventHandlerMapKey struct {
	instKey  *rbxfile.Instance
	eventKey string
	handler  EventHandler
}

type EventHandlerMap struct {
	handlers []eventHandlerMapKey
}

func (m *EventHandlerMap) Bind(packetType *rbxfile.Instance, event string, handler EventHandler) int {
	m.handlers = append(m.handlers, eventHandlerMapKey{
		instKey:  packetType,
		eventKey: event,
		handler:  handler,
	})

	addedIndex := len(m.handlers) - 1

	return addedIndex
}
func (m *EventHandlerMap) Unbind(index int) {
	m.handlers = append(m.handlers[index:], m.handlers[:index+1]...)
}
func (m *EventHandlerMap) Fire(instance *rbxfile.Instance, name string, event *ReplicationEvent) {
	for _, handler := range m.handlers {
		if name == "RemoteOnInvokeSuccess" || name == "RemoteOnInvokeError" {
			println("receiving important remote!", name, NewInstancePath(instance).String())
		}
		if handler.eventKey == name && handler.instKey == instance {
			handler.handler(event)
		}
	}
}
func NewEventHandlerMap() *EventHandlerMap {
	return &EventHandlerMap{}
}
