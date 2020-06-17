#!/bin/dash

set -eux

GOPATH=`go env GOPATH`
WINDIVERT_HOME=/mnt/c/Users/aleks/Downloads/WinDivert-2.2.0-A/x64/
WINDIVERT_BIND=$GOPATH/src/github.com/Gskartwii/windivert-go

cd $GOPATH/src/github.com/Gskartwii/roblox-dissector

[ -f $WINDIVERT_BIND/libwindivert.dll ] || cp $WINDIVERT_HOME/WinDivert.dll $WINDIVERT_BIND/libwindivert.dll

powershell.exe -Command "Start-Process sc.exe -ArgumentList 'stop windivert' -Verb RunAs" || [ $? -eq 36 ] && echo "WinDivert not running"
x86_64-w64-mingw32-windres icon.rc -o icon_win64.syso
$GOPATH/bin/qtdeploy -docker -tags=divert -qt_version=5.12.8 -qt_api=5.12.0 build windows_64_static > /dev/null
cp $WINDIVERT_HOME/WinDivert64.sys $WINDIVERT_HOME/WinDivert.dll deploy/windows/
cp -R deploy/windows/ /mnt/c/Users/aleks/qtdeploy/roblox-dissector
echo "WinDivert compilation done"
