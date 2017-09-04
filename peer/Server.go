package peer
import "net"
import "fmt"
import "math/rand"
import "time"
import "sort"

type Client struct {
	Context *CommunicationContext
	Address *net.UDPAddr
	Reader *PacketReader
	Writer *PacketWriter
	Server *ServerPeer
	MustACK []int
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
	println("Sending acks")
	acks := client.MustACK
	client.MustACK = []int{}
	var ackStructure []ACKRange
	sort.Ints(acks)

	for _, ack := range acks {
		println("Must ack", ack)
		if len(ackStructure) == 0 {
			ackStructure = append(ackStructure, ACKRange{uint32(ack), uint32(ack)})
			continue
		}

		inserted := false
		for _, ackRange := range ackStructure {
			if int(ackRange.Max) == ack {
				inserted = true
			}
			if int(ackRange.Max + 1) == ack {
				ackRange.Max++
				inserted = true
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

				client.Writer.WriteGeneric(3, response, 2)

				response2 := &Packet00Layer{
					SendPingTime: mainLayer.SendPingTime + 10,
				}
				client.Writer.WriteGeneric(0, response2, 2)
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

				client.Writer.WriteGeneric(0x10, response, 2)
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

				client.Writer.WriteGeneric(0x93, response, 3)
			} else if packetType == 0x82 {
                response := server.Dictionaries

                client.Writer.WriteGeneric(0x82, response, 3)
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
		println("Write", payload[0])
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
