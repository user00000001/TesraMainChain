.PHONY: build cluster test

default: build

build:
	go build -o bin/swarm    github.com/TesraSupernet/TesraMainChain/cmd/swarm
	go build -o bin/tesramain     github.com/TesraSupernet/TesraMainChain/cmd/tesramain
	go build -o bin/bootnode github.com/TesraSupernet/TesraMainChain/cmd/bootnode

cluster: build
	scripts/boot-cluster.sh

test:
	go test -v github.com/TesraSupernet/TesraMainChain/swarm/...
