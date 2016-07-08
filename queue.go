package main

import "sync"

type queue interface {
	enqueue(*notificationUPP)
	dequeue() *notificationUPP
	items() []*notificationUPP
}

type circularBuffer struct {
	mutex    *sync.Mutex
	buffer   []*notificationUPP
	capacity int
}

func newCircularBuffer(capacity int) queue {
	return &circularBuffer{&sync.Mutex{}, make([]*notificationUPP, 0, capacity), capacity}
}

func (cb *circularBuffer) enqueue(n *notificationUPP) {
	wasRemoved := true
	for wasRemoved {
		wasRemoved = false
		for i, e := range cb.buffer {
			if (e.ID == n.ID && e.Type == n.Type) {
				cb.buffer = append(cb.buffer[:i], cb.buffer[i+1:]...)
				wasRemoved = true
				break
			}
		}
	}
	cb.mutex.Lock()
	overflow := len(cb.buffer) - cb.capacity
	if (overflow >= 0) {
		for j := 0; j <= overflow; j++ {
			cb.dequeue()
		}
	}
	cb.buffer = append(cb.buffer, n)
	cb.mutex.Unlock()
}

func (cb *circularBuffer) dequeue() *notificationUPP {
	if len(cb.buffer) > 0 {
		i := cb.buffer[0]
		cb.buffer = cb.buffer[1:]
		return i
	}
	return nil
}

func (cb *circularBuffer) items() []*notificationUPP {
	return cb.buffer
}
