#! /bin/bash
trap "rm server;kill 0" EXIT

go build -o ./server ../../../cmd/grpc/main.go
./server -port=8001 &
./server -port=8002 &
./server -port=8003 &

wait