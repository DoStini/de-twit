package main

import (
	"context"
	"flag"
	"fmt"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/peerstore"
	"github.com/multiformats/go-multiaddr"
	"log"
)

func main() {
	port := flag.Int64("port", 4000, "The port of this host")
	bootstrap := flag.String("bootstrap", "", "The bootstrapping node")
	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	h := createNode(*port)

	opts := []dht.Option{
		dht.Mode(dht.ModeServer),
	}

	if *bootstrap != "" {
		addr, err := multiaddr.NewMultiaddr(*bootstrap)
		if err != nil {
			log.Fatalf(err.Error())
		}
		peerInfo, err := peer.AddrInfoFromP2pAddr(addr)
		if err != nil {
			log.Fatalf(err.Error())
		}
		opts = append(opts, dht.BootstrapPeers(*peerInfo))
		h.Peerstore().AddAddr(peerInfo.ID, peerInfo.Addrs[0], peerstore.PermanentAddrTTL)
		err = h.Connect(ctx, *peerInfo)
		if err != nil {
			log.Fatalf(err.Error())
		}
	}

	ctx, eventChannel := dht.RegisterForLookupEvents(ctx)
	kad, err := dht.New(ctx, h, opts...)

	if err != nil {
		log.Println("Err on creating dht")
		log.Fatalf(err.Error())
	}
	err = kad.Bootstrap(ctx)
	if err != nil {
		return
	}

	peer, err := kad.FindPeer(ctx, h.ID())
	if err != nil {
		log.Println(err.Error())
	} else {
		log.Println(peer.ID)
		log.Println(peer.Addrs)
	}
	kad.RoutingTable().Print()

	hostID := h.ID()
	log.Printf("Host MultiAddress: %s/ipfs/%s", h.Addrs()[0].String(), hostID)

	defer func() {
		fmt.Println("test")
		// shut the node down
		if err := h.Close(); err != nil {
			panic(err)
		}
	}()

	for {
		select {
		case event := <-eventChannel:
			log.Println("RECV EVENT")
			log.Println(event.ID)
			log.Println(event.Node)
			log.Println(event.Request)

			peer, err := kad.FindPeer(ctx, h.ID())
			if err != nil {
				log.Println(err.Error())
			} else {
				log.Println(peer.ID)
				log.Println(peer.Addrs)
			}
			kad.RoutingTable().Print()
		}
	}
}
