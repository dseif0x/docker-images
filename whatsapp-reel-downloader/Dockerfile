FROM golang:1.24.3-alpine3.21 AS builder

RUN apk add --no-cache --update gcc g++

WORKDIR $GOPATH/src/mypackage/myapp/
COPY ./ .

RUN CGO_ENABLED=1 go install

FROM alpine:latest

COPY --from=builder /go/bin/whatsapp-reel-downloader /usr/local/bin/whatsapp-reel-downloader

ENTRYPOINT ["/usr/local/bin/whatsapp-reel-downloader"]