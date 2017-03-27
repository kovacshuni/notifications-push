package main

import (
	"net/http"
	_ "net/http/pprof"
	"os"
	"regexp"
	"strconv"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"

	fthealth "github.com/Financial-Times/go-fthealth/v1a"
	"github.com/Financial-Times/notifications-push/dispatcher"
	"github.com/Financial-Times/notifications-push/resources"
	"github.com/Financial-Times/service-status-go/httphandlers"
	"github.com/jawher/mow.cli"

	"github.com/wvanbergen/kafka/consumergroup"
	"github.com/wvanbergen/kazoo-go"

	kafka "github.com/Shopify/sarama"
)

const heartbeatPeriod = 30 * time.Second

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
		Value:  "content",
		Desc:   "The resource of which notifications are produced (e.g., content or lists)",
		EnvVar: "NOTIFICATIONS_RESOURCE",
	})
	zookeeperHost := app.String(cli.StringOpt{
		Name:   "zookeeper_host",
		Value:  "localhost",
		Desc:   "Zookeerper host",
		EnvVar: "ZOOKEEPER_HOST",
	})
	zookeeperPort := app.String(cli.StringOpt{
		Name:   "zookeeper_port",
		Value:  "2181",
		Desc:   "Zookeerper port",
		EnvVar: "ZOOKEEPER_PORT",
	})
	group := app.String(cli.StringOpt{
		Name:   "group",
		Value:  "test-notifications-push",
		Desc:   "consumer group",
		EnvVar: "GROUP_ID",
	})
	apiBaseURL := app.String(cli.StringOpt{
		Name:   "api_base_url",
		Value:  "http://api.ft.com",
		Desc:   "The API base URL where resources are accessible",
		EnvVar: "API_BASE_URL",
	})
	topic := app.String(cli.StringOpt{
		Name:   "topic",
		Value:  "PostPublicationEvents",
		Desc:   "Kafka topic to read from.",
		EnvVar: "TOPIC",
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
	whitelist := app.String(cli.StringOpt{
		Name:   "whitelist",
		Desc:   `The whitelist for incoming notifications - i.e. ^http://.*-transformer-(pr|iw)-uk-.*\.svc\.ft\.com(:\d{2,5})?/content/[\w-]+.*$`,
		EnvVar: "WHITELIST",
	})

	app.Action = func() {
		// consumerConfig := queueConsumer.QueueConfig{
		// 	Addrs:            strings.Split(*consumerAddrs, ","),
		// 	Group:            *consumerGroupID,
		// 	Topic:            *topic,
		// 	Queue:            *consumerHost,
		// 	AuthorizationKey: *consumerAuthorizationKey,
		// 	AutoCommitEnable: *consumerAutoCommitEnable,
		// 	BackoffPeriod:    *backoff,
		// }

		whitelistR, err := regexp.Compile(*whitelist)
		if err != nil {
			log.WithError(err).Fatal("Whitelist regex MUST compile!")
			return
		}

		mapper := dispatcher.NotificationMapper{
			Resource:   *resource,
			APIBaseURL: *apiBaseURL,
		}
		//queueHandler := consumer.NewMessageQueueHandler(whitelistR, mapper, dispatcher)
		//consumer := queueConsumer.NewBatchedConsumer(consumerConfig, queueHandler.HandleMessage, http.Client

		config := consumergroup.NewConfig()
		config.Offsets.Initial = kafka.OffsetNewest
		config.Offsets.ProcessingTimeout = 10 * time.Second

		var zookeeperNodes []string
		zookeeperNodes, config.Zookeeper.Chroot = kazoo.ParseConnectionString(*zookeeperHost + ":" + *zookeeperPort)

		consumer, consumerErr := consumergroup.JoinConsumerGroup(*group, []string{*topic}, zookeeperNodes, config)
		if consumerErr != nil {
			log.Fatalln(consumerErr)
		}

		if err != nil {
			panic(err)
		}

		defer func() {
			if cErr := consumer.Close(); cErr != nil {
				log.Fatal(cErr)
			}
		}()

		history := dispatcher.NewHistory(*historySize)
		dispatcher := dispatcher.NewDispatcher(time.Duration(*delay)*time.Second, heartbeatPeriod, history, consumer, whitelistR, mapper)

		go server(":"+strconv.Itoa(*port), *resource, dispatcher, history, consumer)

		pushService := newPushService(dispatcher)
		pushService.start()
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func server(listen string, resource string, dispatcher dispatcher.Dispatcher, history dispatcher.History, consumer *consumergroup.ConsumerGroup) {
	notificationsPushPath := "/" + resource + "/notifications-push"

	r := mux.NewRouter()

	r.HandleFunc(notificationsPushPath, resources.Push(dispatcher)).Methods("GET")
	r.HandleFunc("/__history", resources.History(history)).Methods("GET")
	r.HandleFunc("/__stats", resources.Stats(dispatcher)).Methods("GET")

	hc := resources.NewNotificationsPushHealthcheck(consumer)

	r.HandleFunc("/__health", fthealth.Handler("Dependent services healthcheck", "Checks if all the dependent services are reachable and healthy.", hc.Check()))
	r.HandleFunc(httphandlers.GTGPath, hc.GTG)
	r.HandleFunc(httphandlers.BuildInfoPath, httphandlers.BuildInfoHandler)
	r.HandleFunc(httphandlers.PingPath, httphandlers.PingHandler)

	http.Handle("/", r)

	err := http.ListenAndServe(listen, nil)
	log.Fatal(err)
}
