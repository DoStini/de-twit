package main

import (
	"fmt"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/multiformats/go-multiaddr"
	"log"
	"math/rand"
)

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

func generatePrivateKey(seed int64) crypto.PrivKey {
	randBytes := rand.New(rand.NewSource(seed))
	prvKey, _, err := crypto.GenerateKeyPairWithReader(crypto.RSA, 2048, randBytes)

	if err != nil {
		log.Fatalf("Could not generate Private Key: %v", err)
	}

	return prvKey
}
