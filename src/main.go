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

	log.Printf("Created Node at: %s/p2p/%s", host.Addrs()[0].String(), hostID)
	log.Printf("Node ID: %s", hostID)

	defer func() {
		if err := host.Close(); err != nil {
			panic(err)
		}
	}()

	for {
		time.Sleep(time.Second * 3)
		kad.RoutingTable().Print()

		storeData := []byte(time.Now().String())

		key := "key"

		value, err := kad.GetValue(ctx, key)

		if err != nil {
			log.Println(err)
		}

		log.Println("got the thing", string(value))
		err = kad.PutValue(ctx, key, storeData)
		if err != nil {
			log.Println(key)
			log.Println("here", err)
		}
	}
}
