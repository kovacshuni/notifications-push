package main

import (
	"net/http"
	"testing"
	"time"
)

func TestBuildSubscriber(t *testing.T) {
	testHeaders := http.Header{}
	testHeaders["X-Forward-For"] = []string{"1.2.3.4", "5.6.7.8"}
	testHeaders["X-Forward-Port"] = []string{"12345"}

	subscriber := buildSubscriber(testHeaders)

	if subscriber.addr != "1.2.3.4:12345" || subscriber.since.After(time.Now()) {
		t.Error("Expected success!")
	}
}
