package main

import (
	timeline2 "de-twit-go/src/timeline"
	pb "de-twit-go/src/timelinepb"
	"fmt"
	"google.golang.org/protobuf/types/known/timestamppb"
	"log"
	"os"
	"path/filepath"
	"time"
)

func main() {
	path := filepath.Join("storage", "andre")

	err := os.MkdirAll(path, os.ModePerm)
	if err != nil {
		log.Fatalln("Error creating folders: ", err)
	}

	timeline := timeline2.Timeline{
		Path: filepath.Join(path, "storage-julio"),
	}
	timeline.Posts = make([]*pb.Post, 0, 75)

	for i := 0; i < 50; i++ {
		timeline.Posts = append(timeline.Posts, &pb.Post{
			Id:          fmt.Sprintf("%d", i),
			Text:        fmt.Sprintf("This is post %d at %d seconds", i, 0),
			User:        "julio",
			LastUpdated: timestamppb.New(time.Unix(0, int64(i))),
		})
	}
	for i := 50; i < 100; i += 2 {
		timeline.Posts = append(timeline.Posts, &pb.Post{
			Id:          fmt.Sprintf("%d", i),
			Text:        fmt.Sprintf("This is post %d at %d seconds", i, 1),
			User:        "julio",
			LastUpdated: timestamppb.New(time.Unix(1, int64(i))),
		})
	}
	err = timeline.WriteFile()
	log.Println(err)

	path = filepath.Join("storage", "nuno")

	err = os.MkdirAll(path, os.ModePerm)
	if err != nil {
		log.Fatalln("Error creating folders: ", err)
	}

	timeline = timeline2.Timeline{
		Path: filepath.Join(path, "storage-julio"),
	}
	timeline.Posts = make([]*pb.Post, 0, 75)

	for i := 0; i < 50; i += 2 {
		timeline.Posts = append(timeline.Posts, &pb.Post{
			Id:          fmt.Sprintf("%d", i),
			Text:        fmt.Sprintf("This is post %d at %d seconds", i, 1),
			User:        "julio",
			LastUpdated: timestamppb.New(time.Unix(1, int64(i))),
		})
	}
	for i := 50; i < 100; i++ {
		timeline.Posts = append(timeline.Posts, &pb.Post{
			Id:          fmt.Sprintf("%d", i),
			Text:        fmt.Sprintf("This is post %d at %d seconds", i, 0),
			User:        "julio",
			LastUpdated: timestamppb.New(time.Unix(0, int64(i))),
		})
	}
	err = timeline.WriteFile()
	log.Println(err)
}
