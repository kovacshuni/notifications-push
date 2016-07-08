package main

import "sync"

type uniqueue interface {
	enqueue(*notificationUPP)
	dequeue() *notificationUPP
	items() []*notificationUPP
}

type arrayUniqueue struct {
	mutex    *sync.Mutex
	buffer   []*notificationUPP
	capacity int
}

func newCircularBuffer(capacity int) uniqueue {
	return &arrayUniqueue{&sync.Mutex{}, make([]*notificationUPP, 0, capacity), capacity}
}

func (cb *arrayUniqueue) enqueue(n *notificationUPP) {
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

func (cb *arrayUniqueue) dequeue() *notificationUPP {
	if len(cb.buffer) > 0 {
		i := cb.buffer[0]
		cb.buffer = cb.buffer[1:]
		return i
	}
	return nil
}

func (cb *arrayUniqueue) items() []*notificationUPP {
	return cb.buffer
}
