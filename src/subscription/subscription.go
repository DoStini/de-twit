package subscription

import (
	"context"
	"fmt"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"google.golang.org/protobuf/proto"
	"log"
	"os"
	pb "src/timelinepb"
	"sync"
)

const UpdateBufferSize = 128

type subscriptionMap struct {
	sync.RWMutex
	m map[string]chan struct{}
}

type PubSubUpdate struct {
	PubS *pubsub.PubSub
	UserTopic *pubsub.Topic
	updateChan chan *pb.Post
	subscriptions subscriptionMap
	self peer.ID
	ctx context.Context
}

func (psu *PubSubUpdate) handleEvents() {

}

func (psu *PubSubUpdate) ListenOnTopic(topic string) error {
	subTopic, err := psu.PubS.Join(fmt.Sprintf("%s", topic))
	if err != nil {
		return err
	}
	subscription, err := subTopic.Subscribe()
	if err != nil {
		return err
	}

	cancelChannel := make(chan struct{})

	psu.subscriptions.Lock()
	psu.subscriptions.m[topic] = cancelChannel
	psu.subscriptions.Unlock()

	go psu.listenOnTopic(subscription, cancelChannel)

	return nil
}

func (psu *PubSubUpdate) listenOnTopic(subscription *pubsub.Subscription, cancelChannel chan struct{}) {
	var logger *log.Logger
	logger = psu.ctx.Value("logger").(*log.Logger)
	if logger == nil {
		logger = log.New(os.Stdin, "listen: ", log.Ltime|log.Lshortfile)
	}

	for {
		select {
		case <- cancelChannel:
			return
		default:
			message, err := subscription.Next(psu.ctx)
			if err != nil {
				logger.Println(err)
				return
			}
			// only forward messages delivered by others
			if message.ReceivedFrom == psu.self {
				continue
			}

			post := new(pb.Post)
			err = proto.Unmarshal(message.Data, post)
			if err != nil {
				logger.Println(err)
				continue
			}
			logger.Println(post.Text)

			// send valid messages onto the Messages channel
			psu.updateChan <- post
		}
	}
}

func MakePubSub(ctx context.Context, h host.Host, username string) (*PubSubUpdate, error) {
	ps, err := pubsub.NewGossipSub(ctx, h)
	if err != nil {
		return nil, err
	}

	ut, err := ps.Join(username)
	if err != nil {
		return nil, err
	}

	return &PubSubUpdate{
		PubS: ps,
		UserTopic: ut,
		updateChan: make(chan *pb.Post, UpdateBufferSize),
		subscriptions: subscriptionMap{m: make(map[string]chan struct{})},
		self: h.ID(),
		ctx: ctx,
	}, nil
}
