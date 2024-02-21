# prometheus-sns-lambda-webhook

Receive prometheus alerts via AWS SNS and push them to the webhook.

## TL;DR

At the time of writing, Amazon-managed Prometheus / Alert Manager does not
support webhooks for receiver configs.

This repo implements a workaround via SNS.

```shell
./prometheus-sns-lambda-webhook lambda \
  --webhook-url https://some.host.com/
```
