package dispatcher

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var m = NewMonitorSubscriber("192.168.1.2")
var s = NewStandardSubscriber("192.168.1.3")

var n1 = Notification{
	APIURL:           "http://api.ft.com/content/7998974a-1e97-11e6-b286-cddde55ca122",
	ID:               "http://www.ft.com/thing/7998974a-1e97-11e6-b286-cddde55ca122",
	Type:             "http://www.ft.com/thing/ThingChangeType/UPDATE",
	PublishReference: "tid_test1",
	LastModified:     "2016-11-02T10:54:22.234Z",
}

var n2 = Notification{
	APIURL:           "http://api.ft.com/content/7998974a-1e97-11e6-b286-cddde55ca122",
	ID:               "http://www.ft.com/thing/7998974a-1e97-11e6-b286-cddde55ca122",
	Type:             "http://www.ft.com/thing/ThingChangeType/DELETE",
	PublishReference: "tid_test2",
	LastModified:     "2016-11-02T10:55:24.244Z",
}

var expectedN1StdMsg = `{"apiUrl":"http://api.ft.com/content/7998974a-1e97-11e6-b286-cddde55ca122","id":"http://www.ft.com/thing/7998974a-1e97-11e6-b286-cddde55ca122","type":"http://www.ft.com/thing/ThingChangeType/UPDATE"}`
var expectedN2StdMsg = `{"apiUrl":"http://api.ft.com/content/7998974a-1e97-11e6-b286-cddde55ca122","id":"http://www.ft.com/thing/7998974a-1e97-11e6-b286-cddde55ca122","type":"http://www.ft.com/thing/ThingChangeType/DELETE"}`

var expectedN1MonitorMsg = `{"apiUrl":"http://api.ft.com/content/7998974a-1e97-11e6-b286-cddde55ca122","id":"http://www.ft.com/thing/7998974a-1e97-11e6-b286-cddde55ca122","type":"http://www.ft.com/thing/ThingChangeType/UPDATE","publishReference":"tid_test1","lastModified":"2016-11-02T10:54:22.234Z"}`
var expectedN2MonitorMsg = `{"apiUrl":"http://api.ft.com/content/7998974a-1e97-11e6-b286-cddde55ca122","id":"http://www.ft.com/thing/7998974a-1e97-11e6-b286-cddde55ca122","type":"http://www.ft.com/thing/ThingChangeType/DELETE","publishReference":"tid_test2","lastModified":"2016-11-02T10:55:24.244Z"}`

func TestShoudDispatchNotificationsToMultipleSubscribers(t *testing.T) {
	h := NewHistory(10)
	d := NewDispatcher(2, h)

	go d.Start()

	d.Register(s)
	d.Register(m)

	d.Send(n1)
	d.Send(n2)

	actualN1StdMsg := <-s.NotificationChannel()
	assert.Equal(t, expectedN1StdMsg, actualN1StdMsg, "First notification standard message dispatched properly")
	actualN2StdMsg := <-s.NotificationChannel()
	assert.Equal(t, expectedN2StdMsg, actualN2StdMsg, "Second notification standard message dispatched properly")

	actualN1MonitorMsg := <-m.NotificationChannel()
	assert.Equal(t, expectedN1MonitorMsg, actualN1MonitorMsg, "First notification monitor message dispatched properly")
	actualN2MonitorMsg := <-m.NotificationChannel()
	assert.Equal(t, expectedN2MonitorMsg, actualN2MonitorMsg, "Second notification monitor message dispatched properly")
}

func TestShouldReturnAllTheSubscribers(t *testing.T) {}

//delay
