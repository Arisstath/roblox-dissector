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

// PacketLogicHandler is a generic struct for connections which
// should implements DataModelHandlers
// TODO: Add Logger to this struct?
type PacketLogicHandler struct {
	*ConnectedPeer
	Replicator     *Replicator
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

	GenericEvents *emitter.Emitter
}

func newPacketLogicHandler(context *CommunicationContext, withClient bool) PacketLogicHandler {
	return PacketLogicHandler{
		ConnectedPeer: NewConnectedPeer(context, withClient),
		Replicator:    NewReplicator(),

		remoteIndices: make(map[*datamodel.Instance]uint32),
		remoteLock:    &sync.Mutex{},

		Context:   context,
		DataModel: context.DataModel,

		GenericEvents: emitter.New(0),
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

func (logicHandler *PacketLogicHandler) cleanup() {
	logicHandler.Connected = false
	// Note: these will NOT close the channel!
	if logicHandler.ackTicker != nil {
		logicHandler.ackTicker.Stop()
	}
	if logicHandler.dataPingTicker != nil {
		logicHandler.dataPingTicker.Stop()
	}
}

// DisconnectionSource is a type describing what caused a disconnection
type DisconnectionSource uint

const (
	// LocalDisconnection represents a disconnection caused by
	// the local peer (i.e. by PacketLogicHandler.Disconnect())
	LocalDisconnection DisconnectionSource = iota
	// RemoteDisconnection represents a disconnection caused by
	// the remote peer
	RemoteDisconnection
)

// Disconnect sends a "-1" disconnection reason packet
// to the remote peer. Note that it doesn't close the
// underlying connection
func (logicHandler *PacketLogicHandler) Disconnect() {
	if logicHandler.Connected {
		logicHandler.WritePacket(&Packet15Layer{
			Reason: -1,
		})
		<-logicHandler.GenericEvents.Emit("disconnected", LocalDisconnection, int32(-1))
		logicHandler.cleanup()
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

	<-logicHandler.GenericEvents.Emit("disconnected", RemoteDisconnection, mainLayer.Reason)
	logicHandler.cleanup()
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

	logicHandler.Replicator.Bind(logicHandler.DefaultPacketReader, logicHandler.DefaultPacketWriter)

	// do NOT call PacketReader.BindDefaultHandlers() here!
	// ServerClients are packet readers which don't want that
}

// WriteDataPackets sends the given Packet83Subpackets to the peer
func (logicHandler *PacketLogicHandler) WriteDataPackets(packets ...Packet83Subpacket) error {
	err := logicHandler.WritePacket(&Packet83Layer{
		SubPackets: packets,
	})
	return err
}

// SendHackFlag attempts to fire the StatsAvailable event on the given player
func (logicHandler *PacketLogicHandler) SendHackFlag(player *datamodel.Instance, flag string) {
	player.FireEvent("StatsAvailable", rbxfile.ValueString(flag))
}
