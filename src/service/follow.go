package service

import (
	"bufio"
	"context"
	"errors"
	"log"
	"sort"
	timeline "src/timeline"
	"src/timelinepb"
	pb "src/timelinepb"

	"github.com/ipfs/go-cid"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"google.golang.org/protobuf/proto"
)

func Follow(ctx context.Context, targetCid cid.Cid, host host.Host, kad *dht.IpfsDHT) (*timelinepb.Timeline, error) {
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
			receivedTimelines = append(receivedTimelines, &t)
		}
	}

	logger.Println("Responses: ", peerResps)

	err := kad.Provide(ctx, targetCid, true)
	if err != nil {
		return nil, err
	}

	if len(receivedTimelines) == 0 {
		return nil, errors.New("user not found")
	}

	timeline := timelinepb.Timeline{Posts: make([]*timelinepb.Post, 0)}

	mergeTimelines(&timeline, receivedTimelines)

	return &timeline, nil
}

func mergeTimelines(t *pb.Timeline, timelines []*timeline.Timeline) error {
	var posts []*pb.Post

	t.Posts = make([]*pb.Post, 0)

	for _, timeline := range timelines {
		posts = append(posts, timeline.Posts...)
	}

	sort.SliceStable(posts, func(i, j int) bool {
		return posts[i].LastUpdated.AsTime().Nanosecond() < posts[j].LastUpdated.AsTime().Nanosecond()
	})

	for _, post := range posts {
		if containsPost(t.Posts, post) {
			continue
		}
		t.Posts = append(t.Posts, post)
	}

	return nil
}

func containsPost(posts []*pb.Post, p *pb.Post) bool {
	for _, post := range posts {
		if post.Id == p.Id && post.User == p.User {
			return true
		}
	}
	return false
}
