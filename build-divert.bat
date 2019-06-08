sc stop WinDivert1.4
(qtdeploy -tags=divert -fast -p 8 build windows) && (copy WinDivert64.sys deploy\windows\) && (copy WinDivert.dll deploy\windows\)