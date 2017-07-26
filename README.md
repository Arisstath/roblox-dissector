# Roblox UDP Dissector
This is a WIP tool that allows dissection of Roblox UDP protocol communication in PCAP files.  
To install this, you must first install [TheRecipe's Go bindings for Qt](github.com/therecipe/qt). 
Then fetch the repo and its dependencies: `go get -v github.com/gskartwii/roblox-dissector/...`  
And compile: `$GOPATH/bin/qtdeploy build $GOPATH/src/github.com/gskartwii/roblox-dissector`  
A directory named `deploy` should now exist in `$GOPATH/src/github/gskartwii/roblox-dissector`.

Usage:  
`roblox-dissector -name /path/to/dump.pcap [-ipv4]`  
The `name` option is required. If you get a panic instantly after trying to start the program, also pass the `ipv4` option. It may be required for PCAP files produced by WinPcap.

Yes, I promise I will improve the startup process by getting rid of the command line flags altogether. I'm also looking into adding a HTTPS traffic proxy which would be similar to Fiddler while working properly.
