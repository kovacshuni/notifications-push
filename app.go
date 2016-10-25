package main

import (
	"encoding/json"

	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	log "github.com/Sirupsen/logrus"

	queueConsumer "github.com/Financial-Times/message-queue-gonsumer/consumer"
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
	app := cli.App("notifications-push", "Proactively notifies subscribers about new content or lists publishes/modifications.")
	resource := app.String(cli.StringOpt{
		Name:   "notifications_resourse",
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
		Desc:   "The API base URL where the content and lists are accessible",
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
	nCap := app.Int(cli.IntOpt{
		Name:   "notifications_capacity",
		Value:  200,
		Desc:   "the nr of recent notifications to be saved and returned on the /notifications endpoint",
		EnvVar: "NOTIFICATIONS_CAPACITY",
	})
	delay := app.Int(cli.IntOpt{
		Name:   "notifications_delay",
		Value:  30,
		Desc:   "The time to delay each notification before forwarding to any subscribers (in seconds).",
		EnvVar: "NOTIFICATIONS_DELAY",
	})

	app.Action = func() {
		dispatcher := newDispatcher()
		go dispatcher.distributeEvents()

		consumerConfig := queueConsumer.QueueConfig{}
		consumerConfig.Addrs = strings.Split(*consumerAddrs, ",")
		consumerConfig.Group = *consumerGroupID
		consumerConfig.Topic = *topic
		consumerConfig.Queue = *consumerHost
		consumerConfig.AuthorizationKey = *consumerAuthorizationKey
		consumerConfig.AutoCommitEnable = *consumerAutoCommitEnable
		consumerConfig.BackoffPeriod = *backoff

		log.Infof("Config: [\n\tresourse: [%v]\n\tconsumerAddrs: [%v]\n\tconsumerGroupID: [%v]\n\ttopic: [%v]\n\tconsumerAutoCommitEnable: [%v]\n\tapiBaseURL: [%v]\n\tnotifications_capacity: [%v]\n]", *resource, *consumerAddrs, *consumerGroupID, *topic, *consumerAutoCommitEnable, *apiBaseURL, *nCap)

		notificationsCache := newUnique(*nCap)
		h := newHandler(*resource, dispatcher, &notificationsCache, *apiBaseURL)
		hc := &healthcheck{client: http.Client{}, consumerConf: consumerConfig}

		notificationsPushPath := "/" + *resource + "/notifications-push"
		notificationsPath := "/" + *resource + "/notifications"

		http.HandleFunc(notificationsPushPath, h.notificationsPush)
		http.HandleFunc(notificationsPath, h.notifications)
		http.HandleFunc("/stats", h.stats)
		http.HandleFunc("/__health", hc.healthcheck())
		http.HandleFunc("/__gtg", hc.gtg)
		go func() {
			err := http.ListenAndServe(":"+strconv.Itoa(*port), nil)
			log.Fatal(err)
		}()

		whiteList := regexp.MustCompile(`^http://.*-transformer-(pr|iw)-uk-.*\.svc\.ft\.com(:\d{2,5})?/(` + *resource + `)/[\w-]+.*$`)

		app := notificationsApp{
			dispatcher,
			&consumerConfig,
			notificationBuilder{*apiBaseURL, *resource},
			&notificationsCache,
			*delay,
			whiteList,
		}
		app.consumeMessages()
	}
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

type notificationsApp struct {
	eventDispatcher     *eventDispatcher
	consumerConfig      *queueConsumer.QueueConfig
	notificationBuilder notificationBuilder
	notificationsCache  *uniqueue
	delay               int
	whiteList           *regexp.Regexp
}

func (app notificationsApp) consumeMessages() {
	consumer := queueConsumer.NewConsumer(*app.consumerConfig, app.receiveEvents, http.Client{})

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		log.Println("Started consuming.")
		consumer.Start()
		log.Println("Finished consuming.")
		wg.Done()
	}()

	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch
	log.Println("Termination signal received. Quitting consumeMessages function.")
	consumer.Stop()
	wg.Wait()
}

func (app notificationsApp) receiveEvents(msg queueConsumer.Message) {
	tid := msg.Headers["X-Request-Id"]
	if strings.HasPrefix(tid, "SYNTH") {
		return
	}
	log.Infof("Received event: tid=[%v].", tid)
	var cmsPubEvent cmsPublicationEvent
	err := json.Unmarshal([]byte(msg.Body), &cmsPubEvent)
	if err != nil {
		log.Warnf("Skipping event: tid=[%v], msg=[%v]: [%v].", tid, msg.Body, err)
		return
	}
	uuid := cmsPubEvent.UUID
	if !app.whiteList.MatchString(cmsPubEvent.ContentURI) {
		log.Infof("Skipping event: tid=[%v]. Invalid resourceUri=[%v]", tid, cmsPubEvent.ContentURI)
		return
	}

	n := app.notificationBuilder.buildNotification(cmsPubEvent)
	if n == nil {
		log.Warnf("Skipping event: tid=[%v]. Cannot build notification for msg=[%#v]", tid, cmsPubEvent)
		return
	}
	bytes, err := json.Marshal([]*notification{n})
	if err != nil {
		log.Warnf("Skipping event: tid=[%v]. Notification [%#v]: [%v]", tid, n, err)
		return
	}

	go func() {
		// wait for the content or lists to be ingested before notifying the clients. Delay is a CLI arg.
		time.Sleep(time.Duration(app.delay) * time.Second)
		log.Infof("Notifying clients about tid=[%v] uuid=[%v].", tid, uuid)
		app.eventDispatcher.incoming <- string(bytes[:])

		uppN := buildUPPNotification(n, tid, cmsPubEvent.LastModified)
		app.notificationsCache.enqueue(uppN)
	}()
}
