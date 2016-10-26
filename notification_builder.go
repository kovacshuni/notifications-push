package main

import (
	"errors"
)

const changeType = "http://www.ft.com/thing/ThingChangeType"

type notificationBuilder struct {
	apiBaseURL string
	resource   string
}

func (nb notificationBuilder) buildNotification(cmsPubEvent cmsPublicationEvent, transactionId string) (notification, error) {
	if cmsPubEvent.UUID == "" {
		return notification{}, errors.New("CMS publication event does not contain a UUID")
	}

	var eventType string
	if cmsPubEvent.hasEmptyPayload() {
		eventType = "DELETE"
	} else {
		eventType = "UPDATE"
	}

	return notification{
		Type:             "http://www.ft.com/thing/ThingChangeType/" + eventType,
		ID:               "http://www.ft.com/thing/" + cmsPubEvent.UUID,
		APIURL:           nb.apiBaseURL + "/" + nb.resource + "/" + cmsPubEvent.UUID,
		PublishReference: transactionId,
		LastModified:     cmsPubEvent.LastModified,
	}, nil
}
