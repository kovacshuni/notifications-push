package consumer

import (
	"errors"
	"regexp"

	"github.com/Financial-Times/notifications-push/dispatcher"
)

// NotificationMapper maps CmsPublicationEvents to Notifications
type NotificationMapper struct {
	APIBaseURL string
	Resource   string
}

// UUIDRegexp enables to check if a string matches a UUID
var UUIDRegexp = regexp.MustCompile("[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12}")

// MapNotification maps the given event to a new notification.
func (n NotificationMapper) MapNotification(event PublicationEvent, transactionID string) (dispatcher.Notification, error) {
	UUID := UUIDRegexp.FindString(event.ContentURI)
	if UUID == "" {
		return dispatcher.Notification{}, errors.New("ContentURI does not contain a UUID")
	}

	var eventType string
	var scoop bool
	var title = ""

	if event.HasEmptyPayload() {
		eventType = "DELETE"
	} else {
		eventType = "UPDATE"
		title, scoop = extractDataFromPayload(event)
	}

	return dispatcher.Notification{
		Type:             "http://www.ft.com/thing/ThingChangeType/" + eventType,
		ID:               "http://www.ft.com/thing/" + UUID,
		APIURL:           n.APIBaseURL + "/" + n.Resource + "/" + UUID,
		PublishReference: transactionID,
		LastModified:     event.LastModified,
		Title:            title,
		Standout:         dispatcher.Standout{Scoop: scoop},
		ContentType:      n.GetNotificationContentType(event),
	}, nil
}

func extractDataFromPayload(event PublicationEvent) (string, bool) {
	scoop := false
	notificationPayloadMap, ok := event.Payload.(map[string]interface{})
	if !ok {
		return "", scoop
	}

	var title = ""
	if notificationPayloadMap["title"] != nil {
		title = notificationPayloadMap["title"].(string)
	}

	var standout = notificationPayloadMap["standout"]
	if standout != nil {
		standoutMap, ok := standout.(map[string]interface{})
		if ok && standoutMap["scoop"] != nil {
			scoop = standoutMap["scoop"].(bool)
		}
	}

	return title, scoop
}

// GetNotificationContentType extracts the content type for a PublicationEvent
func (n NotificationMapper) GetNotificationContentType(event PublicationEvent) string {
	if event.HasEmptyPayload() {
		return ""
	}

	notificationPayloadMap, ok := event.Payload.(map[string]interface{})
	if ok && notificationPayloadMap["type"] != nil {
		return notificationPayloadMap["type"].(string)
	}

	return ""
}
