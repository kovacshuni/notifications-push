package main

import (
	"testing"
)

func TestNewCircularBuffer(t *testing.T) {
	maxSize := 5
	nh := newNotificationHistory(maxSize)
	items := nh.history()
	if len(items) != 0 {
		t.Errorf("Expected empty buffer. Actual: [%d]", len(items))
	}
	if cap(items) != maxSize {
		t.Errorf("Expected buffer with capacity: [%d]. Actual: [%d]", maxSize, cap(items))
	}
}

func TestEnqueue_EnqueuingSizePlusOneNrOfItems_LengthDoesNotOverflow(t *testing.T) {
	nh := newNotificationHistory(2)
	nh.add(notification{ID: "1", PublishReference: "a", Type: changeType + "UPDATE", LastModified: "2016-07-11T13:00:00.000Z"})
	nh.add(notification{ID: "2", PublishReference: "b", Type: changeType + "UPDATE", LastModified: "2016-07-11T14:00:00.000Z"})
	nh.add(notification{ID: "3", PublishReference: "c", Type: changeType + "UPDATE", LastModified: "2016-07-11T15:00:00.000Z"})

	if len(nh.history()) != 2 {
		t.Errorf("Expected length to not change. Actual: [%d]", len(nh.history()))
	}
}

func TestEnqueue_EnqueuingSizePlusOneNrOfItems_FirstItemInsertedIsDequeued(t *testing.T) {
	nh := newNotificationHistory(2)
	nh.add(notification{ID: "1", PublishReference: "a", Type: changeType + "UPDATE", LastModified: "2016-07-11T13:00:00.000Z"})
	nh.add(notification{ID: "2", PublishReference: "b", Type: changeType + "UPDATE", LastModified: "2016-07-11T14:00:00.000Z"})
	nh.add(notification{ID: "3", PublishReference: "c", Type: changeType + "UPDATE", LastModified: "2016-07-11T15:00:00.000Z"})

	for _, n := range nh.history() {
		if n.ID == "1" {
			t.Errorf("Expected item [%d] to be dequeued. Actual items: [%v]", 1, nh.history())
		}
	}
}
