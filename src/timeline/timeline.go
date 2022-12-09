package timeline

import (
	"context"
	"errors"
	"fmt"
	"github.com/ipfs/go-cid"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"src/common"
	pb "src/timelinepb"
	"strings"
)

type PB = pb.Timeline

type Timeline struct {
	PB
	Path string
}

type OwnTimeline struct {
	Timeline
	topic *pubsub.Topic
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

func ReadFollowingTimelines(ctx context.Context, storagePath string) (map[cid.Cid]*Timeline, []cid.Cid, error) {
	timelines := make(map[cid.Cid]*Timeline)
	storedCids := make([]cid.Cid, 0)

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

		storedCids = append(storedCids, fileCid)
		storedTimeline := Timeline{Path: path}

		err = ReadTimelinePbFromFile(path, &storedTimeline.PB)
		if err != nil {
			return err
		}

		timelines[fileCid] = &storedTimeline

		return nil
	})

	if err != nil {
		return nil, nil, err
	}

	return timelines, storedCids, nil
}

func ReadTimelinePbFromFile(path string, buffer *PB) error {
	timelineBytes, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	return ReadTimelinePb(timelineBytes, buffer)
}

func ReadTimelinePb(timelineBytes []byte, buffer *PB) error {
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

	// TODO: GET TIMELINE FROM SUBSCRIBERS
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

func (t *Timeline) AddPost(text string) error {
	post := pb.Post{
		Text:        text,
		Id:          fmt.Sprintf("%d", len(t.Posts)),
		LastUpdated: timestamppb.Now(),
	}

	return t.addPost(&post)
}

func (t *OwnTimeline) AddPost(text string) {
	post := pb.Post{
		Text:        text,
		Id:          fmt.Sprintf("%d", len(t.Posts)),
		LastUpdated: timestamppb.Now(),
	}

	err := t.Timeline.addPost(&post)

	out, err := proto.Marshal(&post)

	err = t.topic.Publish(context.Background(), out)
	if err != nil {
		log.Println("Failed to publish: ", err)
	}
}
