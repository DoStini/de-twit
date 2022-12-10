package main

import (
	"bufio"
	"context"
	"de-twit-go/src/common"
	"de-twit-go/src/dht"
	"de-twit-go/src/postupdater"
	"de-twit-go/src/service"
	"de-twit-go/src/timeline"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

type InputCommands struct {
	port       int64
	serverPort int64
	bootstrap  string
	storage    string
	username   string
}

func parseCommands() InputCommands {
	port := flag.Int64("port", 4000, "The port of this host")
	servePort := flag.Int64("serve", 5000, "The port used for http serving")
	bootstrap := flag.String("bootstrap", "", "The bootstrapping file")
	storage := flag.String("storage", "", "The directory where program files are stored")
	username := flag.String("username", "", "The username")
	flag.Parse()

	if *username == "" {
		flag.Usage()
		os.Exit(1)
	}
	if *storage == "" {
		*storage = filepath.Join("storage", fmt.Sprintf("%s", *username))
	}

	return InputCommands{
		port:       *port,
		serverPort: *servePort,
		bootstrap:  *bootstrap,
		storage:    *storage,
		username:   *username,
	}
}

func main() {
	inputCommands := parseCommands()

	logFile, err := os.OpenFile(fmt.Sprintf("logs/log-%s.log", inputCommands.username), os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf(err.Error())
	}
	logger = log.New(logFile, fmt.Sprintf("node:%s  |  ", inputCommands.username), log.Ltime|log.Lshortfile)

	defer logFile.Close()

	ctx, cancel := context.WithCancel(context.Background())
	ctx = context.WithValue(ctx, "logger", logger)

	defer cancel()

	f, err := os.OpenFile(inputCommands.bootstrap, os.O_RDONLY, 0644)
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

	kad, host, err := dht.NewDHT(ctx, inputCommands.port, bootstrapNodes)
	if err != nil {
		logger.Fatalf("Error creating DHT: %s\n", err.Error())
	}

	hostID := host.ID()
	logger.Printf("Created Node at: %s/p2p/%s", host.Addrs()[0].String(), hostID)
	logger.Printf("Node ID: %s", hostID)

	defer func() {
		if err := host.Close(); err != nil {
			panic(err)
		}
	}()

	postUpdater, err := postupdater.NewPostUpdater(ctx, host, inputCommands.username)
	if err != nil {
		logger.Fatalln(err)
	}

	storedTimeline, err := timeline.CreateOrReadTimeline(inputCommands.storage, postUpdater.UserTopic)
	if err != nil {
		logger.Fatalf(err.Error())
		return
	}

	nodeCid, err := common.GenerateCid(ctx, inputCommands.username)
	if err != nil {
		logger.Fatalf(err.Error())
		return
	}

	err = kad.Provide(ctx, nodeCid, true)
	if err != nil {
		logger.Fatalf(err.Error())
		return
	}

	followingTimelines, err := timeline.ReadFollowingTimelines(ctx, inputCommands.storage)
	if err != nil {
		logger.Fatalf(err.Error())
		return
	}

	for idx, followingCid := range followingTimelines.FollowingCids {
		// TODO: UPDATE TIMELINES, OR START GOROUTINE (not sure if needed) THAT DOES AS SUCH
		// TODO: RIGHT NOW, ALL UPDATE TIMELINE DOES IS JUST CONNECT TO PEERS
		// TODO: SO THAT PUB SUB IS RECONNECTED
		timeline.UpdateTimeline(ctx, followingCid, kad)
		// END: UPDATE TIMELINES

		err := postUpdater.ListenOnFollowingTopic(followingTimelines.FollowingNames[idx], followingTimelines)
		if err != nil {
			logger.Println(err.Error())
			continue
		}

		// TODO: MOVE TO GOROUTINE
		err = kad.Provide(ctx, followingCid, true)
		if err != nil {
			logger.Fatalf(err.Error())
			return
		}
	}

	followingTimelines.Timelines[nodeCid] = &storedTimeline.Timeline

	service.RegisterStreamHandler(ctx, host, nodeCid, followingTimelines)
	r, err := service.NewHTTP(ctx)
	if err != nil {
		logger.Fatalln(err)
	}
	r.RegisterGetRouting(kad)
	r.RegisterPostFollow(nodeCid, inputCommands.storage, kad, followingTimelines, postUpdater)
	r.RegisterPostUnfollow(followingTimelines, postUpdater)
	r.RegisterPostCreate(inputCommands.username, storedTimeline)

	err = r.Run(fmt.Sprintf(":%d", inputCommands.serverPort))
	if err != nil {
		logger.Fatalln(err)
	}
}
