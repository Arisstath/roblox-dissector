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
	Context *CommunicationContext

	ackTicker      *time.Ticker
	dataPingTicker *time.Ticker

	remoteIndices map[*datamodel.Instance]uint32
	remoteLock    *sync.Mutex

	Connection   *net.UDPConn
	pingInterval int

	DataModel *datamodel.DataModel
	Connected bool
}

func newPacketLogicHandler(context *CommunicationContext, withClient bool) PacketLogicHandler {
	return PacketLogicHandler{
		ConnectedPeer: NewConnectedPeer(context, withClient),

		remoteIndices: make(map[*datamodel.Instance]uint32),
		remoteLock:    &sync.Mutex{},

		Context:   context,
		DataModel: context.DataModel,
	}
}

func (logicHandler *PacketLogicHandler) createReader() {
	logicHandler.DefaultPacketReader.SetContext(logicHandler.Context)
}

// only used by server and Studio? client must use ClientPacketLogic.go
func (logicHandler *PacketLogicHandler) startDataPing() {
	// boot up dataping
	logicHandler.dataPingTicker = time.NewTicker(time.Duration(logicHandler.pingInterval) * time.Millisecond)
	go func() {
		for {
			<-logicHandler.dataPingTicker.C

			logicHandler.WritePacket(&Packet83Layer{
				[]Packet83Subpacket{&Packet83_05{
					Timestamp:     uint64(time.Now().UnixNano() / int64(time.Millisecond)),
					PacketVersion: 0,
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
	if logicHandler.Connected {
		logicHandler.WritePacket(&Packet15Layer{
			Reason: -1,
		})
	}
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

func (logicHandler *PacketLogicHandler) disconnectHandler(packetType uint8, layers *PacketLayers) {
	mainLayer := layers.Main.(*Packet15Layer)
	fmt.Printf("Received disconnect with reason %d\n", mainLayer.Reason)

	logicHandler.disconnectInternal()
}

func (logicHandler *PacketLogicHandler) sendPing() {
	packet := &Packet00Layer{
		SendPingTime: uint64(time.Now().UnixNano() / int64(time.Millisecond)),
	}

	_, err := logicHandler.WritePacket(packet)
	if err != nil {
		println("Failed to write ping: ", err.Error())
	}
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
	dataHandlers := logicHandler.DataHandler
	dataHandlers.Bind(5, logicHandler.dataPingHandler)

	basicHandlers := logicHandler.FullReliableHandler
	basicHandlers.Bind(0x15, logicHandler.disconnectHandler)
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

func (logicHandler *PacketLogicHandler) SendHackFlag(player *datamodel.Instance, flag string) error {
	return logicHandler.SendEvent(player, "StatsAvailable", rbxfile.ValueString(flag))
}
