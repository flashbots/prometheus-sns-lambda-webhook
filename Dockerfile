# stage: build

FROM golang:1.22-alpine as builder

ARG VERSION=development

WORKDIR /go/src/flashbots/prometheus-sns-lambda-webhook
COPY go.* ./

RUN go mod download
COPY . ./

ENV CGO_ENABLED=0
RUN go build \
			-ldflags "-X main.version=${VERSION}" \
			-o ./bin/prometheus-sns-lambda-webhook \
		github.com/flashbots/prometheus-sns-lambda-webhook/cmd

# stage: run

FROM gcr.io/distroless/static-debian12 as runner

WORKDIR /app

COPY --from=builder /go/src/flashbots/prometheus-sns-lambda-webhook/bin/prometheus-sns-lambda-webhook ./

ENTRYPOINT [ "./prometheus-sns-lambda-webhook" ]
