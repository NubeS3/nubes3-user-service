package message_queue

import (
	"github.com/Nubes3/common/models/nats"
	"github.com/Nubes3/nubes3-user-service/internal/aggregate"
	n "github.com/nats-io/nats.go"
)

var sub *n.Subscription

func CreateMessageSubcribe() (func(), error) {
	var err error
	sub, err = nats.Nc.QueueSubscribe(nats.UserSubj, "user_nubes3_q", aggregate.UserMessageRequestHandler)

	if err != nil {
		return nil, err
	}
	return cleanup, nil
}

func cleanup() {
	_ = sub.Unsubscribe()
}
