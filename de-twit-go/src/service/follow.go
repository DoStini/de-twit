package service

import (
	"bufio"
	"context"
	"de-twit-go/src/timeline"
	"de-twit-go/src/timelinepb"
	"encoding/binary"
	"errors"
	"github.com/ipfs/go-cid"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"google.golang.org/protobuf/proto"
	"io"
	"log"
	"sort"
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
		err = func() error {
			defer func(stream network.Stream) {
				err := stream.Close()
				if err != nil {
					logger.Println(err)
				}
			}(stream)

			if err != nil {
				logger.Println(err.Error())
				return nil
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

			sizeBuf, err := rw.ReadBytes(0)
			if err != nil {
				return err
			}
			if len(sizeBuf) == 1 && sizeBuf[0] == 0 {
				logger.Println("Received Nothing")
				return nil
			}

			size, i := binary.Varint(sizeBuf)
			if i <= 0 {
				return errors.New("value larger than 64 bits")
			}

			limitedReader := io.LimitReader(rw.Reader, size)
			resp := make([]byte, size)
			_, err = limitedReader.Read(resp)
			if err != nil {
				return err
			}

			var t timeline.Timeline

			err = proto.Unmarshal(resp, &t)
			if err != nil {
				logger.Println(err.Error())
				peerResps = append(peerResps, string(resp))
			} else {
				peerResps = append(peerResps, t.String())
				receivedTimelines = append(receivedTimelines, &t)
			}

			return nil
		}()
		if err != nil {
			return nil, err
		}
	}

	logger.Println("Responses: ", peerResps)

	if len(receivedTimelines) == 0 {
		return nil, errors.New("user not found")
	}

	finalTimeline := timelinepb.Timeline{Posts: make([]*timelinepb.Post, 0)}

	err := mergeTimelines(&finalTimeline, receivedTimelines)

	return &finalTimeline, err
}

func mergeTimelines(t *timelinepb.Timeline, timelines []*timeline.Timeline) error {
	posts := make([]*timelinepb.Post, 0)

	for _, curTimeline := range timelines {
		posts = append(posts, curTimeline.Posts...)
	}

	sort.SliceStable(posts, func(i, j int) bool {
		return posts[i].LastUpdated.AsTime().After(posts[j].LastUpdated.AsTime())
	})

	for _, post := range posts {
		if !containsPost(t.Posts, post) {
			t.Posts = append(t.Posts, post)
		}
	}

	return nil
}

func containsPost(posts []*timelinepb.Post, p *timelinepb.Post) bool {
	for _, post := range posts {
		if post.Id == p.Id && post.User == p.User {
			return true
		}
	}
	return false
}
