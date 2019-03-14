package main

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/tls"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/Gskartwii/roblox-dissector/peer"
)

// Requires you to patch the player with memcheck bypass and rbxsig ignoring! But it could work...
func captureFromPlayerProxy(settings *PlayerProxySettings, captureJobContext context.Context, injectPacket chan peer.RakNetPacket, packetViewer *MyPacketListView, commContext *peer.CommunicationContext) {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	mux := http.NewServeMux()
	acceptNewJoinAshx := true
	proxyContext, proxyContextCancel := context.WithCancel(captureJobContext)

	mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		req.URL.Host = "209.206.41.230"
		req.URL.Scheme = "https"
		if req.URL.Path == "/Game/Join.ashx" {
			println("patching join.ashx for gzip encoding")
			req.Header.Set("Accept-Encoding", "none")
		}

		if req.URL.Path == "/Game/Join.ashx" {
			println("patching join.ashx gzip")
			req.Header.Set("Accept-Encoding", "none")
		}
		resp, err := transport.RoundTrip(req)
		//fmt.Printf("Request: %s/%s %s %v\n%s %v\n", req.Host, req.URL.String(), req.Method, req.Header, resp.Status, resp.Header)
		if err != nil {
			println("error:", err.Error())
			return
		}
		if resp.StatusCode == 403 { // CSRF check fail?
			req.Header.Set("X-Csrf-Token", resp.Header.Get("X-Csrf-Token"))
			println("Set csrftoken:", resp.Header.Get("X-Csrf-Token"))
			println("retrying")
			resp, err = transport.RoundTrip(req)
			//fmt.Printf("Request: %s/%s %s %v\n%s %v\n", req.Host, req.URL.String(), req.Method, req.Header, resp.Status, resp.Header)
			if err != nil {
				println("error:", err.Error())
				return
			}
		}
		defer resp.Body.Close()

		for k, vv := range resp.Header {
			for _, v := range vv {
				w.Header().Add(k, v)
			}
		}

		if req.URL.Path == "/Game/Join.ashx" && acceptNewJoinAshx {
			acceptNewJoinAshx = false
			w.Header().Set("Content-Encoding", "gzip")
			response, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				println("joinashx err:", err.Error())
				return
			}

			newBuffer := bytes.NewBuffer(make([]byte, 0, len(response)))
			result := regexp.MustCompile(`MachineAddress":"\d+.\d+.\d+.\d+","ServerPort":\d+`).ReplaceAll(response, []byte(`MachineAddress":"127.0.0.1","ServerPort":53640`))
			args := regexp.MustCompile(`MachineAddress":"(\d+.\d+.\d+.\d+)","ServerPort":(\d+)`).FindSubmatch(response)
			//println("joinashx response:", string(result))

			serverAddr := string(args[1]) + ":" + string(args[2])
			go captureFromInjectionProxy("127.0.0.1:53640", serverAddr, proxyContext, injectPacket, packetViewer, commContext)

			compressStream := gzip.NewWriter(newBuffer)
			_, err = compressStream.Write(result)
			if err != nil {
				println("joinashx gz w err:", err.Error())
				return
			}
			err = compressStream.Close()
			if err != nil {
				println("joinashx gz close err:", err.Error())
				return
			}
			w.Header().Set("Content-Length", strconv.Itoa(newBuffer.Len()))
			w.WriteHeader(resp.StatusCode)

			w.Write(newBuffer.Bytes())
		} else {
			//println("dumping to", "dumps/" + strings.Replace(req.URL.Path, "/", "_", -1))
			dumpfile, err := os.Create("dumps/" + strings.Replace(req.URL.Path, "/", "_", -1))
			if err != nil {
				println("fail:", err.Error())
				return
			}
			defer dumpfile.Close()
			tee := io.TeeReader(resp.Body, dumpfile)

			io.Copy(w, tee)
		}
	})
	server := &http.Server{Addr: ":443", Handler: mux}

	// HTTP listener must run on its own thread!
	go func() {
		err := server.ListenAndServeTLS(settings.Certfile, settings.Keyfile)
		if err != nil {
			println("listen err:", err.Error())
			return
		}
	}()
	go func() {
		<-captureJobContext.Done()
		println("closing proxy server")
		server.Close()
		proxyContextCancel()
	}()
}
