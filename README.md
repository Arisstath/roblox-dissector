# Sala
## The Essential Roblox Network Suite

[![GitHub Workflow Status](https://img.shields.io/github/workflow/status/Gskartwii/roblox-dissector/Go)](https://github.com/Gskartwii/roblox-dissector/actions)
[![Documentation](https://godoc.org/github.com/Gskartwii/roblox-dissector?status.svg)](https://godoc.org/github.com/Gskartwii/roblox-dissector)
[![Go Report](https://goreportcard.com/badge/github.com/Gskartwii/roblox-dissector)](https://goreportcard.com/report/github.com/Gskartwii/roblox-dissector)
[![Release v0.7.5beta](https://img.shields.io/badge/release-v0.7.5beta-blue.svg)](https://github.com/Gskartwii/roblox-dissector/releases)
[![Discord Chat](https://img.shields.io/discord/564392147038502912)](https://discord.gg/zPbprKb)

Sala is a suite of tools to aid developers, hackers\* and designers in understanding the internal workings of Roblox networking. Currently, Sala is in beta. Contributions are welcome. Check out the GitHub Wiki for documentation.

\* The word “hacker” is used here in the same sense as Atom is a “hackable” editor.

## Getting builds
Some releases are available under the [releases tab](https://github.com/Gskartwii/roblox-dissector/releases). Nightly builds can be found by going to Actions and picking the latest commit and downloading `windows-binary` from under "Artifacts". Instructions for compiling the project can be found on the Wiki.

## Features
* Read PCAP Files
* Capture packets on the fly
* View multiple capture sessions at a time
* Decode/encode most Roblox packets
* Dump DataModels based on capture (with some limitations)
    - Only replicated instances can be dumped
    - Locally available scripts are dumped as *.rbxc files. You need a script decompiler to view them.
* Capture in WinDivert proxy mode (experimental).
* [Versatile API](https://godoc.org/github.com/Gskartwii/roblox-dissector/peer)

## Screenshots
![Sala provides a offline interface for exploring PCAP files.](https://user-images.githubusercontent.com/6651822/55480485-0fc92880-5629-11e9-93eb-8431f85dd793.png)

![DataModel browser](https://user-images.githubusercontent.com/6651822/55480533-35563200-5629-11e9-9b7d-b5ed892a2ff0.png)

## About the name
_Sala_ < Finnish _salaisuus_ ‘secret’ (noun). The name refers to how Sala introduces its user to many obscure aspects of the Roblox network protocol.
