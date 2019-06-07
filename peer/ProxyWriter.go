package peer

import (
	"context"
	"net"
	"time"

	"github.com/olebedev/emitter"
	"github.com/robloxapi/rbxfile"
)

// ProxyHalf describes a proxy connection to a connected peer.
type ProxyHalf struct {
	*ConnectedPeer
	fakePackets []uint32
}

// NewProxyHalf initializes a new ProxyHalf
func NewProxyHalf(context *CommunicationContext, withClient bool) *ProxyHalf {
	return &ProxyHalf{
		ConnectedPeer: NewConnectedPeer(context, withClient),
		fakePackets:   nil,
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

	SecuritySettings SecurityHandler
	RuntimeContext   context.Context

	ackTicker *time.Ticker
}

func (writer *ProxyWriter) startAcker() {
	writer.ackTicker = time.NewTicker(16 * time.Millisecond)
	go func() {
		for {
			select {
			case <-writer.ackTicker.C:
				err := writer.ClientHalf.sendACKs()
				if err != nil {
					println("client ack error", err.Error())
				}
				err = writer.ServerHalf.sendACKs()
				if err != nil {
					println("server ack error", err.Error())
				}
			case <-writer.RuntimeContext.Done():
				return
			}
		}
	}()

}

// NewProxyWriter creates and initializes a new ProxyWriter
func NewProxyWriter(ctx context.Context) *ProxyWriter {
	context := NewCommunicationContext()
	writer := &ProxyWriter{
		RuntimeContext: ctx,
	}
	clientHalf := NewProxyHalf(context, true)
	serverHalf := NewProxyHalf(context, false)

	// Set FromServer/Client appropriately
	clientHalf.DefaultPacketWriter.LayerEmitter.On("*", func(e *emitter.Event) {
		e.Args[0].(*PacketLayers).Root = RootLayer{
			FromServer:  true,
			Logger:      nil,
			Source:      writer.ServerAddr,
			Destination: writer.ClientAddr,
		}
	}, emitter.Void)
	serverHalf.DefaultPacketWriter.LayerEmitter.On("*", func(e *emitter.Event) {
		e.Args[0].(*PacketLayers).Root = RootLayer{
			FromClient:  true,
			Logger:      nil,
			Source:      writer.ServerAddr,
			Destination: writer.ClientAddr,
		}
	}, emitter.Void)

	clientHalf.DefaultPacketReader.LayerEmitter.On("offline", func(e *emitter.Event) {
		layers := e.Args[0].(*PacketLayers)
		println("client offline", layers.PacketType)
		if layers.PacketType == 5 {
			println("recv 5, protocol type", layers.Main.(*Packet05Layer).ProtocolVersion)
		}
		serverHalf.WriteOffline(layers.Main)
	}, emitter.Void)
	serverHalf.DefaultPacketReader.LayerEmitter.On("offline", func(e *emitter.Event) {
		layers := e.Args[0].(*PacketLayers)
		println("server offline", layers.PacketType)
		clientHalf.WriteOffline(layers.Main)
	}, emitter.Void)

	clientHalf.DefaultPacketReader.LayerEmitter.On("reliability", func(e *emitter.Event) {
		layers := e.Args[0].(*PacketLayers)
		clientHalf.mustACK = append(clientHalf.mustACK, int(layers.RakNet.DatagramNumber))
	}, emitter.Void)
	serverHalf.DefaultPacketReader.LayerEmitter.On("reliability", func(e *emitter.Event) {
		layers := e.Args[0].(*PacketLayers)
		serverHalf.mustACK = append(serverHalf.mustACK, int(layers.RakNet.DatagramNumber))
	}, emitter.Void)

	clientHalf.DefaultPacketReader.LayerEmitter.On("full-reliable", func(e *emitter.Event) {
		layers := e.Args[0].(*PacketLayers)
		packetType := layers.PacketType
		// FIXME: No streaming support
		//println("client fullreliable", packetType)
		var err error
		if packetType == 0x15 {
			println("Disconnected by client!!")
			return
		}
		if layers.Error != nil {
			println("client error: ", layers.Error.Error())
			return
		}
		if layers.Main == nil {
			println("Dropping unknown packettype", packetType)
			return
		}
		switch packetType {
		case 0x83:
			mainLayer := layers.Main.(*Packet83Layer)
			modifiedSubpackets := mainLayer.SubPackets[:0] // in case packets need to be dropped
			for _, subpacket := range mainLayer.SubPackets {
				switch subpacket.(type) {
				// this packet will be sent if a debugger is detected
				// we may want to debug the app, so hence we drop it
				case *Packet83_07:
					evtPacket := subpacket.(*Packet83_07)
					if evtPacket.Schema.Name == "StatsAvailable" {
						println("(RobloxApp has detected a hacker) permanently dropping statspacket, codename = ", evtPacket.Event.Arguments[0].String())
					} else {
						modifiedSubpackets = append(modifiedSubpackets, subpacket)
					}
				case *Packet83_02:
					instPacket := subpacket.(*Packet83_02)
					if instPacket.Schema.Name == "Player" {
						println("patching osplatform", instPacket.Instance.Name())
						// patch OSPlatform!
						instPacket.Instance.Set("OsPlatform", rbxfile.ValueString(writer.SecuritySettings.OSPlatform()))
					}
					modifiedSubpackets = append(modifiedSubpackets, subpacket)
				case *Packet83_09:
					// patch id response
					println("patching id resp")
					pmcPacket := subpacket.(*Packet83_09)
					switch pmcPacket.Subpacket.(type) {
					case *Packet83_09_06:
						pmcSubpacket := pmcPacket.Subpacket.(*Packet83_09_06)
						pmcSubpacket.Response = writer.SecuritySettings.GenerateIDResponse(pmcSubpacket.Challenge)
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

			err = serverHalf.WritePacket(mainLayer)
		case 0x8A:
			mainLayer := layers.Main.(*Packet8ALayer)
			writer.SecuritySettings.PatchTicketPacket(mainLayer)

			err = serverHalf.WritePacket(mainLayer)
		case 0x85:
			mainLayer := layers.Main.(*Packet85Layer)
			err = serverHalf.WriteTimestamped(layers.Timestamp, mainLayer)
		default:
			println("passthrough packet: ", packetType)
			err = serverHalf.WritePacket(layers.Main.(RakNetPacket))
		}
		if err != nil {
			println("client error:", err.Error())
		}
	}, emitter.Void)

	serverHalf.DefaultPacketReader.LayerEmitter.On("full-reliable", func(e *emitter.Event) {
		layers := e.Args[0].(*PacketLayers)
		packetType := layers.PacketType
		if layers.Error != nil {
			println("server error: ", layers.Error.Error())
			return
		}
		if layers.Main == nil {
			println("dropping nil packet??", packetType)
			return
		}
		err := clientHalf.WritePacket(layers.Main.(RakNetPacket))
		if err != nil {
			println("server serialize error: ", err.Error())
			return
		}

		if packetType == 0x15 {
			println("Disconnected by server!!")
		}
	}, emitter.Void)
	clientHalf.DefaultPacketReader.ErrorEmitter.On("*", func(e *emitter.Event) {
		println("client error on topic", e.OriginalTopic+":", e.Args[0].(*PacketLayers).Error.Error())
	}, emitter.Void)
	serverHalf.DefaultPacketReader.ErrorEmitter.On("*", func(e *emitter.Event) {
		println("server error on topic", e.OriginalTopic+":", e.Args[0].(*PacketLayers).Error.Error())
	}, emitter.Void)
	// nop ack handler

	// bind default packet handlers so the DataModel is updated accordingly
	clientHalf.BindDataModelHandlers()
	serverHalf.BindDataModelHandlers()

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
