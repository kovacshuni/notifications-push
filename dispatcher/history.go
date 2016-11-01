package dispatcher

import (
	"sort"
	"sync"
	"time"
)

// History contains the last x notifications pushed out to subscribers.
type History interface {
	Push(notification Notification)
	Get() []Notification
}

type inMemoryHistory struct {
	size          int
	mutex         *sync.RWMutex
	notifications []Notification
}

// NewHistory creates a new history type
func NewHistory(size int) History {
	return &inMemoryHistory{size, &sync.RWMutex{}, []Notification{}}
}

func (i *inMemoryHistory) Push(n Notification) {
	i.mutex.Lock()
	defer i.mutex.Unlock()

	i.notifications = append(i.notifications, n)
	sort.Sort(sort.Reverse(byTimestamp(i.notifications)))

	length := len(i.notifications)
	if length > i.size {
		i.notifications = i.notifications[:length-1] // remove the last entry
	}
}

func (i *inMemoryHistory) Get() []Notification {
	i.mutex.RLock()
	defer i.mutex.RUnlock()

	return i.notifications
}

type byTimestamp []Notification

func (notifications byTimestamp) Len() int { return len(notifications) }

func (notifications byTimestamp) Swap(i, j int) {
	notifications[i], notifications[j] = notifications[j], notifications[i]
}

func (notifications byTimestamp) Less(i, j int) bool {
	ti, _ := time.Parse(time.RFC3339Nano, notifications[i].LastModified)
	tj, _ := time.Parse(time.RFC3339Nano, notifications[j].LastModified)
	return ti.Before(tj)
}
