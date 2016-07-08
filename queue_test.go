package main

import "testing"

func TestNewCircularBuffer(t *testing.T) {
	maxSize := 5
	cb := newCircularBuffer(maxSize)
	items := cb.items()
	if len(items) != 0 {
		t.Errorf("Expected empty buffer. Actual: [%d]", len(items))
	}
	if cap(items) != maxSize {
		t.Errorf("Expected buffer with capacity: [%d]. Actual: [%d]", maxSize, cap(items))
	}
}

func TestEnqueue_EnqueuingSizePlusOneNrOfItems_LengthDoesNotOverflow(t *testing.T) {
	cb := newCircularBuffer(2)
	cb.enqueue(&notificationUPP{ID: "1"})
	cb.enqueue(&notificationUPP{ID: "2"})
	cb.enqueue(&notificationUPP{ID: "3"})

	if len(cb.items()) != 2 {
		t.Errorf("Expected length to not change. Actual: [%d]", len(cb.items()))
	}
}

func TestEnqueue_EnqueuingSizePlusOneNrOfItems_FirstItemInsertedIsDequeued(t *testing.T) {
	cb := newCircularBuffer(2)
	cb.enqueue(&notificationUPP{ID: "1"})
	cb.enqueue(&notificationUPP{ID: "2"})
	cb.enqueue(&notificationUPP{ID: "3"})

	for _, i := range cb.items() {
		if i.ID == "1" {
			t.Errorf("Expected item [%d] to be dequeued. Actual items: [%v]", 1, cb.items())
		}
	}
}

func TestDequeue_ReadingFromEmptyCircularBuffer_ResultIsNil(t *testing.T) {
	cb := newCircularBuffer(2)

	if cb.dequeue() != nil {
		t.Errorf("Expected getting [nil] from dequeueing an empty list.")
	}
}

func TestDequeue_EnqueuingSizePlusTwoNrOfItems_DequeuingOrderIsPreserved(t *testing.T) {
	cb := newCircularBuffer(2)
	cb.enqueue(&notificationUPP{ID: "1"})
	cb.enqueue(&notificationUPP{ID: "2"})
	cb.enqueue(&notificationUPP{ID: "3"})
	cb.enqueue(&notificationUPP{ID: "4"})

	third := cb.dequeue()
	fourth := cb.dequeue()

	if third.ID != "3" || fourth.ID != "4" {
		t.Errorf("Dequeued elements are out of order.")
	}
}

func TestDequeue_NoMoreThanCapacity(t *testing.T) {
	cb := newCircularBuffer(2)
	cb.enqueue(&notificationUPP{ID: "1"})
	cb.enqueue(&notificationUPP{ID: "2"})
	cb.enqueue(&notificationUPP{ID: "3"})
	cb.enqueue(&notificationUPP{ID: "4"})

	if len(cb.items()) != 2 {
		t.Errorf("Capacity is not maintained.")
	}
}

func TestEnqueue_WontInsertSameID(t *testing.T) {
	cb := newCircularBuffer(3)
	cb.enqueue(&notificationUPP{ID: "1", PublishReference: "a", Type: changeType + "UPDATE"})
	cb.enqueue(&notificationUPP{ID: "2", PublishReference: "b", Type: changeType + "UPDATE"})
	cb.enqueue(&notificationUPP{ID: "3", PublishReference: "c", Type: changeType + "UPDATE"})
	cb.enqueue(&notificationUPP{ID: "4", PublishReference: "d", Type: changeType + "UPDATE"})
	cb.enqueue(&notificationUPP{ID: "3", PublishReference: "e", Type: changeType + "UPDATE"})
	cb.enqueue(&notificationUPP{ID: "4", PublishReference: "f", Type: changeType + "UPDATE"})

	if len(cb.items()) != 3 {
		t.Errorf("Capacity is not maintained.")
	}
	i0 := cb.items()[0]
	if !(i0.ID == "2" && i0.PublishReference == "b") {
		t.Errorf("Element 0 in items is not correct. Actual: %v", *i0)
	}
	i1 := cb.items()[1]
	if !(i1.ID == "3" && i1.PublishReference == "e") {
		t.Errorf("Element 1 in items is not correct. Actual: %v", *i1)
	}
	i2 := cb.items()[2]
	if !(i2.ID == "4" && i2.PublishReference == "f") {
		t.Errorf("Element 2 in items is not correct. Actual: %v", *i2)
	}
}

func TestEnqueue_WillInsertSameIDIfTypeDiffers(t *testing.T) {
	cb := newCircularBuffer(3)
	cb.enqueue(&notificationUPP{ID: "1", PublishReference: "a", Type: changeType + "UPDATE"})
	cb.enqueue(&notificationUPP{ID: "2", PublishReference: "b", Type: changeType + "UPDATE"})
	cb.enqueue(&notificationUPP{ID: "3", PublishReference: "c", Type: changeType + "UPDATE"})
	cb.enqueue(&notificationUPP{ID: "4", PublishReference: "d", Type: changeType + "UPDATE"})
	cb.enqueue(&notificationUPP{ID: "3", PublishReference: "e", Type: changeType + "DELETE"})
	cb.enqueue(&notificationUPP{ID: "4", PublishReference: "f", Type: changeType + "UPDATE"})

	if len(cb.items()) != 3 {
		t.Errorf("Capacity is not maintained.")
	}
	i0 := cb.items()[0]
	if !(i0.ID == "3" && i0.PublishReference == "c" && i0.Type == changeType + "UPDATE") {
		t.Errorf("Element 0 in items is not correct. Actual: %v", *i0)
	}
	i1 := cb.items()[1]
	if !(i1.ID == "3" && i1.PublishReference == "e" && i1.Type == changeType + "DELETE") {
		t.Errorf("Element 1 in items is not correct. Actual: %v", *i1)
	}
	i2 := cb.items()[2]
	if !(i2.ID == "4" && i2.PublishReference == "f" && i2.Type == changeType + "UPDATE") {
		t.Errorf("Element 2 in items is not correct. Actual: %v", *i2)
	}
}