package consumer

import (
	"regexp"
	"testing"

	"github.com/Financial-Times/notifications-push/test/mocks"

	"github.com/Financial-Times/kafka-client-go/kafka"
	"github.com/stretchr/testify/mock"
)

var defaultWhitelist = regexp.MustCompile(`^http://.*-transformer-(pr|iw)-uk-.*\.svc\.ft\.com(:\d{2,5})?/(lists)/[\w-]+.*$`)

func TestSyntheticMessage(t *testing.T) {
	mapper := NotificationMapper{
		APIBaseURL: "test.api.ft.com",
		Resource:   "lists",
	}

	dispatcher := new(mocks.MockDispatcher)
	handler := NewMessageQueueHandler(defaultWhitelist, mapper, dispatcher)

	msg := kafka.NewFTMessage(map[string]string{"X-Request-Id": "SYNTH_tid"},
		`{"UUID": "a uuid", "ContentURI": "http://list-transformer-pr-uk-up.svc.ft.com:8080/lists/blah/55e40823-6804-4264-ac2f-b29e11bf756a"}`)

	handler.HandleMessage(msg)
	dispatcher.AssertNotCalled(t, "Send")
}

func TestFailedCMSMessageParse(t *testing.T) {
	mapper := NotificationMapper{
		APIBaseURL: "test.api.ft.com",
		Resource:   "lists",
	}

	dispatcher := new(mocks.MockDispatcher)
	handler := NewMessageQueueHandler(defaultWhitelist, mapper, dispatcher)

	msg := kafka.NewFTMessage(map[string]string{"X-Request-Id": "tid_summin"}, "")

	handler.HandleMessage(msg)
	dispatcher.AssertNotCalled(t, "Send")
}

func TestWhitelist(t *testing.T) {
	mapper := NotificationMapper{
		APIBaseURL: "test.api.ft.com",
		Resource:   "lists",
	}

	dispatcher := new(mocks.MockDispatcher)
	handler := NewMessageQueueHandler(defaultWhitelist, mapper, dispatcher)

	msg := kafka.NewFTMessage(map[string]string{"X-Request-Id": "tid_summin"},
		`{"ContentURI": "something which wouldn't match"}`)

	handler.HandleMessage(msg)
	dispatcher.AssertNotCalled(t, "Send")
}

func TestFailsConversionToNotification(t *testing.T) {
	mapper := NotificationMapper{
		APIBaseURL: "test.api.ft.com",
		Resource:   "list",
	}

	dispatcher := new(mocks.MockDispatcher)

	handler := NewMessageQueueHandler(defaultWhitelist, mapper, dispatcher)

	msg := kafka.NewFTMessage(map[string]string{"X-Request-Id": "tid_summin"},
		`{"ContentURI": "http://list-transformer-pr-uk-up.svc.ft.com:8080/lists/blah/55e40823-6804-4264-ac2f-b29e11bf756a" + }`)

	handler.HandleMessage(msg)
	dispatcher.AssertNotCalled(t, "Send")
}

func TestHandleMessage(t *testing.T) {
	mapper := NotificationMapper{
		APIBaseURL: "test.api.ft.com",
		Resource:   "lists",
	}

	dispatcher := new(mocks.MockDispatcher)
	dispatcher.On("Send", mock.AnythingOfType("[]dispatcher.Notification")).Return()

	handler := NewMessageQueueHandler(defaultWhitelist, mapper, dispatcher)

	msg := kafka.NewFTMessage(map[string]string{"X-Request-Id": "tid_summin"},
		`{"UUID": "a uuid", "ContentURI": "http://list-transformer-pr-uk-up.svc.ft.com:8080/lists/blah/55e40823-6804-4264-ac2f-b29e11bf756a"}`)

	handler.HandleMessage(msg)
	dispatcher.AssertExpectations(t)
}

func TestDiscardStandardCarouselPublicationEvents(t *testing.T) {
	mapper := NotificationMapper{
		APIBaseURL: "test.api.ft.com",
		Resource:   "lists",
	}

	dispatcher := new(mocks.MockDispatcher)
	handler := NewMessageQueueHandler(defaultWhitelist, mapper, dispatcher)

	msg1 := kafka.NewFTMessage(map[string]string{"X-Request-Id": "tid_fzy2uqund8_carousel_1485954245"},
		`{"UUID": "a uuid", "ContentURI": "http://list-transformer-pr-uk-up.svc.ft.com:8080/lists/blah/55e40823-6804-4264-ac2f-b29e11bf756a"}`)

	msg2 := kafka.NewFTMessage(map[string]string{"X-Request-Id": "republish_-10bd337c-66d4-48d9-ab8a-e8441fa2ec98_carousel_1493606135"},
		`{"UUID": "a uuid", "ContentURI": "http://list-transformer-pr-uk-up.svc.ft.com:8080/lists/blah/55e40823-6804-4264-ac2f-b29e11bf756a"}`)

	msg3 := kafka.NewFTMessage(map[string]string{"X-Request-Id": "tid_ofcysuifp0_carousel_1488384556_gentx"},
		`{"UUID": "a uuid", "ContentURI": "http://list-transformer-pr-uk-up.svc.ft.com:8080/lists/blah/55e40823-6804-4264-ac2f-b29e11bf756a"}`,
	)
	handler.HandleMessage(msg1)
	handler.HandleMessage(msg2)
	handler.HandleMessage(msg3)

	dispatcher.AssertNotCalled(t, "Send")
}

func TestDiscardCarouselPublicationEventsWithGeneratedTransactionID(t *testing.T) {
	mapper := NotificationMapper{
		APIBaseURL: "test.api.ft.com",
		Resource:   "lists",
	}

	dispatcher := new(mocks.MockDispatcher)
	handler := NewMessageQueueHandler(defaultWhitelist, mapper, dispatcher)

	msg := kafka.NewFTMessage(map[string]string{"X-Request-Id": "tid_fzy2uqund8_carousel_1485954245_gentx"},
		`{"UUID": "a uuid", "ContentURI": "http://list-transformer-pr-uk-up.svc.ft.com:8080/lists/blah/55e40823-6804-4264-ac2f-b29e11bf756a"}`,
	)

	handler.HandleMessage(msg)
	dispatcher.AssertNotCalled(t, "Send")
}
