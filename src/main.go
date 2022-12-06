package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/libp2p/go-libp2p"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
	"log"
)

// MakePeer takes a fully-encapsulated address and converts it to a
// peer ID / Multiaddress pair
func MakePeer(dest string) (peer.ID, multiaddr.Multiaddr) {
	ipfsAddr, err := multiaddr.NewMultiaddr(dest)
	if err != nil {
		log.Fatalf("Err on creating host: %v", err)
	}
	log.Printf("Parsed: ipfsAddr = %s", ipfsAddr)

	peerIDStr, err := ipfsAddr.ValueForProtocol(multiaddr.P_IPFS)
	if err != nil {
		log.Fatalf("Err on creating peerIDStr: %v", err)
	}
	log.Printf("Parsed: PeerIDStr = %s", peerIDStr)

	targetPeerAddr, err := multiaddr.NewMultiaddr(fmt.Sprintf("/ipfs/%s", peerIDStr))
	log.Printf("Created targetPeerAddr = %v", targetPeerAddr)

	targetAddr := ipfsAddr.Decapsulate(targetPeerAddr)
	log.Printf("Decapsulated = %v", targetAddr)

	return peer.ID(peerIDStr), targetAddr
}

func addressWithPort(port int64) (multiaddr.Multiaddr, error) {
	return multiaddr.NewMultiaddr(fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", port))
}

func createNode(port int64) host.Host {
	privKey := generatePrivateKey(port)

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
	port := flag.Int64("port", 4000, "The port of this host")
	//node := flag.String()
	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	h := createNode(*port)
	hostID := h.ID()

	kad, err := dht.New(ctx, h, dht.BootstrapPeers(peer.AddrInfo{ID: hostID}))
	if err != nil {
		log.Fatalf("Err on creating dht")
	}
	err = kad.Bootstrap(ctx)
	if err != nil {
		log.Fatalf(err.Error())
	}

	// kad.FindPeer(ctx, h.ID())

	log.Printf("Host MultiAddress: %s/ipfs/%s", h.Addrs()[0].String(), hostID)
	MakePeer(fmt.Sprintf("%s/ipfs/%s", h.Addrs()[0].String(), hostID))

	// shut the node down
	if err := h.Close(); err != nil {
		panic(err)
	}
}
