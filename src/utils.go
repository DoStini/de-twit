package main

import (
	"github.com/libp2p/go-libp2p/core/crypto"
	"log"
	"math/rand"
)

func generatePrivateKey(seed int64) crypto.PrivKey {
	randBytes := rand.New(rand.NewSource(seed))
	prvKey, _, err := crypto.GenerateKeyPairWithReader(crypto.RSA, 2048, randBytes)

	if err != nil {
		log.Fatalf("Could not generate Private Key: %v", err)
	}

	return prvKey
}
