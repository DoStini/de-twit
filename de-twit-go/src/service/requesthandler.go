package service

import (
	"context"
	"de-twit-go/src/common"
	"de-twit-go/src/postupdater"
	"de-twit-go/src/timeline"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/ipfs/go-cid"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"net/http"
	"path/filepath"
)

type postRequest struct {
	Text string `json:"text" binding:"required"`
}

type HTTPServer struct {
	*gin.Engine
	ctx context.Context
}

func (r *HTTPServer) RegisterGetRouting(kad *dht.IpfsDHT) {
	r.GET("/routing/info", func(c *gin.Context) {
		kad.RoutingTable().Print()

		c.String(http.StatusOK, "ok")
	})
}

func (r *HTTPServer) RegisterPostFollow(
	nodeCid cid.Cid,
	storage string,
	kad *dht.IpfsDHT,
	followingTimelines *timeline.FollowingTimelines,
	postUpdater *postupdater.PostUpdater,
) {
	logger := common.GetLogger(r.ctx)
	host := kad.Host()

	r.POST("/:user/follow", func(c *gin.Context) {
		user := c.Param("user")

		targetCid, err := common.GenerateCid(r.ctx, user)
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

			receivedTimeline, err := Follow(r.ctx, targetCid, host, kad)

			if err != nil {
				logger.Println(err.Error())
				c.String(http.StatusInternalServerError, err.Error())
				return nil, err
			}

			receivedTimeline.Path = filepath.Join(storage, fmt.Sprintf("storage-%s", user))
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
		err = postUpdater.ListenOnFollowingTopic(user, followingTimelines)
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
}

func NewHTTP(
	ctx context.Context,
) *HTTPServer {
	return &HTTPServer{
		Engine: gin.Default(),
		ctx:    ctx,
	}
}
