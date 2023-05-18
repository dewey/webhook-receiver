FROM golang:1.20-alpine as builder

RUN apk add --no-cache \
    # Important: required for go-sqlite3
    gcc \
    # Required for Alpine
    musl-dev \
    git \
    bash

ENV GO111MODULE=on
ENV CGO_ENABLED=1
ENV GOGC=off

# Add our code
ADD ./ $GOPATH/src/github.com/dewey/webhook-receiver

# build
WORKDIR $GOPATH/src/github.com/dewey/webhook-receiver
RUN cd $GOPATH/src/github.com/dewey/webhook-receiver && go build -mod=vendor -v -o /webhook-receiver ./cmd/api/

# multistage
FROM alpine:latest

# https://stackoverflow.com/questions/33353532/does-alpine-linux-handle-certs-differently-than-busybox#33353762
RUN apk --update upgrade && \
    apk add curl ca-certificates && \
    update-ca-certificates && \
    rm -rf /var/cache/apk/*

COPY --from=builder /webhook-receiver /usr/bin/webhook-receiver

# Run the image as a non-root user
RUN adduser -D whr
RUN chmod 0755 /usr/bin/webhook-receiver

USER whr

# Run the app. CMD is required to run on Heroku
CMD webhook-receiver 