package dispatcher

import (
	"encoding/json"
	"regexp"
	"strings"
)

// NotificationQueueMessage is a wrapper for the queue consumer message type
type NotificationQueueMessage struct {
	message
}

// HasSynthTransactionID checks if the message is synthetic
func (msg NotificationQueueMessage) HasSynthTransactionID() bool {
	tid := msg.TransactionID()
	return strings.HasPrefix(tid, "SYNTH")
}

// HasCarouselTransactionID checks if the message is genearted by the publish carousel
func (msg NotificationQueueMessage) HasCarouselTransactionID() bool {
	return carouselTransactionIDRegExp.MatchString(msg.TransactionID())
}

var carouselTransactionIDRegExp = regexp.MustCompile(`^(tid_[\S]+)_carousel_[\d]{10}.*$`) //`^(tid_[\S]+)_carousel_[\d]{10}.*$`

// TransactionID returns the message TID
func (msg NotificationQueueMessage) TransactionID() string {
	return msg.Headers["X-Request-Id"]
}

// ToPublicationEvent converts the message to a CmsPublicationEvent
func (msg NotificationQueueMessage) ToPublicationEvent() (event PublicationEvent, err error) {
	err = json.Unmarshal([]byte(msg.Body), &event)
	return event, err
}

// PublicationEvent is the data structure that represents a publication event consumed from Kafka
type PublicationEvent struct {
	ContentURI   string
	UUID         string
	Payload      interface{}
	LastModified string
}

// Matches is a method that returns True if the ContentURI of a publication event
// matches a whiteList regexp
func (e PublicationEvent) Matches(whiteList *regexp.Regexp) bool {
	return whiteList.MatchString(e.ContentURI)
}

// HasEmptyPayload is a method that returns true if the PublicationEvent has an empty playload
func (e PublicationEvent) HasEmptyPayload() bool {
	switch v := e.Payload.(type) {
	case nil:
		return true
	case string:
		if len(v) == 0 {
			return true
		}
	case map[string]interface{}:
		if len(v) == 0 {
			return true
		}
	}
	return false
}
