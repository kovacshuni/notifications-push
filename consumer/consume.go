package consumer

import (
	"regexp"

	queueConsumer "github.com/Financial-Times/message-queue-gonsumer/consumer"
	"github.com/Financial-Times/notifications-push/dispatcher"
	log "github.com/Sirupsen/logrus"
)

type MessageQueueHandler interface {
	HandleMessage(queueMsg []queueConsumer.Message)
}

type simpleMessageQueueHandler struct {
	whiteList  *regexp.Regexp
	mapper     NotificationMapper
	dispatcher dispatcher.Dispatcher
}

// NewMessageQueueHandler returns a new message handler
func NewMessageQueueHandler(resource string, mapper NotificationMapper, dispatcher dispatcher.Dispatcher) MessageQueueHandler {
	whiteList := regexp.MustCompile(`^http://.*-transformer-(pr|iw)-uk-.*\.svc\.ft\.com(:\d{2,5})?/(` + resource + `)/[\w-]+.*$`)
	return &simpleMessageQueueHandler{
		whiteList:  whiteList,
		mapper:     mapper,
		dispatcher: dispatcher,
	}
}

func (m simpleMessageQueueHandler) HandleMessage(msgs []queueConsumer.Message) {
	var batch []dispatcher.Notification
	for _, queueMsg := range msgs {
		msg := NotificationQueueMessage{queueMsg}

		if msg.HasSynthTransactionID() {
			return
		}

		cmsPubEvent, err := msg.ToCmsPublicationEvent()
		if err != nil {
			log.Warnf("Skipping event: tid=[%v], msg=[%v]: [%v].", msg.TransactionID(), msg.Body, err)
			return
		}

		if !cmsPubEvent.Matches(m.whiteList) {
			log.Infof("Skipping event: tid=[%v]. Invalid resourceUri=[%v]", msg.TransactionID(), cmsPubEvent.ContentURI)
			return
		}

		notification, err := m.mapper.MapNotification(cmsPubEvent, msg.TransactionID())
		if err != nil {
			log.Warnf("Skipping event: tid=[%v]. Cannot build notification for msg=[%#v] : [%v]", msg.TransactionID(), cmsPubEvent, err)
			return
		}

		log.WithField("tid", notification.PublishReference).WithField("id", notification.ID).Infof("Received event. Waiting configured delay (%vs).", 10)
		batch = append(batch, notification)
	}

	m.dispatcher.Send(batch...)
}
