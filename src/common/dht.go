package common

import (
	"context"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/peerstore"
	"github.com/multiformats/go-multiaddr"
	"log"
	"sync"
)

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

func StartDHT(ctx context.Context, port int64, bootstrapNodes []string) (*dht.IpfsDHT, host.Host, error) {
	logger := ctx.Value("logger").(*log.Logger)

	var bootstrapNodeInfos []peer.AddrInfo

	h := createNode(ctx, port)

	opts := []dht.Option{
		dht.Mode(dht.ModeServer),
		dht.Validator(NullValidator{}),
		dht.ProtocolPrefix("/p2p"),
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
