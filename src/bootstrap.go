package main

import (
	"context"
	"flag"
	"fmt"
)

func main() {
	port := flag.Int64("port", 4000, "The port of this host")
	bootstrap := flag.String("bootstrap", "", "The bootstrapping node")
	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	h := createNode(4001)

	// otherKey := generatePrivateKey(4000)

	// print the node's listening addresses
	fmt.Println("Listen addresses:", h.Addrs())

	// shut the node down
	if err := h.Close(); err != nil {
		panic(err)
	}
}
