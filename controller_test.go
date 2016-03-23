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

func TestGetClientAddr_XForwardedHeadersPopulated(t *testing.T) {
	testHeaders := http.Header{}
	testHeaders["X-Forwarded-For"] = []string{"1.2.3.4", "5.6.7.8"}
	testRequest := &http.Request{
		Header: testHeaders,
	}

	addr := getClientAddr(testRequest)

	if addr != "1.2.3.4" {
		t.Errorf("Expected: [1.2.3.4]. Actual: [%v]", addr)
	}
}

func TestGetClientAddr_XForwardedHeadersMissing(t *testing.T) {
	testRequest := &http.Request{
		RemoteAddr: "10.10.10.10:10101",
	}

	addr := getClientAddr(testRequest)

	if addr != "10.10.10.10:10101" {
		t.Error("Expected: [10.10.10.10:10101]. Actual: [%v]", addr)
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
		go testRequest(ts.URL + "/content/notifications")
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
