# Roblox UDP Dissector
This is a WIP tool that allows dissection of Roblox UDP protocol communication in PCAP files.  
Some releases are available under the [releases tab](https://github.com/Gskartwii/roblox-dissector/releases).  
To compile this, you must first install [TheRecipe's Go bindings for Qt](https://github.com/therecipe/qt). 
Then fetch the repo and its dependencies: `go get -v github.com/gskartwii/roblox-dissector/...`  
And compile: `%GOPATH%/bin/qtdeploy build %GOPATH%/src/github.com/gskartwii/roblox-dissector`  
A directory named `deploy` should now exist in `$GOPATH/src/github/gskartwii/roblox-dissector`.

---

Features not yet implemented:

* HTTPS proxy similar to Fiddler
* TeleportService support

```
-----BEGIN PGP SIGNED MESSAGE-----
Hash: SHA1

This software, roblox-dissector, was written by the Roblox user "gskw".
-----BEGIN PGP SIGNATURE-----
Version: GnuPG v1

iQEcBAEBAgAGBQJZx7fZAAoJEMMNCRxmuvnmBUMH/3yIzPedT1iVnYQuedEl1/9H
H9fLxSJb9H4WEE9bS10eDdKrb8XwUkLnY9tSaZawwNA3Ku1I47gn4+1KCuLp7V3I
q8zf8vvzBKxN8eQYz0q4tN87JzF6bmNA8wfv5qCZPAZ+GXc8bM4xKeRiB7+C3+yB
3I3e33oqAp+eS/0f/yj52bofzb0d7M7BdLvlkBQs+BbWZP4ShlnjfK8w864e2Xin
xxr8kqHetg6eKPckNvCIO1DdvAB7+k24lCjw3aqwp/YIKwVo+LP0yxsS4zq17HEo
0NKrEMeIhG0tr9Xqs5o8Kov9ieV9aP/JZ1UCEzswA/oXz7fNbfVfhzcfjGGUYgE=
=W5Wq
-----END PGP SIGNATURE-----
```

