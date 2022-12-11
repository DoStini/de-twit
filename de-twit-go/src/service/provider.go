package service

import (
	"context"
	"de-twit-go/src/common"
	dht2 "de-twit-go/src/dht"
	"de-twit-go/src/timeline"
	"github.com/ipfs/go-cid"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"time"
)

func RegisterProvideRoutine(ctx context.Context, kad *dht.IpfsDHT, followingTimelines *timeline.FollowingTimelines, nodeCid cid.Cid) {
	logger := common.GetLogger(ctx)

	go func() {
		ticker := time.NewTicker(dht2.ProviderTTL / 2)

		// timer but first tick is instantaneous
		for ; true; <-ticker.C {
			go func() {
				err := kad.Provide(ctx, nodeCid, true)
				if err != nil {
					logger.Println(err.Error())
				}
			}()

			followingTimelines.RLock()
			for _, followingCid := range followingTimelines.FollowingCids {
				go func(followingCid cid.Cid) {
					err := kad.Provide(ctx, followingCid, true)
					if err != nil {
						logger.Println(err.Error())
					}
				}(followingCid)
			}
			followingTimelines.RUnlock()
		}
	}()
}
