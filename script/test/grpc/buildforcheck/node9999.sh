#!/bin/bash
trap "rm server;kill 0" EXIT

go build -o ./server ../../../../cmd/grpc/main.go
./server -port 9999