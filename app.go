package main

import (
	"net/http"
	_ "net/http/pprof"
	"os"
	"strconv"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"

	fthealth "github.com/Financial-Times/go-fthealth/v1a"
	queueConsumer "github.com/Financial-Times/message-queue-gonsumer/consumer"
	"github.com/Financial-Times/notifications-push/consumer"
	"github.com/Financial-Times/notifications-push/dispatcher"
	"github.com/Financial-Times/notifications-push/resources"
	"github.com/Financial-Times/service-status-go/httphandlers"
	"github.com/jawher/mow.cli"
)

func init() {
	f := &log.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: time.RFC3339Nano,
	}

	log.SetFormatter(f)
}

func main() {
	app := cli.App("notifications-push", "Proactively notifies subscribers about new publishes/modifications.")
	resource := app.String(cli.StringOpt{
		Name:   "notifications_resource",
		Value:  "",
		Desc:   "The resource of which notifications are produced (e.g., content or lists)",
		EnvVar: "NOTIFICATIONS_RESOURCE",
	})
	consumerAddrs := app.String(cli.StringOpt{
		Name:   "consumer_proxy_addr",
		Value:  "",
		Desc:   "Comma separated kafka proxy hosts for message consuming.",
		EnvVar: "QUEUE_PROXY_ADDRS",
	})
	consumerHost := app.String(cli.StringOpt{
		Name:   "consumer_host_header",
		Value:  "",
		Desc:   "Host header for consumer proxy.",
		EnvVar: "QUEUE_HOST",
	})
	consumerGroupID := app.String(cli.StringOpt{
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
	apiBaseURL := app.String(cli.StringOpt{
		Name:   "api_base_url",
		Value:  "http://api.ft.com",
		Desc:   "The API base URL where resources are accessible",
		EnvVar: "API_BASE_URL",
	})
	topic := app.String(cli.StringOpt{
		Name:   "topic",
		Value:  "",
		Desc:   "Kafka topic to read from.",
		EnvVar: "TOPIC",
	})
	backoff := app.Int(cli.IntOpt{
		Name:   "backoff",
		Value:  4,
		Desc:   "The backoff time for the queue gonsumer.",
		EnvVar: "CONSUMER_BACKOFF",
	})
	port := app.Int(cli.IntOpt{
		Name:   "port",
		Value:  8080,
		Desc:   "application port",
		EnvVar: "PORT",
	})
	historySize := app.Int(cli.IntOpt{
		Name:   "notification_history_size",
		Value:  200,
		Desc:   "the number of recent notifications to be saved and returned on the /__history endpoint",
		EnvVar: "NOTIFICATION_HISTORY_SIZE",
	})
	delay := app.Int(cli.IntOpt{
		Name:   "notifications_delay",
		Value:  30,
		Desc:   "The time to delay each notification before forwarding to any subscribers (in seconds).",
		EnvVar: "NOTIFICATIONS_DELAY",
	})

	app.Action = func() {
		consumerConfig := queueConsumer.QueueConfig{
			Addrs:            strings.Split(*consumerAddrs, ","),
			Group:            *consumerGroupID,
			Topic:            *topic,
			Queue:            *consumerHost,
			AuthorizationKey: *consumerAuthorizationKey,
			AutoCommitEnable: *consumerAutoCommitEnable,
			BackoffPeriod:    *backoff,
		}

		history := dispatcher.NewHistory(*historySize)
		dispatcher := dispatcher.NewDispatcher(time.Duration(*delay)*time.Second, history)

		mapper := consumer.NotificationMapper{
			Resource:   *resource,
			APIBaseURL: *apiBaseURL,
		}

		queueHandler := consumer.NewMessageQueueHandler(*resource, mapper, dispatcher)
		consumer := queueConsumer.NewBatchedConsumer(consumerConfig, queueHandler.HandleMessage, http.Client{})

		go server(":"+strconv.Itoa(*port), *resource, dispatcher, history, consumerConfig)

		pushService := newPushService(dispatcher, consumer)
		pushService.start()
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func server(listen string, resource string, dispatcher dispatcher.Dispatcher, history dispatcher.History, consumerConfig queueConsumer.QueueConfig) {
	notificationsPushPath := "/" + resource + "/notifications-push"

	r := mux.NewRouter()

	r.HandleFunc(notificationsPushPath, resources.Push(dispatcher)).Methods("GET")
	r.HandleFunc("/__history", resources.History(history))
	r.HandleFunc("/__stats", resources.Stats(dispatcher))

	hc := resources.NewNotificationsPushHealthcheck(consumerConfig)

	r.HandleFunc("/__health", fthealth.Handler("Dependent services healthcheck", "Checks if all the dependent services are reachable and healthy.", hc.Check()))
	r.HandleFunc(httphandlers.GTGPath, hc.GTG)
	r.HandleFunc(httphandlers.BuildInfoPath, httphandlers.BuildInfoHandler)
	r.HandleFunc(httphandlers.PingPath, httphandlers.PingHandler)

	http.Handle("/", r)

	err := http.ListenAndServe(listen, nil)
	log.Fatal(err)
}
