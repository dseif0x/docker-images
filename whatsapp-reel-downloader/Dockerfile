FROM --platform=${BUILDPLATFORM:-linux/amd64} golang:1.24.3-alpine3.21 AS builder

ARG TARGETPLATFORM
ARG BUILDPLATFORM
ARG TARGETOS
ARG TARGETARCH

RUN apk add --no-cache --update gcc g++

WORKDIR $GOPATH/src/mypackage/myapp/
COPY ./ .
RUN go get -d -v
RUN CGO_ENABLED=1 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -ldflags="-w -s" -o /go/bin/whatsapp-reel-downloader


FROM alpine:latest
WORKDIR /go/bin
COPY --from=builder /go/bin/whatsapp-reel-downloader /go/bin/whatsapp-reel-downloader
ENTRYPOINT ["/go/bin/whatsapp-reel-downloader"]