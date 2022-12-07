package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"src/common"
	"src/timeline"
)


func main() {
	port := flag.Int64("port", 4000, "The port of this host")
	bootstrap := flag.String("bootstrap", "", "The bootstrapping file")
	storage := flag.String("storage", "", "The directory where program files are stored")
	flag.Parse()

	if *storage == "" {
		*storage = fmt.Sprintf("timeline/%d", *port)
	}

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

	_, host := common.StartDHT(ctx, *port, bootstrapNodes)

	hostID := host.ID()
	logger.Printf("Created Node at: %s/p2p/%s", host.Addrs()[0].String(), hostID)
	logger.Printf("Node ID: %s", hostID)

	defer func() {
		if err := host.Close(); err != nil {
			panic(err)
		}
	}()

	/* go func(kad *dht.IpfsDHT) {
		ticker := time.NewTicker(2000 * time.Millisecond)

		for {
			select {
			case <-ticker.C:
				logger.Println("ROUTING TABLE:")
				kad.RoutingTable().Print()
			}
		}
	}(kad) */

	storedTimeline := timeline.CreateOrReadTimeline(*storage)

	inputScanner := bufio.NewScanner(os.Stdin)

	fmt.Println("Waiting for input")
	for inputScanner.Scan() {
		line := inputScanner.Text()

		fmt.Println("current posts")
		for _, post := range storedTimeline.Posts {
			timeline.PrintPost(post)
		}

		switch line {
		case "post":
			fmt.Println("Post text?")
			inputScanner.Scan()
			text := inputScanner.Text()

			storedTimeline.AddPost(text)
		case "exit":
			fmt.Println("Exiting Application")
			os.Exit(0)
		}

		fmt.Println("Waiting for input")
	}
}
