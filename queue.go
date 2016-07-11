package main

import (
	"sync"
	"time"
)

type uniqueue struct {
	mutex    *sync.Mutex
	buffer   []*notificationUPP
	capacity int
}

func newUnique(capacity int) uniqueue {
	return uniqueue{&sync.Mutex{}, make([]*notificationUPP, 0, capacity), capacity}
}

func (cb *uniqueue) enqueue(n *notificationUPP) {
	if n.LastModified == "" {
		warnLogger.Printf("Incoming notification must have a last modified date: %v", n)
		return
	}
	newestRawDate, err := time.Parse(time.RFC3339, n.LastModified)
	if err != nil {
		warnLogger.Printf("Incoming notification has malformed date: %v", n.LastModified)
		return
	}
	for i, e := range cb.buffer {
		if e.ID == n.ID && e.Type == n.Type {
			eDate, err := time.Parse(time.RFC3339Nano, e.LastModified)
			if err != nil {
				warnLogger.Printf("Notification in cache has malformed date: %v", n)
				return
			}
			if eDate.Before(newestRawDate) {
				cb.buffer = append(cb.buffer[:i], cb.buffer[i+1:]...)
				break
			} else {
				return
			}
		}
	}
	cb.mutex.Lock()
	overflow := len(cb.buffer) - cb.capacity
	if overflow >= 0 {
		for j := 0; j <= overflow; j++ {
			cb.dequeue()
		}
	}
	cb.buffer = append(cb.buffer, n)
	cb.mutex.Unlock()
}

func (cb *uniqueue) dequeue() *notificationUPP {
	if len(cb.buffer) > 0 {
		i := cb.buffer[0]
		cb.buffer = cb.buffer[1:]
		return i
	}
	return nil
}

func (cb uniqueue) items() []*notificationUPP {
	return cb.buffer
}
