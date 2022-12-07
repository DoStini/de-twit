package service

import (
	"context"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/peer"
	"log"
	"src/common"
)

func Follow(ctx context.Context, kad *dht.IpfsDHT, user string) bool {
	logger := ctx.Value("logger").(*log.Logger)

	var peers []peer.AddrInfo

	c, err := common.GenerateCid(ctx, user)
	if err != nil {
		logger.Fatalf(err.Error())
		return false
	}

	peerChan := kad.FindProvidersAsync(ctx, c, 5)
	for p := range peerChan {
		peers = append(peers, p)
	}

	for _, currPeer := range peers {
		logger.Println(currPeer.String())
	}

	err = kad.Provide(ctx, c, true)
	if err != nil {
		logger.Fatalf(err.Error())
		return false
	}

	return true
}
