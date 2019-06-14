package peer

import (
	"context"
	"fmt"
	"math/rand"
	"net"

	"github.com/Gskartwii/roblox-dissector/datamodel"
	"github.com/olebedev/emitter"
	"github.com/robloxapi/rbxfile"
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

var noLocalDefaults = map[string](map[string]rbxfile.Value){
	"StarterGui": map[string]rbxfile.Value{
		"Archivable":            rbxfile.ValueBool(true),
		"Name":                  rbxfile.ValueString("StarterGui"),
		"ResetPlayerGuiOnSpawn": rbxfile.ValueBool(false),
		"RobloxLocked":          rbxfile.ValueBool(false),
		// TODO: Set token ID correctly here_
		"ScreenOrientation":  datamodel.ValueToken{Value: 0},
		"ShowDevelopmentGui": rbxfile.ValueBool(true),
		"Tags":               rbxfile.ValueBinaryString(""),
	},
	"Workspace": map[string]rbxfile.Value{
		"Archivable": rbxfile.ValueBool(true),
		// TODO: Set token ID correctly here_
		"AutoJointsMode":             datamodel.ValueToken{Value: 0},
		"CollisionGroups":            rbxfile.ValueString(""),
		"ExpSolverEnabled_Replicate": rbxfile.ValueBool(true),
		"ExplicitAutoJoints":         rbxfile.ValueBool(true),
		"FallenPartsDestroyHeight":   rbxfile.ValueFloat(-500.0),
		"FilteringEnabled":           rbxfile.ValueBool(true),
		"Gravity":                    rbxfile.ValueFloat(196.2),
		"ModelInPrimary":             rbxfile.ValueCFrame{},
		"Name":                       rbxfile.ValueString("Workspace"),
		"PrimaryPart":                datamodel.ValueReference{Instance: nil, Reference: datamodel.NullReference},
		"RobloxLocked":               rbxfile.ValueBool(false),
		"StreamingEnabled":           rbxfile.ValueBool(false),
		"StreamingMinRadius":         rbxfile.ValueInt(0),
		"StreamingTargetRadius":      rbxfile.ValueInt(0),
		"Tags":                       rbxfile.ValueBinaryString(""),
		"TerrainWeldsFixed":          rbxfile.ValueBool(true),
	},
	"StarterPack": map[string]rbxfile.Value{
		"Archivable":   rbxfile.ValueBool(true),
		"Name":         rbxfile.ValueString("StarterPack"),
		"RobloxLocked": rbxfile.ValueBool(false),
		"Tags":         rbxfile.ValueBinaryString(""),
	},
	"TeleportService": map[string]rbxfile.Value{
		"Archivable":   rbxfile.ValueBool(true),
		"Name":         rbxfile.ValueString("Teleport Service"), // intentional
		"RobloxLocked": rbxfile.ValueBool(false),
		"Tags":         rbxfile.ValueBinaryString(""),
	},
	"LocalizationService": map[string]rbxfile.Value{
		"Archivable":           rbxfile.ValueBool(true),
		"IsTextScraperRunning": rbxfile.ValueBool(false),
		"LocaleManifest":       rbxfile.ValueString("en-us"),
		"Name":                 rbxfile.ValueString("LocalizationService"),
		"RobloxLocked":         rbxfile.ValueBool(false),
		"ShouldUseCloudTable":  rbxfile.ValueBool(false),
		"Tags":                 rbxfile.ValueBinaryString(""),
		"WebTableContents":     rbxfile.ValueString(""),
	},
	"Players": map[string]rbxfile.Value{
		"Archivable":         rbxfile.ValueBool(true),
		"MaxPlayersInternal": rbxfile.ValueInt(6),
		"Name":               rbxfile.ValueString("Players"),
		"PreferredPlayersInternal": rbxfile.ValueInt(6),
		"RespawnTime":              rbxfile.ValueFloat(5.0),
		"RobloxLocked":             rbxfile.ValueBool(false),
		"Tags":                     rbxfile.ValueBinaryString(""),
	},
}

// normalizeTypes changes the types of instances from binary format types to network types
func normalizeTypes(children []*datamodel.Instance, schema *NetworkSchema) {
	for _, instance := range children {
		defaultValues, ok := noLocalDefaults[instance.ClassName]
		if ok {
			for _, prop := range schema.SchemaForClass(instance.ClassName).Properties {
				if _, ok = instance.Properties[prop.Name]; !ok {
					println("Adding missing default value", instance.ClassName, prop.Name)
					instance.Properties[prop.Name] = defaultValues[prop.Name]
				}
			}
		}

		// hack: color is saved in the wrong format
		if instance.ClassName == "Part" {
			color := instance.Get("Color")
			if color != nil {
				instance.Set("Color3uint8", color)
				delete(instance.Properties, "Color")
			}
		}

		for name, prop := range instance.Properties {
			propSchema := schema.SchemaForClass(instance.ClassName).SchemaForProp(name)
			if propSchema == nil {
				fmt.Printf("Warning: %s.%s doesn't exist in schema! Stripping this property.\n", instance.ClassName, name)
				delete(instance.Properties, name)
				continue
			}
			switch propSchema.Type {
			case PropertyTypeProtectedString0,
				PropertyTypeProtectedString1,
				PropertyTypeProtectedString2,
				PropertyTypeProtectedString3:
				// This type may be encoded correctly depending on the format
				if _, ok = prop.(rbxfile.ValueString); ok {
					instance.Properties[name] = rbxfile.ValueProtectedString(prop.(rbxfile.ValueString))
				}
			case PropertyTypeContent:
				// This type may be encoded correctly depending on the format
				if _, ok = prop.(rbxfile.ValueString); ok {
					instance.Properties[name] = rbxfile.ValueContent(prop.(rbxfile.ValueString))
				}
			case PropertyTypeEnum:
				instance.Properties[name] = datamodel.ValueToken{ID: propSchema.EnumID, Value: prop.(datamodel.ValueToken).Value}
			case PropertyTypeBinaryString:
				// This type may be encoded correctly depending on the format
				if _, ok = prop.(rbxfile.ValueString); ok {
					instance.Properties[name] = rbxfile.ValueBinaryString(prop.(rbxfile.ValueString))
				}
			case PropertyTypeColor3uint8:
				if _, ok = prop.(rbxfile.ValueColor3); ok {
					propc3 := prop.(rbxfile.ValueColor3)
					instance.Properties[name] = rbxfile.ValueColor3uint8{R: uint8(propc3.R * 255), G: uint8(propc3.G * 255), B: uint8(propc3.B * 255)}
				}
			case PropertyTypeBrickColor:
				if _, ok = prop.(rbxfile.ValueInt); ok {
					instance.Properties[name] = rbxfile.ValueBrickColor(prop.(rbxfile.ValueInt))
				}
			}
		}
		normalizeTypes(instance.Children, schema)
	}
}

func normalizeChildren(instances []*datamodel.Instance, schema *NetworkSchema) {
	for _, inst := range instances {
		newChildren := make([]*datamodel.Instance, 0, len(inst.Children))
		for _, child := range inst.Children {
			class := schema.SchemaForClass(child.ClassName)
			if class == nil {
				fmt.Printf("Warning: %s doesn't exist in schema! Stripping this instance.\n", child.ClassName)
				continue
			}

			newChildren = append(newChildren, child)
		}

		inst.Children = newChildren
		normalizeChildren(inst.Children, schema)
	}
}

func normalizeServices(root *datamodel.DataModel, schema *NetworkSchema) {
	newInstances := make([]*datamodel.Instance, 0, len(root.Instances))
	for _, serv := range root.Instances {
		class := schema.SchemaForClass(serv.ClassName)
		if class == nil {
			fmt.Printf("Warning: %s doesn't exist in schema! Stripping this instance.\n", serv.ClassName)
			continue
		}

		newInstances = append(newInstances, serv)
	}

	root.Instances = newInstances
}

// NormalizeDataModel changes a DataModel loaded by rbxfile so that it can be used by the server
func NormalizeDataModel(dataModel *datamodel.DataModel, schema *NetworkSchema) {
	normalizeServices(dataModel, schema)
	// Clear children of some services if they exist
	players := dataModel.FindService("Players")
	if players != nil {
		players.Children = nil
	}
	joints := dataModel.FindService("JointsService")
	if joints != nil {
		joints.Children = nil
	}
	normalizeServices(dataModel, schema)
	normalizeChildren(dataModel.Instances, schema)
	normalizeTypes(dataModel.Instances, schema)
}
