package main

import (
	"os"
	"os/signal"
	"sync"
	"syscall"

	queueConsumer "github.com/Financial-Times/message-queue-gonsumer/consumer"
	"github.com/Financial-Times/notifications-push/dispatcher"
	log "github.com/Sirupsen/logrus"
)

type pushService struct {
	dispatcher dispatcher.Dispatcher
	consumer   queueConsumer.Consumer
}

func newPushService(d dispatcher.Dispatcher, consumer queueConsumer.Consumer) *pushService {
	return &pushService{
		dispatcher: d,
		consumer:   consumer,
	}
}

func (p *pushService) start() {
	go p.dispatcher.Start()

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		log.Println("Started consuming.")
		p.consumer.Start()
		log.Println("Finished consuming.")
		wg.Done()
	}()

	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch
	log.Info("Termination signal received. Quitting message consumer and notification dispatcher function.")
	p.consumer.Stop()
	p.dispatcher.Stop()
	wg.Wait()
}
