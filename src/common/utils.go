package common

import (
	"context"
	"fmt"
	"github.com/ipfs/go-cid"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/multiformats/go-multiaddr"
	"github.com/multiformats/go-multicodec"
	"github.com/multiformats/go-multihash"
	"log"
	"math/rand"
)

func addressWithPort(port int64) (multiaddr.Multiaddr, error) {
	return multiaddr.NewMultiaddr(fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", port))
}

func createNode(ctx context.Context, port int64) host.Host {
	logger := ctx.Value("logger").(*log.Logger)

	privKey := generatePrivateKey(ctx, port)

	addr, err := addressWithPort(port)
	if err != nil {
		logger.Fatal(err)
	}

	h, err := libp2p.New(
		libp2p.Identity(privKey),
		libp2p.ListenAddrs(
			addr,
		),
	)
	if err != nil {
		logger.Fatalf("Err on creating host: %v", err)
	}

	return h
}

func generatePrivateKey(ctx context.Context, seed int64) crypto.PrivKey {
	logger := ctx.Value("logger").(*log.Logger)
	randBytes := rand.New(rand.NewSource(seed))
	prvKey, _, err := crypto.GenerateKeyPairWithReader(crypto.RSA, 2048, randBytes)

	if err != nil {
		logger.Fatalf("Could not generate Private Key: %v", err)
	}

	return prvKey
}

func GenerateCid(ctx context.Context, key string) (cid.Cid, error) {
	logger := ctx.Value("logger").(*log.Logger)

	pref := cid.Prefix{
		Version:  1,
		Codec:    uint64(multicodec.Raw),
		MhType:   multihash.SHA2_256,
		MhLength: -1, // default length
	}
	c, err := pref.Sum([]byte(key))
	if err != nil {
		logger.Fatalf(err.Error())
	}
	return c, err
}
