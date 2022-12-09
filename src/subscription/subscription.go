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

type postUpdate struct {
	post *pb.Post
	user string
}

type subscriptionMap struct {
	sync.RWMutex
	m map[string]*pubsub.Subscription
}

type PubSubUpdate struct {
	PubS *pubsub.PubSub
	UserTopic *pubsub.Topic
	updateChan chan *postUpdate
	subscriptions subscriptionMap
	self peer.ID
	ctx context.Context
}

// this is just to make handling post updates run sequentially
func (psu *PubSubUpdate) handleEvents() {
	var logger *log.Logger
	logger = psu.ctx.Value("logger").(*log.Logger)
	if logger == nil {
		logger = log.New(os.Stdin, "listen: ", log.Ltime|log.Lshortfile)
	}

	for {
		select {
		case postUpdate := <- psu.updateChan:
			logger.Printf("Wake up, new post from %s just dropped!", postUpdate.user)
			logger.Println(postUpdate.post.Text)
		}
	}
}

func (psu *PubSubUpdate) StopListeningTopic(topic string) {
	psu.subscriptions.RLock()
	subscription := psu.subscriptions.m[topic]
	psu.subscriptions.RUnlock()

	if subscription == nil {
		return
	}

	subscription.Cancel()
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

	psu.subscriptions.Lock()
	psu.subscriptions.m[topic] = subscription
	psu.subscriptions.Unlock()

	go psu.listenOnTopic(subscription)
	return nil
}

func (psu *PubSubUpdate) listenOnTopic(subscription *pubsub.Subscription) {
	var logger *log.Logger
	logger = psu.ctx.Value("logger").(*log.Logger)
	if logger == nil {
		logger = log.New(os.Stdin, "listen: ", log.Ltime|log.Lshortfile)
	}

	for {
			message, err := subscription.Next(psu.ctx)
			if err != nil {
				logger.Println("My darling, I have died and gone to heaven.")
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

			// send valid messages onto the Messages channel
			psu.updateChan <- &postUpdate{post: post, user: subscription.Topic()}
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

	psu := &PubSubUpdate{
		PubS: ps,
		UserTopic: ut,
		updateChan: make(chan *postUpdate, UpdateBufferSize),
		subscriptions: subscriptionMap{m: make(map[string]*pubsub.Subscription)},
		self: h.ID(),
		ctx: ctx,
	}

	go psu.handleEvents()
	return psu, nil
}
