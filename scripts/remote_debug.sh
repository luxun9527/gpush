#!/bin/bash

flag=$1

projectPath='/root/smb/gpush'

source /etc/profile


if [ ! -e "$projectPath/bin" ];then
  mkdir $projectPath/bin
fi


if [ $flag='socket' ]; then
    cd "$projectPath/cmd/socket"
    go build -gcflags "all=-N -l" -o "$projectPath/bin/socket"
    dlv --listen=:2345 --headless=true --api-version=2 --accept-multiclient exec "$projectPath/bin/socket" -- -config=/root/smb/gpush/config/socket/config.toml
fi

