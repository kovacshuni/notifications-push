package main

import (
	"regexp"

	queueConsumer "github.com/Financial-Times/message-queue-gonsumer/consumer"
	log "github.com/Sirupsen/logrus"
)

type messageQueueHandler struct {
	whiteList              *regexp.Regexp
	notificationBuilder    notificationBuilder
	notificationDispatcher notificationDispatcher
}

func newMessageQueueHandler(resource string, apiBaseURL string, dispatcher notificationDispatcher) *messageQueueHandler {
	whiteList := regexp.MustCompile(`^http://.*-transformer-(pr|iw)-uk-.*\.svc\.ft\.com(:\d{2,5})?/(` + resource + `)/[\w-]+.*$`)
	return &messageQueueHandler{
		whiteList:              whiteList,
		notificationBuilder:    notificationBuilder{apiBaseURL, resource},
		notificationDispatcher: dispatcher,
	}
}

func (mqh messageQueueHandler) handleMessages(queueMsg queueConsumer.Message) {
	msg := notificationQueueMessage{queueMsg}

	if msg.hasSynthTransactionId() {
		return
	}
	log.Infof("Received event: tid=[%v].", msg.transactionId())

	cmsPubEvent, err := msg.cmsPublicationEvent()
	if err != nil {
		log.Warnf("Skipping event: tid=[%v], msg=[%v]: [%v].", msg.transactionId(), msg.Body, err)
		return
	}

	if !cmsPubEvent.matches(mqh.whiteList) {
		log.Infof("Skipping event: tid=[%v]. Invalid resourceUri=[%v]", msg.transactionId(), cmsPubEvent.ContentURI)
		return
	}

	n, err := mqh.notificationBuilder.buildNotification(cmsPubEvent, msg.transactionId())
	if err != nil {
		log.Warnf("Skipping event: tid=[%v]. Cannot build notification for msg=[%#v] : [%v]", msg.transactionId(), cmsPubEvent, err)
		return
	}

	mqh.notificationDispatcher.dispatch(n)
}
