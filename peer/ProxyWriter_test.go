package peer

import (
	"bytes"
	"context"
	"fmt"
	"net"

	"github.com/olebedev/emitter"
)

// ExampleProxyWriter provides an example on how to use the ProxyWriter struct.
func ExampleProxyWriter() {
	clientAddr, _ := net.ResolveUDPAddr("udp", "127.0.0.1:53640")
	serverAddr, _ := net.ResolveUDPAddr("udp", "30.40.50.60:50000")
	proxy := NewProxyWriter(context.TODO())
	proxy.ClientAddr = clientAddr
	proxy.ServerAddr = serverAddr

	proxy.ClientHalf.Output.On("udp", func(e *emitter.Event) {
		payload := e.Args[0].([]byte)
		// the proxy is requesting us to write this payload to the client
		fmt.Printf("Write %X to client (%s)\n", payload, proxy.ClientAddr)
	})
	proxy.ServerHalf.Output.On("udp", func(e *emitter.Event) {
		payload := e.Args[0].([]byte)
		// the proxy is requesting us to write this payload to the server
		fmt.Printf("Write %X to server (%s)\n", payload, proxy.ServerAddr)
	})

	var clientHandshake bytes.Buffer

	// write packet id and offline message id
	clientHandshake.WriteByte(0x05)
	clientHandshake.Write(OfflineMessageID)
	packet := &Packet05Layer{
		ProtocolVersion:  5,
		MTUPaddingLength: 10,
	}
	packet.Serialize(nil, &extendedWriter{&clientHandshake}) // pretend we have a packet

	// ProxyClient should be called for packets coming in from the client
	proxy.ProxyClient(clientHandshake.Bytes(), &PacketLayers{
		Root: RootLayer{
			FromClient:  true,
			Source:      clientAddr,
			Destination: serverAddr,
		},
	})
	// Output: Write 0500FFFF00FEFEFEFEFDFDFDFD123456780500000000000000000000 to server (30.40.50.60:50000)
}
