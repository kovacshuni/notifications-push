package main

import (
	"encoding/json"
	"strings"

	queueConsumer "github.com/Financial-Times/message-queue-gonsumer/consumer"
)

type notificationQueueMessage struct {
	queueConsumer.Message
}

func (msg notificationQueueMessage) hasSynthTransactionId() bool {
	tid := msg.transactionId()
	return strings.HasPrefix(tid, "SYNTH")
}

func (msg notificationQueueMessage) transactionId() string {
	return msg.Headers["X-Request-Id"]
}

func (msg notificationQueueMessage) cmsPublicationEvent() (cmsPublicationEvent, error) {
	var event cmsPublicationEvent
	err := json.Unmarshal([]byte(msg.Body), &event)
	return event, err
}
