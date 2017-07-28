# Roblox UDP Dissector
This is a WIP tool that allows dissection of Roblox UDP protocol communication in PCAP files.  
To install this, you must first install [TheRecipe's Go bindings for Qt](github.com/therecipe/qt). 
Then fetch the repo and its dependencies: `go get -v github.com/gskartwii/roblox-dissector/...`  
And compile: `$GOPATH/bin/qtdeploy build $GOPATH/src/github.com/gskartwii/roblox-dissector`  
A directory named `deploy` should now exist in `$GOPATH/src/github/gskartwii/roblox-dissector`.

Usage:  
`roblox-dissector -name /path/to/dump.pcap [-ipv4] [-live interface_name] [-promisc]`  
The `name` option is required unless you want to a live capture. If you get a panic instantly after trying to start the program, also pass the `ipv4` option. It may be required for PCAP files produced by RawCap.

If you want to capture traffic to localhost on Windows, you must use RawCap and `tail`. Run the commands like so:

```
/path/to/rawcap.exe -f 127.0.0.1 /path/to/rawcap_output.pcap&
tail -f -c +1 -f /path/to/rawcap_output.pcap | /path/to/roblox-dissector.exe -name -
```

When using a live capture, the `name` option should not be passed. Instead, pass an interface name to `live`. To find interface names on Windows, follow this [guide](http://shad0wbq.blogspot.com/2006/06/windump-finding-pcap-device-mapping.html).

If you want to capture all traffic on your local network, including other devices, pass the `promisc` flag.

Yes, I promise I will improve the startup process by getting rid of the command line flags altogether. I'm also looking into adding a HTTPS traffic proxy which would be similar to Fiddler while working properly.

Code for 0x8A packets exists, but not publicly due to security reasons. This may change in the future.
