package service

import (
	"context"
	"de-twit-go/src/common"
	"de-twit-go/src/postupdater"
	"de-twit-go/src/timeline"
	"de-twit-go/src/timelinepb"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/ipfs/go-cid"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"io"
	"log"
	"net/http"
	"path/filepath"
)

type Event struct {
	// Events are pushed to this channel by the main events-gathering routine
	Message chan string

	// New client connections
	NewClients chan chan string

	// Closed client connections
	ClosedClients chan chan string

	// Total client connections
	TotalClients map[chan string]bool
}

type ClientChan chan string

func NewServer() (event *Event) {
	event = &Event{
		Message:       make(chan string),
		NewClients:    make(chan chan string),
		ClosedClients: make(chan chan string),
		TotalClients:  make(map[chan string]bool),
	}

	go event.listen()

	return
}

func (stream *Event) serveHTTP() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Initialize client channel
		clientChan := make(ClientChan)

		// Send new connection to event server
		stream.NewClients <- clientChan

		defer func() {
			// Send closed connection to event server
			stream.ClosedClients <- clientChan
		}()

		c.Set("clientChan", clientChan)

		c.Next()
	}
}

func (stream *Event) listen() {
	for {
		select {
		// Add new available client
		case client := <-stream.NewClients:
			stream.TotalClients[client] = true
			log.Printf("Client added. %d registered clients", len(stream.TotalClients))

		// Remove closed client
		case client := <-stream.ClosedClients:
			delete(stream.TotalClients, client)
			close(client)
			log.Printf("Removed client. %d registered clients", len(stream.TotalClients))

		// Broadcast message to client
		case eventMsg := <-stream.Message:
			for clientMessageChan := range stream.TotalClients {
				clientMessageChan <- eventMsg
			}
		}
	}
}

type postRequest struct {
	Text string `json:"text" binding:"required"`
}

type HTTPServer struct {
	*gin.Engine
	ctx context.Context
}

type errorResponse struct {
	errorCode int
	reason    string
}

func (e *errorResponse) Error() string {
	return e.reason
}

func (r *HTTPServer) RegisterGetRouting(kad *dht.IpfsDHT) {
	r.GET("/routing/info", func(c *gin.Context) {
		kad.RoutingTable().Print()

		c.String(http.StatusOK, "")
	})
}

func (r *HTTPServer) RegisterPostFollow(
	nodeCid cid.Cid,
	storage string,
	kad *dht.IpfsDHT,
	followingTimelines *timeline.FollowingTimelines,
	postUpdater *postupdater.PostUpdater,
	httpHandler func(post *timelinepb.Post),
) {
	logger := common.GetLogger(r.ctx)
	host := kad.Host()

	r.POST("/:user/follow", func(c *gin.Context) {
		user := c.Param("user")

		targetCid, err := common.GenerateCid(r.ctx, user)
		if err != nil {
			logger.Println("PostFollow: Couldn't Generate content id: ", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "couldn't generate content id for username"})
			return
		}

		if targetCid == nodeCid {
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "can't follow own profile"})
			return
		}

		receivedTimeline, resErr := func() (*timeline.Timeline, *errorResponse) {
			followingTimelines.Lock()
			defer followingTimelines.Unlock()

			if common.Contains(followingTimelines.FollowingCids, targetCid) {
				return nil, &errorResponse{errorCode: http.StatusUnprocessableEntity, reason: "already following"}
			}

			receivedTimeline, err := Follow(r.ctx, targetCid, host, kad)
			if err != nil {
				logger.Println("PostFollow: Couldn't Follow: ", err.Error())
				return nil, &errorResponse{errorCode: http.StatusInternalServerError, reason: err.Error()}
			}

			receivedTimeline.Path = filepath.Join(storage, fmt.Sprintf("storage-%s", user))

			err = receivedTimeline.WriteFile()
			if err != nil {
				logger.Println("PostFollow: Couldn't Write Timeline: ", err.Error())
				return nil, &errorResponse{errorCode: http.StatusInternalServerError, reason: err.Error()}
			}

			followingTimelines.FollowingCids = append(followingTimelines.FollowingCids, targetCid)
			followingTimelines.Timelines[targetCid] = receivedTimeline

			return receivedTimeline, nil
		}()
		if resErr != nil {
			c.JSON(resErr.errorCode, gin.H{"error": resErr.reason})
			return
		}

		// after follow, peers should be connected, so they belong on the same pub subnetwork
		err = postUpdater.ListenOnFollowingTopic(user, followingTimelines, httpHandler)
		if err != nil {
			logger.Println("PostFollow: Couldn't Listen on topic", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		err = kad.Provide(r.ctx, targetCid, true)
		if err != nil {
			logger.Println("PostFollow: Couldn't Provide", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		posts := make([]string, 0)
		for _, post := range receivedTimeline.Posts {
			posts = append(posts, fmt.Sprintf("%s: %s", post.GetLastUpdated().String(), post.GetText()))
		}

		c.JSON(http.StatusOK, posts)
	})
}

func (r *HTTPServer) RegisterPostUnfollow(
	followingTimelines *timeline.FollowingTimelines,
	postUpdater *postupdater.PostUpdater,
) {
	logger := common.GetLogger(r.ctx)

	r.POST("/:user/unfollow", func(c *gin.Context) {
		user := c.Param("user")

		targetCid, err := common.GenerateCid(r.ctx, user)
		if err != nil {
			logger.Println("PostFollow: Couldn't Generate content id: ", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "couldn't generate content id for username"})
			return
		}

		resErr := func() *errorResponse {
			followingTimelines.Lock()
			defer followingTimelines.Unlock()

			targetIndex := common.FindIndex(followingTimelines.FollowingCids, targetCid)
			if targetIndex == -1 {
				return &errorResponse{errorCode: http.StatusUnprocessableEntity, reason: "not following"}
			}

			err := postUpdater.StopListeningTopic(user)
			if err != nil {
				return &errorResponse{errorCode: http.StatusBadRequest, reason: err.Error()}
			}

			targetTimeline := followingTimelines.Timelines[targetCid]

			delete(followingTimelines.Timelines, targetCid)
			followingTimelines.FollowingCids = common.RemoveIndex(followingTimelines.FollowingCids, targetIndex)

			err = targetTimeline.DeleteFile()
			if err != nil {
				logger.Println("PostUnfollow: Couldn't delete timeline ", err)
				return &errorResponse{errorCode: http.StatusInternalServerError, reason: "couldn't delete timeline"}
			}

			return nil
		}()
		if resErr != nil {
			c.JSON(resErr.errorCode, gin.H{"error": resErr.reason})
			return
		}

		c.String(http.StatusOK, "")
	})
}

func (r *HTTPServer) RegisterPostCreate(username string, storedTimeline *timeline.OwnTimeline) {
	logger := common.GetLogger(r.ctx)

	r.POST("/post/create", func(c *gin.Context) {
		var json postRequest

		if err := c.ShouldBindJSON(&json); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		storedTimeline.Lock()
		err := storedTimeline.AddPost(json.Text, username)
		if err != nil {
			storedTimeline.Unlock()
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		storedTimeline.Unlock()

		logger.Println("Current OwnTimeline: ")

		for _, post := range storedTimeline.Posts {
			logger.Println(post.Text)
			logger.Printf("Posted at %s", post.LastUpdated.String())
		}

		c.String(http.StatusOK, "")
	})
}

func (r *HTTPServer) RegisterGetTimeline(timelines *timeline.FollowingTimelines) gin.IRoutes {
	return r.GET("/timeline", func(c *gin.Context) {
		posts := make([]timelinepb.Post, 0)

		for _, currentTimeline := range timelines.Timelines {
			for _, post := range currentTimeline.Posts {
				posts = append(posts, *post)
			}
		}

		c.JSON(http.StatusOK, posts)
		return
	})
}

func (r *HTTPServer) RegisterGetTimelineStream(stream *Event) {
	r.GET("/timeline/stream", stream.serveHTTP(), func(c *gin.Context) {
		v, ok := c.Get("clientChan")
		if !ok {
			return
		}
		clientChan, ok := v.(ClientChan)
		if !ok {
			return
		}
		c.Stream(func(w io.Writer) bool {
			if msg, ok := <-clientChan; ok {
				c.SSEvent("message", msg)
				return true
			}
			return false
		})
	})
}

func NewHTTP(
	ctx context.Context,
) *HTTPServer {
	return &HTTPServer{
		Engine: gin.Default(),
		ctx:    ctx,
	}
}
