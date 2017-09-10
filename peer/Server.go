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

type Client struct {
	Context *CommunicationContext
	Address *net.UDPAddr
	Reader *PacketReader
	Writer *PacketWriter
	Server *ServerPeer
	MustACK []int
    InstanceID uint32
}

type ServerPeer struct {
	Connection *net.UDPConn
	Clients map[string]*Client
	Address *net.UDPAddr
	GUID uint64
    Dictionaries *Packet82Layer
    Schema *StaticSchema
}

func (client *Client) SendACKs() {
	if len(client.MustACK) == 0 {
		return
	}
	acks := client.MustACK
	client.MustACK = []int{}
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

	client.Writer.WriteRakNet(result)
}

func (client *Client) Receive(buf []byte) {
	packet := &UDPPacket{
		Stream: BufferToStream(buf),
		Source: *client.Address,
		Destination: *client.Server.Address,
	}
	client.Reader.ReadPacket(buf, packet)
}

func newClient(addr *net.UDPAddr, server *ServerPeer) *Client {
	var client *Client
	context := NewCommunicationContext() // Peers will be detected by RakNet parser
	packetReader := &PacketReader{
		SimpleHandler: func(packetType byte, packet *UDPPacket, layers *PacketLayers) {
			if packetType == 0x5 {
				response := &Packet06Layer{
					GUID: server.GUID,
					UseSecurity: false,
					MTU: 1492,
				}

				client.Writer.WriteSimple(6, response)
			} else if packetType == 0x7 {
				response := &Packet08Layer{
					MTU: 1492,
					UseSecurity: false,
					IPAddress: addr,
				}

				client.Writer.WriteSimple(8, response)
			}
		},
		ReliableHandler: func(packetType byte, packet *UDPPacket, layers *PacketLayers) {
			rakNetLayer := layers.RakNet
			client.MustACK = append(client.MustACK, int(rakNetLayer.DatagramNumber))
		},
		FullReliableHandler: func(packetType byte, packet *UDPPacket, layers *PacketLayers) {
			if packetType == 0x0 {
				mainLayer := layers.Main.(Packet00Layer)
				response := &Packet03Layer{
					SendPingTime: mainLayer.SendPingTime,
					SendPongTime: mainLayer.SendPingTime + 10,
				}

				client.Writer.WriteGeneric(context, 3, response, 2)

				response2 := &Packet00Layer{
					SendPingTime: mainLayer.SendPingTime + 10,
				}
				client.Writer.WriteGeneric(context, 0, response2, 2)
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

				client.Writer.WriteGeneric(context, 0x10, response, 2)
			} else if packetType == 0x90 {
				response := &Packet93Layer{
					UnknownBool1: true,
					UnknownBool2: true,
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

				client.Writer.WriteGeneric(context, 0x93, response, 3)
			} else if packetType == 0x82 {
                response := server.Dictionaries

                client.Writer.WriteGeneric(context, 0x82, response, 3)

                response2 := &Packet97Layer{*server.Schema}
                context.StaticSchema = server.Schema
                context.ESchemaParsed.Broadcast()

                client.Writer.WriteGeneric(context, 0x97, response2, 3)

                dataModel := &rbxfile.Root{}
                context.DataModel = dataModel

				var workspace *rbxfile.Instance
				var replicatedStorage *rbxfile.Instance

                initInstances := &Packet81Layer{
                    Items: make([]*Packet81LayerItem, len(services)),
                    Bools: [5]bool{true, false, false, false, true},
                    Int1: 0,
                    Int2: 0,
                    String1: []byte("RBX0123456789ABCDEF"),
                }
                for i, className := range services {
                    classID := context.StaticSchema.ClassesByName[className]
                    instance := &rbxfile.Instance{
                        ClassName: className,
                        Reference: strconv.Itoa(int(client.InstanceID)),
                        IsService: true,
                        Properties: make(map[string]rbxfile.Value),
                    }
                    client.InstanceID++

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
                    context.RefStringsByReferent[instance.Reference] = "RBX0123456789ABCDEF"
                }

                client.Writer.WriteGeneric(context, 0x81, initInstances, 3)

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

                client.Writer.WriteGeneric(context, 0x83, replicationResponse, 3)

				onlyWorkspaceJoinData := &Packet83_0B{make([]*rbxfile.Instance, 0)}
				InputObject := &rbxfile.Instance{
					ClassName: "InputObject",
					Reference: strconv.Itoa(int(client.InstanceID)),
					IsService: false,
					Properties: make(map[string]rbxfile.Value),
				}
				client.InstanceID++
				onlyWorkspaceJoinData.Instances = append(onlyWorkspaceJoinData.Instances, InputObject)
				context.RefStringsByReferent[InputObject.Reference] = "RBX0123456789ABCDEF"
				workspace.AddChild(InputObject)

				Explosion := &rbxfile.Instance{
					ClassName: "Explosion",
					Reference: strconv.Itoa(int(client.InstanceID)),
					IsService: false,
					Properties: make(map[string]rbxfile.Value),
				}
				client.InstanceID++
				onlyWorkspaceJoinData.Instances = append(onlyWorkspaceJoinData.Instances, Explosion)
				context.RefStringsByReferent[Explosion.Reference] = "RBX0123456789ABCDEF"
				workspace.AddChild(Explosion)

				ParabolaAdornment := &rbxfile.Instance{
					ClassName: "ParabolaAdornment",
					Reference: strconv.Itoa(int(client.InstanceID)),
					IsService: false,
					Properties: make(map[string]rbxfile.Value),
				}
				client.InstanceID++
				onlyWorkspaceJoinData.Instances = append(onlyWorkspaceJoinData.Instances, ParabolaAdornment)
				context.RefStringsByReferent[ParabolaAdornment.Reference] = "RBX0123456789ABCDEF"
				workspace.AddChild(ParabolaAdornment)

				Terrain := &rbxfile.Instance{
					ClassName: "Terrain",
					Reference: strconv.Itoa(int(client.InstanceID)),
					IsService: false,
					Properties: make(map[string]rbxfile.Value),
				}
				client.InstanceID++
				onlyWorkspaceJoinData.Instances = append(onlyWorkspaceJoinData.Instances, Terrain)
				context.RefStringsByReferent[Terrain.Reference] = "RBX0123456789ABCDEF"
				workspace.AddChild(Terrain)

				Status := &rbxfile.Instance{
					ClassName: "Status",
					Reference: strconv.Itoa(int(client.InstanceID)),
					IsService: false,
					Properties: make(map[string]rbxfile.Value),
				}
				client.InstanceID++
				onlyWorkspaceJoinData.Instances = append(onlyWorkspaceJoinData.Instances, Status)
				context.RefStringsByReferent[Status.Reference] = "RBX0123456789ABCDEF"
				workspace.AddChild(Status)

				PlayerGui := &rbxfile.Instance{
					ClassName: "PlayerGui",
					Reference: strconv.Itoa(int(client.InstanceID)),
					IsService: false,
					Properties: make(map[string]rbxfile.Value),
				}
				client.InstanceID++
				onlyWorkspaceJoinData.Instances = append(onlyWorkspaceJoinData.Instances, PlayerGui)
				context.RefStringsByReferent[PlayerGui.Reference] = "RBX0123456789ABCDEF"
				workspace.AddChild(PlayerGui)

				client.Writer.WriteGeneric(context, 0x83, &Packet83Layer{
					[]Packet83Subpacket{onlyWorkspaceJoinData},
				}, 3)

				allDefaultsJoinData := &Packet83_0B{make([]*rbxfile.Instance, 0, len(context.StaticSchema.Instances))}

				humanoid := &rbxfile.Instance{
					ClassName: "Humanoid",
					Reference: strconv.Itoa(int(client.InstanceID)),
					IsService: false,
					Properties: make(map[string]rbxfile.Value),
				}
				client.InstanceID++
				allDefaultsJoinData.Instances = append(allDefaultsJoinData.Instances, humanoid)
				context.RefStringsByReferent[humanoid.Reference] = "RBX0123456789ABCDEF"
				replicatedStorage.AddChild(humanoid)
				animator := &rbxfile.Instance{
					ClassName: "Animator",
					Reference: strconv.Itoa(int(client.InstanceID)),
					IsService: false,
					Properties: make(map[string]rbxfile.Value),
				}
				client.InstanceID++
				allDefaultsJoinData.Instances = append(allDefaultsJoinData.Instances, animator)
				context.RefStringsByReferent[animator.Reference] = "RBX0123456789ABCDEF"
				humanoid.AddChild(animator)

				part := &rbxfile.Instance{
					ClassName: "Part",
					Reference: strconv.Itoa(int(client.InstanceID)),
					IsService: false,
					Properties: make(map[string]rbxfile.Value),
				}
				client.InstanceID++
				allDefaultsJoinData.Instances = append(allDefaultsJoinData.Instances, part)
				context.RefStringsByReferent[part.Reference] = "RBX0123456789ABCDEF"
				replicatedStorage.AddChild(part)
				attachment := &rbxfile.Instance{
					ClassName: "Attachment",
					Reference: strconv.Itoa(int(client.InstanceID)),
					IsService: false,
					Properties: make(map[string]rbxfile.Value),
				}
				client.InstanceID++
				allDefaultsJoinData.Instances = append(allDefaultsJoinData.Instances, attachment)
				context.RefStringsByReferent[attachment.Reference] = "RBX0123456789ABCDEF"
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
                        Reference: strconv.Itoa(int(client.InstanceID)),
                        IsService: false,
                        Properties: make(map[string]rbxfile.Value),
                    }
                    client.InstanceID++

					allDefaultsJoinData.Instances = append(allDefaultsJoinData.Instances, instance)
                    context.RefStringsByReferent[instance.Reference] = "RBX0123456789ABCDEF"

					replicatedStorage.AddChild(instance)
				}

				client.Writer.WriteGeneric(context, 0x83, &Packet83Layer{
					[]Packet83Subpacket{allDefaultsJoinData},
				}, 3)
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
	packetWriter.OutputHandler = func(payload []byte) {
		num, err := server.Connection.WriteToUDP(payload, addr)
		if err != nil {
			fmt.Printf("Wrote %d bytes, err: %s", num, err.Error())
		}
	}

	client = &Client{
		Reader: packetReader,
		Writer: packetWriter,
		Context: context,
		Address: addr,
		Server: server,
        InstanceID: 1000,
	}

	ackTicker := time.NewTicker(17)
	go func() {
		for {
			<- ackTicker.C
			client.SendACKs()
		}
	}()
	return client
}

func StartServer(port uint16, dictionaries *Packet82Layer, schema *StaticSchema) error {
	server := &ServerPeer{Clients: make(map[string]*Client)}

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

		thisClient, ok := server.Clients[client.String()]
		if !ok {
			thisClient = newClient(client, server)
			server.Clients[client.String()] = thisClient
		}
		thisClient.Receive(buf[:n])
	}
}
