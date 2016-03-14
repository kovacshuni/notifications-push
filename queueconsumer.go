package main

import (
	queueConsumer "github.com/Financial-Times/message-queue-gonsumer/consumer"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

func (n NotificationsApp) consumeMessages() {
	infoLogger.Println("Entered consumeMessages() function.")
	consumerConfig := n.consumerConfig

	consumer := queueConsumer.NewConsumer(*consumerConfig, n.eventDispatcher.receiveEvents, http.Client{})

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		infoLogger.Println("Started consuming.")
		consumer.Start()
		infoLogger.Println("Finished consuming.")
		wg.Done()
	}()

	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch
	infoLogger.Println("Termination signal received. Quitting consumeMessages function.")
	consumer.Stop()
	wg.Wait()
}
