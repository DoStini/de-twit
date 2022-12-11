package service

import (
	"bufio"
	"context"
	"de-twit-go/src/common"
	"de-twit-go/src/timeline"
	"de-twit-go/src/timelinepb"
	"encoding/binary"
	"fmt"
	"github.com/ipfs/go-cid"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"google.golang.org/protobuf/proto"
	"log"
	"os"
)

func RegisterStreamHandler(ctx context.Context, host host.Host, nodeCid cid.Cid, followingTimelines *timeline.FollowingTimelines) {
	var logger *log.Logger
	logger = ctx.Value("logger").(*log.Logger)
	if logger == nil {
		logger = log.New(os.Stdin, "listen: ", log.Ltime|log.Lshortfile)
	}

	host.SetStreamHandler("/p2p/1.0.0", func(stream network.Stream) {
		rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))
		defer func(stream network.Stream) {
			err := stream.Close()
			if err != nil {
				logger.Println(err)
			}
		}(stream)

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
		var followingTimeline *timeline.Timeline

		followingTimelines.RLock()
		if nodeCid == requestedCid || common.Contains(followingTimelines.FollowingCids, requestedCid) {
			followingTimeline = followingTimelines.Timelines[requestedCid]
		}
		followingTimelines.RUnlock()

		if followingTimeline != nil {
			replyMessage := &timelinepb.GetReply{
				Status:   timelinepb.ReplyStatus_OK,
				Timeline: &followingTimeline.PB,
			}
			size := proto.Size(replyMessage)
			reply = make([]byte, 0, size+binary.MaxVarintLen64+1)

			reply = append(binary.AppendVarint(reply, int64(size)), 0)

			buf, err := proto.Marshal(replyMessage)
			if err != nil {
				logger.Println("Failed to encode post:", err)
				followingTimelines.RUnlock()
				return
			}

			reply = append(reply, buf...)
		} else {
			logger.Println(fmt.Sprintf("Node not following %s anymore", requestedCid))
			replyMessage := &timelinepb.GetReply{
				Status:   timelinepb.ReplyStatus_NOT_FOLLOWING,
				Timeline: nil,
			}
			size := proto.Size(replyMessage)
			reply = make([]byte, 0, size+binary.MaxVarintLen64+1)

			reply = append(binary.AppendVarint(reply, int64(size)), 0)

			buf, err := proto.Marshal(replyMessage)
			if err != nil {
				logger.Println("Failed to encode post:", err)
				followingTimelines.RUnlock()
				return
			}

			reply = append(reply, buf...)
		}

		_, err = rw.Write(reply)
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
