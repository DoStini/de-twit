package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/libp2p/go-libp2p/core/network"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"src/common"
	"src/postupdater"
	"src/service"
	"src/timeline"
)

func main() {
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

	host.SetStreamHandler("/p2p/1.0.0", func(stream network.Stream) {
		rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))

		_, err = rw.WriteString(fmt.Sprintf("resp from %d\n", *port))
		if err != nil {
			logger.Fatalf(err.Error())
			return
		}

		err = rw.Flush()
		if err != nil {
			logger.Fatalf(err.Error())
			return
		}
	})

	postUpdater, err := postupdater.NewPostUpdater(ctx, host, *username)
	if err != nil {
		logger.Fatalln(err)
	}

	storedTimeline := timeline.CreateOrReadTimeline(*storage, postUpdater.UserTopic)

	c, err := common.GenerateCid(ctx, *username)
	if err != nil {
		logger.Fatalf(err.Error())
		return
	}

	err = kad.Provide(ctx, c, true)
	if err != nil {
		logger.Fatalf(err.Error())
		return
	}

	r := gin.Default()
	r.GET("/routing/info", func(c *gin.Context) {
		kad.RoutingTable().Print()

		c.String(http.StatusOK, "ok")
	})

	r.POST("/:user/subscribe", func(c *gin.Context) {
		user := c.Param("user")

		posts, err := service.Follow(ctx, host, kad, user)

		if err != nil {
			c.String(http.StatusInternalServerError, err.Error())
			return
		}

		// after follow, peers should be connected, so they belong on the same pub subnetwork
		err = postUpdater.ListenOnTopic(user, func(postUpdate *postupdater.PostUpdate) {
			logger.Printf("Hey baby, new post from %s just dropped!\n", postUpdate.User)
			logger.Println(postUpdate.Post.Text)
		})
		if err != nil {
			logger.Println(err)
			c.String(http.StatusInternalServerError, "%s", err)

			return
		}

		c.JSON(http.StatusOK, posts)
	})

	r.POST("/:user/unfollow", func(c *gin.Context) {
		user := c.Param("user")
		err := postUpdater.StopListeningTopic(user)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{})
	})

	r.POST("/post/create", func(c *gin.Context) {
		var json PostRequest

		if err = c.ShouldBindJSON(&json); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		storedTimeline.AddPost(json.Text)

		logger.Println("Current Timeline: ")

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
