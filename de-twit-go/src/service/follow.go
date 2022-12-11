package service

import (
	"bufio"
	"context"
	dht2 "de-twit-go/src/dht"
	"de-twit-go/src/timeline"
	"de-twit-go/src/timelinepb"
	"encoding/binary"
	"errors"
	"github.com/ipfs/go-cid"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"google.golang.org/protobuf/proto"
	"io"
	"log"
	"sync"
)

func FindPosts(ctx context.Context, targetCid cid.Cid, kad *dht.IpfsDHT) (*timelinepb.Timeline, error) {
	logger := ctx.Value("logger").(*log.Logger)
	host := kad.Host()

	var timelineLock sync.RWMutex
	var receivedTimelines []*timelinepb.Timeline
	var peerResps []string

	dht2.HandleWithProviders(ctx, targetCid, kad, func(info peer.AddrInfo) error {
		stream, err := host.NewStream(ctx, info.ID, "/p2p/1.0.0")
		if err != nil {
			logger.Println(err.Error())
			return nil
		}
		err = func() error {
			defer func(stream network.Stream) {
				err := stream.Close()
				if err != nil {
					logger.Println(err)
				}
			}(stream)

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

			var reply timelinepb.GetReply

			err = proto.Unmarshal(resp, &reply)
			if err != nil {
				logger.Println(err.Error())
				timelineLock.Lock()

				peerResps = append(peerResps, string(resp))

				timelineLock.Unlock()
				return err
			}

			if reply.Status == timelinepb.ReplyStatus_OK {
				timelineLock.Lock()

				peerResps = append(peerResps, reply.Timeline.String())
				receivedTimelines = append(receivedTimelines, reply.Timeline)

				timelineLock.Unlock()
			} else {
				timelineLock.Lock()

				logger.Println("Reply Status: Not Following")
				peerResps = append(peerResps, reply.Timeline.String())

				timelineLock.Unlock()
			}

			return nil
		}()
		if err != nil {
			return err
		}

		return nil
	})

	logger.Println("Responses: ", peerResps)
	if len(receivedTimelines) == 0 {
		return nil, errors.New("user not found")
	}

	finalTimeline := timelinepb.Timeline{Posts: make([]*timelinepb.Post, 0)}
	err := timeline.MergeTimelines(&finalTimeline, receivedTimelines)

	return &finalTimeline, err
}
