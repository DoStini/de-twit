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

func PrintPost(post *pb.Post) {
	fmt.Println("POST ----")
	fmt.Println(post.Id)
	fmt.Println(post.Text)
	fmt.Println(post.LastUpdated.String())
}

func CreateOrReadTimeline(storagePath string, topic *pubsub.Topic) *Timeline {
	var storedTimeline Timeline
	path := filepath.Join(storagePath, "storage")

	timelineBytes, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		err := os.MkdirAll(storagePath, os.ModePerm)
		if err != nil {
			log.Fatalln("Error creating folders: ", err)
		}

		timelineFile, err := os.OpenFile(path, os.O_CREATE | os.O_WRONLY, 0644)
		if err != nil {
			log.Fatalln("Error creating storage file: ", err)
		}

		// TODO: GET TIMELINE FROM SUBSCRIBERS, AND TIMELINE OF SUBSCRIBED
		storedTimeline = Timeline{Timeline: pb.Timeline{Posts: []*pb.Post{} }}
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
	} else if err == nil {
		// TODO: POSSIBLY, MERGE SUBSCRIBED TIMELINES FROM OTHER SUBSCRIBERS
		err := proto.Unmarshal(timelineBytes, &storedTimeline.Timeline)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		log.Fatalf(err.Error())
	}

	storedTimeline.path = path
	storedTimeline.topic = topic
	return &storedTimeline
}

type Timeline struct {
	pb.Timeline
	path string
	topic *pubsub.Topic
}

func (t *Timeline) AddPost(text string) {
	post := pb.Post{
		Text:        text,
		Id:          fmt.Sprintf("%d", len(t.Posts)),
		LastUpdated: timestamppb.Now(),
	}

	t.Posts = append(t.Posts, &post)

	out, err := proto.Marshal(&post)
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
