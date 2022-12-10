package service

import (
	"context"
	"de-twit-go/src/common"
	"de-twit-go/src/postupdater"
	"de-twit-go/src/timeline"
	pb "de-twit-go/src/timelinepb"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/ipfs/go-cid"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

type postRequest struct {
	Text string `json:"text" binding:"required"`
}

func StartHTTP(
	ctx context.Context,
	kad *dht.IpfsDHT,
	nodeCid cid.Cid,
	storedTimeline *timeline.OwnTimeline,
	followingTimelines *timeline.FollowingTimelines,
	postUpdater *postupdater.PostUpdater,
	storage string, username string, serverPort int64,
) error {
	var logger *log.Logger
	logger = ctx.Value("logger").(*log.Logger)
	if logger == nil {
		logger = log.New(os.Stdin, "listen: ", log.Ltime|log.Lshortfile)
	}

	host := kad.Host()
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

			receivedTimeline, err := Follow(ctx, targetCid, host, kad)

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

	return r.Run(fmt.Sprintf(":%d", serverPort))
}
