package main

import (
	"context"
	"fmt"
)

func main2() {
	_, cancel := context.WithCancel(context.Background())
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
