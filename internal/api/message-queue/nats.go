package message_queue

import (
	"encoding/json"
	"github.com/Nubes3/common/models/nats"
	n "github.com/nats-io/nats.go"
	"github.com/prometheus/common/log"
)

func CreateMessageReplier() {
	nats.Nc.Subscribe(nats.UserSubj, func(msg *n.Msg) {
		message := nats.Msg{}
		err := json.Unmarshal(msg.Data, &message)
		if err != nil {
			//TODO log
			log.Error("Unknown format: " + string(msg.Data))
		}

		if message.ReqType == nats.GetById {

		}

		if message.ReqType == nats.Resolve {

		}
	})
}
