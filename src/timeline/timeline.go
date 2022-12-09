package timeline

import (
	"context"
	"errors"
	"fmt"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
	"log"
	"os"
	"path/filepath"
	pb "src/timelinepb"
)

type timelinePB = pb.Timeline

type Timeline struct {
	timelinePB
	path string
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

	readTimeline, err := ReadTimelinePb(timelineBytes)
	if err != nil {
		return nil, err
	}

	storedTimeline := new(OwnTimeline)
	storedTimeline.timelinePB = *readTimeline
	storedTimeline.topic = topic
	storedTimeline.path = path

	return storedTimeline, nil
}

func ReadTimelinePb(timelineBytes []byte) (*timelinePB, error) {
	readTimeline := new(timelinePB)
	err := proto.Unmarshal(timelineBytes, readTimeline)
	if err != nil {
		return nil, err
	}

	return readTimeline, nil
}

//
//func ReadTimelineFromFile(storagePath string) (*Timeline, error) {
//
//}

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
	storedTimeline.path = path
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

func (t *OwnTimeline) AddPost(text string) {
	post := pb.Post{
		Text:        text,
		Id:          fmt.Sprintf("%d", len(t.Posts)),
		LastUpdated: timestamppb.Now(),
	}

	t.Posts = append(t.Posts, &post)

	out, err := proto.Marshal(&t.Timeline)
	if err != nil {
		log.Fatalln("Failed to encode post:", err)
	}
	if err := os.WriteFile(t.path, out, 0644); err != nil {
		log.Fatalln("Failed to write storage:", err)
	}

	err = t.topic.Publish(context.Background(), out)
	if err != nil {
		log.Println("Failed to publish: ", err)
	}
}
