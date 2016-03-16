# notifications-push

Proactively notifies subscribers about new content publishes/modifications.

## Build & Run the binary

```bash
go build

./notifications-push \
    --consumer_proxy_addr="https://kafka-proxy-iw-uk-p-1.glb.ft.com,https://kafka-proxy-iw-uk-p-2.glb.ft.com" \
    --consumer_group_id="notifications-push" \
    --consumer_authorization_key "can't tell, get it by etcdctl get /ft/_credentials/kafka-bridge/authorization_key" \
    --topic="CmsPublicationEvents"
```

## Build & Run with Docker

```bash
docker build -t coco/notifications-push .

docker run --env QUEUE_PROXY_ADDRS="https://kafka-proxy-iw-uk-p-1.glb.ft.com,https://kafka-proxy-iw-uk-p-2.glb.ft.com" \
    --env GROUP_ID="notifications-push" \
    --env AUTHORIZATION_KEY="can't tell, get it by etcdctl get /ft/_credentials/kafka-bridge/authorization_key" \
    --env TOPIC="CmsPublicationEvents" \
    coco/notifications-push
```
