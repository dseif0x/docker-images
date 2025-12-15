FROM golang:1.24.3-alpine3.21 AS builder

RUN apk add --no-cache --update gcc g++

WORKDIR $GOPATH/src/mypackage/myapp/
COPY ./ .

RUN CGO_ENABLED=1 go install

FROM node:alpine AS builder-node
RUN apk add --no-cache git
RUN git clone https://github.com/Okramjimmy/Instagram-reels-downloader.git
WORKDIR /Instagram-reels-downloader

RUN npm install
RUN npm run build

FROM alpine:latest

RUN apk add --no-cache --no-scripts nodejs npm dumb-init

COPY --from=builder /go/bin/whatsapp-reel-downloader /usr/local/bin/whatsapp-reel-downloader
COPY --from=builder-node /Instagram-reels-downloader/ /app/public

ENTRYPOINT ["dumb-init", "--", "sh", "-lc", "(cd /app/public && npm run start) & exec /usr/local/bin/whatsapp-reel-downloader"]
