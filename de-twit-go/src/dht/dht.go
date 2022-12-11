package dht

import (
	"context"
	"de-twit-go/src/common"
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
	"sync/atomic"
	"time"
)

var minProviders int32 = 5

var ProviderTTL = time.Hour

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

	providers.ProvideValidity = ProviderTTL
	providerStore, err := providers.NewProviderManager(
		ctx,
		h.ID(),
		h.Peerstore(),
		sync2.MutexWrap(datastore.NewMapDatastore()),
		providers.CleanupInterval(ProviderTTL/2),
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

func HandleWithProviders(ctx context.Context, cid cid.Cid, kad *dht.IpfsDHT, work func(peer.AddrInfo) error) {
	var count atomic.Int32
	logger := common.GetLogger(ctx)
	ctx, cancel := context.WithCancel(ctx)

	var wg sync.WaitGroup
	peerChan := kad.FindProvidersAsync(ctx, cid, 0)
	guard := make(chan struct{}, minProviders)

	for p := range peerChan {
		wg.Add(1)
		go func(info peer.AddrInfo) {
			defer wg.Done()

			guard <- struct{}{}
			defer func() { <-guard }()
			if count.Load() >= minProviders {
				cancel()
				return
			}

			if err := kad.Host().Connect(ctx, info); err != nil {
				logger.Println(err.Error())
				return
			}

			err := work(info)
			if err != nil {
				logger.Println(err.Error())
				return
			}

			count.Add(1)
		}(p)
	}
	wg.Wait()
	cancel()

	logger.Printf("Connected to %d peers\n", count.Load())
}
