package peer

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/Gskartwii/roblox-dissector/datamodel"
	"github.com/olebedev/emitter"
	"github.com/robloxapi/rbxfile"
)

// TODO: Add Logger to this struct?
type PacketLogicHandler struct {
	*ConnectedPeer
	Context        *CommunicationContext
	RunningContext context.Context

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

// only used by server and Studio? client must use ClientPacketLogic.go
func (logicHandler *PacketLogicHandler) startDataPing() {
	// boot up dataping
	logicHandler.dataPingTicker = time.NewTicker(time.Duration(logicHandler.pingInterval) * time.Millisecond)
	go func() {
		for {
			select {
			case <-logicHandler.dataPingTicker.C:
				logicHandler.WritePacket(&Packet83Layer{
					[]Packet83Subpacket{&Packet83_05{
						Timestamp:     uint64(time.Now().UnixNano() / int64(time.Millisecond)),
						PacketVersion: 0,
					}},
				})
			case <-logicHandler.RunningContext.Done():
				return
			}
		}
	}()
}

func (logicHandler *PacketLogicHandler) startAcker() {
	logicHandler.ackTicker = time.NewTicker(500 * time.Millisecond)
	go func() {
		for {
			select {
			case <-logicHandler.ackTicker.C:
				err := logicHandler.sendACKs()
				if err != nil {
					println("ACK Error:", err.Error())
				}
			case <-logicHandler.RunningContext.Done():
				return
			}
		}
	}()
}

func (logicHandler *PacketLogicHandler) defaultReliabilityLayerHandler(e *emitter.Event) {
	logicHandler.mustACK = append(logicHandler.mustACK, int(e.Args[0].(*PacketLayers).RakNet.DatagramNumber))
}

func (logicHandler *PacketLogicHandler) disconnectInternal() {
	// Note: these will NOT close the channel!
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
func (logicHandler *PacketLogicHandler) dataPingHandler(e *emitter.Event) {
	logicHandler.sendDataPingBack()
}

func (logicHandler *PacketLogicHandler) disconnectHandler(e *emitter.Event) {
	mainLayer := e.Args[0].(*Packet15Layer)
	fmt.Printf("Received disconnect with reason %d\n", mainLayer.Reason)

	logicHandler.disconnectInternal()
}

func (logicHandler *PacketLogicHandler) sendPing() {
	packet := &Packet00Layer{
		SendPingTime: uint64(time.Now().UnixNano() / int64(time.Millisecond)),
	}

	err := logicHandler.WritePacket(packet)
	if err != nil {
		println("Failed to write ping: ", err.Error())
	}
}

func (logicHandler *PacketLogicHandler) sendPong(pingTime uint64) {
	response := &Packet03Layer{
		SendPingTime: pingTime,
		SendPongTime: uint64(time.Now().UnixNano() / int64(time.Millisecond)),
	}

	err := logicHandler.WritePacket(response)
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
	logicHandler.DefaultPacketReader.LayerEmitter.On("reliability", logicHandler.defaultReliabilityLayerHandler, emitter.Void)
	dataHandlers := logicHandler.DataEmitter
	dataHandlers.On("ID_REPLIC_PING", logicHandler.dataPingHandler, emitter.Void)

	basicHandlers := logicHandler.PacketEmitter
	basicHandlers.On("ID_DISCONNECTION_NOTIFICATION", logicHandler.disconnectHandler, emitter.Void)

	// do NOT call PacketReader.BindDefaultHandlers() here!
	// ServerClients are packet readers which don't want that
}

func (logicHandler *PacketLogicHandler) WriteDataPackets(packets ...Packet83Subpacket) error {
	err := logicHandler.WritePacket(&Packet83Layer{
		SubPackets: packets,
	})
	return err
}

func (logicHandler *PacketLogicHandler) SendEvent(instance *datamodel.Instance, name string, arguments ...rbxfile.Value) error {
	instance.FireEvent(name, arguments...)
	return logicHandler.WriteDataPackets(
		&Packet83_07{
			Instance: instance,
			Schema:   logicHandler.Context.StaticSchema.SchemaForClass(instance.ClassName).SchemaForEvent(name),
			Event:    &ReplicationEvent{arguments},
		},
	)
}

func (logicHandler *PacketLogicHandler) ReplicationInstance(inst *datamodel.Instance, deleteOnDisconnect bool) *ReplicationInstance {
	repInstance := &ReplicationInstance{}
	repInstance.DeleteOnDisconnect = deleteOnDisconnect
	repInstance.Instance = inst
	repInstance.Parent = inst.Parent()
	repInstance.Schema = logicHandler.Context.StaticSchema.SchemaForClass(inst.ClassName)
	inst.PropertiesMutex.RLock()
	repInstance.Properties = make(map[string]rbxfile.Value, len(inst.Properties))
	for name, value := range inst.Properties {
		repInstance.Properties[name] = value
	}
	inst.PropertiesMutex.RUnlock()

	return repInstance
}

func (logicHandler *PacketLogicHandler) constructInstanceList(list []*ReplicationInstance, instance *datamodel.Instance) []*ReplicationInstance {
	for _, child := range instance.Children {
		list = append(list, logicHandler.ReplicationInstance(child, false))
		list = logicHandler.constructInstanceList(list, child)
	}
	return list
}

func (logicHandler *PacketLogicHandler) ReplicateJoinData(rootInstance *datamodel.Instance, replicateProperties, replicateChildren bool) error {
	list := []*ReplicationInstance{}
	// HACK: Replicating some instances to the client without including properties
	// may result in an error and a disconnection.
	// Here's a bad workaround
	rootInstance.PropertiesMutex.RLock()
	if replicateProperties && len(rootInstance.Properties) != 0 {
		list = append(list, logicHandler.ReplicationInstance(rootInstance, false))
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
			list = append(list, logicHandler.ReplicationInstance(rootInstance, false))
		}
	}
	rootInstance.PropertiesMutex.RUnlock()
	var joinDataObject *Packet83_0B
	// FIXME: This may result in the joindata becoming too large
	// for the client to handle! We need to split it up into
	// multiple segments
	if replicateChildren {
		joinDataObject = &Packet83_0B{
			Instances: logicHandler.constructInstanceList(list, rootInstance),
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

func (logicHandler *PacketLogicHandler) ReplicateInstance(inst *datamodel.Instance, deleteOnDisconnect bool) error {
	return logicHandler.WriteDataPackets(&Packet83_02{logicHandler.ReplicationInstance(inst, deleteOnDisconnect)})
}
