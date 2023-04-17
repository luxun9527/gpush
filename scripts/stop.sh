#!/bin/bash

pid=$(ps x | grep "ws" | grep -v grep | awk '{print $1}')
echo $pid
kill $pid
pid1=$(ps x | grep "proxy" | grep -v grep | awk '{print $1}')
echo $pid1
kill $pid1