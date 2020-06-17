#!/bin/dash
set -eux

sc.exe stop windivert || [ $? -eq 36 ] && echo "WinDivert not running"
powershell.exe -Command "Start-Process C:\\Users\\aleks\\qtdeploy\\roblox-dissector\\roblox-dissector.exe -Verb RunAs"
