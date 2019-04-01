# Roblox UDP Dissector
This is a WIP tool that allows dissection of Roblox UDP protocol communication in PCAP files.  
Some releases are available under the [releases tab](https://github.com/Gskartwii/roblox-dissector/releases).  
To compile this, you must first install [TheRecipe's Go bindings for Qt](https://github.com/therecipe/qt). 
Then fetch the repo and its dependencies: `go get -v github.com/gskartwii/roblox-dissector/...`  
And compile: 

```
cd %GOPATH%/src/github.com/gskartwii/roblox-dissector
%GOPATH%/bin/qtdeploy build windows
```

A directory named `deploy` should now exist in `$GOPATH/src/github/gskartwii/roblox-dissector`.  
After the first compilation, you can pass the `-fast` flag to `qtdeploy` to speed up compilation, provided that you don't remove the `deploy` directory:

```
%GOPATH%/bin/qtdeploy -fast build windows
```