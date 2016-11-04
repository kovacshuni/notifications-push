package dispatcher

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestHistory(t *testing.T) {
	history := NewHistory(2)

	history.Push(Notification{
		ID:           "note1",
		LastModified: time.Now().Format(time.RFC3339Nano),
	})

	history.Push(Notification{
		ID:           "note2",
		LastModified: time.Now().Format(time.RFC3339Nano),
	})

	notifications := history.Notifications()
	assert.Equal(t, 2, len(notifications), "Should be size 2")

	history.Push(Notification{
		ID:           "note3",
		LastModified: time.Now().Format(time.RFC3339Nano),
	})

	notifications = history.Notifications()
	assert.Equal(t, 2, len(notifications), "Should still be size 2")

	for _, n := range notifications {
		t.Log(n.ID)
	}

	assert.Equal(t, "note3", notifications[0].ID, "Should be the last pushed notification")
}
