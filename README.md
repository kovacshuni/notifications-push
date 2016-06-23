notifications-push
==================

[![Circle CI](https://circleci.com/gh/Financial-Times/notifications-push/tree/master.png?style=shield)](https://circleci.com/gh/Financial-Times/notifications-push/tree/master)

Proactively notifies subscribers about new content publishes/modifications.
Also able to tell you the last n events in a non-proactive/normal request-response way.
Can tell you the statistics of who are the current subscribers for push.


How to Build & Run the binary
-----------------------------

1. Build and test:

        go build
        go test

2. Run:

    1. via environment variables:

            export QUEUE_PROXY_ADDRS="https://kafka-proxy-iw-uk-p-1.glb.ft.com,https://kafka-proxy-iw-uk-p-2.glb.ft.com" \
                && export GROUP_ID="notifications-push" \
                && export AUTHORIZATION_KEY="$(ssh semantic-tunnel-up.ft.com etcdctl get /ft/_credentials/kafka-bridge/authorization_key)" \
                && ./notifications-push

    2. or via command-line parameters:

            ./notifications-push \
                --consumer_proxy_addr="https://kafka-proxy-iw-uk-p-1.glb.ft.com,https://kafka-proxy-iw-uk-p-2.glb.ft.com" \
                --consumer_group_id="notifications-push" \
                --consumer_authorization_key "$(ssh semantic-tunnel-up.ft.com etcdctl get /ft/_credentials/kafka-bridge/authorization_key)" \
                --topic="CmsPublicationEvents"


How to Use
----------

1. Go to [http://localhost:8080/content/notifications-push](http://localhost:8080/content/notifications-push)
2. You should see a continuously line-by-line streamed response like:
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

NB. The empty `[]` lines are heartbeats.

3. Go to [http://localhost:8080/content/notifications](http://localhost:8080/content/notifications)
4. You should see the last 200 (or some other number) of events like:
```
[
  {
    "apiUrl": "http://api.ft.com/content/815e74c4-32ea-11e6-bda0-04585c31b153",
    "id": "http://www.ft.com/thing/815e74c4-32ea-11e6-bda0-04585c31b153",
    "type": "http://www.ft.com/thing/ThingChangeType/UPDATE",
    "lastModified": "2016-06-22T17:07:38.442Z",
    "publishReference": "tid_8bbcyrrs06"
  },
  {
    "apiUrl": "http://api.ft.com/content/28286998-3897-11e6-a780-b48ed7b6126f",
    "id": "http://www.ft.com/thing/28286998-3897-11e6-a780-b48ed7b6126f",
    "type": "http://www.ft.com/thing/ThingChangeType/UPDATE",
    "lastModified": "2016-06-22T17:59:44.477Z",
    "publishReference": "tid_1ni36jnvsp"
  }
]
```
5. Go to [http://localhost:8080/stats](http://localhost:8080/stats)
6. You should see something like:
```
{
  "nrOfSubscribers": 1,
  "subscribers": [
    {
      "addr": "127.0.0.1:61047",
      "since": "Jun 23 13:32:33.502",
      "connectionDuration": "8m14.226344658s"
    }
  ]
}
```

How to Build & Run with Docker
------------------------------
```
    docker build -t coco/notifications-push .

    docker run --env QUEUE_PROXY_ADDRS="https://kafka-proxy-iw-uk-p-1.glb.ft.com,https://kafka-proxy-iw-uk-p-2.glb.ft.com" \
        --env GROUP_ID="notifications-push" \
        --env AUTHORIZATION_KEY="can't tell, get it by etcdctl get /ft/_credentials/kafka-bridge/authorization_key" \
        --env TOPIC="CmsPublicationEvents" \
        coco/notifications-push
```


Useful Links
------------

* Production: https://prod-up-read.ft.com/content/notifications-push (needs credentials)
