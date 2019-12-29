# Sala
## The Essential Roblox Network Suite

[![Build Status](https://travis-ci.org/Gskartwii/roblox-dissector.svg?branch=master)](https://travis-ci.org/Gskartwii/roblox-dissector)
[![Documentation](https://godoc.org/github.com/Gskartwii/roblox-dissector?status.svg)](https://godoc.org/github.com/Gskartwii/roblox-dissector)
[![Go Report](https://goreportcard.com/badge/github.com/Gskartwii/roblox-dissector)](https://goreportcard.com/report/github.com/Gskartwii/roblox-dissector)
[![Release v0.7.4beta](https://img.shields.io/badge/release-v0.7.4beta-blue.svg)](https://github.com/Gskartwii/roblox-dissector/releases)

[Discord Chat](https://discord.gg/zPbprKb)

Sala is a suite of tools to aid developers, hackers\* and designers in understanding the internal workings of Roblox networking. Currently, Sala development is entering the beta stage. Contributions are welcome.

\* The word “hacker” is used here in the same sense as Atom is a “hackable” editor.

## Getting builds
Some releases are available under the [releases tab](https://github.com/Gskartwii/roblox-dissector/releases).  
To compile Sala, you must first install [TheRecipe's Go bindings for Qt](https://github.com/therecipe/qt). Preferably, use Qt 5.12.2 and 64-bit MinGW version 7.3. 
Then fetch the repo and its dependencies: `go get -v github.com/Gskartwii/roblox-dissector/...`  
And compile: 

```
cd %GOPATH%/src/github.com/Gskartwii/roblox-dissector
%GOPATH%/bin/qtdeploy build windows
```

A directory named `deploy` should now exist in `$GOPATH/src/github/gskartwii/roblox-dissector`.  
After the first compilation, you can pass the `-fast` flag to `qtdeploy` to speed up compilation, provided that you don't remove the `deploy` directory:

```
%GOPATH%/bin/qtdeploy -fast build windows
```

If you have the right setup, you can also pass build tags to deploy. Use `-tags=lumikide` to compile with Lumikide support (if you don't know what it is, you won't need it) and `-tags=divert` to compile with WinDivert support.

## Features
* Read PCAP Files
* Capture packets on the fly
* View multiple capture sessions at a time
* Decode/encode most\* Roblox packets
* Dump DataModels based on capture (with some limitations)
    - Only replicated instances can be dumped
    - Locally available scripts are dumped as *.rbxc files. You need a script decompiler to view them.
* Capture in proxy mode: inject packets on the fly! (experimental)
* [Versatile API](https://godoc.org/github.com/Gskartwii/roblox-dissector/peer)

\* Packets not yet supported: StreamingEnabled, Smooth Terrain, certain anti-cheat packets (ID_ROCKY)
Packets for which encoding is not supported: memcheck hash packets

### Planned features
* StreamingEnabled support
* Smooth Terrain support
* Possibly ID_ROCKY and memcheck hash support
* Create and save PCAP files
* Support platforms other than Windows
* Custom Roblox server

## Screenshots
![Sala provides a offline interface for exploring PCAP files.](https://user-images.githubusercontent.com/6651822/55480485-0fc92880-5629-11e9-93eb-8431f85dd793.png)

![DataModel browser](https://user-images.githubusercontent.com/6651822/55480533-35563200-5629-11e9-9b7d-b5ed892a2ff0.png)

## About the name
_Sala_ < Finnish _salaisuus_ ‘secret’ (noun). The name refers to how Sala introduces its user to many obscure aspects of the Roblox network protocol.
