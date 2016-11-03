package consumer

import (
	"testing"

	queueConsumer "github.com/Financial-Times/message-queue-gonsumer/consumer"
	"github.com/Financial-Times/notifications-push/test/mocks"

	"github.com/stretchr/testify/mock"
)

func TestSyntheticMessage(t *testing.T) {
	mapper := NotificationMapper{
		APIBaseURL: "test.api.ft.com",
		Resource:   "list",
	}

	dispatcher := new(mocks.MockDispatcher)
	handler := NewMessageQueueHandler("list", mapper, dispatcher)

	msg := []queueConsumer.Message{
		{
			Headers: map[string]string{
				"X-Request-Id": "SYNTH_tid",
			},
			Body: "",
		},
	}

	handler.HandleMessage(msg)
	dispatcher.AssertNotCalled(t, "Send")
}

func TestFailedCMSMessageParse(t *testing.T) {
	mapper := NotificationMapper{
		APIBaseURL: "test.api.ft.com",
		Resource:   "list",
	}

	dispatcher := new(mocks.MockDispatcher)
	handler := NewMessageQueueHandler("list", mapper, dispatcher)

	msg := []queueConsumer.Message{
		{
			Headers: map[string]string{
				"X-Request-Id": "tid_summin",
			},
			Body: "",
		},
	}

	handler.HandleMessage(msg)
	dispatcher.AssertNotCalled(t, "Send")
}

func TestWhitelist(t *testing.T) {
	mapper := NotificationMapper{
		APIBaseURL: "test.api.ft.com",
		Resource:   "list",
	}

	dispatcher := new(mocks.MockDispatcher)
	handler := NewMessageQueueHandler("list", mapper, dispatcher)

	msg := []queueConsumer.Message{
		{
			Headers: map[string]string{
				"X-Request-Id": "tid_summin",
			},
			Body: `{
	         "ContentURI": "something which wouldn't match"
	      }`,
		},
	}

	handler.HandleMessage(msg)
	dispatcher.AssertNotCalled(t, "Send")
}

func TestFailsConversionToNotification(t *testing.T) {
	mapper := NotificationMapper{
		APIBaseURL: "test.api.ft.com",
		Resource:   "list",
	}

	dispatcher := new(mocks.MockDispatcher)

	handler := NewMessageQueueHandler("list", mapper, dispatcher)

	msg := []queueConsumer.Message{
		{
			Headers: map[string]string{
				"X-Request-Id": "tid_summin",
			},
			Body: `{
	         "ContentURI": "http://list-transformer-pr-uk-up.svc.ft.com:8080/list/blah"
	      }`,
		},
	}

	handler.HandleMessage(msg)
	dispatcher.AssertNotCalled(t, "Send")
}

func TestHandleMessage(t *testing.T) {
	mapper := NotificationMapper{
		APIBaseURL: "test.api.ft.com",
		Resource:   "list",
	}

	dispatcher := new(mocks.MockDispatcher)
	dispatcher.On("Send", mock.AnythingOfType("[]dispatcher.Notification")).Return()

	handler := NewMessageQueueHandler("list", mapper, dispatcher)

	msg := []queueConsumer.Message{
		{
			Headers: map[string]string{
				"X-Request-Id": "tid_summin",
			},
			Body: `{
	         "UUID": "a uuid",
	         "ContentURI": "http://list-transformer-pr-uk-up.svc.ft.com:8080/list/blah"
	      }`,
		},
	}

	handler.HandleMessage(msg)
	dispatcher.AssertExpectations(t)
}
