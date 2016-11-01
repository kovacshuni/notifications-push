package consumer

import (
	"errors"

	"github.com/Financial-Times/notifications-push/dispatcher"
)

// NotificationMapper maps CmsPublicationEvents to Notifications
type NotificationMapper struct {
	APIBaseURL string
	Resource   string
}

// MapNotification maps the given event to a new notification.
func (n NotificationMapper) MapNotification(event CmsPublicationEvent, transactionID string) (dispatcher.Notification, error) {
	if event.UUID == "" {
		return dispatcher.Notification{}, errors.New("CMS publication event does not contain a UUID")
	}

	var eventType string
	if event.HasEmptyPayload() {
		eventType = "DELETE"
	} else {
		eventType = "UPDATE"
	}

	return dispatcher.Notification{
		Type:             "http://www.ft.com/thing/ThingChangeType/" + eventType,
		ID:               "http://www.ft.com/thing/" + event.UUID,
		APIURL:           n.APIBaseURL + "/" + n.Resource + "/" + event.UUID,
		PublishReference: transactionID,
		LastModified:     event.LastModified,
	}, nil
}
