package peer
import "net"
import "fmt"
import "math/rand"
import "time"
import "sort"
import "github.com/gskartwii/rbxfile"
import "strconv"

var empty = struct{}{}
var noLocalDefaults = map[string]struct{}{
	"AdService": empty,
	"JointsService": empty,
	"Players": empty,
	"StarterGui": empty,
	"StarterPack": empty,
	"Workspace": empty,
}
var noncreatable = map[string]struct{}{
	"AnalyticsService": empty,
	"AssetService": empty,
    "BadgeService": empty,
	"CacheableContentProvider": empty,
	"ContentProvider": empty,
	"ChangeHistoryService": empty,
    "Chat": empty,
	"CollectionService": empty,
	"ContextActionService": empty,
	"CoreGui": empty,
    "CSGDictionaryService": empty,
	"ControllerService": empty,
	"CookiesService": empty,
	"DataStoreService": empty,
	"Debris": empty,
	"DebugSettings": empty,
	"FlagStandService": empty,
	"FlyweightService": empty,
    "FriendService": empty,
	"GamepadService": empty,
	"GamePassService": empty,
	"GameSettings": empty,
	"Geometry": empty,
	"GlobalSettings": empty,
	"GoogleAnalyticsConfiguration": empty,
	"GroupService": empty,
	"GuiRoot": empty,
	"GuidRegistryService": empty,
	"GuiService": empty,
	"HapticService": empty,
	"Hopper": empty,
	"HttpRbxApiService": empty,
	"HttpService": empty,
    "InsertService": empty,
	"InstancePacketCache": empty,
	"JointsService": empty,
	"KeyframeSequenceProvider": empty,
    "Lighting": empty,
	"LobbyService": empty,
    "LocalizationService": empty,
    "LogService": empty,
	"LoginService": empty,
	"LuaSettings": empty,
	"LuaWebService": empty,
    "MarketplaceService": empty,
	"MeshContentProvider": empty,
	"NetworkClient": empty,
	"NetworkServer": empty,
	"NetworkSettings": empty,
	"NonReplicatedCSGDictionaryService": empty,
	"NotificationService": empty,
	"OneQuarterClusterPacketCacheBase": empty,
	"ParallelRampPart": empty,
	"PathfindingService": empty,
	"PersonalServerService": empty,
	"PhysicsPacketCache": empty,
	"PhysicsService": empty,
	"PhysicsSettings": empty,
	"Platform": empty,
	"Players": empty,
    "PointsService": empty,
	"PrismPart": empty,
	"PyramidPart": empty,
	"RenderHooksService": empty,
	"RenderSettings": empty,
    "ReplicatedFirst": empty,
    "ReplicatedStorage": empty,
	"RightAngleRampPart": empty,
    "RobloxReplicatedStorage": empty,
	"RunService": empty,
	"RuntimeScriptService": empty,
	"ScriptContext": empty,
	"ScriptService": empty,
	"Selection": empty,
	"ServerScriptService": empty,
	"ServerStorage": empty,
	"SolidModelContentProvider": empty,
    "SoundService": empty,
	"SpawnerService": empty,
	"StarterGui": empty,
	"StarterPack": empty,
    "StarterPlayer": empty,
	"Stats": empty,
	"Studio": empty,
	"TaskScheduler": empty,
    "Teams": empty,
	"TeleportService": empty,
    "TestService": empty,
	"TextService": empty,
	"TextureContentProvider": empty,
	"ThirdPartyUserService": empty,
	"TimerService": empty,
	"TouchInputService": empty,
	"TouchInputUserService": empty,
	"TweenService": empty,
	"UserGameSettings": empty,
	"UserInputService": empty,
	"UserSettings": empty,
	"VirtualUser": empty,
	"Visit": empty,
	"VRService": empty,
    "Workspace": empty,

	"InputObject": empty,
	"ParabolaAdornment": empty,
	"Camera": empty,
	"Terrain": empty,
	"TouchInterest": empty,
	"Status": empty,
	"PlayerGui": empty,
}

var services = []string{
	"AdService",
	"BadgeService",
	"CSGDictionaryService",
	"Chat",
	"FriendService",
	"InsertService",
	"JointsService",
	"Lighting",
	"LocalizationService",
	"LogService",
	"MarketplaceService",
	"Players",
	"PointsService",
	"ReplicatedFirst",
	"ReplicatedStorage",
	"RobloxReplicatedStorage",
	"SoundService",
	"StarterGui",
	"StarterPack",
	"StarterPlayer",
	"Teams",
	"TestService",
	"Workspace",
}

type client struct {
	context *CommunicationContext
	address *net.UDPAddr
	reader *PacketReader
	writer *PacketWriter
	server *ServerPeer
	mustACK []int
    instanceID uint32
}

// ServerPeer describes a server that is hosted by roblox-dissector.
type ServerPeer struct {
	Connection *net.UDPConn
	clients map[string]*client
	Address *net.UDPAddr
	GUID uint64
    Dictionaries *Packet82Layer
    Schema *StaticSchema
}

func (client *client) sendACKs() {
	if len(client.mustACK) == 0 {
		return
	}
	acks := client.mustACK
	client.mustACK = []int{}
	var ackStructure []ACKRange
	sort.Ints(acks)

	for _, ack := range acks {
		if len(ackStructure) == 0 {
			ackStructure = append(ackStructure, ACKRange{uint32(ack), uint32(ack)})
			continue
		}

		inserted := false
		for i, ackRange := range ackStructure {
			if int(ackRange.Max) == ack {
				inserted = true
                break
			}
			if int(ackRange.Max + 1) == ack {
				ackStructure[i].Max++
				inserted = true
                break
			}
		}
		if inserted {
			continue
		}

		ackStructure = append(ackStructure, ACKRange{uint32(ack), uint32(ack)})
	}

	result := &RakNetLayer{
		IsValid: true,
		IsACK: true,
		ACKs: ackStructure,
	}

	client.writer.WriteRakNet(result, client.address)
}

func (client *client) receive(buf []byte) {
	packet := UDPPacketFromBytes(buf)
	packet.Source = *client.address
	packet.Destination = *client.server.Address
	client.reader.ReadPacket(buf, packet)
}

func newClient(addr *net.UDPAddr, server *ServerPeer) *client {
	var myClient *client
	context := NewCommunicationContext() // Peers will be detected by RakNet parser
	packetReader := &PacketReader{
		SimpleHandler: func(packetType byte, packet *UDPPacket, layers *PacketLayers) {
			if packetType == 0x5 {
				response := &Packet06Layer{
					GUID: server.GUID,
					UseSecurity: false,
					MTU: 1492,
				}

				myClient.writer.WriteSimple(6, response, addr)
			} else if packetType == 0x7 {
				response := &Packet08Layer{
					MTU: 1492,
					UseSecurity: false,
					IPAddress: addr,
				}

				myClient.writer.WriteSimple(8, response, addr)
			}
		},
		ReliableHandler: func(packetType byte, packet *UDPPacket, layers *PacketLayers) {
			rakNetLayer := layers.RakNet
			myClient.mustACK = append(myClient.mustACK, int(rakNetLayer.DatagramNumber))
		},
		FullReliableHandler: func(packetType byte, packet *UDPPacket, layers *PacketLayers) {
			if packetType == 0x0 {
				mainLayer := layers.Main.(Packet00Layer)
				response := &Packet03Layer{
					SendPingTime: mainLayer.SendPingTime,
					SendPongTime: mainLayer.SendPingTime + 10,
				}

				myClient.writer.WriteGeneric(context, 3, response, 2, addr)

				response2 := &Packet00Layer{
					SendPingTime: mainLayer.SendPingTime + 10,
				}
				myClient.writer.WriteGeneric(context, 0, response2, 2, addr)
			} else if packetType == 0x9 {
				mainLayer := layers.Main.(Packet09Layer)
				incomingTimestamp := mainLayer.Timestamp

				nullIP, _ := net.ResolveUDPAddr("udp", "255.255.255.255:0")
				loIP, _ := net.ResolveUDPAddr("udp", "127.0.0.1:0")
				response := &Packet10Layer{
					IPAddress: addr,
					SendPingTime: incomingTimestamp,
					SendPongTime: incomingTimestamp + 10,
					SystemIndex: 0,
					Addresses: [10]*net.UDPAddr{
						loIP,
						nullIP,
						nullIP,
						nullIP,
						nullIP,
						nullIP,
						nullIP,
						nullIP,
						nullIP,
						nullIP,
					},
				}

				myClient.writer.WriteGeneric(context, 0x10, response, 2, addr)
			} else if packetType == 0x90 {
				response := &Packet93Layer{
					ProtocolSchemaSync: true,
					ApiDictionaryCompression: true,
					Params: map[string]bool{
						"BodyColorsColor3PropertyReplicationEnabled": false,
						"PartColor3Uint8Enabled": false,
						"SendAdditionalNonAdjustedTimeStamp": true,
						"UseNewProtocolForStreaming": true,
						"UseNewPhysicsSender": false,
						"FixWeldedHumanoidsDeath": false,
						"UseNetworkSchema2": true,
					},
				}

				myClient.writer.WriteGeneric(context, 0x93, response, 3, addr)
			} else if packetType == 0x82 {
                response := server.Dictionaries

                myClient.writer.WriteGeneric(context, 0x82, response, 3, addr)

                response2 := &Packet97Layer{*server.Schema}
                context.StaticSchema = server.Schema

                myClient.writer.WriteGeneric(context, 0x97, response2, 3, addr)

                dataModel := &rbxfile.Root{}
                context.DataModel = dataModel

				var workspace *rbxfile.Instance
				var replicatedStorage *rbxfile.Instance

                initInstances := &Packet81Layer{
                    Items: make([]*Packet81LayerItem, len(services)),
					DistributedPhysicsEnabled: true,
					StreamJob: false,
					FilteringEnabled: false,
					AllowThirdPartySales: false,
					CharacterAutoSpawn: false,
                    Int1: 0,
                    Int2: 0,
                    ReferentString: []byte("RBX0123456789ABCDEF"),
                }
                for i, className := range services {
                    classID := context.StaticSchema.ClassesByName[className]
                    instance := &rbxfile.Instance{
                        ClassName: className,
                        Reference: strconv.Itoa(int(myClient.instanceID)),
                        IsService: true,
                        Properties: make(map[string]rbxfile.Value),
                    }
                    myClient.instanceID++

                    item := &Packet81LayerItem{
                        ClassID: uint16(classID),
                        Instance: instance,
                        Bool1: false,
                        Bool2: false,
                    }
					if className == "Workspace" {
						workspace = instance
					} else if className == "ReplicatedStorage" {
						replicatedStorage = instance
					}

                    initInstances.Items[i] = item
                }

                myClient.writer.WriteGeneric(context, 0x81, initInstances, 3, addr)

				joinData := &Packet83_0B{make([]*rbxfile.Instance, 0, len(services) + 1)}
                replicationResponse := &Packet83Layer{
                    SubPackets: []Packet83Subpacket{
                        &Packet83_10{
                            TagId: 12,
                        },
						joinData,
                        &Packet83_05{
                            false,
                            294470000,
                            0,
                            0,
                        },
                        &Packet83_10{
                            TagId: 13,
                        },
                    },
                }

                for _, item := range initInstances.Items {
					
					if _, ok := noLocalDefaults[item.Instance.ClassName]; ok {
						continue
					}
					joinData.Instances = append(joinData.Instances, item.Instance)
                }

                myClient.writer.WriteGeneric(context, 0x83, replicationResponse, 3, addr)

				onlyWorkspaceJoinData := &Packet83_0B{make([]*rbxfile.Instance, 0)}
				InputObject := &rbxfile.Instance{
					ClassName: "InputObject",
					Reference: strconv.Itoa(int(myClient.instanceID)),
					IsService: false,
					Properties: make(map[string]rbxfile.Value),
				}
				myClient.instanceID++
				onlyWorkspaceJoinData.Instances = append(onlyWorkspaceJoinData.Instances, InputObject)
				workspace.AddChild(InputObject)

				Explosion := &rbxfile.Instance{
					ClassName: "Explosion",
					Reference: strconv.Itoa(int(myClient.instanceID)),
					IsService: false,
					Properties: make(map[string]rbxfile.Value),
				}
				myClient.instanceID++
				onlyWorkspaceJoinData.Instances = append(onlyWorkspaceJoinData.Instances, Explosion)
				workspace.AddChild(Explosion)

				ParabolaAdornment := &rbxfile.Instance{
					ClassName: "ParabolaAdornment",
					Reference: strconv.Itoa(int(myClient.instanceID)),
					IsService: false,
					Properties: make(map[string]rbxfile.Value),
				}
				myClient.instanceID++
				onlyWorkspaceJoinData.Instances = append(onlyWorkspaceJoinData.Instances, ParabolaAdornment)
				workspace.AddChild(ParabolaAdornment)

				Terrain := &rbxfile.Instance{
					ClassName: "Terrain",
					Reference: strconv.Itoa(int(myClient.instanceID)),
					IsService: false,
					Properties: make(map[string]rbxfile.Value),
				}
				myClient.instanceID++
				onlyWorkspaceJoinData.Instances = append(onlyWorkspaceJoinData.Instances, Terrain)
				workspace.AddChild(Terrain)

				Status := &rbxfile.Instance{
					ClassName: "Status",
					Reference: strconv.Itoa(int(myClient.instanceID)),
					IsService: false,
					Properties: make(map[string]rbxfile.Value),
				}
				myClient.instanceID++
				onlyWorkspaceJoinData.Instances = append(onlyWorkspaceJoinData.Instances, Status)
				workspace.AddChild(Status)

				PlayerGui := &rbxfile.Instance{
					ClassName: "PlayerGui",
					Reference: strconv.Itoa(int(myClient.instanceID)),
					IsService: false,
					Properties: make(map[string]rbxfile.Value),
				}
				myClient.instanceID++
				onlyWorkspaceJoinData.Instances = append(onlyWorkspaceJoinData.Instances, PlayerGui)
				workspace.AddChild(PlayerGui)

				myClient.writer.WriteGeneric(context, 0x83, &Packet83Layer{
					[]Packet83Subpacket{onlyWorkspaceJoinData},
				}, 3, addr)

				allDefaultsJoinData := &Packet83_0B{make([]*rbxfile.Instance, 0, len(context.StaticSchema.Instances))}

				humanoid := &rbxfile.Instance{
					ClassName: "Humanoid",
					Reference: strconv.Itoa(int(myClient.instanceID)),
					IsService: false,
					Properties: make(map[string]rbxfile.Value),
				}
				myClient.instanceID++
				allDefaultsJoinData.Instances = append(allDefaultsJoinData.Instances, humanoid)
				replicatedStorage.AddChild(humanoid)
				animator := &rbxfile.Instance{
					ClassName: "Animator",
					Reference: strconv.Itoa(int(myClient.instanceID)),
					IsService: false,
					Properties: make(map[string]rbxfile.Value),
				}
				myClient.instanceID++
				allDefaultsJoinData.Instances = append(allDefaultsJoinData.Instances, animator)
				humanoid.AddChild(animator)

				part := &rbxfile.Instance{
					ClassName: "Part",
					Reference: strconv.Itoa(int(myClient.instanceID)),
					IsService: false,
					Properties: make(map[string]rbxfile.Value),
				}
				myClient.instanceID++
				allDefaultsJoinData.Instances = append(allDefaultsJoinData.Instances, part)
				replicatedStorage.AddChild(part)
				attachment := &rbxfile.Instance{
					ClassName: "Attachment",
					Reference: strconv.Itoa(int(myClient.instanceID)),
					IsService: false,
					Properties: make(map[string]rbxfile.Value),
				}
				myClient.instanceID++
				allDefaultsJoinData.Instances = append(allDefaultsJoinData.Instances, attachment)
				part.AddChild(attachment)


				for _, class := range context.StaticSchema.Instances {
					if _, ok := noncreatable[class.Name]; ok {
						continue
					}
					if class.Name == "Humanoid" || class.Name == "Animator" || class.Name == "Attachment" || class.Name == "DebuggerBreakpoint" || class.Name == "Part" || class.Name == "ScriptDebugger" || class.Name == "DebuggerManager" || class.Name == "DebuggerWatch" || class.Name == "Player" {
						continue
					}

                    instance := &rbxfile.Instance{
                        ClassName: class.Name,
                        Reference: strconv.Itoa(int(myClient.instanceID)),
                        IsService: false,
                        Properties: make(map[string]rbxfile.Value),
                    }
                    myClient.instanceID++

					allDefaultsJoinData.Instances = append(allDefaultsJoinData.Instances, instance)

					replicatedStorage.AddChild(instance)
				}

				myClient.writer.WriteGeneric(context, 0x83, &Packet83Layer{
					[]Packet83Subpacket{allDefaultsJoinData},
				}, 3, addr)
            }
		},
		ErrorHandler: func(err error) {
			println(err.Error())
		},
		Context: context,
	}
	packetWriter := NewPacketWriter()
	packetWriter.ErrorHandler = func(err error) {
		println(err.Error())
	}
	packetWriter.OutputHandler = func(payload []byte, dest *net.UDPAddr) {
		num, err := server.Connection.WriteToUDP(payload, dest)
		if err != nil {
			fmt.Printf("Wrote %d bytes, err: %s", num, err.Error())
		}
	}
	

	myClient = &client{
		reader: packetReader,
		writer: packetWriter,
		context: context,
		address: addr,
		server: server,
        instanceID: 1000,
	}

	ackTicker := time.NewTicker(17)
	go func() {
		for {
			<- ackTicker.C
			myClient.sendACKs()
		}
	}()
	return myClient
}

// StartServer attempts to start a server within roblox-dissector for _Studio_ clients
// to connect to.
// Dictionaries and schema should come from gobs dumped by the dissector.
func StartServer(port uint16, dictionaries *Packet82Layer, schema *StaticSchema) error {
	server := &ServerPeer{clients: make(map[string]*client)}

	var err error
	server.Address, err = net.ResolveUDPAddr("udp", fmt.Sprintf(":%d", port))
	if err != nil {
		return err
	}

	conn, err := net.ListenUDP("udp", server.Address)
	defer conn.Close()
	if err != nil {
		return err
	}
	server.Connection = conn
	server.GUID = rand.Uint64()
    server.Dictionaries = dictionaries
    server.Schema = schema

	buf := make([]byte, 1492)

	for {
		n, client, err := conn.ReadFromUDP(buf)
		if err != nil {
			println("Err:", err.Error())
			continue
		}

		thisClient, ok := server.clients[client.String()]
		if !ok {
			thisClient = newClient(client, server)
			server.clients[client.String()] = thisClient
		}
		thisClient.receive(buf[:n])
	}
}
