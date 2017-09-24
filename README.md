# Roblox UDP Dissector
This is a WIP tool that allows dissection of Roblox UDP protocol communication in PCAP files.  
To install this, you must first install [TheRecipe's Go bindings for Qt](https://github.com/therecipe/qt). 
Then fetch the repo and its dependencies: `go get -v github.com/gskartwii/roblox-dissector/...`  
And compile: `%GOPATH%/bin/qtdeploy build %GOPATH%/src/github.com/gskartwii/roblox-dissector`  
A directory named `deploy` should now exist in `$GOPATH/src/github/gskartwii/roblox-dissector`.

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
6. Run a client,with the server address set to `localme`.
7. `roblox-dissector` should now be able to capture the traffic.

I'm also into adding a HTTPS traffic proxy which would be similar to Fiddler while working properly.

Code for 0x8A packets exists, but not publicly due to security reasons. This may change in the future.

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
