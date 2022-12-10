package service

import (
	"bufio"
	"context"
	"fmt"
	"github.com/ipfs/go-cid"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"google.golang.org/protobuf/proto"
	"log"
	"os"
	"src/common"
	"src/timeline"
)

func RegisterStreamHandler(ctx context.Context, host host.Host, nodeCid cid.Cid, followingTimelines *timeline.FollowingTimelines) {
	var logger *log.Logger
	logger = ctx.Value("logger").(*log.Logger)
	if logger == nil {
		logger = log.New(os.Stdin, "listen: ", log.Ltime|log.Lshortfile)
	}

	host.SetStreamHandler("/p2p/1.0.0", func(stream network.Stream) {
		rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))

		resp, err := rw.ReadBytes(0)
		if err != nil {
			logger.Println(err.Error())
		}

		cidResp := resp[:len(resp)-1]

		requestedCid, err := cid.Cast(cidResp)
		if err != nil {
			logger.Println(err.Error())
			stream.Close()
			return
		}

		var reply []byte

		followingTimelines.RLock()
		if nodeCid == requestedCid || common.Contains(followingTimelines.FollowingCids, requestedCid) {
			reply, err = proto.Marshal(followingTimelines.Timelines[requestedCid])
			if err != nil {
				logger.Println("Failed to encode post:", err)
				followingTimelines.RUnlock()
				return
			}
		} else {
			logger.Println(fmt.Sprintf("Node not following %s anymore", requestedCid))
			reply = []byte(fmt.Sprintf("%s-NOT-FOLLOWING", nodeCid))
		}
		followingTimelines.RUnlock()

		_, err = rw.Write(append(reply, 0))
		if err != nil {
			logger.Println(err.Error())
			return
		}

		err = rw.Flush()
		if err != nil {
			logger.Println(err.Error())
			return
		}
	})
}