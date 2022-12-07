package main

import (
	"bufio"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"flag"
	"fmt"
	"github.com/ipfs/go-cid"
	"github.com/multiformats/go-multicodec"
	multihash "github.com/multiformats/go-multihash/core"
	"log"
	"os"
	"time"
)


func main() {
	port := flag.Int64("port", 4000, "The port of this host")
	bootstrap := flag.String("bootstrap", "", "The bootstrapping file")
	flag.Parse()

	logFile, err := os.OpenFile(fmt.Sprintf("logs/log-%d.log", *port), os.O_CREATE | os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf(err.Error())
	}
	logger = log.New(logFile, fmt.Sprintf("node:%d  |  ", *port), log.Ltime | log.Lshortfile)

	defer logFile.Close()

	ctx, cancel := context.WithCancel(context.Background())
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

	kad, host := startDHT(ctx, *port, bootstrapNodes)

	hostID := host.ID()
	logger.Printf("Created Node at: %s/p2p/%s", host.Addrs()[0].String(), hostID)
	logger.Printf("Node ID: %s", hostID)

	defer func() {
		if err := host.Close(); err != nil {
			panic(err)
		}
	}()

	if *port >= 4000 && *port < 4500{
		logger.Println("testput")

		hash := sha256.New()
		hash.Write([]byte("key"))
		key := hash.Sum(nil)

		err := kad.PutValue(ctx, hex.EncodeToString(key), []byte("stuff"))
		if err != nil {
			return
		}
	} else if *port >= 4500  && *port < 5000 {
		logger.Println("testprovide")

		pref := cid.Prefix{
			Version: 1,
			Codec: uint64(multicodec.Raw),
			MhType: multihash.SHA2_256,
			MhLength: -1, // default length
		}
		c, err := pref.Sum([]byte("key2"))
		if err != nil {
			logger.Fatalf(err.Error())
		}

		err = kad.Provide(ctx, c, true)
		if err != nil {
			return
		}
	} else if *port == 6000 {
		logger.Println("test")
		hash := sha256.New()
		hash.Write([]byte("key"))
		key := hash.Sum(nil)

		start := time.Now()
		peers, err := kad.GetClosestPeers(ctx, hex.EncodeToString(key))
		elapsed := time.Since(start)
		logger.Printf("CLOSEST PEERS took %s", elapsed)

		if err != nil {
			return
		}
		logger.Println(peers)
	} else if *port == 6001 {
		logger.Printf("Finding providers")

		pref := cid.Prefix{
			Version: 1,
			Codec: uint64(multicodec.Raw),
			MhType: multihash.SHA2_256,
			MhLength: -1, // default length
		}
		c, err := pref.Sum([]byte("key2"))
		if err != nil {
			logger.Fatalf(err.Error())
		}


		logger.Printf("Finding providers")
		start := time.Now()
		peers, err := kad.FindProviders(ctx, c)
		elapsed := time.Since(start)
		logger.Printf("FIND PROVIDERS took %s", elapsed)

		if err != nil {
			return
		}
		logger.Println(peers)
	}


	ticker := time.NewTicker(2000 * time.Millisecond)

	for {
		select {
		case <- ticker.C:
			logger.Println("ROUTING TABLE:")
			kad.RoutingTable().Print()
		}
	}
}
