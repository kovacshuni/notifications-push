package dispatcher

import (
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var delay = 2 * time.Second
var historySize = 10

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

var expectedN1StdMsg = `[{"apiUrl":"http://api.ft.com/content/7998974a-1e97-11e6-b286-cddde55ca122","id":"http://www.ft.com/thing/7998974a-1e97-11e6-b286-cddde55ca122","type":"http://www.ft.com/thing/ThingChangeType/UPDATE"}]`
var expectedN2StdMsg = `[{"apiUrl":"http://api.ft.com/content/7998974a-1e97-11e6-b286-cddde55ca122","id":"http://www.ft.com/thing/7998974a-1e97-11e6-b286-cddde55ca122","type":"http://www.ft.com/thing/ThingChangeType/DELETE"}]`

var expectedN1MonitorMsg = `[{"apiUrl":"http://api.ft.com/content/7998974a-1e97-11e6-b286-cddde55ca122","id":"http://www.ft.com/thing/7998974a-1e97-11e6-b286-cddde55ca122","type":"http://www.ft.com/thing/ThingChangeType/UPDATE","publishReference":"tid_test1","lastModified":"2016-11-02T10:54:22.234Z"}]`
var expectedN2MonitorMsg = `[{"apiUrl":"http://api.ft.com/content/7998974a-1e97-11e6-b286-cddde55ca122","id":"http://www.ft.com/thing/7998974a-1e97-11e6-b286-cddde55ca122","type":"http://www.ft.com/thing/ThingChangeType/DELETE","publishReference":"tid_test2","lastModified":"2016-11-02T10:55:24.244Z"}]`

func TestShoudDispatchNotificationsToMultipleSubscribers(t *testing.T) {
	h := NewHistory(historySize)
	d := NewDispatcher(delay, h)

	m := NewMonitorSubscriber("192.168.1.2")
	s := NewStandardSubscriber("192.168.1.3")

	go d.Start()
	defer d.Stop()

	d.Register(s)
	d.Register(m)

	d.Send(n1, n2)

	actualhbMessage := <-s.NotificationChannel()
	assert.Equal(t, heartbeatMsg, actualhbMessage, "First message is a heartbeat")
	actualN1StdMsg := <-s.NotificationChannel()
	assert.Equal(t, expectedN1StdMsg, actualN1StdMsg, "First notification standard message dispatched properly")
	actualN2StdMsg := <-s.NotificationChannel()
	assert.Equal(t, expectedN2StdMsg, actualN2StdMsg, "Second notification standard message dispatched properly")

	actualhbMessage = <-m.NotificationChannel()
	assert.Equal(t, heartbeatMsg, actualhbMessage, "First message is a heartbeat")
	actualN1MonitorMsg := <-m.NotificationChannel()
	assert.Equal(t, expectedN1MonitorMsg, actualN1MonitorMsg, "First notification monitor message dispatched properly")
	actualN2MonitorMsg := <-m.NotificationChannel()
	assert.Equal(t, expectedN2MonitorMsg, actualN2MonitorMsg, "Second notification monitor message dispatched properly")
}

func TestAddAndDeleteOfSubscribers(t *testing.T) {
	h := NewHistory(historySize)
	d := NewDispatcher(delay, h)

	m := NewMonitorSubscriber("192.168.1.2")
	s := NewStandardSubscriber("192.168.1.3")

	go d.Start()
	defer d.Stop()

	d.Register(s)
	d.Register(m)

	assert.Contains(t, d.Subscribers(), s, "Dispatcher contains standard subscriber")
	assert.Contains(t, d.Subscribers(), m, "Dispatcher contains monitor subscriber")
	assert.Equal(t, 2, len(d.Subscribers()), "Dispatcher has 2 subscribers")

	d.Close(s)

	assert.NotContains(t, d.Subscribers(), s, "Dispatcher does not contain standard subscriber")
	assert.Contains(t, d.Subscribers(), m, "Dispatcher contains monitor subscriber")
	assert.Equal(t, 1, len(d.Subscribers()), "Dispatcher has 1 subscriber")

	d.Close(m)

	assert.NotContains(t, d.Subscribers(), s, "Dispatcher does not contain standard subscriber")
	assert.NotContains(t, d.Subscribers(), m, "Dispatcher does not contain monitor subscriber")
	assert.Equal(t, 0, len(d.Subscribers()), "Dispatcher has no subscribers")

	d.Register(m)

	assert.NotContains(t, d.Subscribers(), s, "Dispatcher does not contain standard subscriber")
	assert.Contains(t, d.Subscribers(), m, "Dispatcher contains monitor subscriber")
	assert.Equal(t, 1, len(d.Subscribers()), "Dispatcher has 1 subscriber")

}

func TestDispatchDelay(t *testing.T) {
	h := NewHistory(historySize)
	d := NewDispatcher(delay, h)

	s := NewStandardSubscriber("192.168.1.3")

	go d.Start()
	defer d.Stop()

	d.Register(s)

	actualhbMessage := <-s.NotificationChannel()

	start := time.Now()
	go d.Send(n1)

	actualN1StdMsg := <-s.NotificationChannel()

	stop := time.Now()

	actualDelay := stop.Sub(start)

	assert.Equal(t, heartbeatMsg, actualhbMessage, "First message is a heartbeat")
	assert.Equal(t, expectedN1StdMsg, actualN1StdMsg, "First notification standard message dispatched properly")
	assert.InEpsilon(t, delay.Nanoseconds(), actualDelay.Nanoseconds(), 0.05, "The delay is correct with 0.05 relative error")
}

func TestHeartbeat(t *testing.T) {
	h := NewHistory(10)
	d := NewDispatcher(delay, h)

	s := NewStandardSubscriber("192.168.1.3")

	start := time.Now()
	go d.Start()
	defer d.Stop()
	d.Register(s)

	actualHbMsg := <-s.NotificationChannel()
	assert.Equal(t, heartbeatMsg, actualHbMsg, "The first heartbeat message is correct")

	actualHbMsg = <-s.NotificationChannel()
	actualHbDelay := time.Since(start)
	assert.InEpsilon(t, heartbeatPeriod.Nanoseconds(), actualHbDelay.Nanoseconds(), 0.05, "The first heartbeat delay is correct with 0.05 relative error")
	assert.Equal(t, heartbeatMsg, actualHbMsg, "The second heartbeat message is correct")

	start = start.Add(heartbeatPeriod)
	actualHbMsg = <-s.NotificationChannel()
	actualHbDelay = time.Since(start)

	assert.InEpsilon(t, heartbeatPeriod.Nanoseconds(), actualHbDelay.Nanoseconds(), 0.05, "The second heartbeat delay is correct with 0.05 relative error")
	assert.Equal(t, heartbeatMsg, actualHbMsg, "The third heartbeat message is correct")

	start = start.Add(heartbeatPeriod)
	actualHbMsg = <-s.NotificationChannel()
	actualHbDelay = time.Since(start)
	assert.InEpsilon(t, heartbeatPeriod.Nanoseconds(), actualHbDelay.Nanoseconds(), 0.05, "The third heartbeat delay is correct with 0.05 relative error")
	assert.Equal(t, heartbeatMsg, actualHbMsg, "The fourth heartbeat message is correct")
}

func TestHeartbeatWithNotifications(t *testing.T) {
	h := NewHistory(historySize)
	d := NewDispatcher(delay, h)

	s := NewStandardSubscriber("192.168.1.3")

	start := time.Now()
	go d.Start()
	defer d.Stop()
	d.Register(s)

	actualHbMsg := <-s.NotificationChannel()
	assert.Equal(t, heartbeatMsg, actualHbMsg, "The first heartbeat message is correct")

	actualHbMsg = <-s.NotificationChannel()
	actualHbDelay := time.Since(start)
	assert.InEpsilon(t, heartbeatPeriod.Nanoseconds(), actualHbDelay.Nanoseconds(), 0.05, "The first heartbeat delay is correct with 0.05 relative error")
	assert.Equal(t, heartbeatMsg, actualHbMsg, "The second heartbeat message is correct")

	// send a notification
	start = start.Add(heartbeatPeriod)
	randDuration1 := time.Duration(rand.Intn(int(heartbeatPeriod.Seconds()-delay.Seconds()))) * time.Second
	time.Sleep(randDuration1)
	d.Send(n1)
	actualN1StdMsg := <-s.NotificationChannel()
	assert.Equal(t, expectedN1StdMsg, actualN1StdMsg, "First notification message dispatched properly")

	// waiting for the second heartbeat
	actualHbMsg = <-s.NotificationChannel()
	actualHbDelay = time.Since(start.Add(delay))
	assert.InEpsilon(t, randDuration1.Nanoseconds()+heartbeatPeriod.Nanoseconds(), actualHbDelay.Nanoseconds(), 0.05, "The second heartbeat delay is correct with 0.05 relative error")
	assert.Equal(t, heartbeatMsg, actualHbMsg, "The third heartbeat message is correct")

	// send a notification
	start = time.Now()
	randDuration2 := time.Duration(rand.Intn(int(heartbeatPeriod.Seconds()-delay.Seconds()))) * time.Second
	time.Sleep(randDuration2)
	d.Send(n2)
	actualN2StdMsg := <-s.NotificationChannel()
	assert.Equal(t, expectedN2StdMsg, actualN2StdMsg, "Second notification message dispatched properly")

	// send two notifications
	randDuration3 := time.Duration(rand.Intn(int(heartbeatPeriod.Seconds()-delay.Seconds()))) * time.Second
	time.Sleep(randDuration3)
	d.Send(n1, n2)
	actualN1StdMsg = <-s.NotificationChannel()
	assert.Equal(t, expectedN1StdMsg, actualN1StdMsg, "Third notification message dispatched properly")
	actualN2StdMsg = <-s.NotificationChannel()
	assert.Equal(t, expectedN2StdMsg, actualN2StdMsg, "Fourth notification message dispatched properly")

	// waiting for the third heartbeat
	actualHbMsg = <-s.NotificationChannel()
	actualHbDelay = time.Since(start.Add(2 * delay))
	assert.InEpsilon(t, randDuration2.Nanoseconds()+randDuration3.Nanoseconds()+heartbeatPeriod.Nanoseconds(), actualHbDelay.Nanoseconds(), 0.05, "The third heartbeat delay is correct with 0.05 relative error")
	assert.Equal(t, heartbeatMsg, actualHbMsg, "The fourth heartbeat message is correct")
}

func TestDispatchedNotificationsInHistory(t *testing.T) {
	h := NewHistory(historySize)
	d := NewDispatcher(delay, h)

	go d.Start()
	defer d.Stop()

	d.Send(n1, n2)
	time.Sleep(time.Duration(delay.Seconds()+1) * time.Second)

	assert.Equal(t, n1, h.Notifications()[1], "First notification appears in history in the correct order")
	assert.Equal(t, n2, h.Notifications()[0], "Second notification appears in history in the correct order")
	assert.Len(t, h.Notifications(), 2, "History contains 2 notifications")

	for i := 0; i < historySize; i++ {
		d.Send(n2)
	}
	time.Sleep(time.Duration(delay.Seconds()+1) * time.Second)

	assert.Len(t, h.Notifications(), historySize, "History contains 10 notifications")
	assert.NotContains(t, h.Notifications(), n1, "History does not contain old notification")
}
