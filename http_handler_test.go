package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"

	log "github.com/Sirupsen/logrus"
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
		t.Errorf("Expected: [10.10.10.10:10101]. Actual: [%v]", addr)
	}
}

func TestIntegration_NotificationsPushRequestsServed_NrOfClientsReflectedOnStatsEndpoint(t *testing.T) {
	//setting up test controller
	dispatcher := newNotificationDispatcher(10, 200)

	h := newHttpHandler(dispatcher)
	go h.dispatcher.start()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "notifications-push") {
			h.notificationsPush(w, r)
			return
		}
		h.stats(w, r)
	}))
	defer func() {
		ts.Close()
	}()
	nrOfRequests := 10
	for i := 0; i < nrOfRequests; i++ {
		go testRequest(ts.URL + "/notifications-push")
	}
	time.Sleep(time.Second)
	resp, err := http.Get(ts.URL + "/__stats")
	if err != nil {
		t.Error(err)
	}
	defer func() {
		err = resp.Body.Close()
		if err != nil {
			t.Error(err)
		}
	}()

	var stats map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&stats)
	if err != nil {
		t.Error(err)
	}
	actualNrOfReqs, ok := stats["nrOfSubscribers"].(float64)
	if !ok || int(actualNrOfReqs) != nrOfRequests {
		t.Errorf("Expected: [%v]. Found: [%v]", nrOfRequests, int(actualNrOfReqs))
	}
}

func TestNotifications_NotificationsInHistoryMatchReponseNotifications(t *testing.T) {
	not0 := notification{
		APIURL:           "http://localhost:8080/content/16ecb25e-3c63-11e6-8716-a4a71e8140b0",
		ID:               "http://www.ft.com/thing/16ecb25e-3c63-11e6-8716-a4a71e8140b0",
		Type:             "http://www.ft.com/thing/ThingChangeType/UPDATE",
		PublishReference: "test1",
		LastModified:     "2016-06-27T14:56:00.988Z",
	}
	not1 := notification{
		APIURL:           "http://localhost:8080/content/26ecb25e-3c63-11e6-8716-a4a71e8140b0",
		ID:               "http://www.ft.com/thing/26ecb25e-3c63-11e6-8716-a4a71e8140b0",
		Type:             "http://www.ft.com/thing/ThingChangeType/DELETE",
		PublishReference: "test2",
		LastModified:     "2016-06-27T14:57:00.988Z",
	}
	expectedNotificationHistory := []notification{not0, not1}

	d := newNotificationDispatcher(10, 2)
	h := newHttpHandler(d)
	go h.dispatcher.start()

	h.dispatcher.dispatch(not0)
	h.dispatcher.dispatch(not1)

	req, err := http.NewRequest("GET", "http://localhost:8080/content/__history", nil)
	if err != nil {
		t.Errorf("[%v]", err)
	}
	w := httptest.NewRecorder()
	h.history(w, req)

	var actualNotificationHistory []notification
	err = json.Unmarshal(w.Body.Bytes(), actualNotificationHistory)
	if err != nil {
		t.Errorf("[%v]", err)
	}
	if !reflect.DeepEqual(expectedNotificationHistory, actualNotificationHistory) {
		t.Errorf("Expected: [%v]. Actual: [%v]", expectedNotificationHistory, actualNotificationHistory)
	}
}

func testRequest(url string) {
	resp, err := http.Get(url)
	if err != nil {
		log.Warn(err)
	}
	defer func() {
		time.Sleep(time.Second * 2)
		err = resp.Body.Close()
		if err != nil {
			log.Warn(err)
		}
	}()
}
