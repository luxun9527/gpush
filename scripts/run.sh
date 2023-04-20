#!/bin/bash
dir=$(pwd)
chmod 777 "$dir/scripts/stop.sh"
"$dir/scripts/stop.sh"
if [ ! -e "./bin" ];then
  mkdir bin
fi
wsdir="$dir/cmd/socket"
cd $wsdir
go build -o "$dir/bin/socket"
proxydir="$dir/cmd/proxy"
cd $proxydir
go build -o "$dir/bin/proxy"
echo '开始执行'
wsConfigDir="$dir/config/ws/config.toml"
proxyConfigDir="$dir/config/proxy/config.toml"
echo $wsConfigDir
nohup "$dir/bin/proxy" --config="$proxyConfigDir"   &
echo $proxyConfigDir
nohup "$dir/bin/socket" --config="$wsConfigDir"    &

