package main

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"regexp"

	"github.com/Gskartwii/roblox-dissector/peer"

	windivert "github.com/gskartwii/windivert-go"
	"github.com/olebedev/emitter"
)

type HTTPLayer struct {
	OriginalHost   string
	OriginalScheme string
	Request        *http.Request
	RequestBody    []byte
	Response       *http.Response
	ResponseBody   []byte
}

type HTTPConversation struct {
	Name              string
	ExpectingJoinAshx bool
	LayerEmitter      *emitter.Emitter
	ErrorEmitter      *emitter.Emitter
}

func (conv *HTTPConversation) Layers() *emitter.Emitter {
	return conv.LayerEmitter
}
func (conv *HTTPConversation) Errors() *emitter.Emitter {
	return conv.ErrorEmitter
}

type ProxiedPacket struct {
	Payload []byte
	Layers  *peer.PacketLayers
}

func (captureContext *CaptureContext) CaptureFromWinDivertProxy(ctx context.Context, realServerAddr string) error {
	dstAddr, err := net.ResolveUDPAddr("udp", realServerAddr)
	if err != nil {
		return err
	}

	filter := fmt.Sprintf("(ip.SrcAddr == %s and udp.SrcPort == %d) or (ip.DstAddr == %s and udp.DstPort == %d)",
		dstAddr.IP.String(), dstAddr.Port,
		dstAddr.IP.String(), dstAddr.Port)

	divertConnection, err := windivert.Open(filter, windivert.LayerNetwork, 405, 0)
	if err != nil {
		return err
	}
	var ifIdx, subIfIdx uint32

	proxyWriter := peer.NewProxyWriter(ctx)
	proxyWriter.ServerAddr = dstAddr
	proxyWriter.SecuritySettings = peer.Win10Settings()

	proxyWriter.ClientHalf.Output.On("udp", func(e *emitter.Event) { // writes TO client
		p := e.Args[0].([]byte)
		err := divertConnection.SendUDP(p, proxyWriter.ServerAddr, proxyWriter.ClientAddr, false, ifIdx, subIfIdx)
		if err != nil {
			fmt.Println("write fail to client %s/%d/%d: %s", proxyWriter.ClientAddr.String(), ifIdx, subIfIdx, err.Error())
			return
		}
	}, emitter.Void)
	proxyWriter.ServerHalf.Output.On("udp", func(e *emitter.Event) { // writes TO server
		p := e.Args[0].([]byte)
		err := divertConnection.SendUDP(p, proxyWriter.ClientAddr, proxyWriter.ServerAddr, true, ifIdx, subIfIdx)
		if err != nil {
			fmt.Println("write fail to server %d/%d: %s", ifIdx, subIfIdx, err.Error())
			return
		}
	}, emitter.Void)

	clientConversation := NewProviderConversation(proxyWriter.ClientHalf.DefaultPacketWriter, proxyWriter.ClientHalf.DefaultPacketReader)
	serverConversation := NewProviderConversation(proxyWriter.ServerHalf.DefaultPacketReader, proxyWriter.ServerHalf.DefaultPacketWriter)
	captureContext.AddConversation(clientConversation)
	captureContext.AddConversation(serverConversation)

	packetChan := make(chan ProxiedPacket, 100)

	go func() {
		var pktSrcAddr, pktDstAddr *net.UDPAddr
		var winDivertAddr *windivert.Address
		var err error
		var udpPayload []byte
		for {
			payload := make([]byte, 1500)
			winDivertAddr, _, err = divertConnection.Recv(payload)
			if err != nil {
				fmt.Printf("divert recv fail: %s\n", err.Error())
				return
			}
			ifIdx = winDivertAddr.InterfaceIndex
			subIfIdx = winDivertAddr.SubInterfaceIndex

			pktSrcAddr, pktDstAddr, udpPayload, err = windivert.ExtractUDP(payload)
			if err != nil {
				fmt.Printf("parse udp fail: %s\n", err.Error())
				return
			}

			// First packet must be from client
			if proxyWriter.ClientAddr == nil && pktDstAddr.String() == dstAddr.String() {
				// If so, assign the client to the src address
				proxyWriter.ClientAddr = pktSrcAddr
				clientConversation.ClientAddress = pktSrcAddr
				serverConversation.ClientAddress = pktSrcAddr
			}

			layers := &peer.PacketLayers{
				Root: peer.RootLayer{
					Source:      pktSrcAddr,
					Destination: pktDstAddr,
					// TODO: Can this be improved?
					FromClient: proxyWriter.ClientAddr.String() == pktSrcAddr.String(),
					FromServer: proxyWriter.ServerAddr.String() == pktSrcAddr.String(),
				},
			}
			if payload[0] > 0x8 {
				select {
				case packetChan <- ProxiedPacket{Layers: layers, Payload: udpPayload}:
				case <-ctx.Done():
					return
				}
			} else { // Need priority for join packets
				proxyWriter.ProxyClient(udpPayload, layers)
			}
		}
	}()
	for {
		select {
		case newPacket := <-packetChan:
			if newPacket.Layers.Root.FromClient { // from client? handled by client side
				proxyWriter.ProxyClient(newPacket.Payload, newPacket.Layers)
			} else {
				proxyWriter.ProxyServer(newPacket.Payload, newPacket.Layers)
			}
		case <-ctx.Done():
			divertConnection.Close()
			return nil
		}
	}
	return nil
}

func NewHTTPConversation(name string) *HTTPConversation {
	return &HTTPConversation{
		Name:              name,
		ExpectingJoinAshx: true,
		LayerEmitter:      emitter.New(0),
		ErrorEmitter:      emitter.New(0),
	}
}

func (conv *HTTPConversation) CaptureForWinDivert(ctx context.Context, captureCtx *CaptureContext, certFile string, keyFile string) {
	transport := &http.Transport{
		// FIXME: We must set InsecureSkipVerify because the host will be wrong
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	mux := http.NewServeMux()
	acceptNewJoinAshx := true

	mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		thisReq := &HTTPLayer{
			OriginalHost:   req.URL.Host,
			OriginalScheme: req.URL.Scheme,
			Request:        req,
		}

		req.URL.Host = "209.206.41.230"
		req.URL.Scheme = "https"

		// TODO: Bandwidth problem? I can't be bothered to ungzip.
		req.Header.Set("Accept-Encoding", "none")
		if req.URL.Path == "/Game/Join.ashx" || req.URL.Path == "/Game/PlaceLauncher.ashx" {
			// We must use a raw setter, because req.AddCookie() "sanitizes" the values in the wrong way
			req.Header.Set("Cookie", req.Header.Get("Cookie")+"; RBXAppDeviceIdentifier=AppDeviceIdentifier=ROBLOX UWP")
		}

		// TODO: Is storing all request bodies too heavy on memory?
		oldBody := req.Body
		requestBuf := bytes.NewBuffer(make([]byte, 0, req.ContentLength))
		requestTee := io.TeeReader(oldBody, requestBuf)

		req.Body = ioutil.NopCloser(requestTee)
		resp, err := transport.RoundTrip(req)
		if err != nil {
			conv.ErrorEmitter.Emit("err", err, thisReq)
			return
		}
		oldBody.Close()
		thisReq.RequestBody = requestBuf.Bytes()

		if resp.StatusCode == 403 { // CSRF check fail?
			req.Header.Set("X-Csrf-Token", resp.Header.Get("X-Csrf-Token"))
			resp, err = transport.RoundTrip(req)
			if err != nil {
				conv.ErrorEmitter.Emit("err", err, thisReq)
				return
			}
		}
		defer resp.Body.Close()

		for k, vv := range resp.Header {
			for _, v := range vv {
				w.Header().Add(k, v)
			}
		}
		w.WriteHeader(resp.StatusCode)
		thisReq.Response = resp

		if req.URL.Path == "/Game/Join.ashx" && acceptNewJoinAshx { // TODO: Stop using global variables for this
			acceptNewJoinAshx = false
			response, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				conv.ErrorEmitter.Emit("err", err, thisReq)
				return
			}
			thisReq.ResponseBody = response

			args := regexp.MustCompile(`MachineAddress":"(\d+.\d+.\d+.\d+)","ServerPort":(\d+)`).FindSubmatch(response)

			serverAddr := string(args[1]) + ":" + string(args[2])
			go func() {
				err := captureCtx.CaptureFromWinDivertProxy(ctx, serverAddr)
				if err != nil {
					println("windivert error: ", err.Error())
				}
			}()

			w.Write(response)
		} else {
			respBuf := bytes.NewBuffer(make([]byte, 0, resp.ContentLength))
			tee := io.TeeReader(resp.Body, respBuf)

			_, err := io.Copy(w, tee)
			if err != nil {
				conv.ErrorEmitter.Emit("err", err, thisReq)
				return
			}

			thisReq.ResponseBody = respBuf.Bytes()
		}
		<-conv.LayerEmitter.Emit("http", thisReq)
	})
	server := &http.Server{Addr: ":443", Handler: mux}

	// HTTP listener must run on its own thread!
	go func() {
		err := server.ListenAndServeTLS(certFile, keyFile)
		if err != nil {
			println("listen err:", err.Error())
			return
		}
	}()
	go func() {
		<-ctx.Done()
		println("closing proxy server")
		server.Close()
	}()
}

func (captureContext *CaptureContext) HookDivert() *HTTPConversation {
	conversation := NewHTTPConversation("divert-http")
	<-captureContext.ConversationEmitter.Emit("http", conversation)

	return conversation
}

func (session *CaptureSession) CaptureFromDivert(certFile string, keyFile string) {
	conv := session.CaptureContext.HookDivert()

	conv.CaptureForWinDivert(session.Context, session.CaptureContext, certFile, keyFile)
}
