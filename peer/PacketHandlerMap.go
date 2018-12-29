package peer

import (
	"container/list"
	"strings"
	"sync"

	"github.com/robloxapi/rbxfile"
)

type PacketHandlerConnection struct {
	Callback interface{}
	Data     interface{}
	Once     bool
	element  *list.Element
	parent   interface {
		Disconnect(*PacketHandlerConnection)
	}
}

func (conn *PacketHandlerConnection) Disconnect() {
	if conn == nil {
		println("WARNING: Ignoring nil disconnection")
		return
	}
	conn.parent.Disconnect(conn)
}

type PacketHandlerMap struct {
	*sync.Mutex
	m *list.List
}

func (m *PacketHandlerMap) Disconnect(conn *PacketHandlerConnection) {
	m.Lock()
	m.m.Remove(conn.element)
	m.Unlock()
}
func NewPacketHandlerMap() *PacketHandlerMap {
	return &PacketHandlerMap{m: list.New(), Mutex: &sync.Mutex{}}
}

type RawPacketHandlerMap struct {
	*PacketHandlerMap
}

func (m *RawPacketHandlerMap) Bind(packetType uint8, handler ReceiveHandler) *PacketHandlerConnection {
	m.Lock()
	conn := &PacketHandlerConnection{Callback: handler, parent: m, Data: packetType}
	conn.element = m.m.PushBack(conn)
	m.Unlock()
	return conn
}
func (m *RawPacketHandlerMap) Fire(packetType uint8, layers *PacketLayers) {
	m.Lock()
	for e := m.m.Front(); e != nil; e = e.Next() {
		conn := e.Value.(*PacketHandlerConnection)
		if conn.Data.(uint8) == packetType {
			go conn.Callback.(ReceiveHandler)(packetType, layers)
		}
	}
	m.Unlock()
}
func NewRawPacketHandlerMap() *RawPacketHandlerMap {
	return &RawPacketHandlerMap{NewPacketHandlerMap()}
}

type DeleteInstanceHandler func(*rbxfile.Instance)
type DeleteInstanceHandlerMap struct {
	*PacketHandlerMap
}

func (m *DeleteInstanceHandlerMap) Bind(packetType Referent, handler DeleteInstanceHandler) *PacketHandlerConnection {
	m.Lock()
	conn := &PacketHandlerConnection{parent: m, Callback: handler, Data: packetType}
	conn.element = m.m.PushBack(conn)
	m.Unlock()
	return conn
}
func (m *DeleteInstanceHandlerMap) Fire(packetType *rbxfile.Instance) {
	m.Lock()
	for e := m.m.Front(); e != nil; e = e.Next() {
		conn := e.Value.(*PacketHandlerConnection)
		if packetType.Reference == string(conn.Data.(Referent)) {
			go conn.Callback.(DeleteInstanceHandler)(packetType)
		}
	}
	m.Unlock()
}
func NewDeleteInstanceHandlerMap() *DeleteInstanceHandlerMap {
	return &DeleteInstanceHandlerMap{NewPacketHandlerMap()}
}

type DataPacketHandlerMap struct {
	*PacketHandlerMap
}
type DataReceiveHandler func(uint8, *PacketLayers, Packet83Subpacket)

func (m *DataPacketHandlerMap) Bind(packetType uint8, handler DataReceiveHandler) *PacketHandlerConnection {
	m.Lock()
	conn := &PacketHandlerConnection{parent: m, Callback: handler, Data: packetType}
	conn.element = m.m.PushBack(conn)
	m.Unlock()
	return conn
}
func (m *DataPacketHandlerMap) Fire(packetType uint8, layers *PacketLayers, subpacket Packet83Subpacket) {
	m.Lock()
	for e := m.m.Front(); e != nil; e = e.Next() {
		conn := e.Value.(*PacketHandlerConnection)
		if conn.Data.(uint8) == packetType {
			go conn.Callback.(DataReceiveHandler)(packetType, layers, subpacket)
		}
	}
	m.Unlock()
}
func NewDataHandlerMap() *DataPacketHandlerMap {
	return &DataPacketHandlerMap{NewPacketHandlerMap()}
}

type NewInstanceHandler func(*rbxfile.Instance)
type NewInstanceHandlerMap struct {
	*PacketHandlerMap
}

// Bind assumes that you hold the lock yourself
func (m *NewInstanceHandlerMap) Bind(packetType *InstancePath, handler NewInstanceHandler) *PacketHandlerConnection {
	conn := &PacketHandlerConnection{
		parent:   m,
		Callback: handler,
		Data:     packetType,
	}
	conn.element = m.m.PushBack(conn)

	return conn
}
func (m *NewInstanceHandlerMap) Fire(instance *rbxfile.Instance) {
	m.Lock()
	unbindMe := make([]*PacketHandlerConnection, 0)
	for e := m.m.Front(); e != nil; e = e.Next() {
		conn := e.Value.(*PacketHandlerConnection)
		if conn.Data.(*InstancePath).Matches(instance) {
			go conn.Callback.(NewInstanceHandler)(instance)
			if conn.Once {
				unbindMe = append(unbindMe, conn)
			}
		}
	}
	// Must do this after the loop
	for _, val := range unbindMe {
		m.m.Remove(val.element)
	}
	m.Unlock()
}
func NewNewInstanceHandlerMap() *NewInstanceHandlerMap {
	return &NewInstanceHandlerMap{NewPacketHandlerMap()}
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
	if len(path.p) == 0 {
		return path.root == instance
	}
	index := len(path.p) - 1
	if instance.Name() != path.p[index] && path.p[index] != "*" {
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
				return true
			}
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
	if index != 0 || (path.root != nil && path.root != instance) {
		return false
	}
	return true
}

type PropertyHandler func(value rbxfile.Value)
type propertyHandlerData struct {
	instKey *rbxfile.Instance
	propKey string
}

type PropertyHandlerMap struct {
	*PacketHandlerMap
}

func (m *PropertyHandlerMap) Bind(instance *rbxfile.Instance, property string, handler PropertyHandler) *PacketHandlerConnection {
	m.Lock()
	conn := &PacketHandlerConnection{
		parent:   m,
		Callback: handler,
		Data:     propertyHandlerData{instance, property},
	}
	conn.element = m.m.PushBack(conn)

	m.Unlock()
	return conn
}
func (m *PropertyHandlerMap) Fire(instance *rbxfile.Instance, name string, value rbxfile.Value) {
	m.Lock()
	for e := m.m.Front(); e != nil; e = e.Next() {
		conn := e.Value.(*PacketHandlerConnection)
		data := conn.Data.(propertyHandlerData)
		if data.instKey == instance && data.propKey == name {
			go conn.Callback.(PropertyHandler)(value)
		}
	}
	m.Unlock()
}
func NewPropertyHandlerMap() *PropertyHandlerMap {
	return &PropertyHandlerMap{NewPacketHandlerMap()}
}

// Handler that receives events
type EventHandler func(event *ReplicationEvent)
type eventHandlerData struct {
	instKey  *rbxfile.Instance
	eventKey string
}

type EventHandlerMap struct {
	*PacketHandlerMap
}

func (m *EventHandlerMap) Bind(packetType *rbxfile.Instance, event string, handler EventHandler) *PacketHandlerConnection {
	m.Lock()
	conn := &PacketHandlerConnection{
		parent:   m,
		Callback: handler,
		Data: eventHandlerData{
			instKey:  packetType,
			eventKey: event,
		},
	}
	conn.element = m.m.PushBack(conn)
	m.Unlock()
	return conn
}
func (m *EventHandlerMap) Fire(instance *rbxfile.Instance, name string, event *ReplicationEvent) {
	m.Lock()
	for e := m.m.Front(); e != nil; e = e.Next() {
		conn := e.Value.(*PacketHandlerConnection)
		data := conn.Data.(eventHandlerData)
		if data.eventKey == name && data.instKey == instance {
			go conn.Callback.(EventHandler)(event)
		}
	}
	m.Unlock()
}
func NewEventHandlerMap() *EventHandlerMap {
	return &EventHandlerMap{NewPacketHandlerMap()}
}
