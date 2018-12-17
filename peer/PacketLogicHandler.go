package peer

import (
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/gskartwii/rbxfile"
)

type PacketLogicHandler struct {
	*ConnectedPeer
	Context          *CommunicationContext
	handlers         *RawPacketHandlerMap
	dataHandlers     *DataPacketHandlerMap
	instanceHandlers *NewInstanceHandlerMap
	deleteHandlers   *DeleteInstanceHandlerMap
	eventHandlers    *EventHandlerMap
	propHandlers     *PropertyHandlerMap

	ackTicker      *time.Ticker
	dataPingTicker *time.Ticker

	remoteIndices map[*rbxfile.Instance]uint32
	remoteLock    *sync.Mutex

	Connection   *net.UDPConn
	pingInterval int
}

func newPacketLogicHandler(context *CommunicationContext) PacketLogicHandler {
	return PacketLogicHandler{
		ConnectedPeer:    NewConnectedPeer(context),
		handlers:         NewRawPacketHandlerMap(),
		dataHandlers:     NewDataHandlerMap(),
		instanceHandlers: NewNewInstanceHandlerMap(),
		deleteHandlers:   NewDeleteInstanceHandlerMap(),
		eventHandlers:    NewEventHandlerMap(),
		propHandlers:     NewPropertyHandlerMap(),

		remoteIndices: make(map[*rbxfile.Instance]uint32),
		remoteLock:    &sync.Mutex{},

		Context: context,
	}
}

func (logicHandler *PacketLogicHandler) RegisterPacketHandler(packetType uint8, handler ReceiveHandler) {
	logicHandler.handlers.Bind(packetType, handler)
}
func (logicHandler *PacketLogicHandler) RegisterDataHandler(packetType uint8, handler DataReceiveHandler) {
	logicHandler.dataHandlers.Bind(packetType, handler)
}
func (logicHandler *PacketLogicHandler) RegisterInstanceHandler(path *InstancePath, handler NewInstanceHandler) *PacketHandlerConnection {
	logicHandler.instanceHandlers.Lock()
	conn := logicHandler.instanceHandlers.Bind(path, handler)
	logicHandler.instanceHandlers.Unlock()
	return conn
}

func (logicHandler *PacketLogicHandler) defaultAckHandler(layers *PacketLayers) {
	// nop
	if layers.Error != nil {
		println("ack error: ", layers.Error.Error())
	}
}
func (logicHandler *PacketLogicHandler) defaultReliabilityLayerHandler(layers *PacketLayers) {
	logicHandler.mustACK = append(logicHandler.mustACK, int(layers.RakNet.DatagramNumber))
	if layers.Error != nil {
		println("reliabilitylayer error: ", layers.Error.Error())
	}
}
func (logicHandler *PacketLogicHandler) defaultSimpleHandler(packetType byte, layers *PacketLayers) {
	if layers.Error == nil {
		go logicHandler.handlers.Fire(packetType, layers) // Let the reader continue its job while the packet is processed
	} else {
		println("simple error: ", layers.Error.Error())
	}
}
func (logicHandler *PacketLogicHandler) defaultReliableHandler(packetType byte, layers *PacketLayers) {
	// nop
	if layers.Error != nil {
		println("reliable error: ", layers.Error.Error())
	}
}
func (logicHandler *PacketLogicHandler) defaultFullReliableHandler(packetType byte, layers *PacketLayers) {
	if layers.Error == nil {
		go logicHandler.handlers.Fire(packetType, layers)
	} else {
		println("simple error: ", layers.Error.Error())
	}
}

func (logicHandler *PacketLogicHandler) createReader() {
	logicHandler.ACKHandler = logicHandler.defaultAckHandler
	logicHandler.ReliabilityLayerHandler = logicHandler.defaultReliabilityLayerHandler
	logicHandler.SimpleHandler = logicHandler.defaultSimpleHandler
	logicHandler.ReliableHandler = logicHandler.defaultReliableHandler
	logicHandler.FullReliableHandler = logicHandler.defaultFullReliableHandler

	logicHandler.DefaultPacketReader.SetContext(logicHandler.Context)
}

func (logicHandler *PacketLogicHandler) startDataPing() {
	// boot up dataping
	logicHandler.dataPingTicker = time.NewTicker(time.Duration(logicHandler.pingInterval) * time.Millisecond)
	go func() {
		for {
			<-logicHandler.dataPingTicker.C

			logicHandler.WritePacket(&Packet83Layer{
				[]Packet83Subpacket{&Packet83_05{
					Timestamp:  uint64(time.Now().UnixNano() / int64(time.Millisecond)),
					IsPingBack: false,
				}},
			})
		}
	}()
}

func (logicHandler *PacketLogicHandler) startAcker() {
	logicHandler.ackTicker = time.NewTicker(500 * time.Millisecond)
	go func() {
		for {
			<-logicHandler.ackTicker.C
			logicHandler.sendACKs()
		}
	}()
}

func (logicHandler *PacketLogicHandler) disconnectInternal() error {
	if logicHandler.ackTicker != nil {
		logicHandler.ackTicker.Stop()
	}
	if logicHandler.dataPingTicker != nil {
		logicHandler.dataPingTicker.Stop()
	}
	return logicHandler.Connection.Close()
}

func (logicHandler *PacketLogicHandler) Disconnect() {
	logicHandler.WritePacket(&Packet15Layer{
		Reason: 0xFFFFFFFF,
	})

	logicHandler.disconnectInternal()
}

func (logicHandler *PacketLogicHandler) deleteHandler(packetType uint8, layers *PacketLayers, subpacket Packet83Subpacket) {
	mainpacket := subpacket.(*Packet83_01)

	logicHandler.deleteHandlers.Fire(mainpacket.Instance)
}
func (logicHandler *PacketLogicHandler) newInstanceHandler(packetType uint8, layers *PacketLayers, subpacket Packet83Subpacket) {
	mainpacket := subpacket.(*Packet83_02)

	logicHandler.instanceHandlers.Fire(mainpacket.Child)
}
func (logicHandler *PacketLogicHandler) joinDataHandler(packetType uint8, layers *PacketLayers, subpacket Packet83Subpacket) {
	mainpacket := subpacket.(*Packet83_0B)

	for _, inst := range mainpacket.Instances {
		logicHandler.instanceHandlers.Fire(inst)
	}
}
func (logicHandler *PacketLogicHandler) propHandler(packetType uint8, layers *PacketLayers, item Packet83Subpacket) {
	mainPacket := item.(*Packet83_03)

	logicHandler.propHandlers.Fire(mainPacket.Instance, mainPacket.PropertyName, mainPacket.Value)
}
func (logicHandler *PacketLogicHandler) eventHandler(packetType uint8, layers *PacketLayers, item Packet83Subpacket) {
	mainPacket := item.(*Packet83_07)

	logicHandler.eventHandlers.Fire(mainPacket.Instance, mainPacket.EventName, mainPacket.Event)
}

func (logicHandler *PacketLogicHandler) sendDataPingBack() {
	response := &Packet83_06{
		Timestamp:  uint64(time.Now().UnixNano() / int64(time.Millisecond)),
		IsPingBack: true,
	}

	err := logicHandler.WriteDataPackets(response)
	if err != nil {
		println("Failed to send datapingback:", err.Error())
	}
}
func (logicHandler *PacketLogicHandler) dataPingHandler(packetType uint8, layers *PacketLayers, item Packet83Subpacket) {
	logicHandler.sendDataPingBack()
}

func (logicHandler *PacketLogicHandler) dataHandler(packetType uint8, layers *PacketLayers) {
	mainLayer := layers.Main.(*Packet83Layer)
	for _, item := range mainLayer.SubPackets {
		logicHandler.dataHandlers.Fire(item.Type(), layers, item)
	}
}

func (logicHandler *PacketLogicHandler) disconnectHandler(packetType uint8, layers *PacketLayers) {
	mainLayer := layers.Main.(*Packet15Layer)
	fmt.Printf("Received disconnect with reason %d\n", mainLayer.Reason)

	logicHandler.disconnectInternal()
}

func (logicHandler *PacketLogicHandler) sendPong(pingTime uint64) {
	response := &Packet03Layer{
		SendPingTime: pingTime,
		SendPongTime: uint64(time.Now().UnixNano() / int64(time.Millisecond)),
	}

	_, err := logicHandler.WritePacket(response)
	if err != nil {
		println("Failed to write pong: ", err.Error())
	}
}
func (logicHandler *PacketLogicHandler) pingHandler(packetType byte, layers *PacketLayers) {
	mainLayer := layers.Main.(*Packet00Layer)

	logicHandler.sendPong(mainLayer.SendPingTime)
}

func (logicHandler *PacketLogicHandler) bindDefaultHandlers() {
	// common to all peers
	dataHandlers := logicHandler.dataHandlers
	dataHandlers.Bind(1, logicHandler.deleteHandler)
	dataHandlers.Bind(2, logicHandler.newInstanceHandler)
	dataHandlers.Bind(3, logicHandler.propHandler)
	dataHandlers.Bind(5, logicHandler.dataPingHandler)
	dataHandlers.Bind(7, logicHandler.eventHandler)
	dataHandlers.Bind(0xB, logicHandler.joinDataHandler)

	basicHandlers := logicHandler.handlers
	basicHandlers.Bind(0x15, logicHandler.disconnectHandler)
	basicHandlers.Bind(0x83, logicHandler.dataHandler)
}

func (logicHandler *PacketLogicHandler) WriteDataPackets(packets ...Packet83Subpacket) error {
	_, err := logicHandler.WritePacket(&Packet83Layer{
		SubPackets: packets,
	})
	return err
}

func (logicHandler *PacketLogicHandler) FindService(name string) *rbxfile.Instance {
	if logicHandler.Context == nil || logicHandler.Context.DataModel == nil {
		return nil
	}
	for _, service := range logicHandler.Context.DataModel.Instances {
		if service.ClassName == name {
			return service
		}
	}
	return nil
}

func (logicHandler *PacketLogicHandler) WaitForChild(instance *rbxfile.Instance, path ...string) <-chan *rbxfile.Instance {
	logicHandler.instanceHandlers.Lock()
	retChannel := make(chan *rbxfile.Instance, 1)
	currInstance := instance
	lastInstance := instance
	currPath := path
	if currInstance != nil {
		for i := 0; i < len(path); i++ {
			currInstance = currInstance.FindFirstChild(path[i], false)
			if currInstance == nil {
				currPath = path[i:]
				break
			}
			lastInstance = currInstance
		}
	}
	if currInstance == nil {
		fmt.Printf("Must create instancepath %s %v\n", lastInstance.GetFullName(), currPath)
		logicHandler.instanceHandlers.Bind(&InstancePath{currPath, lastInstance}, func(instance *rbxfile.Instance) {
			fmt.Printf("Received %s %v %s\n", lastInstance.GetFullName(), currPath, instance.GetFullName())
			retChannel <- instance
		}).Once = true
	} else {
		retChannel <- currInstance
	}
	logicHandler.instanceHandlers.Unlock()
	return retChannel
}

func (logicHandler *PacketLogicHandler) WaitForRefProp(instance *rbxfile.Instance, name string) <-chan *rbxfile.Instance {
	retChannel := make(chan *rbxfile.Instance, 1)
	instance.PropertiesMutex.RLock()
	if instance.Properties[name] != nil && instance.Properties[name].(rbxfile.ValueReference).Instance != nil {
		retChannel <- instance.Properties[name].(rbxfile.ValueReference).Instance
	} else {
		var connection *PacketHandlerConnection
		connection = logicHandler.propHandlers.Bind(instance, name, func(value rbxfile.Value) {
			if value.(rbxfile.ValueReference).Instance != nil {
				retChannel <- value.(rbxfile.ValueReference).Instance
				connection.Disconnect()
			}
		})
	}
	instance.PropertiesMutex.RUnlock()
	return retChannel
}

func (logicHandler *PacketLogicHandler) WaitForInstance(path ...string) <-chan *rbxfile.Instance { // returned channels are output only
	service := logicHandler.FindService(path[0])
	if service == nil {
		return logicHandler.WaitForChild(nil, path...)
	}
	return logicHandler.WaitForChild(service, path[1:]...)
}

func (logicHandler *PacketLogicHandler) MakeEventChan(instance *rbxfile.Instance, name string) (*PacketHandlerConnection, chan *ReplicationEvent) {
	newChan := make(chan *ReplicationEvent)
	connection := logicHandler.eventHandlers.Bind(instance, name, func(evt *ReplicationEvent) {
		newChan <- evt
	})
	return connection, newChan
}

func (logicHandler *PacketLogicHandler) SendEvent(instance *rbxfile.Instance, name string, arguments ...rbxfile.Value) error {
	return logicHandler.WriteDataPackets(
		&Packet83_07{
			Instance:  instance,
			EventName: name,
			Event:     &ReplicationEvent{arguments},
		},
	)
}

func (logicHandler *PacketLogicHandler) MakeChildChan(instance *rbxfile.Instance) chan *rbxfile.Instance {
	newChan := make(chan *rbxfile.Instance)

	go func() { // we don't want this to block
		for _, child := range instance.Children {
			newChan <- child
		}
	}()

	path := NewInstancePath(instance)
	path.p = append(path.p, "*") // wildcard

	logicHandler.instanceHandlers.Bind(path, func(inst *rbxfile.Instance) {
		newChan <- inst
	})

	return newChan
}

type GroupDeleteChan struct {
	C         chan *rbxfile.Instance
	binding   *PacketHandlerConnection
	referents []Referent
}

func (channel *GroupDeleteChan) AddInstances(instances ...*rbxfile.Instance) {
	for _, inst := range instances {
		channel.referents = append(channel.referents, Referent(inst.Reference))
	}
}
func (channel *GroupDeleteChan) Destroy() {
	channel.binding.Disconnect()
}
func (logicHandler *PacketLogicHandler) MakeGroupDeleteChan(instances []*rbxfile.Instance) *GroupDeleteChan {
	channel := &GroupDeleteChan{
		C:         make(chan *rbxfile.Instance),
		referents: make([]Referent, len(instances)),
	}

	for i, inst := range instances {
		channel.referents[i] = Referent(inst.Reference)
	}

	channel.binding = logicHandler.dataHandlers.Bind(1, func(packetType uint8, layers *PacketLayers, subpacket Packet83Subpacket) {
		mainpacket := subpacket.(*Packet83_01)
		for _, inst := range channel.referents {
			if string(inst) == mainpacket.Instance.Reference {
				channel.C <- mainpacket.Instance
				break
			}
		}
	})

	return channel
}
