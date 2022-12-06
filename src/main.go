package main

import (
	"context"
	"flag"
	"log"
	"time"
)

func main() {
	port := flag.Int64("port", 4000, "The port of this host")
	bootstrap := flag.String("bootstrap", "", "The bootstrapping node")
	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var bootstrapNodes []string
	if *bootstrap != "" {
		bootstrapNodes = append(bootstrapNodes, *bootstrap)
	}

	kad, host := startDHT(ctx, *port, bootstrapNodes)

	hostID := host.ID()
	log.Printf("Created Node at: %s/ipfs/%s", host.Addrs()[0].String(), hostID)
	log.Printf("Node ID: %s", hostID)

	defer func() {
		if err := host.Close(); err != nil {
			panic(err)
		}
	}()

	for {
		time.Sleep(time.Second * 3)
		kad.RoutingTable().Print()
	}
}
