package peer

import (
	"fmt"
	"strings"
	"sync"

	"github.com/gskartwii/rbxfile"
)

type RawPacketHandlerMap struct {
	*sync.Mutex
	m        map[uint8](map[int]ReceiveHandler)
	uniqueId int
}

func (m *RawPacketHandlerMap) Bind(packetType uint8, handler ReceiveHandler) int {
	m.Lock()
	if m.m[packetType] == nil {
		m.m[packetType] = make(map[int]ReceiveHandler)
	}
	m.m[packetType][m.uniqueId] = handler
	m.uniqueId++
	m.Unlock()
	return m.uniqueId - 1
}
func (m *RawPacketHandlerMap) Unbind(packetType uint8, index int) {
	m.Lock()
	delete(m.m[packetType], index)
	m.Unlock()
}
func (m *RawPacketHandlerMap) Fire(packetType uint8, layers *PacketLayers) {
	m.Lock()
	for _, handler := range m.m[packetType] {
		handler(packetType, layers)
	}
	m.Unlock()
}
func NewRawPacketHandlerMap() *RawPacketHandlerMap {
	return &RawPacketHandlerMap{m: map[uint8](map[int]ReceiveHandler){}, Mutex: &sync.Mutex{}}
}

type DeleteInstanceHandler func(*rbxfile.Instance)
type DeleteInstanceHandlerMap struct {
	*sync.Mutex
	m        map[Referent](map[int]DeleteInstanceHandler)
	uniqueId int
}

func (m *DeleteInstanceHandlerMap) Bind(packetType Referent, handler DeleteInstanceHandler) int {
	m.Lock()
	if m.m[packetType] == nil {
		m.m[packetType] = make(map[int]DeleteInstanceHandler)
	}
	m.m[packetType][m.uniqueId] = handler
	m.uniqueId++
	m.Unlock()
	return m.uniqueId - 1
}
func (m *DeleteInstanceHandlerMap) Unbind(packetType Referent, index int) {
	m.Lock()
	delete(m.m[packetType], index)
	m.Unlock()
}
func (m *DeleteInstanceHandlerMap) Fire(packetType *rbxfile.Instance) {
	m.Lock()
	for _, handler := range m.m[Referent(packetType.Reference)] {
		handler(packetType)
	}
	m.Unlock()
}
func NewDeleteInstanceHandlerMap() *DeleteInstanceHandlerMap {
	return &DeleteInstanceHandlerMap{m: map[Referent](map[int]DeleteInstanceHandler){}, Mutex: &sync.Mutex{}}
}

type DataPacketHandlerMap struct {
	*sync.Mutex
	m        map[uint8]map[int]DataReceiveHandler
	uniqueId int
}

type DataReceiveHandler func(uint8, *PacketLayers, Packet83Subpacket)

func (m *DataPacketHandlerMap) Bind(packetType uint8, handler DataReceiveHandler) int {
	m.Lock()
	if m.m[packetType] == nil {
		m.m[packetType] = make(map[int]DataReceiveHandler)
	}
	m.m[packetType][m.uniqueId] = handler
	m.uniqueId++
	m.Unlock()
	return m.uniqueId - 1
}
func (m *DataPacketHandlerMap) Unbind(packetType uint8, index int) {
	m.Lock()
	delete(m.m[packetType], index)
	m.Unlock()
}
func (m *DataPacketHandlerMap) Fire(packetType uint8, layers *PacketLayers, subpacket Packet83Subpacket) {
	m.Lock()
	for _, handler := range m.m[packetType] {
		handler(packetType, layers, subpacket)
	}
	m.Unlock()
}
func NewDataHandlerMap() *DataPacketHandlerMap {
	return &DataPacketHandlerMap{m: map[uint8](map[int]DataReceiveHandler){}, Mutex: &sync.Mutex{}}
}

type NewInstanceHandler func(*rbxfile.Instance) bool
type newInstanceHandlerMapKey struct {
	key     *InstancePath
	handler NewInstanceHandler
}

type NewInstanceHandlerMap struct {
	*sync.Mutex
	handlers map[int]newInstanceHandlerMapKey
	uniqueId int
}

func (m *NewInstanceHandlerMap) Bind(packetType *InstancePath, handler NewInstanceHandler) int {
	m.Lock()
	m.handlers[m.uniqueId] = newInstanceHandlerMapKey{
		key:     packetType,
		handler: handler,
	}

	m.uniqueId++
	m.Unlock()
	return m.uniqueId - 1
}
func (m *NewInstanceHandlerMap) Unbind(index int) {
	m.Lock()
	delete(m.handlers, index)
	m.Unlock()
}
func (m *NewInstanceHandlerMap) Fire(instance *rbxfile.Instance) {
	//println("fire", NewInstancePath(instance).String())
	m.Lock()
	unbindUs := []int{}
	for index, handler := range m.handlers {
		if handler.key.Matches(instance) {
			if handler.handler(instance) {
				unbindUs = append(unbindUs, index)
			}
		}
	}
	for _, index := range unbindUs {
		println("unbinding key", m.handlers[index].key.String())
		delete(m.handlers, index)
	}
	m.Unlock()
}
func NewNewInstanceHandlerMap() *NewInstanceHandlerMap {
	return &NewInstanceHandlerMap{handlers: map[int]newInstanceHandlerMapKey{}, Mutex: &sync.Mutex{}}
}

// InstancePath describes a single instance's path
type InstancePath struct {
	p    []string
	root *rbxfile.Instance
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

	return &InstancePath{path, nil}
}

// Returns a string representation where the names are joined with dots
func (path *InstancePath) String() string {
	return strings.Join(path.p, ".")
}

// Checks if the path is the same as an Instance's
func (path *InstancePath) Matches(instance *rbxfile.Instance) bool {
	fmt.Printf("matches called %s %s %v\n", path.root.GetFullName(), instance.GetFullName(), path.p)
	if len(path.p) == 0 {
		fmt.Printf("zero length %s %s %v\n", path.root.GetFullName(), instance.GetFullName(), path.p)
		return path.root == instance
	}
	index := len(path.p) - 1
	if instance.Name() != path.p[index] && path.p[index] != "*" {
		fmt.Printf("first failed %s %s %v\n", path.root.GetFullName(), instance.GetFullName(), path.p)
		return false
	}

	// HACK: when the instance is a service, we pretend it has no parent
	// this is because the DataModel is never properly replicated/created, it is only
	// used as a parent for services in JoinData packets
	for instance.Parent() != nil && !instance.IsService {
		instance = instance.Parent()
		index--
		if index < 0 {
			if instance == path.root { // path index -1 is oob for the path, so we look at root which is parent to the path
				fmt.Printf("@0, instance equal %s %s %v\n", path.root.GetFullName(), instance.GetFullName(), path.p)
				return true
			}
			fmt.Printf("below zero %s %s %v\n", path.root.GetFullName(), instance.GetFullName(), path.p)
			return false
		}
		if !instance.IsService {
			if instance.Name() != path.p[index] && path.p[index] != "*" { // wildcard handling
				fmt.Printf("not service, didn't match %s %s %v\n", path.root.GetFullName(), instance.GetFullName(), path.p)
				return false
			}
		} else {
			if instance.ClassName != path.p[index] && path.p[index] != "*" {
				// HACK: use classname detection for services
				// this is to circumvent the anti-hack system found in Jailbreak
				fmt.Printf("service didn't match %s %s %v\n", path.root.GetFullName(), instance.GetFullName(), path.p)
				return false
			}
		}
	}
	if index != 0 || (path.root != nil && path.root != instance) {
		fmt.Printf("last thing didn't match %s %s %v\n", path.root.GetFullName(), instance.GetFullName(), path.p)
		return false
	}
	fmt.Printf("resolved %s %s %v\n", path.root.GetFullName(), instance.GetFullName(), path.p)
	return true
}

type PropertyHandler func(value rbxfile.Value) bool
type propertyHandlerMapKey struct {
	instKey *rbxfile.Instance
	propKey string
	handler PropertyHandler
}

type PropertyHandlerMap struct {
	*sync.Mutex
	handlers map[int]propertyHandlerMapKey
	uniqueId int
}

func (m *PropertyHandlerMap) Bind(instance *rbxfile.Instance, property string, handler PropertyHandler) int {
	m.Lock()
	m.handlers[m.uniqueId] = propertyHandlerMapKey{
		instKey: instance,
		propKey: property,
		handler: handler,
	}

	m.uniqueId++
	m.Unlock()
	return m.uniqueId
}
func (m *PropertyHandlerMap) Unbind(index int) {
	m.Lock()
	delete(m.handlers, index)
	m.Unlock()
}
func (m *PropertyHandlerMap) Fire(instance *rbxfile.Instance, name string, value rbxfile.Value) {
	m.Lock()
	unbindUs := []int{}
	for index, handler := range m.handlers {
		if handler.instKey == instance && handler.propKey == name {
			if handler.handler(value) {
				unbindUs = append(unbindUs, index)
			}
		}
	}
	for _, index := range unbindUs {
		println("unbinding key", m.handlers[index].propKey)
		delete(m.handlers, index)
	}
	m.Unlock()
}
func NewPropertyHandlerMap() *PropertyHandlerMap {
	return &PropertyHandlerMap{handlers: map[int]propertyHandlerMapKey{}, Mutex: &sync.Mutex{}}
}

// Handler that receives events
type EventHandler func(event *ReplicationEvent)
type eventHandlerMapKey struct {
	instKey  *rbxfile.Instance
	eventKey string
	handler  EventHandler
}

type EventHandlerMap struct {
	*sync.Mutex
	handlers map[int]eventHandlerMapKey
	uniqueId int
}

func (m *EventHandlerMap) Bind(packetType *rbxfile.Instance, event string, handler EventHandler) int {
	m.Lock()
	m.handlers[m.uniqueId] = eventHandlerMapKey{
		instKey:  packetType,
		eventKey: event,
		handler:  handler,
	}

	m.uniqueId++
	m.Unlock()
	return m.uniqueId - 1
}
func (m *EventHandlerMap) Unbind(index int) {
	m.Lock()
	delete(m.handlers, index)
	m.Unlock()
}
func (m *EventHandlerMap) Fire(instance *rbxfile.Instance, name string, event *ReplicationEvent) {
	m.Lock()
	for _, handler := range m.handlers {
		if name == "RemoteOnInvokeSuccess" || name == "RemoteOnInvokeError" {
			println("receiving important remote!", name, NewInstancePath(instance).String())
		}
		if handler.eventKey == name && handler.instKey == instance {
			handler.handler(event)
		}
	}
	m.Unlock()
}
func NewEventHandlerMap() *EventHandlerMap {
	return &EventHandlerMap{handlers: map[int]eventHandlerMapKey{}, Mutex: &sync.Mutex{}}
}
