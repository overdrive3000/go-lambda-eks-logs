#!/bin/bash

SLEEP=$((1 + $RANDOM % 10))

sleep $SLEEP

echo "let's crash by OOM"
/usr/bin/stress --vm 2 --vm-bytes 256M
