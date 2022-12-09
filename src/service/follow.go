package service

import (
	"bufio"
	"context"
	"github.com/ipfs/go-cid"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"google.golang.org/protobuf/proto"
	"log"
	"src/common"
	timeline "src/timeline"
)

func Follow(ctx context.Context, host host.Host, kad *dht.IpfsDHT, followingCids []cid.Cid, user string) (*timeline.OwnTimeline, *cid.Cid, error) {
	logger := ctx.Value("logger").(*log.Logger)

	var peers []peer.AddrInfo

	c, err := common.GenerateCid(ctx, user)
	if err != nil {
		return nil, nil, err
	}

	peerChan := kad.FindProvidersAsync(ctx, c, 5)
	for p := range peerChan {
		peers = append(peers, p)
	}

	var peerResps []string

	var receivedTimelines []*timeline.OwnTimeline

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

		_, err = rw.Write(append(c.Bytes(), 0))
		if err != nil {
			return nil, nil, err
		}

		err = rw.Flush()
		if err != nil {
			return nil, nil, err
		}

		resp, err := rw.ReadBytes(0)
		if err != nil {
			return nil, nil, err
		}

		resp = resp[:len(resp)-1]

		var t timeline.OwnTimeline

		err = proto.Unmarshal(resp, &t.Timeline)
		if err != nil {
			logger.Println(err.Error())
			peerResps = append(peerResps, string(resp))
		} else {
			peerResps = append(peerResps, t.String())
		}

		receivedTimelines = append(receivedTimelines, &t)
	}

	err = kad.Provide(ctx, c, true)
	if err != nil {
		return nil, nil, err
	}

	return receivedTimelines[0], &c, nil
}

func Unfollow(ctx context.Context, followingCids *[]cid.Cid, user string) error {

	c, err := common.GenerateCid(ctx, user)
	if err != nil {
		return err
	}

	index := common.FindIndex(*followingCids, c)
	*followingCids = common.RemoveIndex(*followingCids, index)

	return nil
}
