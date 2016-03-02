# notifications-push

Proactively notifies subscribers about new content publishes/modifications.

## Build & Run locally

```bash
go build
./notifications-push \
    --consumer_proxy_addr="https://kafka-proxy-iw-uk-p-1.glb.ft.com,https://kafka-proxy-iw-uk-p-2.glb.ft.com" \
    --consumer_group_id="notifications-push" \
    --consumer_authorization_key="can't tell, get it by etcdctl get /ft/_credentials/kafka-bridge/authorization_key" \
    --topic="CmsPublicationEvents"
```

Docker?

```
QUEUE_PROXY_ADDRS="https://kafka-proxy-iw-uk-p-1.glb.ft.com,https://kafka-proxy-iw-uk-p-2.glb.ft.com"
GROUP_ID="notifications-push"
AUTHORIZATION_KEY="can't tell, get it by etcdctl get /ft/_credentials/kafka-bridge/authorization_key"
TOPIC="CmsPublicationEvents"

```
