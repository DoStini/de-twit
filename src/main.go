package main

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"src/common"
	"src/postupdater"
	"src/service"
	"src/timeline"
	pb "src/timelinepb"
)

type InputCommands struct {
	port int64
	serverPort int64
	bootstrap string
	storage string
	username string
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

	kad, host, err := common.StartDHT(ctx, inputCommands.port, bootstrapNodes)
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

	// TODO: MOVE TO GOROUTINE
	for _, followingCid := range followingTimelines.FollowingCids {
		err := kad.Provide(ctx, followingCid, true)
		if err != nil {
			logger.Fatalf(err.Error())
			return
		}
	}

	followingTimelines.Timelines[nodeCid] = &storedTimeline.Timeline

	service.RegisterStreamHandler(ctx, host, nodeCid, followingTimelines)

	r := gin.Default()
	r.GET("/routing/info", func(c *gin.Context) {
		kad.RoutingTable().Print()

		c.String(http.StatusOK, "ok")
	})

	r.POST("/:user/follow", func(c *gin.Context) {
		user := c.Param("user")

		targetCid, err := common.GenerateCid(ctx, user)
		if err != nil {
			c.String(http.StatusInternalServerError, "Can't generate content id for username")
			return
		}

		if targetCid == nodeCid {
			c.String(http.StatusUnprocessableEntity, "Can't follow own profile")
			return
		}

		receivedTimeline, err := func() (*timeline.Timeline, error) {
			followingTimelines.Lock()
			defer followingTimelines.Unlock()

			if common.Contains(followingTimelines.FollowingCids, targetCid) {
				c.String(http.StatusUnprocessableEntity, "Already following")
				return nil, errors.New("already following")
			}

			receivedTimeline, err := service.Follow(ctx, targetCid, host, kad)

			if err != nil {
				logger.Println(err.Error())
				c.String(http.StatusInternalServerError, err.Error())
				return nil, err
			}

			receivedTimeline.Path = filepath.Join(inputCommands.storage, fmt.Sprintf("storage-%s", user))
			err = receivedTimeline.WriteFile()

			if err != nil {
				logger.Println(err.Error())
				c.String(http.StatusInternalServerError, err.Error())
				return nil, err
			}

			followingTimelines.FollowingCids = append(followingTimelines.FollowingCids, targetCid)
			followingTimelines.Timelines[targetCid] = receivedTimeline

			return receivedTimeline, nil
		}()
		if err != nil {
			return
		}

		// after follow, peers should be connected, so they belong on the same pub subnetwork
		err = postUpdater.ListenOnTopic(user, func(postUpdate *pb.Post) {
			logger.Printf("Hey baby, new post from %s just dropped!\n", postUpdate.User)
			logger.Println(postUpdate.Text)

			targetCid, err := common.GenerateCid(ctx, postUpdate.User)
			if err != nil {
				logger.Printf("Couldn't process message: %s\n", err)
				return
			}

			err = func() error {
				followingTimelines.RLock()
				defer followingTimelines.RUnlock()

				if !common.Contains(followingTimelines.FollowingCids, targetCid) {
					return errors.New("not following")
				}
				targetTimeline := followingTimelines.Timelines[targetCid]
				err := targetTimeline.AddPost(postUpdate.Id, postUpdate.Text, postUpdate.User, postUpdate.LastUpdated)
				if err != nil {
					return err
				}

				return nil
			}()
			if err != nil {
				logger.Printf("Couldn't process message: %s\n", err)
				return
			}
		})
		if err != nil {
			logger.Println(err)
			c.String(http.StatusInternalServerError, "%s", err)
			return
		}

		posts := make([]string, 0)
		for _, post := range receivedTimeline.Posts {
			posts = append(posts, fmt.Sprintf("%s: %s", post.GetLastUpdated().String(), post.GetText()))
		}

		c.JSON(http.StatusOK, posts)
	})

	r.POST("/:user/unfollow", func(c *gin.Context) {
		user := c.Param("user")

		targetCid, err := common.GenerateCid(ctx, user)
		if err != nil {
			c.String(http.StatusInternalServerError, "Can't generate content id for username")
			return
		}

		err = func() error {
			followingTimelines.Lock()
			defer followingTimelines.Unlock()

			targetIndex := common.FindIndex(followingTimelines.FollowingCids, targetCid)
			if targetIndex == -1 {
				c.String(http.StatusUnprocessableEntity, "Not following")
				return errors.New("not following")
			}

			err := postUpdater.StopListeningTopic(user)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return err
			}

			targetTimeline := followingTimelines.Timelines[targetCid]

			delete(followingTimelines.Timelines, targetCid)
			followingTimelines.FollowingCids = common.RemoveIndex(followingTimelines.FollowingCids, targetIndex)

			err = targetTimeline.DeleteFile()
			if err != nil {
				logger.Println(err.Error())
				c.String(http.StatusInternalServerError, err.Error())
				return err
			}

			return nil
		}()

		c.String(http.StatusOK, "")
		// TODO: DISCONNECT PUB SUB
	})

	r.POST("/post/create", func(c *gin.Context) {
		var json PostRequest

		if err = c.ShouldBindJSON(&json); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		storedTimeline.Lock()
		err := storedTimeline.AddPost(json.Text, inputCommands.username)
		if err != nil {
			storedTimeline.Unlock()
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		storedTimeline.Unlock()

		logger.Println("Current OwnTimeline: ")

		for _, post := range storedTimeline.Posts {
			logger.Println(post.Text)
			logger.Printf("Posted at %s", post.LastUpdated.String())
		}
	})

	err = r.Run(fmt.Sprintf(":%d", inputCommands.serverPort))
	if err != nil {
		logger.Fatalf(err.Error())
		return
	}
}

type PostRequest struct {
	Text string `json:"text" binding:"required"`
}
