package subscription

import (
	"context"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/host"
)

const UpdateBufferSize = 128

type PubSubUpdate struct {
	PubS *pubsub.PubSub
	UserTopic *pubsub.Topic
	updateChan chan *pubsub.Message
}

//func (psu *PubSubUpdate) handleEvents

func MakePubSub(ctx *context.Context, host *host.Host, username string) (*PubSubUpdate, error) {
	ps, err := pubsub.NewGossipSub(*ctx, *host)
	if err != nil {
		return nil, err
	}

	ut, err := ps.Join(username)
	if err != nil {
		return nil, err
	}

	return &PubSubUpdate{PubS: ps, UserTopic: ut, updateChan: make(chan *pubsub.Message, UpdateBufferSize)}, nil
}