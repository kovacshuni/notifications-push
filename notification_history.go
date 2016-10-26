package main

import (
	"sort"
	"sync"
	"time"
)

type notificationHistory struct {
	mutex         *sync.RWMutex
	notifications []notification
	size          int
}

func newNotificationHistory(size int) *notificationHistory {
	return &notificationHistory{&sync.RWMutex{}, []notification{}, size}
}

func (nh *notificationHistory) add(n notification) {
	nh.mutex.Lock()
	defer nh.mutex.Unlock()

	nh.notifications = append(nh.notifications, n)
	sort.Sort(byTimestamp(nh.notifications))

	if len(nh.notifications) > nh.size {
		nh.notifications = nh.notifications[1:]
	}

}

func (nh *notificationHistory) history() []notification {
	nh.mutex.RLock()
	defer nh.mutex.RUnlock()

	return nh.notifications
}

type byTimestamp []notification

func (notifications byTimestamp) Len() int { return len(notifications) }

func (notifications byTimestamp) Swap(i, j int) {
	notifications[i], notifications[j] = notifications[j], notifications[i]
}

func (notifications byTimestamp) Less(i, j int) bool {
	ti, _ := time.Parse(time.RFC3339Nano, notifications[i].LastModified)
	tj, _ := time.Parse(time.RFC3339Nano, notifications[j].LastModified)
	return ti.After(tj)
}
