FROM golang:1.25.3-alpine3.21 AS builder

RUN apk add --no-cache --update gcc g++

WORKDIR $GOPATH/src/mypackage/myapp/
COPY ./ .

RUN CGO_ENABLED=1 go install

FROM alpine:latest

RUN apk add --no-cache ca-certificates dumb-init

COPY --from=builder /go/bin/whatsapp-reel-downloader /usr/local/bin/whatsapp-reel-downloader

ENTRYPOINT ["dumb-init", "--", "/usr/local/bin/whatsapp-reel-downloader"]
