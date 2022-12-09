package service

import (
	"bufio"
	"context"
	"errors"
	"github.com/ipfs/go-cid"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"google.golang.org/protobuf/proto"
	"log"
	"src/common"
	timeline "src/timeline"
)

func Follow(ctx context.Context, targetCid cid.Cid, host host.Host, kad *dht.IpfsDHT) (*timeline.Timeline, error) {
	logger := ctx.Value("logger").(*log.Logger)

	var peers []peer.AddrInfo

	peerChan := kad.FindProvidersAsync(ctx, targetCid, 5)
	for p := range peerChan {
		peers = append(peers, p)
	}

	var peerResps []string

	var receivedTimelines []*timeline.Timeline

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

		_, err = rw.Write(append(targetCid.Bytes(), 0))
		if err != nil {
			return nil, err
		}

		err = rw.Flush()
		if err != nil {
			return nil, err
		}

		resp, err := rw.ReadBytes(0)
		if err != nil {
			return nil, err
		}

		resp = resp[:len(resp)-1]

		var t timeline.Timeline

		err = proto.Unmarshal(resp, &t)
		if err != nil {
			logger.Println(err.Error())
			peerResps = append(peerResps, string(resp))
		} else {
			peerResps = append(peerResps, t.String())
		}

		receivedTimelines = append(receivedTimelines, &t)
	}

	err := kad.Provide(ctx, targetCid, true)
	if err != nil {
		return nil, err
	}

	return receivedTimelines[0], nil
}

func Unfollow(targetCid cid.Cid, followingCids []cid.Cid, timelines map[cid.Cid]*timeline.Timeline) ([]cid.Cid, map[cid.Cid]*timeline.Timeline, error) {
	index := common.FindIndex(followingCids, targetCid)

	if index == -1 {
		return nil, nil, errors.New("target timeline not found")
	}

	delete(timelines, targetCid)
	followingCids = common.RemoveIndex(followingCids, index)

	return followingCids, timelines, nil
}
