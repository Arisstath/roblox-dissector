package peer

import (
	"container/list"
	"sync"
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
