package peer

import (
	"context"
	"net"
	"time"

	"github.com/gskartwii/rbxfile"
)

// ProxyHalf describes a proxy connection to a connected peer.
type ProxyHalf struct {
	*ConnectedPeer
	fakePackets []uint32
}

func NewProxyHalf(context *CommunicationContext) *ProxyHalf {
    return &ProxyHalf{
        ConnectedPeer: NewConnectedPeer(context),
        fakePackets: nil,
    }
}

func (w *ProxyHalf) rotateDN(old uint32) uint32 {
	for i := len(w.fakePackets) - 1; i >= 0; i-- {
		fakepacket := w.fakePackets[i]
		if old >= fakepacket {
			old++
		}
	}
	return old
}
func (w *ProxyHalf) rotateACK(ack ACKRange) (bool, ACKRange) {
	fakepackets := w.fakePackets
	for i := len(fakepackets) - 1; i >= 0; i-- {
		fakepacket := fakepackets[i]
		if ack.Max >= fakepacket {
			ack.Max--
		}
		if ack.Min > fakepacket {
			ack.Min--
		}
	}
	return ack.Min > ack.Max, ack
}
func (w *ProxyHalf) rotateACKs(acks []ACKRange) (bool, []ACKRange) {
	newacks := make([]ACKRange, 0, len(acks))
	for i := 0; i < len(acks); i++ {
		dropthis, newack := w.rotateACK(acks[i])
		if !dropthis {
			newacks = append(newacks, newack)
		}
	}
	return len(newacks) == 0, newacks
}

// ProxyWriter describes a proxy that connects two peers.
// ProxyWriters have injection capabilities.
type ProxyWriter struct {
	// ClientHalf only does communications with the client
	// ClientHalf receives from client, ClientHalf sends to client
	ClientHalf *ProxyHalf
	// The above also applies to ServerHalf
	ServerHalf *ProxyHalf
	ClientAddr *net.UDPAddr
	ServerAddr *net.UDPAddr

	SecuritySettings SecuritySettings
	RuntimeContext   context.Context
	CancelFunc       context.CancelFunc

	ackTicker *time.Ticker
}

func (writer *ProxyWriter) startAcker() {
	writer.ackTicker = time.NewTicker(16 * time.Millisecond)
	go func() {
		for {
			select {
			case <-writer.ackTicker.C:
				writer.ClientHalf.sendACKs()
				writer.ServerHalf.sendACKs()
			case <-writer.RuntimeContext.Done():
				return
			}
		}
	}()

}

// NewProxyWriter creates and initializes a new ProxyWriter
func NewProxyWriter(context *CommunicationContext) *ProxyWriter {
	writer := &ProxyWriter{}
	clientHalf := NewProxyHalf(context)
	serverHalf := NewProxyHalf(context)

	clientHalf.SimpleHandler = func(packetType byte, layers *PacketLayers) {
		if layers.Error != nil {
			println("client error: ", layers.Error.Error())
			return
		}
		println("client simple", packetType)
		if packetType == 5 {
			println("recv 5, protocol type", layers.Main.(*Packet05Layer).ProtocolVersion)
		}
		serverHalf.WriteSimple(layers.Main.(RakNetPacket))
	}
	serverHalf.SimpleHandler = func(packetType byte, layers *PacketLayers) {
		if layers.Error != nil {
			println("server error: ", layers.Error.Error())
			return
		}
		println("server simple", packetType)
		clientHalf.WriteSimple(layers.Main.(RakNetPacket))
	}

	clientHalf.ReliabilityLayerHandler = func(layers *PacketLayers) {
		if layers.Error != nil {
			println("client error: ", layers.Error.Error())
			return
		}
		clientHalf.mustACK = append(clientHalf.mustACK, int(layers.RakNet.DatagramNumber))
	}
	serverHalf.ReliabilityLayerHandler = func(layers *PacketLayers) {
		if layers.Error != nil {
			println("server error: ", layers.Error.Error())
			return
		}
		serverHalf.mustACK = append(serverHalf.mustACK, int(layers.RakNet.DatagramNumber))
	}

	clientHalf.FullReliableHandler = func(packetType byte, layers *PacketLayers) {
		// FIXME: No streaming support
		//println("client fullreliable", packetType)
		var err error
		if packetType == 0x15 {
			println("Disconnected by client!!")
			writer.CancelFunc()
			return
		}
		if layers.Error != nil {
			println("client error: ", layers.Error.Error())
			return
		}
		if layers.Main == nil || (packetType != 0x83 && packetType != 0x8A && packetType != 0x85 && packetType != 0x86) {
			relPacket := layers.Reliability
			// packets that fail to parse: pass through untouched
			// FIXME: this may prove problematic
			//println("client sent reliable, serverHalf writing", packetType, packet.Source.String(), packet.Destination.String())
			err = serverHalf.WriteReliablePacket(
				relPacket.SplitBuffer.data,
				relPacket,
			)
			if err != nil {
				println(err.Error())
			}
			return
		}
		switch packetType {
		case 0x83:
			mainLayer := layers.Main.(*Packet83Layer)
			modifiedSubpackets := mainLayer.SubPackets[:0] // in case packets need to be dropped
			for _, subpacket := range mainLayer.SubPackets {
				switch subpacket.(type) {
				case *Packet83_02:
					instPacket := subpacket.(*Packet83_02)
					println("patching osplatform", instPacket.Child.Name())
					if instPacket.Child.ClassName == "Player" {
						// patch OsPlatform!
						instPacket.Child.Properties["OsPlatform"] = rbxfile.ValueString(writer.SecuritySettings.OsPlatform)
					}
					modifiedSubpackets = append(modifiedSubpackets, subpacket)
				case *Packet83_09:
					// patch id response
					println("patching id resp")
					pmcPacket := subpacket.(*Packet83_09)
					if pmcPacket.SubpacketType == 6 {
						pmcSubpacket := pmcPacket.Subpacket.(*Packet83_09_06)
						pmcSubpacket.Int2 = writer.SecuritySettings.IdChallengeResponse - pmcSubpacket.Int1
						modifiedSubpackets = append(modifiedSubpackets, subpacket)
					} // if not type 6, drop it!
				case *Packet83_12:
					println("permanently dropping hash packet")
					// IMPORTANT! We don't drop the entire hash packet 0x83 containers!
					// Under heavy stress, the Roblox client may pack everything inside the container,
					// including hash packets.
					// It used to seem that this was not the case, but I was proven wrong.
				case *Packet83_05:
					pingPacket := subpacket.(*Packet83_05)
					pingPacket.SendStats = 0
					pingPacket.ExtraStats = 0
					modifiedSubpackets = append(modifiedSubpackets, subpacket)
				case *Packet83_06:
					pingPacket := subpacket.(*Packet83_06)
					pingPacket.SendStats = 0
					pingPacket.ExtraStats = 0
					modifiedSubpackets = append(modifiedSubpackets, subpacket)
				default:
					modifiedSubpackets = append(modifiedSubpackets, subpacket)
				}
			}
			mainLayer.SubPackets = modifiedSubpackets

			_, err = serverHalf.WritePacket(mainLayer)
		case 0x8A:
			mainLayer := layers.Main.(*Packet8ALayer)
			mainLayer.DataModelHash = writer.SecuritySettings.DataModelHash
			mainLayer.SecurityKey = writer.SecuritySettings.SecurityKey
			mainLayer.Platform = writer.SecuritySettings.OsPlatform
			mainLayer.GoldenHash = writer.SecuritySettings.GoldenHash

			_, err = serverHalf.WritePacket(mainLayer)
		case 0x85:
			mainLayer := layers.Main.(*Packet85Layer)
			_, err = serverHalf.WritePhysics(layers.Timestamp, mainLayer)
		case 0x86:
			mainLayer := layers.Main.(*Packet86Layer)
			_, err = serverHalf.WritePacket(mainLayer)
		case 0x87:
			mainLayer := layers.Main.(*Packet87Layer)
			_, err = serverHalf.WritePacket(mainLayer)
		}
		if err != nil {
			println("client error:", err.Error())
		}
	}
	serverHalf.FullReliableHandler = func(packetType byte, layers *PacketLayers) {
		if layers.Error != nil {
			println("server error: ", layers.Error.Error())
			return
		}
		relPacket := layers.Reliability
		//println("server sent reliable, clientHalf writing", packetType, packet.Source.String(), packet.Destination.String())
		clientHalf.WriteReliablePacket(
			relPacket.SplitBuffer.data,
			relPacket,
		)

		if packetType == 0x15 {
			println("Disconnected by server!!")
			writer.CancelFunc()
		}
	}
	clientHalf.ReliableHandler = func(packetType byte, layers *PacketLayers) {
		if layers.Error != nil {
			println("client error: ", layers.Error.Error())
			return
		}
	}
	serverHalf.ReliableHandler = func(packetType byte, layers *PacketLayers) {
		if layers.Error != nil {
			println("server error: ", layers.Error.Error())
			return
		}
	}

	clientHalf.ACKHandler = func(layers *PacketLayers) {
		/*drop, newacks := serverHalf.rotateACKs(layer.ACKs)
		if !drop {
			layer.ACKs = newacks
			serverHalf.Writer.WriteRakNet(layer, writer.ServerAddr)
		}*/
	}
	serverHalf.ACKHandler = func(layers *PacketLayers) {
		/*drop, newacks := clientHalf.rotateACKs(layer.ACKs)
		if !drop {
			layer.ACKs = newacks
			clientHalf.Writer.WriteRakNet(layer, writer.ClientAddr)
		}*/
	}

	clientHalf.DefaultPacketWriter.SetToClient(true) // writes TO client!
	clientHalf.DefaultPacketReader.SetIsClient(true) // reads FROM client!
    // Caches will have been assigned by NewConnectedPeer()

	writer.ClientHalf = clientHalf
	writer.ServerHalf = serverHalf

	writer.startAcker()

	return writer
}

// ProxyClient should be called when the client sends a packet.
func (writer *ProxyWriter) ProxyClient(payload []byte, layers *PacketLayers) {
	writer.ClientHalf.ReadPacket(payload, layers)
}

// ProxyServer should be called when the server sends a packet.
func (writer *ProxyWriter) ProxyServer(payload []byte, layers *PacketLayers) {
	writer.ServerHalf.ReadPacket(payload, layers)
}

// InjectServer should be called when an injected packet should be sent to
// the server. [WIP]
/*func (writer *ProxyWriter) InjectServer(packet RakNetPacket) {
	olddn := writer.ServerHalf.Writer.datagramNumber
	writer.ServerHalf.Writer.WriteGeneric(
		writer.ServerHalf.Reader.Context,
		0x83,
		packet,
		0,
		writer.ServerAddr,
	) // Unreliable packets, might improve this sometime
	for i := olddn; i < writer.ServerHalf.Writer.datagramNumber; i++ {
		println("adding fakepacket", i)
		writer.ServerHalf.fakePackets = append(writer.ServerHalf.fakePackets, i)
	}
}
*/
