package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/ipfs/go-cid"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/network"
	"google.golang.org/protobuf/proto"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"src/common"
	"src/service"
	"src/timeline"
)

func main() {
	port := flag.Int64("port", 4000, "The port of this host")
	servePort := flag.Int64("serve", 5000, "The port used for http serving")
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

	var followingCids []cid.Cid

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

	ps, err := pubsub.NewGossipSub(ctx, host)
	if err != nil {
		logger.Fatalln(err)
	}
	topic, err := ps.Join(*username)
	if err != nil {
		logger.Fatalf(err.Error())
		return
	}

	storedTimeline, err := timeline.CreateOrReadTimeline(*storage, topic)
	if err != nil {
		logger.Fatalf(err.Error())
		return
	}

	nodeCid, err := common.GenerateCid(ctx, *username)
	if err != nil {
		logger.Fatalf(err.Error())
		return
	}

	err = kad.Provide(ctx, nodeCid, true)
	if err != nil {
		logger.Fatalf(err.Error())
		return
	}

	host.SetStreamHandler("/p2p/1.0.0", func(stream network.Stream) {
		rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))

		resp, err := rw.ReadBytes(0)
		if err != nil {
			logger.Println(err.Error())
		}

		cidResp := resp[:len(resp)-1]

		requestedCid, err := cid.Cast(cidResp)
		if err != nil {
			logger.Println(err.Error())
			stream.Close()
			return
		}

		var reply []byte

		if nodeCid == requestedCid || common.Contains(followingCids, requestedCid) {
			reply, err = proto.Marshal(&storedTimeline.Timeline)
			if err != nil {
				logger.Println("Failed to encode post:", err)
				return
			}
		} else {
			logger.Println(fmt.Sprintf("Node not following %s anymore", requestedCid))
			reply = []byte(fmt.Sprintf("%d-NOT-FOLLOWING", *port))
		}

		_, err = rw.Write(append(reply, 0))
		if err != nil {
			logger.Println(err.Error())
			return
		}

		err = rw.Flush()
		if err != nil {
			logger.Println(err.Error())
			return
		}
	})

	r := gin.Default()
	r.GET("/routing/info", func(c *gin.Context) {
		kad.RoutingTable().Print()

		c.String(http.StatusOK, "ok")
	})

	r.POST("/:user/follow", func(c *gin.Context) {
		user := c.Param("user")

		receivedTimeline, targetCid, err := service.Follow(ctx, host, kad, followingCids, user)

		if err != nil {
			c.String(http.StatusInternalServerError, err.Error())
			return
		}

		followingCids = append(followingCids, *targetCid)

		var posts []string
		for _, post := range receivedTimeline.Posts {
			posts = append(posts, fmt.Sprintf("%s: %s", post.GetLastUpdated().String(), post.GetText()))
		}

		c.JSON(http.StatusOK, posts)

		// TODO: SETUP PUB SUB
	})

	r.POST("/:user/unfollow", func(c *gin.Context) {
		user := c.Param("user")

		err := service.Unfollow(ctx, &followingCids, user)

		if err != nil {
			c.String(http.StatusInternalServerError, err.Error())
			return
		}

		c.String(http.StatusOK, "ok")

		// TODO: DISCONNECT PUB SUB
	})

	r.POST("/post/create", func(c *gin.Context) {
		var json PostRequest

		if err = c.ShouldBindJSON(&json); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		storedTimeline.AddPost(json.Text)

		logger.Println("Current OwnTimeline: ")

		for _, post := range storedTimeline.Posts {
			logger.Println(post.Text)
			logger.Printf("Posted at %s", post.LastUpdated.String())
		}
	})

	err = r.Run(fmt.Sprintf(":%d", *servePort))
	if err != nil {
		logger.Fatalf(err.Error())
		return
	}
}

type PostRequest struct {
	Text string `json:"text" binding:"required"`
}
