package main

import (
	"strings"
	"testing"
)

func TestBuildNotification_MissingUUID_NotificationNil(t *testing.T) {
	testCmsPubEvent := cmsPublicationEvent{
		UUID: "",
	}
	n := buildNotification(testCmsPubEvent)
	if n != nil {
		t.Errorf("Expected: [nil]. Actual: [%v]", n)
	}
}

func TestBuildNotification_MissingPayload_EventTypeDELETE(t *testing.T) {
	testCmsPubEvent := cmsPublicationEvent{
		UUID:    "foobar",
		Payload: "",
	}
	n := buildNotification(testCmsPubEvent)
	if !strings.HasSuffix(n.Type, "DELETE") {
		t.Errorf("Expected event type DELETE. Actual type URL: [%v]", n.Type)
	}
}

func TestBuildNotification_HappyScenario(t *testing.T) {
	testCmsPubEvent := cmsPublicationEvent{
		UUID:    "baz",
		Payload: `{ "foo" : "bar" }`,
	}
	n := buildNotification(testCmsPubEvent)
	if !strings.HasSuffix(n.Type, "UPDATE") {
		t.Errorf("Expected event type UPDATE. Actual type URL: [%v]", n.Type)
	}
	if !strings.HasSuffix(n.ID, "baz") {
		t.Errorf("Expected ID suffix: [baz]. Actual ID URL: [%v]", n.ID)
	}
	if !strings.HasSuffix(n.APIURL, "baz") {
		t.Errorf("Expected APIURL suffix: [baz]. Actual APIURL: [%v]", n.APIURL)
	}
}
