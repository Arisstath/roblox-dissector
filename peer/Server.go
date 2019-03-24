package peer

import (
	"fmt"
	"math/rand"
	"net"
	"time"

	"github.com/gskartwii/roblox-dissector/datamodel"
	"github.com/olebedev/emitter"
)

// TODO: Filtering?
type ServerClient struct {
	PacketLogicHandler
	Server  *CustomServer
	Address *net.UDPAddr
}

type CustomServer struct {
	Context            *CommunicationContext
	Connection         *net.UDPConn
	Clients            map[string]*ServerClient
	Address            *net.UDPAddr
	GUID               uint64
	Schema             *StaticSchema
	InstanceDictionary *datamodel.InstanceDictionary
}

func (client *ServerClient) ReadPacket(buf []byte) {
	layers := &PacketLayers{
		Root: RootLayer{
			Source:      client.Address,
			Destination: client.Server.Address,
			FromServer:  false,
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
}

func (client *ServerClient) Init() {
	client.createReader()
	client.bindDefaultHandlers()
	// Write to server's connection
	client.Connection = client.Server.Connection
	client.createWriter()

	client.SetIsClient(false)
	client.SetToClient(true)

	client.Connected = true

	client.startAcker()
}
func NewServerClient(clientAddr *net.UDPAddr, server *CustomServer, context *CommunicationContext) *ServerClient {
	newContext := &CommunicationContext{
		InstancesByReferent: context.InstancesByReferent,
		DataModel:           context.DataModel,
		Server:              server.Address,
		Client:              clientAddr,
		StaticSchema:        context.StaticSchema,
		InstanceTopScope:    context.InstanceTopScope,
	}

	newClient := &ServerClient{
		PacketLogicHandler: newPacketLogicHandler(newContext, true),
		Server:             server,
		Address:            clientAddr,
	}

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

func NewCustomServer(port uint16, schema *StaticSchema, dataModel *datamodel.DataModel) (*CustomServer, error) {
	server := &CustomServer{Clients: make(map[string]*ServerClient)}

	var err error
	server.Address, err = net.ResolveUDPAddr("udp", fmt.Sprintf(":%d", port))
	if err != nil {
		return server, err
	}

	rand.Seed(time.Now().UnixNano())
	server.GUID = rand.Uint64()
	server.Schema = schema
	server.Context = NewCommunicationContext()
	server.Context.DataModel = dataModel
	server.Context.StaticSchema = schema
	server.InstanceDictionary = datamodel.NewInstanceDictionary()

	return server, nil
}
