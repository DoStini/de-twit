package timeline

import (
	"context"
	"de-twit-go/src/common"
	dht2 "de-twit-go/src/dht"
	pb "de-twit-go/src/timelinepb"
	"errors"
	"fmt"
	"github.com/ipfs/go-cid"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/peer"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
)

type PB = pb.Timeline

type Timeline struct {
	PB
	Path string
}

type OwnTimeline struct {
	Timeline
	sync.RWMutex
	topic *pubsub.Topic
}

type FollowingTimelines struct {
	sync.RWMutex
	Timelines      map[cid.Cid]*Timeline
	FollowingCids  []cid.Cid
	FollowingNames []string
}

func CreateOrReadTimeline(storagePath string, topic *pubsub.Topic) (*OwnTimeline, error) {
	path := filepath.Join(storagePath, "storage")

	timelineBytes, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return CreateTimeline(storagePath, path, topic), nil
	}
	if err != nil {
		return nil, err
	}

	storedTimeline := new(OwnTimeline)

	err = ReadTimelinePb(timelineBytes, &storedTimeline.PB)
	if err != nil {
		return nil, err
	}

	storedTimeline.topic = topic
	storedTimeline.Path = path

	return storedTimeline, nil
}

func UpdateTimeline(ctx context.Context, cid cid.Cid, kad *dht.IpfsDHT) {
	// TODO: RIGHT NOW, ALL THAT IS DONE IS JUST CONNECTING TO PROVIDER
	dht2.HandleWithProviders(ctx, cid, kad, func(info peer.AddrInfo) error {
		return nil
	})
}

func ReadFollowingTimelines(ctx context.Context, storagePath string) (*FollowingTimelines, error) {
	followingTimelines := &FollowingTimelines{
		Timelines:      make(map[cid.Cid]*Timeline),
		FollowingCids:  make([]cid.Cid, 0),
		FollowingNames: make([]string, 0),
	}

	err := filepath.Walk(storagePath, func(path string, info fs.FileInfo, err error) error {
		if info.Name() == "storage" {
			return nil
		}

		if info.IsDir() {
			return nil
		}

		parts := strings.Split(info.Name(), "-")

		if len(parts) != 2 {
			return nil
		}

		fileCid, err := common.GenerateCid(ctx, parts[1])
		if err != nil {
			return err
		}

		followingTimelines.FollowingCids = append(followingTimelines.FollowingCids, fileCid)
		followingTimelines.FollowingNames = append(followingTimelines.FollowingNames, parts[1])
		storedTimeline := Timeline{Path: path}

		err = ReadTimelinePbFromFile(path, &storedTimeline.PB)
		if err != nil {
			return err
		}

		followingTimelines.Timelines[fileCid] = &storedTimeline

		return nil
	})

	if err != nil {
		return nil, err
	}

	return followingTimelines, nil
}

func ReadTimelinePbFromFile(path string, buffer *PB) error {
	timelineBytes, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	return ReadTimelinePb(timelineBytes, buffer)
}

func ReadTimelinePb(timelineBytes []byte, buffer *PB) error {
	// TODO: POSSIBLY, MERGE SUBSCRIBED TIMELINES FROM OTHER SUBSCRIBERS
	err := proto.Unmarshal(timelineBytes, buffer)
	if err != nil {
		return err
	}

	return nil
}

func CreateTimeline(storagePath string, path string, topic *pubsub.Topic) *OwnTimeline {
	err := os.MkdirAll(storagePath, os.ModePerm)
	if err != nil {
		log.Fatalln("Error creating folders: ", err)
	}

	timelineFile, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalln("Error creating storage file: ", err)
	}

	// TODO: GET TIMELINE FROM SUBSCRIBERS, AND TIMELINE OF SUBSCRIBED
	storedTimeline := new(OwnTimeline)
	storedTimeline.Posts = []*pb.Post{}
	storedTimeline.Path = path
	storedTimeline.topic = topic

	out, err := proto.Marshal(&storedTimeline.Timeline)
	if err != nil {
		log.Fatalln("Error marshalling storage: ", err)
	}

	_, err = timelineFile.Write(out)
	if err != nil {
		log.Fatalln("Error writing to storage file: ", err)
	}

	err = timelineFile.Close()
	if err != nil {
		log.Println(err)
	}
	return storedTimeline
}

func (t *Timeline) DeleteFile() error {
	return os.Remove(t.Path)
}

func (t *Timeline) WriteFile() error {
	timelineFile, err := os.OpenFile(t.Path, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	defer func(timelineFile *os.File) {
		err := timelineFile.Close()
		if err != nil {

		}
	}(timelineFile)

	out, err := proto.Marshal(t)
	if err != nil {
		return err
	}

	_, err = timelineFile.Write(out)
	return err
}

func (t *Timeline) addPost(post *pb.Post) error {
	t.Posts = append(t.Posts, post)

	out, err := proto.Marshal(t)
	if err != nil {
		return err
	}

	return os.WriteFile(t.Path, out, 0644)
}

func (t *Timeline) AddPost(id string, text string, user string, lastUpdated *timestamppb.Timestamp) error {
	post := pb.Post{
		Text:        text,
		Id:          id,
		User:        user,
		LastUpdated: lastUpdated,
	}

	return t.addPost(&post)
}

func (t *OwnTimeline) AddPost(text string, user string) error {
	post := pb.Post{
		Text:        text,
		Id:          fmt.Sprintf("%d", len(t.Posts)),
		User:        user,
		LastUpdated: timestamppb.Now(),
	}

	err := t.Timeline.addPost(&post)
	if err != nil {
		return err
	}

	out, err := proto.Marshal(&post)
	if err != nil {
		return err
	}

	out, err = proto.Marshal(&post)
	if err != nil {
		log.Fatalln("Failed to encode post:", err)
	}
	err = t.topic.Publish(context.Background(), out)
	if err != nil {
		return err
	}

	return nil
}

type postID struct {
	id   string
	user string
}

func MergeTimelines(t *PB, timelines []*Timeline) error {
	contains := make(map[postID]int)

	for _, curTimeline := range timelines {
		for _, post := range curTimeline.Posts {
			id := postID{user: post.User, id: post.Id}

			if val, ok := contains[id]; ok && post.LastUpdated.AsTime().After(t.Posts[val].LastUpdated.AsTime()) {
				t.Posts[val] = post
			} else if !ok {
				contains[id] = len(t.Posts)
				t.Posts = append(t.Posts, post)
			}
		}
	}

	sort.SliceStable(t.Posts, func(i, j int) bool {
		return t.Posts[i].LastUpdated.AsTime().After(t.Posts[j].LastUpdated.AsTime())
	})

	return nil
}
