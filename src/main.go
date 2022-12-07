package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"src/common"
	"time"
)

func main() {
	port := flag.Int64("port", 4000, "The port of this host")
	bootstrap := flag.String("bootstrap", "", "The bootstrapping file")
	flag.Parse()

	logFile, err := os.OpenFile(fmt.Sprintf("logs/log-%d.log", *port), os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf(err.Error())
	}
	logger = log.New(logFile, fmt.Sprintf("node:%d  |  ", *port), log.Ltime|log.Lshortfile)

	defer logFile.Close()

	ctx, cancel := context.WithCancel(context.Background())
	ctx = context.WithValue(ctx, "logger", logger)

	defer cancel()

	f, err := os.OpenFile(*bootstrap, os.O_RDONLY, 0644)
	if err != nil {
		logger.Fatalf(err.Error())
	}

	var bootstrapNodes []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		s := scanner.Text()
		bootstrapNodes = append(bootstrapNodes, s)
	}

	err = f.Close()
	if err != nil {
		logger.Fatalf(err.Error())
	}

	kad, host := common.StartDHT(ctx, *port, bootstrapNodes)

	hostID := host.ID()
	logger.Printf("Created Node at: %s/p2p/%s", host.Addrs()[0].String(), hostID)
	logger.Printf("Node ID: %s", hostID)

	defer func() {
		if err := host.Close(); err != nil {
			panic(err)
		}
	}()

	ticker := time.NewTicker(2000 * time.Millisecond)

	for {
		select {
		case <-ticker.C:
			logger.Println("ROUTING TABLE:")
			kad.RoutingTable().Print()
		}
	}
}
