package resources

import (
	"bufio"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Financial-Times/notifications-push/dispatcher"
	"github.com/Financial-Times/notifications-push/test/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var start func(sub dispatcher.Subscriber)

func TestPushStandardSubscriber(t *testing.T) {
	d := new(MockDispatcher)

	d.On("Register", mock.AnythingOfType("*dispatcher.StandardSubscriber")).Return()
	d.On("Close", mock.AnythingOfType("*dispatcher.StandardSubscriber")).Return()

	w := NewStreamResponseRecorder()
	req, err := http.NewRequest("GET", "/content/notifications-push", nil)
	if err != nil {
		t.Fatal(err)
	}

	req.Header.Set("X-Forwarded-For", "some-host, some-other-host-that-isnt-used")

	start = func(sub dispatcher.Subscriber) {
		sub.NotificationChannel() <- "hi"
		time.Sleep(10 * time.Millisecond)
		w.closer <- true

		assert.True(t, time.Now().After(sub.Since()))
		assert.Equal(t, "some-host", sub.Address())
	}

	Push(d)(w, req)

	assert.Equal(t, "text/event-stream", w.Header().Get("Content-Type"), "Should be SSE")
	assert.Equal(t, "no-cache, no-store, must-revalidate", w.Header().Get("Cache-Control"))
	assert.Equal(t, "keep-alive", w.Header().Get("Connection"))
	assert.Equal(t, "no-cache", w.Header().Get("Pragma"))
	assert.Equal(t, "0", w.Header().Get("Expires"))

	reader := bufio.NewReader(w.Body)
	body, _ := reader.ReadString(byte(rune(0))) // read to EOF

	assert.Equal(t, "data: hi\n\n", body)

	assert.Equal(t, 200, w.Code, "Should be OK")
	d.AssertExpectations(t)
}

func TestPushMonitorSubscriber(t *testing.T) {
	d := new(MockDispatcher)

	d.On("Register", mock.AnythingOfType("*dispatcher.MonitorSubscriber")).Return()
	d.On("Close", mock.AnythingOfType("*dispatcher.MonitorSubscriber")).Return()

	w := NewStreamResponseRecorder()
	req, err := http.NewRequest("GET", "/content/notifications-push?monitor=true", nil)
	if err != nil {
		t.Fatal(err)
	}

	req.Header.Set("X-Forwarded-For", "some-host, some-other-host-that-isnt-used")

	start = func(sub dispatcher.Subscriber) {
		sub.NotificationChannel() <- "hi"
		time.Sleep(10 * time.Millisecond)
		w.closer <- true

		assert.True(t, time.Now().After(sub.Since()))
		assert.Equal(t, "some-host", sub.Address())
	}

	Push(d)(w, req)

	assert.Equal(t, "text/event-stream", w.Header().Get("Content-Type"), "Should be SSE")
	assert.Equal(t, "no-cache, no-store, must-revalidate", w.Header().Get("Cache-Control"))
	assert.Equal(t, "keep-alive", w.Header().Get("Connection"))
	assert.Equal(t, "no-cache", w.Header().Get("Pragma"))
	assert.Equal(t, "0", w.Header().Get("Expires"))

	reader := bufio.NewReader(w.Body)
	body, _ := reader.ReadString(byte(rune(0))) // read to EOF

	assert.Equal(t, "data: hi\n\n", body)

	assert.Equal(t, 200, w.Code, "Should be OK")
	d.AssertExpectations(t)
}

func TestPushFailed(t *testing.T) {
	d := new(MockDispatcher)
	w := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "/content/notifications-push", nil)
	if err != nil {
		t.Fatal(err)
	}

	Push(d)(w, req)
	assert.Equal(t, 500, w.Code)
}

type MockDispatcher struct {
	mocks.MockDispatcher
}

func (m *MockDispatcher) Register(sub dispatcher.Subscriber) {
	m.Called(sub)
	go start(sub)
}

func NewStreamResponseRecorder() *StreamResponseRecorder {
	w := httptest.NewRecorder()
	closer := make(chan bool)
	return &StreamResponseRecorder{w, closer}
}

type StreamResponseRecorder struct {
	*httptest.ResponseRecorder
	closer chan bool
}

func (r *StreamResponseRecorder) CloseNotify() <-chan bool {
	return r.closer
}
