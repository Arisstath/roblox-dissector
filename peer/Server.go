package peer

import (
	"context"
	"fmt"
	"math/rand"
	"net"

	"github.com/Gskartwii/roblox-dissector/datamodel"
	"github.com/olebedev/emitter"
)

// TODO: Filtering?
type ServerClient struct {
	PacketLogicHandler
	Server  *CustomServer
	Address *net.UDPAddr

	Player *datamodel.Instance
	Index  int

	replicatedInstances []*ReplicationContainer
	handlingChild       *datamodel.Instance
	handlingProp        handledChange
	handlingEvent       handledChange
	handlingRemoval     *datamodel.Instance
}

type CustomServer struct {
	Context            *CommunicationContext
	Connection         *net.UDPConn
	Clients            map[string]*ServerClient
	ClientEmitter      *emitter.Emitter
	Address            *net.UDPAddr
	GUID               uint64
	Schema             *StaticSchema
	InstanceDictionary *datamodel.InstanceDictionary
	RunningContext     context.Context

	PlayerIndex int
}

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

func (client *ServerClient) Init() {
	client.bindDefaultHandlers()
	// Write to server's connection
	client.Connection = client.Server.Connection
	client.createWriter()

	client.Connected = true

	client.startAcker()
}
func NewServerClient(clientAddr *net.UDPAddr, server *CustomServer, context *CommunicationContext) *ServerClient {
	newContext := &CommunicationContext{
		InstancesByReference: context.InstancesByReference,
		DataModel:            context.DataModel,
		StaticSchema:         context.StaticSchema,
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

func (myServer *CustomServer) Start() error {
	conn, err := net.ListenUDP("udp", myServer.Address)
	defer conn.Close()
	if err != nil {
		return err
	}
	myServer.Connection = conn

	buf := make([]byte, 1492)

	for {
		n, client, err := conn.ReadFromUDP(buf)
		if err != nil {
			return err
		}

		thisClient, ok := myServer.Clients[client.String()]
		if !ok {
			thisClient = NewServerClient(client, myServer, myServer.Context)
			myServer.Clients[client.String()] = thisClient
			thisClient.Init()

			<-myServer.ClientEmitter.Emit("client", thisClient)
		}
		thisClient.ReadPacket(buf[:n])
	}
}

func (myServer *CustomServer) Stop() {
	for _, client := range myServer.Clients {
		client.Disconnect()
	}
	myServer.Connection.Close()
}

func NewCustomServer(ctx context.Context, port uint16, schema *StaticSchema, dataModel *datamodel.DataModel, dict *datamodel.InstanceDictionary) (*CustomServer, error) {
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
	server.Context.StaticSchema = schema
	server.InstanceDictionary = dict
	server.Context.InstanceTopScope = server.InstanceDictionary.Scope
	server.ClientEmitter = emitter.New(0)

	return server, nil
}
