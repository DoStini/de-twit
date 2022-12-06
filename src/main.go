package main

import (
	"context"
	"fmt"
	"github.com/libp2p/go-libp2p"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/multiformats/go-multiaddr"
	"log"
)

func addressWithPort(port int64) (multiaddr.Multiaddr, error) {
	return multiaddr.NewMultiaddr(fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", port))
}

func createNode(port int64) host.Host {
	privKey := generatePrivateKey(port)

	fmt.Println(privKey)

	addr, err := addressWithPort(port)
	if err != nil {
		log.Fatal(err)
	}

	h, err := libp2p.New(
		libp2p.Identity(privKey),
		libp2p.ListenAddrs(
			addr,
		),
	)
	if err != nil {
		log.Fatalf("Err on creating host: %v", err)
	}

	return h
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	h := createNode(4000)

	addr, _ := multiaddr.NewMultiaddr("localhost:4001")
	other, _ := libp2p.New(libp2p.Identity(generatePrivateKey(4001)), libp2p.ListenAddrs(addr))
	fmt.Println(h.ID(), other.ID())

	_, err := dht.New(ctx, h, dht.BootstrapPeers())
	if err != nil {
		log.Fatalf("Err on creating dht")
	}

	// h.Peerstore().AddAddr(other.ID(), addr, 1000)

	// kad.FindPeer(ctx, h.ID())

	// print the node's listening addresses
	fmt.Println("Listen addresses:", h.Addrs())

	// shut the node down
	if err := h.Close(); err != nil {
		panic(err)
	}
}
