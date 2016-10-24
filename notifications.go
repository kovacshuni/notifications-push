package main

const changeType = "http://www.ft.com/thing/ThingChangeType/"

type notificationBuilder struct {
	apiBaseURL string
	resource   string
}

type notification struct {
	APIURL string `json:"apiUrl"`
	ID     string `json:"id"`
	Type   string `json:"type"`
}

type notificationUPP struct {
	APIURL           string `json:"apiUrl"`
	ID               string `json:"id"`
	Type             string `json:"type"`
	LastModified     string `json:"lastModified"`
	PublishReference string `json:"publishReference"`
}

type link struct {
	Href string `json:"href"`
	Rel  string `json:"rel"`
}

type notificationsPageUpp struct {
	RequestURL    string            `json:"requestUrl"`
	Notifications []notificationUPP `json:"notifications"`
	Links         []link            `json:"links"`
}

func (nb notificationBuilder) buildNotification(cmsPubEvent cmsPublicationEvent) *notification {
	if cmsPubEvent.UUID == "" {
		return nil
	}

	empty := false
	switch v := cmsPubEvent.Payload.(type) {
	case nil:
		empty = true
	case string:
		if len(v) == 0 {
			empty = true
		}
	case map[string]interface{}:
		if len(v) == 0 {
			empty = true
		}
	}
	eventType := "UPDATE"
	if empty {
		eventType = "DELETE"
	}
	return &notification{
		Type:   changeType + eventType,
		ID:     "http://www.ft.com/thing/" + cmsPubEvent.UUID,
		APIURL: nb.apiBaseURL + "/" + nb.resource + "/" + cmsPubEvent.UUID,
	}
}

func buildUPPNotification(n *notification, tid, lastModified string) *notificationUPP {
	return &notificationUPP{
		APIURL:           n.APIURL,
		ID:               n.ID,
		Type:             n.Type,
		LastModified:     lastModified,
		PublishReference: tid,
	}
}

func newNotificationsPageUpp(notifications []notificationUPP, requestURI string, apiBaseURL string, resource string, internalBaseURL string) notificationsPageUpp {
	return notificationsPageUpp{
		RequestURL:    apiBaseURL + requestURI,
		Notifications: notifications,
		Links: []link{link{
			Href: internalBaseURL + "/" + resource + "/notifications?empty=true",
			Rel:  "next",
		}},
	}
}
