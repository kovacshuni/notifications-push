package main

import (
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/Financial-Times/notifications-push/dispatcher"
	log "github.com/Sirupsen/logrus"
)

type pushService struct {
	dispatcher dispatcher.Dispatcher
}

func newPushService(d dispatcher.Dispatcher) *pushService {
	return &pushService{
		dispatcher: d,
	}
}

func (p *pushService) start() {
	go p.dispatcher.Start()

	var wg sync.WaitGroup
	wg.Add(1)

	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch
	log.Info("Termination signal received. Quitting message consumer and notification dispatcher function.")
	p.dispatcher.Stop()
	wg.Wait()
}
