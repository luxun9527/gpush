#!/bin/bash
dir=$(pwd)
chmod 777 "$dir/scripts/stop.sh"
"$dir/scripts/stop.sh"
socketConfigDir="$dir/config/socket/config.toml"
proxyConfigDir="$dir/config/proxy/config.toml"
nohup "$dir/bin/proxy" --config="$proxyConfigDir"   > /dev/null 2>&1 &
nohup "$dir/bin/socket" --config="$socketConfigDir"  > /dev/null 2>&1 &

