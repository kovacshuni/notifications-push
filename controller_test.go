package main

import (
	"net/http"
	"testing"
)

func TestGetClientAddr_XForwardedHeadersPopulated(t *testing.T) {
	testHeaders := http.Header{}
	testHeaders["X-Forwarded-For"] = []string{"1.2.3.4", "5.6.7.8"}
	testHeaders["X-Forwarded-Port"] = []string{"12345"}

	testRequest := &http.Request{
		Header: testHeaders,
	}

	addr := getClientAddr(testRequest)

	if addr != "1.2.3.4:12345" {
		t.Errorf("Expected: [1.2.3.4:12345]. Actual: [%v]", addr)
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
