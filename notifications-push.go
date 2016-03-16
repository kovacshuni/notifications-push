package main

import (
	"io"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"strings"

	queueConsumer "github.com/Financial-Times/message-queue-gonsumer/consumer"
	"github.com/jawher/mow.cli"
)

const logPattern = log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile | log.LUTC

var infoLogger *log.Logger
var warnLogger *log.Logger
var errorLogger *log.Logger

type NotificationsApp struct {
	eventDispatcher *EventDispatcher
	consumerConfig  *queueConsumer.QueueConfig
	controller      *Controller
}

func main() {
	app := cli.App("notifications-push", "Proactively notifies subscribers about new content publishes/modifications.")
	consumerAddrs := app.String(cli.StringOpt{
		Name:   "consumer_proxy_addr",
		Value:  "",
		Desc:   "Comma separated kafka proxy hosts for message consuming.",
		EnvVar: "QUEUE_PROXY_ADDRS",
	})
	consumerGroupId := app.String(cli.StringOpt{
		Name:   "consumer_group_id",
		Value:  "",
		Desc:   "Kafka qroup id used for message consuming.",
		EnvVar: "GROUP_ID",
	})
	consumerAutoCommitEnable := app.Bool(cli.BoolOpt{
		Name:   "consumer_autocommit_enable",
		Value:  true,
		Desc:   "Enable autocommit for small messages.",
		EnvVar: "CONSUMER_AUTOCOMMIT_ENABLE",
	})
	consumerAuthorizationKey := app.String(cli.StringOpt{
		Name:   "consumer_authorization_key",
		Value:  "",
		Desc:   "The authorization key required to UCS access.",
		EnvVar: "AUTHORIZATION_KEY",
	})
	topic := app.String(cli.StringOpt{
		Name:   "topic",
		Value:  "",
		Desc:   "Kafka topic to read from.",
		EnvVar: "TOPIC",
	})
	app.Action = func() {
		initLogs(os.Stdout, os.Stdout, os.Stderr)

		dispatcher := NewEvents()
		go dispatcher.distributeEvents()

		consumerConfig := queueConsumer.QueueConfig{}
		consumerConfig.Addrs = strings.Split(*consumerAddrs, ",")
		consumerConfig.Group = *consumerGroupId
		consumerConfig.Topic = *topic
		consumerConfig.AuthorizationKey = *consumerAuthorizationKey
		consumerConfig.AutoCommitEnable = *consumerAutoCommitEnable
		consumerConfig.ConcurrentProcessing = true

		infoLogger.Printf("Consumer config: [%#v]", consumerConfig)
		controller := Controller{dispatcher}
		healthcheck := &Healthcheck{client: http.Client{}, consumerConf: consumerConfig}

		notificationsApp := NotificationsApp{dispatcher, &consumerConfig, &controller}

		http.HandleFunc("/", controller.notifications)
		http.HandleFunc("/__health", healthcheck.healthcheck())
		http.HandleFunc("/__gtg", healthcheck.gtg)

		go func() {
			err := http.ListenAndServe(":8080", nil)
			errorLogger.Println(err)
		}()

		notificationsApp.consumeMessages()
	}
	app.Run(os.Args)
}

func initLogs(infoHandle io.Writer, warnHandle io.Writer, errorHandle io.Writer) {
	infoLogger = log.New(infoHandle, "INFO  - ", logPattern)
	warnLogger = log.New(warnHandle, "WARN  - ", logPattern)
	errorLogger = log.New(errorHandle, "ERROR - ", logPattern)
}
