package peer
import "time"
import "net"
import "fmt"
import "sort"

type JoinAshxResponse struct {
	NewClientTicket string
	SessionId string
}

type CustomClient struct {
	Context *CommunicationContext
	Reader *PacketReader
	Writer *PacketWriter
	Address net.UDPAddr
	ServerAddress net.UDPAddr
	Connected bool
	MustACK []int
}

func (client *CustomClient) SendACKs() {
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

	client.Writer.WriteRakNet(result, &client.ServerAddress)
}

func (client *CustomClient) Receive(buf []byte) {
	packet := &UDPPacket{
		Stream: BufferToStream(buf),
		Source: client.ServerAddress,
		Destination: client.Address,
	}
	client.Reader.ReadPacket(buf, packet)
}

func StartClient(addr net.UDPAddr) (*CustomClient, error) {
	context := NewCommunicationContext()
	client := &CustomClient{}
	client.Context = context
	client.ServerAddress = addr

	packetReader := NewPacketReader()
	packetReader.ACKHandler = func(packet *UDPPacket, layer *RakNetLayer) {
		println("received ack")
	}
	packetReader.ReliabilityLayerHandler = func(p *UDPPacket, re *ReliabilityLayer, ra *RakNetLayer) {
		println("receive reliabilitylayer")
	}
	packetReader.SimpleHandler = func(packetType byte, packet *UDPPacket, layers *PacketLayers) {
		if packetType == 0x6 {
			println("receive 6!")
			client.Connected = true

			response := &Packet07Layer{
				GUID: 0x1122334455667788, // hahaha
				MTU: 1492,
				IPAddress: &addr,
			}
			client.Writer.WriteSimple(7, response, &addr)
		} else if packetType == 0x8 {
			println("receive 8!")

			response := &Packet09Layer{
				GUID: 0x1122334455667788,
				Timestamp: 117,
				UseSecurity: false,
				Password: []byte{/*0x37, 0x4F, */0x5E, 0x11, /*0x6C, 0x45*/},
			}
			client.Writer.WriteGeneric(context, 9, response, 2, &addr)
		} else {
			println("receive simple unk", packetType)
		}
	}
	packetReader.ReliableHandler = func(packetType byte, packet *UDPPacket, layers *PacketLayers) {
		rakNetLayer := layers.RakNet
		client.MustACK = append(client.MustACK, int(rakNetLayer.DatagramNumber))
	}
	packetReader.FullReliableHandler = func(packetType byte, packet *UDPPacket, layers *PacketLayers) {
		if packetType == 0x0 {
			mainLayer := layers.Main.(*Packet00Layer)
			response := &Packet03Layer{
				SendPingTime: mainLayer.SendPingTime,
				SendPongTime: mainLayer.SendPingTime + 10,
			}

			client.Writer.WriteGeneric(context, 3, response, 2, &addr)
		} else if packetType == 0x10 {
			println("receive 10!")
			mainLayer := layers.Main.(*Packet10Layer)
			nullIP, _ := net.ResolveUDPAddr("udp", "255.255.255.255:0")
			response := &Packet13Layer{
				IPAddress: mainLayer.IPAddress,
				Addresses: [10]*net.UDPAddr{
					&client.Address,
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
				SendPingTime: mainLayer.SendPingTime,
				SendPongTime: mainLayer.SendPingTime + 10,
			}

			client.Writer.WriteGeneric(context, 3, response, 2, &addr)

			response8A := &Packet8ALayer{
				PlayerId: 0,
				ClientTicket: []byte(""),
				String2: []byte(""),
				Int2: 36,
				SecurityHash: []byte(""),
				Platform: []byte("Win32"),
				String5: []byte("?"),
				SessionId: []byte(""),
				Int3: 5555555,
			}
			client.Writer.WriteGeneric(context, 3, response8A, 2, &addr)
		} else {
			println("receive generic unk", packetType)
		}
	}
	packetReader.ErrorHandler = func(err error) {
		println(err.Error())
	}
	packetReader.Context = context
	client.Reader = packetReader

	conn, err := net.DialUDP("udp", nil, &addr)
	defer conn.Close()
	if err != nil {
		return client, err
	}
	client.Address = *conn.LocalAddr().(*net.UDPAddr)

	packetWriter := NewPacketWriter()
	packetWriter.ErrorHandler = func(err error) {
		println(err.Error())
	}
	packetWriter.OutputHandler = func(payload []byte, dest *net.UDPAddr) {
		num, err := conn.Write(payload)
		if err != nil {
			fmt.Printf("Wrote %d bytes, err: %s\n", num, err.Error())
		}
	}
	client.Writer = packetWriter

	connreqpacket := &Packet05Layer{ProtocolVersion: 5}

	go func() {
		for i := 0; i < 5; i++ {
			if client.Connected {
				println("successfully dialed")
				return
			}
			client.Writer.WriteSimple(5, connreqpacket, &addr)
			time.Sleep(5)
		}
		println("dial failed after 5 attempts")
	}()
	ackTicker := time.NewTicker(17)
	go func() {
		for {
			<- ackTicker.C
			client.SendACKs()
		}
	}()

	buf := make([]byte, 1492)
	for {
		n, _, err := conn.ReadFromUDP(buf)
		if err != nil {
			println("Read err:", err.Error())
			continue
		}

		client.Receive(buf[:n])
	}
}
