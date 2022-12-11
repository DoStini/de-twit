package main

import (
	"bufio"
	"context"
	"de-twit-go/src/common"
	"de-twit-go/src/dht"
	"de-twit-go/src/postupdater"
	"de-twit-go/src/service"
	"de-twit-go/src/timeline"
	pb "de-twit-go/src/timelinepb"
	json2 "encoding/json"
	"flag"
	"fmt"
	"github.com/gin-contrib/cors"
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

	hostCid, err := common.GenerateCid(ctx, inputCommands.username)
	if err != nil {
		logger.Fatalf(err.Error())
		return
	}

	followingTimelines, err := timeline.ReadFollowingTimelines(ctx, inputCommands.storage)
	if err != nil {
		logger.Fatalf(err.Error())
		return
	}

	serverSentStream := service.NewServer()
	serverSentStreamHandler := func(post *pb.Post) {
		json, err := json2.Marshal(post)
		if err != nil {
			logger.Println(err.Error())
			return
		}
		logger.Println(string(json))
		serverSentStream.Message <- string(json)
	}

	func() {
		posts, err := service.FindPosts(ctx, hostCid, kad)
		if err != nil {
			logger.Println(err.Error())
			return
		}

		timelines := make([]*pb.Timeline, 0)
		timelines = append(timelines, posts)
		timelines = append(timelines, &storedTimeline.PB)

		err = timeline.MergeTimelines(&storedTimeline.PB, timelines)
		if err != nil {
			logger.Println(err.Error())
			return
		}

		err = storedTimeline.WriteFile()
		if err != nil {
			logger.Println("PostFollow: Couldn't Write Timeline: ", err.Error())
			return
		}
	}()

	for idx, followingCid := range followingTimelines.FollowingCids {
		posts, err := service.FindPosts(ctx, followingCid, kad)
		if err != nil {
			logger.Println(err.Error())
			continue
		}
		timelines := make([]*pb.Timeline, 0)
		timelines = append(timelines, posts)
		timelines = append(timelines, &followingTimelines.Timelines[followingCid].PB)

		err = timeline.MergeTimelines(&followingTimelines.Timelines[followingCid].PB, timelines)
		if err != nil {
			logger.Println(err.Error())
			continue
		}

		err = postUpdater.ListenOnFollowingTopic(followingTimelines.FollowingNames[idx], followingTimelines, serverSentStreamHandler)
		if err != nil {
			logger.Println(err.Error())
			continue
		}

		err = followingTimelines.Timelines[followingCid].WriteFile()
		if err != nil {
			logger.Println("PostFollow: Couldn't Write Timeline: ", err.Error())
			return
		}
	}

	service.RegisterProvideRoutine(ctx, kad, followingTimelines, hostCid)
	followingTimelines.Timelines[hostCid] = &storedTimeline.Timeline

	service.RegisterStreamHandler(ctx, host, hostCid, followingTimelines)

	r, err := service.NewHTTP(ctx)
	if err != nil {
		logger.Fatalln(err)
	}

	r.Use(cors.New(cors.Config{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{"POST", "PUT", "PATCH", "DELETE"},
		AllowHeaders: []string{"Content-Type,access-control-allow-origin, access-control-allow-headers"},
	}))
	r.RegisterGetRouting(kad)
	r.RegisterPostFollow(hostCid, inputCommands.storage, kad, followingTimelines, postUpdater, serverSentStreamHandler)
	r.RegisterPostUnfollow(followingTimelines, postUpdater)
	r.RegisterPostCreate(inputCommands.username, storedTimeline)
	r.RegisterGetTimeline(followingTimelines)
	r.RegisterGetTimelineStream(serverSentStream)
	r.RegisterGetUser(ctx, followingTimelines, hostCid, kad)

	err = r.Run(fmt.Sprintf(":%d", inputCommands.serverPort))
	if err != nil {
		logger.Fatalln(err)
	}
}
