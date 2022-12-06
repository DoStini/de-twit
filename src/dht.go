package main

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

func startDHT(ctx context.Context, port int64, bootstrapNodes []string) (*dht.IpfsDHT, host.Host) {
	var bootstrapNodeInfos []peer.AddrInfo

	h := createNode(port)

	opts := []dht.Option{
		dht.Mode(dht.ModeServer),
	}

	for _, node := range bootstrapNodes {
		addr, err := multiaddr.NewMultiaddr(node)
		if err != nil {
			log.Println(err.Error())
			continue
		}
		peerInfo, err := peer.AddrInfoFromP2pAddr(addr)
		if err != nil {
			log.Println(err.Error())
			continue
		}

		bootstrapNodeInfos = append(bootstrapNodeInfos, *peerInfo)
		opts = append(opts, dht.BootstrapPeers(*peerInfo))
	}

	kad, err := dht.New(ctx, h, opts...)

	if err != nil {
		log.Println("Err on creating dht")
		log.Fatalf(err.Error())
	}
	err = kad.Bootstrap(ctx)
	if err != nil {
		log.Fatalf(err.Error())
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
				log.Println(err.Error())
			} else {
				log.Println("Connected to node ", nodeInfo)
			}
		}()
	}
	wg.Wait()

	return kad, h
}
