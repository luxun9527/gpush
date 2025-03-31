#!/bin/bash
dir=$(pwd)
if [ ! -e "$dir/bin" ];then
  mkdir bin
fi
cd "$dir/cmd/socket"
go build -o "$dir/bin/socket"
cd "$dir/cmd/proxy"
go build -o "$dir/bin/proxy"

cd "$dir/cmd/stress"
go build -o "$dir/bin/stress"