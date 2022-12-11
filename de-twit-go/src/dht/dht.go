package dht

import (
	"context"
	"de-twit-go/src/common"
	"de-twit-go/src/timeline"
	"fmt"
	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-datastore"
	sync2 "github.com/ipfs/go-datastore/sync"
	"github.com/libp2p/go-libp2p"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p-kad-dht/providers"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/peerstore"
	"github.com/multiformats/go-multiaddr"
	"log"
	"math/rand"
	"sync"
	"time"
)

var providerTTL = time.Hour

type NullValidator struct{}

// Validate always returns success
func (nv NullValidator) Validate(key string, value []byte) error {
	return nil
}

// Select always selects the first record
func (nv NullValidator) Select(key string, values [][]byte) (int, error) {
	strs := make([]string, len(values))
	for i := 0; i < len(values); i++ {
		strs[i] = string(values[i])
	}

	return 0, nil
}

func createNode(ctx context.Context, port int64) host.Host {
	logger := ctx.Value("logger").(*log.Logger)

	privKey := generatePrivateKey(ctx, port)

	addr, err := addressWithPort(port)
	if err != nil {
		logger.Fatal(err)
	}

	h, err := libp2p.New(
		libp2p.Identity(privKey),
		libp2p.ListenAddrs(
			addr,
		),
	)
	if err != nil {
		logger.Fatalf("Err on creating host: %v", err)
	}

	return h
}

func addressWithPort(port int64) (multiaddr.Multiaddr, error) {
	return multiaddr.NewMultiaddr(fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", port))
}

func generatePrivateKey(ctx context.Context, seed int64) crypto.PrivKey {
	logger := ctx.Value("logger").(*log.Logger)
	randBytes := rand.New(rand.NewSource(seed))
	prvKey, _, err := crypto.GenerateKeyPairWithReader(crypto.RSA, 2048, randBytes)

	if err != nil {
		logger.Fatalf("Could not generate Private Key: %v", err)
	}

	return prvKey
}

func NewDHT(ctx context.Context, port int64, bootstrapNodes []string) (*dht.IpfsDHT, host.Host, error) {
	logger := ctx.Value("logger").(*log.Logger)

	var bootstrapNodeInfos []peer.AddrInfo

	h := createNode(ctx, port)

	providers.ProvideValidity = providerTTL
	providerStore, err := providers.NewProviderManager(
		ctx,
		h.ID(),
		h.Peerstore(),
		sync2.MutexWrap(datastore.NewMapDatastore()),
		providers.CleanupInterval(providerTTL/2),
	)
	if err != nil {
		return nil, nil, err
	}

	opts := []dht.Option{
		dht.Mode(dht.ModeServer),
		dht.Validator(NullValidator{}),
		dht.ProtocolPrefix("/p2p"),
		dht.ProviderStore(providerStore),
	}

	for _, node := range bootstrapNodes {
		addr, err := multiaddr.NewMultiaddr(node)
		if err != nil {
			logger.Printf("Warning on DHT creation: %s", err)
			continue
		}
		peerInfo, err := peer.AddrInfoFromP2pAddr(addr)
		if err != nil {
			logger.Printf("Warning on DHT creation: %s", err)
			continue
		}

		bootstrapNodeInfos = append(bootstrapNodeInfos, *peerInfo)
	}
	opts = append(opts, dht.BootstrapPeers(bootstrapNodeInfos...))

	kad, err := dht.New(ctx, h, opts...)
	if err != nil {
		return nil, nil, err
	}
	err = kad.Bootstrap(ctx)
	if err != nil {
		return nil, nil, err
	}

	var wg sync.WaitGroup
	for _, nodeInfo := range bootstrapNodeInfos {
		wg.Add(1)
		nodeInfo := nodeInfo
		go func() {
			defer wg.Done()
			h.Peerstore().AddAddr(nodeInfo.ID, nodeInfo.Addrs[0], peerstore.PermanentAddrTTL)
			err = h.Connect(ctx, nodeInfo)

			if err != nil {
				logger.Printf("Warning on DHT creation: %s", err)
			} else {
				logger.Println("Info on DHT creation: connected to node ", nodeInfo)
			}
		}()
	}
	wg.Wait()

	return kad, h, nil
}

func RegisterProvideRoutine(ctx context.Context, kad *dht.IpfsDHT, followingTimelines *timeline.FollowingTimelines, nodeCid cid.Cid) {
	logger := common.GetLogger(ctx)

	go func() {
		ticker := time.NewTicker(providerTTL / 2)

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
