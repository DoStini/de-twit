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
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"google.golang.org/protobuf/proto"
	"io"
	"log"
	"sync"
)

func Follow(ctx context.Context, targetCid cid.Cid, host host.Host, kad *dht.IpfsDHT) (*timelinepb.Timeline, error) {
	logger := ctx.Value("logger").(*log.Logger)

	var timelineLock sync.RWMutex
	var receivedTimelines []*timeline.Timeline
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
