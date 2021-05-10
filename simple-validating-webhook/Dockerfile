FROM golang:1.16.4-alpine AS builder
ADD . /build
RUN apk add build-base && \
    cd /build  && \
    go test -race -v ./... && \
    GO111MODULE=on CGO_ENABLED=0 GOOS=linux go build -tags netgo -a -v -o /webhook . && \
    ls -ltr /webhook

FROM golang:1.16.4-alpine
LABEL maintainer="prithak"
COPY --from=builder /webhook /usr/bin/webhook

RUN apk update && \
    chmod 0770 /usr/bin/webhook && \
    mkdir -p /source && \
    addgroup -g 1000 -S sre && \
    adduser -u 1000 -S sre -G sre && \
    chown -Rv sre:sre /source && \
    chmod 0775 /source

USER sre
WORKDIR /source

ENTRYPOINT [ "/usr/bin/webhook" ]
