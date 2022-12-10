package service

import (
	"bufio"
	"context"
	dht2 "de-twit-go/src/dht"
	"de-twit-go/src/timeline"
	"errors"
	"github.com/ipfs/go-cid"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"google.golang.org/protobuf/proto"
	"log"
	"sync"
)

func Follow(ctx context.Context, targetCid cid.Cid, host host.Host, kad *dht.IpfsDHT) (*timeline.Timeline, error) {
	logger := ctx.Value("logger").(*log.Logger)

	var timelineLock sync.RWMutex
	var receivedTimelines []*timeline.Timeline
	var peerResps []string

	dht2.DoWithProviders(ctx, targetCid, kad, func(info peer.AddrInfo) error {
		stream, err := host.NewStream(ctx, info.ID, "/p2p/1.0.0")
		if err != nil {
			return err
		}

		rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))

		_, err = rw.Write(append(targetCid.Bytes(), 0))
		if err != nil {
			return err
		}

		err = rw.Flush()
		if err != nil {
			return err
		}

		resp, err := rw.ReadBytes(0)
		if err != nil {
			return err
		}

		resp = resp[:len(resp)-1]

		var t timeline.Timeline

		err = proto.Unmarshal(resp, &t)
		if err != nil {
			logger.Println(err.Error())
			timelineLock.Lock()

			peerResps = append(peerResps, string(resp))

			timelineLock.Unlock()
			return err
		}

		timelineLock.Lock()

		peerResps = append(peerResps, t.String())
		receivedTimelines = append(receivedTimelines, &t)

		timelineLock.Unlock()

		return nil
	})

	logger.Println("Responses: ", peerResps)

	err := kad.Provide(ctx, targetCid, true)
	if err != nil {
		return nil, err
	}

	if len(receivedTimelines) == 0 {
		return nil, errors.New("user not found")
	}

	return receivedTimelines[0], nil
}
