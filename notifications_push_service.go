package main

import (
	"os"
	"os/signal"
	"sync"
	"syscall"

	queueConsumer "github.com/Financial-Times/message-queue-gonsumer/consumer"
	log "github.com/Sirupsen/logrus"
)

type notificationsPushService struct {
	dispatcher *notificationDispatcher
	consumer   queueConsumer.Consumer
}

func newNotificationsPushService(d *notificationDispatcher, consumer queueConsumer.Consumer) *notificationsPushService {
	return &notificationsPushService{
		dispatcher: d,
		consumer:   consumer,
	}
}

func (nps *notificationsPushService) start() {
	go nps.dispatcher.start()

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		log.Println("Started consuming.")
		nps.consumer.Start()
		log.Println("Finished consuming.")
		wg.Done()
	}()

	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch
	log.Println("Termination signal received. Quitting consumeMessages function.")
	nps.consumer.Stop()
	wg.Wait()
}
