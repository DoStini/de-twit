package postupdater

import (
	"context"
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

type PostUpdate struct {
	Post *pb.Post
	User string
}

type Subscription struct {
	sub     *pubsub.Subscription
	topic   *pubsub.Topic
	handler func(*PostUpdate)
}

type subscriptionMap struct {
	sync.RWMutex
	m map[string]*Subscription
}

type PostUpdater struct {
	PubS          *pubsub.PubSub
	UserTopic     *pubsub.Topic
	updateChan    chan *PostUpdate
	subscriptions subscriptionMap
	self          peer.ID
	ctx           context.Context
}

// this is just to make handling post updates run sequentially
func (psu *PostUpdater) handleEvents() {
	for {
		select {
		case postUpdate := <-psu.updateChan:
			psu.subscriptions.RLock()
			subscription := psu.subscriptions.m[postUpdate.User]
			psu.subscriptions.RUnlock()

			subscription.handler(postUpdate)
		}
	}
}

func (psu *PostUpdater) StopListeningTopic(topic string) error {
	psu.subscriptions.RLock()
	subscription := psu.subscriptions.m[topic]
	psu.subscriptions.RUnlock()

	if subscription == nil {
		return nil
	}

	subscription.sub.Cancel()
	err := subscription.topic.Close()
	if err != nil {
		return err
	}

	psu.subscriptions.Lock()
	delete(psu.subscriptions.m, topic)
	psu.subscriptions.Unlock()

	return nil
}

func (psu *PostUpdater) ListenOnTopic(topic string, handler func(*PostUpdate)) error {
	subTopic, err := psu.PubS.Join(topic)
	if err != nil {
		return err
	}
	subscription, err := subTopic.Subscribe()
	if err != nil {
		return err
	}

	psu.subscriptions.Lock()
	psu.subscriptions.m[topic] = &Subscription{
		sub:     subscription,
		topic:   subTopic,
		handler: handler,
	}
	psu.subscriptions.Unlock()

	go psu.listenOnTopic(subscription)
	return nil
}

func (psu *PostUpdater) listenOnTopic(subscription *pubsub.Subscription) {
	var logger *log.Logger
	logger = psu.ctx.Value("logger").(*log.Logger)
	if logger == nil {
		logger = log.New(os.Stdin, "listen: ", log.Ltime|log.Lshortfile)
	}

	for {
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

		// send valid messages onto the Messages channel
		psu.updateChan <- &PostUpdate{Post: post, User: subscription.Topic()}
	}
}

func NewPostUpdater(ctx context.Context, h host.Host, username string) (*PostUpdater, error) {
	ps, err := pubsub.NewGossipSub(ctx, h)
	if err != nil {
		return nil, err
	}

	ut, err := ps.Join(username)
	if err != nil {
		return nil, err
	}

	// Apparently, if node does not subscribe, other peers don't know that it is in network, haven't noticed it
	// but will do it just to be sure
	_, err = ut.Subscribe()
	if err != nil {
		return nil, err
	}

	psu := &PostUpdater{
		PubS:          ps,
		UserTopic:     ut,
		updateChan:    make(chan *PostUpdate, UpdateBufferSize),
		subscriptions: subscriptionMap{m: make(map[string]*Subscription)},
		self:          h.ID(),
		ctx:           ctx,
	}

	go psu.handleEvents()
	return psu, nil
}
