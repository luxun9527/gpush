#!/bin/bash
dir=$(pwd)
chmod 777 "$dir/scripts/stop.sh"
"$dir/scripts/stop.sh"
if [ ! -e "./bin" ];then
  mkdir bin
fi
socketdir="$dir/cmd/socket"
cd $socketdir
go build -o "$dir/bin/socket"
proxydir="$dir/cmd/proxy"
cd $proxydir
go build -o "$dir/bin/proxy"
echo '开始执行'
socketConfigDir="$dir/config/socket/config.toml"
proxyConfigDir="$dir/config/proxy/config.toml"
echo $socketConfigDir
nohup "$dir/bin/proxy" --config="$proxyConfigDir"   > /dev/null 2>&1 &
echo $proxyConfigDir
nohup "$dir/bin/socket" --config="$socketConfigDir"   > /dev/null 2>&1 &

