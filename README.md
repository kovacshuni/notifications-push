notifications-push
==================

[![CircleCI](https://circleci.com/gh/Financial-Times/notifications-push.svg?style=svg)](https://circleci.com/gh/Financial-Times/notifications-push) [![Go Report Card](https://goreportcard.com/badge/github.com/Financial-Times/notifications-push)](https://goreportcard.com/report/github.com/Financial-Times/notifications-push) [![Coverage Status](https://coveralls.io/repos/github/Financial-Times/notifications-push/badge.svg?branch=master)](https://coveralls.io/github/Financial-Times/notifications-push?branch=master) [![codecov](https://codecov.io/gh/Financial-Times/notifications-push/branch/master/graph/badge.svg)](https://codecov.io/gh/Financial-Times/notifications-push)

Notifications-push is a microservice that provides push notifications of publications or changes of articles and lists of content.
The microservice consumes a specific Apache Kafka topic group, then it pushes a notification for each article or list available in the consumed Kafka messages.

How to Build & Run the binary
-----------------------------

1. Build and test:
```
go build
go test ./...
```
2. Run:

* via environment variables:

```
export QUEUE_PROXY_ADDRS="http://ftapp14714-lvpr-uk-t:8080,http://ftapp14721-lvpr-uk-t:8080" \
    && export NOTIFICATIONS_RESOURCE="content" \
    && export GROUP_ID="notifications-push-yourtest" \
    && export AUTHORIZATION_KEY="$(ssh semantic-tunnel-up.ft.com etcdctl get /ft/_credentials/kafka-bridge/authorization_key)" \
    && export TOPIC=CmsPublicationEvents \
    && export API_BASE_URL="http://api.ft.com" \
    && ./notifications-push
```

* or via command-line parameters:

```
./notifications-push \
    --notifications_resourse="content"
    --consumer_proxy_addr="https://kafka-proxy-iw-uk-p-1.glb.ft.com,https://kafka-proxy-iw-uk-p-2.glb.ft.com" \
    --consumer_group_id="notifications-push" \
    --consumer_authorization_key "$(ssh semantic-tunnel-up.ft.com etcdctl get /ft/_credentials/kafka-bridge/authorization_key)" \
    --api-base-url="http://api.ft.com" \
    --topic="CmsPublicationEvents"
```

NB: for the complete list of options run `./notifications-push -h`

HTTP endpoints
----------

### Push stream

By opening a HTTP connection with a GET method to the `/{resource}/notifications-push` endpoint, subscribers can consume the notifications push stream for the resource specified in the configuration (content or lists).
The stream will look like the following one.

```
data: []

data: []

data: [{"apiUrl":"http://api.ft.com/content/648bda7b-1187-3496-b48e-57ecb14d5b0a","id":"http://www.ft.com/thing/648bda7b-1187-3496-b48e-57ecb14d5b0a","type":"http://www.ft.com/thing/ThingChangeType/UPDATE"}]

data: []

data: []

data: []

data: [{"apiUrl":"http://api.ft.com/content/e2e49a44-ef3c-11e5-aff5-19b4e253664a","id":"http://www.ft.com/thing/e2e49a44-ef3c-11e5-aff5-19b4e253664a","type":"http://www.ft.com/thing/ThingChangeType/UPDATE"}]

data: []

data: [{"apiUrl":"http://api.ft.com/content/d38489fa-ecf4-11e5-888e-2eadd5fbc4a4","id":"http://www.ft.com/thing/d38489fa-ecf4-11e5-888e-2eadd5fbc4a4","type":"http://www.ft.com/thing/ThingChangeType/UPDATE"}]

data: [{"apiUrl":"http://api.ft.com/content/648bda7b-1187-3496-b48e-57ecb14d5b0a","id":"http://www.ft.com/thing/648bda7b-1187-3496-b48e-57ecb14d5b0a","type":"http://www.ft.com/thing/ThingChangeType/UPDATE"}]

data: [{"apiUrl":"http://api.ft.com/content/648bda7b-1187-3496-b48e-57ecb14d5b0a","id":"http://www.ft.com/thing/648bda7b-1187-3496-b48e-57ecb14d5b0a","type":"http://www.ft.com/thing/ThingChangeType/UPDATE"}]
```

The empty `[]` lines are heartbeats. Notifications-push will send a heartbeat every 30 seconds to keep the connection active.

The notifications-push stream endpoint allows a `monitor` query parameter. By setting the `monitor` flag as `true`, the push stream returns `publishReference` and `lastModified` attributes in the notification message, which is necessary information for UPP internal monitors such as [PAM](https://github.com/Financial-Times/publish-availability-monitor).


To test the stream endpoint you can run the following CURL commands :
```
curl -X GET "http://localhost:8080/content/notifications-push"
curl -X GET "http://localhost:8080/lists/notifications-push?monitor=true"
curl -X GET "https://<user>@<password>:pre-prod-up.ft.com/lists/notifications-push"
```

**WARNING: In CoCo, this endpoint does not work under `/__notifications-push/` and `/__list-notifications-push/`.**
The reason for this is because Vulcan does not support long polling of HTTP requests. We worked around this issue by forwarding messages through Varnish to a fixed port for both services.


### Notification history
A HTTP GET to the `/__history` endpoint will return the history of the last notifications consumed from the Kakfa queue.
The expected payload should look like the following one:

```
[
	{
		"apiUrl": "http://api.ft.com/content/eabefe3e-a4b9-11e6-8b69-02899e8bd9d1",
		"id": "http://www.ft.com/thing/eabefe3e-a4b9-11e6-8b69-02899e8bd9d1",
		"type": "http://www.ft.com/thing/ThingChangeType/UPDATE",
		"publishReference": "tid_jwhfe7n6dj",
		"lastModified": "2016-11-07T13:59:44.950Z"
	},
	{
		"apiUrl": "http://api.ft.com/content/c92d7b0c-a4c7-11e6-8b69-02899e8bd9d1",
		"id": "http://www.ft.com/thing/c92d7b0c-a4c7-11e6-8b69-02899e8bd9d1",
		"type": "http://www.ft.com/thing/ThingChangeType/UPDATE",
		"publishReference": "tid_owd5zqw11m",
		"lastModified": "2016-11-07T13:59:04.546Z"
	},
	{
		"apiUrl": "http://api.ft.com/content/58437fc0-a4f1-11e6-8898-79a99e2a4de6",
		"id": "http://www.ft.com/thing/58437fc0-a4f1-11e6-8898-79a99e2a4de6",
		"type": "http://www.ft.com/thing/ThingChangeType/DELETE",
		"publishReference": "tid_5z8dxzsesj",
		"lastModified": "2016-11-07T13:58:36.285Z"
	}
]
```

### Stats
A HTTP GET to the `/__stats` endpoint will return the stats about the current subscribers that are consuming the notifications push stream
The expected payload should look like the following one:
```
{
	"nrOfSubscribers": 2,
	"subscribers": [
		{
			"addr": "127.0.0.1:61047",
			"since": "Nov  7 14:26:04.018",
			"connectionDuration": "2m41.693365011s",
			"type": "dispatcher.standardSubscriber"
		},
		{
			"addr": "192.168.1.3:65345",
			"since": "Nov  7 14:26:06.259",
			"connectionDuration": "2m39.453175004",
			"type": "dispatcher.monitorSubscriber"
		}
	]
}
```

How to Build & Run with Docker
------------------------------
```
    docker build -t coco/notifications-push .

    docker run --env QUEUE_PROXY_ADDRS="http://ftapp14714-lvpr-uk-t:8080,http://ftapp14721-lvpr-uk-t:8080" \
        --env GROUP_ID="notifications-push-yourtest" \
        --env AUTHORIZATION_KEY="can't tell, get it by etcdctl get /ft/_credentials/kafka-bridge/authorization_key" \
        --env TOPIC="CmsPublicationEvents" \
        --env API_BASE_URL="http://api.ft.com" \
        coco/notifications-push
```

Clients
-------

Example client code is provided in `bin/client` directory

Useful Links
------------
* Production: https://prod-coco-up-read.ft.com/content/notifications-push (needs credentials)
* Production: https://prod-coco-up-read.ft.com/lists/notifications-push (needs credentials)
