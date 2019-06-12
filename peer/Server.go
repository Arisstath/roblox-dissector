package peer

import (
	"context"
	"fmt"
	"math/rand"
	"net"

	"github.com/Gskartwii/roblox-dissector/datamodel"
	"github.com/olebedev/emitter"
)

// ServerClient represents a local server's connection to a remote
// client
// TODO: Filtering?
type ServerClient struct {
	PacketLogicHandler
	Server  *CustomServer
	Address *net.UDPAddr

	Player *datamodel.Instance
	// Index is the player's index within the server.
	// Among other things, it is used in the determining the player's name
	// (i.e. Player1, Player2, etc.)
	Index int
}

// CustomServer is custom implementation of a Roblox server
type CustomServer struct {
	Context            *CommunicationContext
	Connection         *net.UDPConn
	Clients            map[string]*ServerClient
	ClientEmitter      *emitter.Emitter
	Address            *net.UDPAddr
	GUID               uint64
	Schema             *NetworkSchema
	InstanceDictionary *datamodel.InstanceDictionary
	RunningContext     context.Context

	PlayerIndex int
}

// ReadPacket processes a UDP packet sent by the client
// Its first argument is a byte slice containing the UDP payload
func (client *ServerClient) ReadPacket(buf []byte) {
	layers := &PacketLayers{
		Root: RootLayer{
			Source:      client.Address,
			Destination: client.Server.Address,
			FromClient:  true,
		},
	}
	client.ConnectedPeer.ReadPacket(buf, layers)
}

func (client *ServerClient) createWriter() {
	client.Output.On("udp", func(e *emitter.Event) {
		num, err := client.Connection.WriteToUDP(e.Args[0].([]byte), client.Address)
		if err != nil {
			fmt.Printf("Wrote %d bytes, err: %s\n", num, err.Error())
		}
	}, emitter.Void)
	client.DefaultPacketWriter.LayerEmitter.On("*", func(e *emitter.Event) {
		e.Args[0].(*PacketLayers).Root = RootLayer{
			FromServer:  true,
			Logger:      nil,
			Source:      client.Server.Address,
			Destination: client.Address,
		}
	}, emitter.Void)
}

func (client *ServerClient) init() {
	client.createWriter()
	client.bindDefaultHandlers()
	// Write to server's connection
	client.Connection = client.Server.Connection

	client.Connected = true

	client.startAcker()
}

func newServerClient(clientAddr *net.UDPAddr, server *CustomServer, context *CommunicationContext) *ServerClient {
	newContext := &CommunicationContext{
		InstancesByReference: context.InstancesByReference,
		DataModel:            context.DataModel,
		NetworkSchema:        context.NetworkSchema,
		InstanceTopScope:     context.InstanceTopScope,
	}

	server.PlayerIndex++
	newClient := &ServerClient{
		PacketLogicHandler: newPacketLogicHandler(newContext, true),
		Server:             server,
		Address:            clientAddr,
		Index:              server.PlayerIndex,
	}
	newClient.RunningContext = server.RunningContext

	return newClient
}

func (myServer *CustomServer) bindToDisconnection(client *ServerClient) {
	// HACK: gets priority in the emitter via Use()
	client.GenericEvents.Use("disconnected", func(e *emitter.Event) {
		println("server received client disconnection")
		delete(myServer.Clients, client.Address.String())
	})
}

// Start starts the server's read loop
func (myServer *CustomServer) Start() error {
	conn, err := net.ListenUDP("udp", myServer.Address)
	if err != nil {
		return err
	}
	myServer.Connection = conn
	defer myServer.stop()

	buf := make([]byte, 1492)
	for {
		n, client, err := conn.ReadFromUDP(buf)
		if err != nil {
			return err
		}

		select {
		case <-myServer.RunningContext.Done():
			return myServer.RunningContext.Err()
		default:
		}

		thisClient, ok := myServer.Clients[client.String()]
		if !ok {
			// always check for offline messages, disconnected peers
			// may keep sending packets which must be ignored
			if !IsOfflineMessage(buf[:n]) {
				continue
			}
			thisClient = newServerClient(client, myServer, myServer.Context)
			myServer.Clients[client.String()] = thisClient

			myServer.bindToDisconnection(thisClient)

			thisClient.init()

			<-myServer.ClientEmitter.Emit("client", thisClient)
		}
		thisClient.ReadPacket(buf[:n])
	}
}

func (myServer *CustomServer) stop() {
	for _, client := range myServer.Clients {
		client.Disconnect()
	}
	myServer.Connection.Close()
}

// NewCustomServer initializes a CustomServer
func NewCustomServer(ctx context.Context, port uint16, schema *NetworkSchema, dataModel *datamodel.DataModel, dict *datamodel.InstanceDictionary) (*CustomServer, error) {
	server := &CustomServer{Clients: make(map[string]*ServerClient)}

	var err error
	server.Address, err = net.ResolveUDPAddr("udp", fmt.Sprintf(":%d", port))
	if err != nil {
		return server, err
	}

	server.RunningContext = ctx
	server.GUID = rand.Uint64()
	server.Schema = schema
	server.Context = NewCommunicationContext()
	server.Context.DataModel = dataModel
	server.Context.NetworkSchema = schema
	server.InstanceDictionary = dict
	server.Context.InstanceTopScope = server.InstanceDictionary.Scope
	server.ClientEmitter = emitter.New(0)

	return server, nil
}
