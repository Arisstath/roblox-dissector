package peer

import (
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/gskartwii/roblox-dissector/datamodel"
	"github.com/robloxapi/rbxfile"
)

type PacketLogicHandler struct {
	*ConnectedPeer
	Context      *CommunicationContext
	handlers     *RawPacketHandlerMap
	dataHandlers *DataPacketHandlerMap

	ackTicker      *time.Ticker
	dataPingTicker *time.Ticker

	remoteIndices map[*datamodel.Instance]uint32
	remoteLock    *sync.Mutex

	Connection   *net.UDPConn
	pingInterval int

	DataModel *datamodel.DataModel
}

func newPacketLogicHandler(context *CommunicationContext, withClient bool) PacketLogicHandler {
	return PacketLogicHandler{
		ConnectedPeer: NewConnectedPeer(context, withClient),
		handlers:      NewRawPacketHandlerMap(),
		dataHandlers:  NewDataHandlerMap(),

		remoteIndices: make(map[*datamodel.Instance]uint32),
		remoteLock:    &sync.Mutex{},

		Context:   context,
		DataModel: context.DataModel,
	}
}

func (logicHandler *PacketLogicHandler) RegisterPacketHandler(packetType uint8, handler ReceiveHandler) {
	logicHandler.handlers.Bind(packetType, handler)
}
func (logicHandler *PacketLogicHandler) RegisterDataHandler(packetType uint8, handler DataReceiveHandler) {
	logicHandler.dataHandlers.Bind(packetType, handler)
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

func (logicHandler *PacketLogicHandler) disconnectInternal() {
	if logicHandler.ackTicker != nil {
		logicHandler.ackTicker.Stop()
	}
	if logicHandler.dataPingTicker != nil {
		logicHandler.dataPingTicker.Stop()
	}
}

func (logicHandler *PacketLogicHandler) Disconnect() {
	logicHandler.WritePacket(&Packet15Layer{
		Reason: 0xFFFFFFFF,
	})
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
	dataHandlers.Bind(5, logicHandler.dataPingHandler)

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

func (logicHandler *PacketLogicHandler) SendEvent(instance *datamodel.Instance, name string, arguments ...rbxfile.Value) error {
	instance.FireEvent(name, arguments...)
	return logicHandler.WriteDataPackets(
		&Packet83_07{
			Instance:  instance,
			EventName: name,
			Event:     &ReplicationEvent{arguments},
		},
	)
}

func constructInstanceList(list []*datamodel.Instance, instance *datamodel.Instance) []*datamodel.Instance {
	list = append(list, instance.Children...)
	for _, child := range instance.Children {
		list = constructInstanceList(list, child)
	}
	return list
}

func (logicHandler *PacketLogicHandler) ReplicateJoinData(rootInstance *datamodel.Instance, replicateProperties, replicateChildren bool) error {
	list := []*datamodel.Instance{}
	// HACK: Replicating some instances to the client without including properties
	// may result in an error and a disconnection.
	// Here's a bad workaround
	rootInstance.PropertiesMutex.RLock()
	if replicateProperties && len(rootInstance.Properties) != 0 {
		list = append(list, rootInstance)
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
			list = append(list, rootInstance)
		}
	}
	rootInstance.PropertiesMutex.RUnlock()
	var joinDataObject *Packet83_0B
	// FIXME: This may result in the joindata becoming too large
	// for the client to handle! We need to split it up into
	// multiple segments
	if replicateChildren {
		joinDataObject = &Packet83_0B{
			Instances: constructInstanceList(list, rootInstance),
		}
	} else {
		joinDataObject = &Packet83_0B{
			Instances: list,
		}
	}

	return logicHandler.WriteDataPackets(joinDataObject)
}
