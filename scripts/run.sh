#!/bin/bash
dir=$(pwd)
echo  "$dir"
chmod 777 "$dir/scripts/stop.sh"
"$dir/scripts/stop.sh"
socketConfigDir="$dir/config/socket/config.toml"
proxyConfigDir="$dir/config/proxy/config.toml"
nohup "$dir/bin/proxy" --config="$proxyConfigDir"   > test.log 2>&1 &
nohup "$dir/bin/socket" --config="$socketConfigDir"  > socket.log 2>&1 &

