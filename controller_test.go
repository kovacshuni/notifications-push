package main

import (
	"encoding/json"
	_ "io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestBuildSubscriber(t *testing.T) {
	testHeaders := http.Header{}
	testHeaders["X-Forwarded-For"] = []string{"1.2.3.4", "5.6.7.8"}
	testHeaders["X-Forwarded-Port"] = []string{"12345"}

	subscriber := buildSubscriber(testHeaders)

	if subscriber.Addr != "1.2.3.4:12345" || subscriber.Since.After(time.Now()) {
		t.Error("Expected success!")
	}
}

func TestHTTPEndpoints_IntegrationTest(t *testing.T) {
	//setting up test controller
	controller := controller{newEvents()}
	go controller.dispatcher.distributeEvents()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "notifications") {
			controller.notifications(w, r)
			return
		}
		controller.stats(w, r)
	}))
	defer ts.Close()
	t.Logf("Server url: [%v]", ts.URL)
	for i := 0; i < 10; i++ {
		go testRequest(ts.URL + "/notifications")
	}
	time.Sleep(time.Second)
	res, err := http.Get(ts.URL + "/stats")
	if err != nil {
		t.Error(err)
	}
	var stats map[string]interface{}
	err = json.NewDecoder(res.Body).Decode(&stats)
	if err != nil {
		t.Error(err)
	}
	t.Logf("Stats: [%v]", stats)
	if stats["nrOfSubscribers"] != 10 {
		t.Errorf("Expected: [10]. Found: [%v]", stats["nrOfSubscribers"])
	}
}

func testRequest(url string) {
	_, err := http.Get(url)
	if err != nil {
		log.Println(err)
	}
}
