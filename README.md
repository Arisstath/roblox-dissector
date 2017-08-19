# Roblox UDP Dissector
This is a WIP tool that allows dissection of Roblox UDP protocol communication in PCAP files.  
To install this, you must first install [TheRecipe's Go bindings for Qt](https://github.com/therecipe/qt). 
Then fetch the repo and its dependencies: `go get -v github.com/gskartwii/roblox-dissector/...`  
And compile: `$GOPATH/bin/qtdeploy build $GOPATH/src/github.com/gskartwii/roblox-dissector`  
A directory named `deploy` should now exist in `$GOPATH/src/github/gskartwii/roblox-dissector`.

Usage:  
`roblox-dissector -name /path/to/dump.pcap [-ipv4] [-live interface_name] [-promisc] [-instschema instances.txt -propschema properties.txt -eventschema events.txt]`  
The `name` option is required unless you want to a live capture. If you get a panic instantly after trying to start the program, also pass the `ipv4` option. It may be required for PCAP files produced by RawCap.  
You must also specify the instance, property and event schema if you want to capture traffic using the new static network schema (i.e. RobloxPlayer). To get the files, see [my official dumps](https://github.com/gskartwii/roblox-network-schema).

If you want to capture Roblox Studio traffic on Windows, you must make it flow through your router. Steps:

1. Run ipconfig to find your local IP address.
2. Run notepad.exe as an administrator.
3. Open `C:\\Windows\\System32\\drivers\\etc\\hosts`.
4. Add a line that forwards traffic for some domain to your local IP address, for example:

```
localme 192.168.1.4
```

5. Start a Roblox Studio server normally, but no clients.
6. Make `roblox-dissector` run a live capture (see below).
6. Run a client using the following command in `cmd` (replace :

```
%LOCALAPPDATA%\Roblox\Versions\version-CURRENTVERSION\RobloxStudioBeta.exe -ide -avatar -script "game:GetService'NetworkClient':PlayerConnect(0, 'localme', 53640)"
```

7. `roblox-dissector` should now be able to capture the traffic.

When using a live capture, the `name` option should not be passed. Instead, pass an interface name to `live`. To find interface names on Windows, follow this [guide](http://shad0wbq.blogspot.com/2006/06/windump-finding-pcap-device-mapping.html).

If you want to capture all traffic on your local network, including other devices, pass the `promisc` flag.

Yes, I promise I will improve the startup process by getting rid of the command line flags altogether. I'm also looking into adding a HTTPS traffic proxy which would be similar to Fiddler while working properly.

Code for 0x8A packets exists, but not publicly due to security reasons. This may change in the future.
