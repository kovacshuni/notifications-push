notifications-push
==================

[![Circle CI](https://circleci.com/gh/Financial-Times/notifications-push/tree/master.png?style=shield)](https://circleci.com/gh/Financial-Times/notifications-push/tree/master)

Proactively notifies subscribers about new content publishes/modifications.


How to Build & Run the binary
-----------------------------

1. Build:

        go build

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


NB. The empty `[]` lines are heartbeats.


How to Build & Run with Docker
------------------------------

    docker build -t coco/notifications-push .

    docker run --env QUEUE_PROXY_ADDRS="https://kafka-proxy-iw-uk-p-1.glb.ft.com,https://kafka-proxy-iw-uk-p-2.glb.ft.com" \
        --env GROUP_ID="notifications-push" \
        --env AUTHORIZATION_KEY="can't tell, get it by etcdctl get /ft/_credentials/kafka-bridge/authorization_key" \
        --env TOPIC="CmsPublicationEvents" \
        coco/notifications-push


Useful Links
------------

* Production: https://prod-up-read.ft.com/content/notifications-push (needs credentials)
