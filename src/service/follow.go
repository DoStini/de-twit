package service

import (
	"bufio"
	"context"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"log"
	"src/common"
)

func Follow(ctx context.Context, host host.Host, kad *dht.IpfsDHT, user string) ([]string, error) {
	logger := ctx.Value("logger").(*log.Logger)

	var peers []peer.AddrInfo

	c, err := common.GenerateCid(ctx, user)
	if err != nil {
		logger.Fatalf(err.Error())
		return nil, err
	}

	peerChan := kad.FindProvidersAsync(ctx, c, 5)
	for p := range peerChan {
		peers = append(peers, p)
	}

	var peerResps []string

	for _, currPeer := range peers {
		if currPeer.ID == host.ID() {
			continue
		}

		if err := host.Connect(ctx, currPeer); err != nil {
			logger.Println(err.Error())
			continue
		}
		stream, err := host.NewStream(ctx, currPeer.ID, "/p2p/1.0.0")
		if err != nil {
			logger.Println(err.Error())
			continue
		}

		rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))

		resp, err := rw.ReadString('\n')
		if err != nil {
			return nil, err
		}

		peerResps = append(peerResps, string(resp))
	}

	err = kad.Provide(ctx, c, true)
	if err != nil {
		logger.Fatalf(err.Error())
		return nil, err
	}

	return peerResps, nil
}
