package main

import (
	"strings"
	"testing"
)

func TestBuildNotification_MissingUUID_NotificationNil(t *testing.T) {
	nb := notificationBuilder{"http://test.api.ft.com", "content"}
	testCmsPubEvent := cmsPublicationEvent{
		UUID: "",
	}
	n, err := nb.buildNotification(testCmsPubEvent, "tid_test")
	if err != nil {
		t.Errorf("Expected: [nil]. Actual: [%v]", n)
	}
}

func TestBuildNotification_MissingPayload_EventTypeDELETE(t *testing.T) {
	nb := notificationBuilder{"http://test.api.ft.com", "content"}
	testCmsPubEvent := cmsPublicationEvent{
		UUID:    "foobar",
		Payload: nil,
	}
	n, _ := nb.buildNotification(testCmsPubEvent, "tid_test")
	if !strings.HasSuffix(n.Type, "DELETE") {
		t.Errorf("Expected event type DELETE. Actual type URL: [%v]", n.Type)
	}
}

func TestBuildNotification_PayloadIsEmptyString_EventTypeDELETE(t *testing.T) {
	nb := notificationBuilder{"http://test.api.ft.com", "content"}
	testCmsPubEvent := cmsPublicationEvent{
		UUID:    "foobar",
		Payload: "",
	}
	n, _ := nb.buildNotification(testCmsPubEvent, "tid_tst")
	if !strings.HasSuffix(n.Type, "DELETE") {
		t.Errorf("Expected event type DELETE. Actual type URL: [%v]", n.Type)
	}
}

func TestBuildNotification_PayloadIsEmptyMap_EventTypeDELETE(t *testing.T) {
	nb := notificationBuilder{"http://test.api.ft.com", "content"}
	testCmsPubEvent := cmsPublicationEvent{
		UUID:    "foobar",
		Payload: map[string]interface{}{},
	}
	n, _ := nb.buildNotification(testCmsPubEvent, "tid_test")
	if !strings.HasSuffix(n.Type, "DELETE") {
		t.Errorf("Expected event type DELETE. Actual type URL: [%v]", n.Type)
	}
}

func TestBuildNotification_HappyScenario(t *testing.T) {
	nb := notificationBuilder{"http://test.api.ft.com", "content"}
	testCmsPubEvent := cmsPublicationEvent{
		UUID:    "baz",
		Payload: []byte(`{ "foo" : "bar" }`),
	}
	n, _ := nb.buildNotification(testCmsPubEvent, "tid_test")
	if !strings.HasSuffix(n.Type, "UPDATE") {
		t.Errorf("Expected event type UPDATE. Actual type URL: [%v]", n.Type)
	}
	if !strings.HasSuffix(n.ID, "baz") {
		t.Errorf("Expected ID suffix: [baz]. Actual ID URL: [%v]", n.ID)
	}
	if !strings.HasSuffix(n.APIURL, "baz") || !strings.HasPrefix(n.APIURL, nb.apiBaseURL) {
		t.Errorf("Expected: APIURL suffix: [baz], prefix: [%v]. Actual APIURL: [%v]", nb.apiBaseURL, n.APIURL)
	}
}
