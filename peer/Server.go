package peer
import "net"
import "fmt"
import "math/rand"

type Client struct {
	Context *CommunicationContext
	Address *net.UDPAddr
	Reader *PacketReader
	Writer *PacketWriter
	Server *ServerPeer
}

type ServerPeer struct {
	Connection *net.UDPConn
	Clients map[string]*Client
	Address *net.UDPAddr
	GUID uint64
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
		},
		FullReliableHandler: func(packetType byte, packet *UDPPacket, layers *PacketLayers) {
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

	return client
}

func StartServer(port uint16) error {
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
