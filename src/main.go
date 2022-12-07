package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"log"
	"os"
	"path/filepath"
	"src/common"
	"src/timeline"
)


func main() {
	port := flag.Int64("port", 4000, "The port of this host")
	bootstrap := flag.String("bootstrap", "", "The bootstrapping file")
	storage := flag.String("storage", "", "The directory where program files are stored")
	username := flag.String("username", "", "The port of this host")
	flag.Parse()

	if *username == "" {
		flag.Usage()
		os.Exit(1)
	}
	if *storage == "" {
		*storage = filepath.Join("storage", fmt.Sprintf("%s", *username))
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

	ps, err := pubsub.NewGossipSub(ctx, host)
	if err != nil {
		logger.Fatalln(err)
	}
	topic, err := ps.Join(*username)
	if err != nil {
		return
	}

	storedTimeline := timeline.CreateOrReadTimeline(*storage, topic)
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
