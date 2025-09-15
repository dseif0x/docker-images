FROM --platform=${BUILDPLATFORM:-linux/amd64} golang:1.24.3-alpine3.21 AS builder

ARG TARGETPLATFORM
ARG BUILDPLATFORM
ARG TARGETOS
ARG TARGETARCH

WORKDIR $GOPATH/src/mypackage/myapp/
COPY ./ .
RUN go get -d -v
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -ldflags="-w -s" -o /go/bin/k8s-ups-drainer


FROM alpine:latest
WORKDIR /go/bin
COPY --from=builder /go/bin/k8s-ups-drainer /go/bin/k8s-ups-drainer
ENTRYPOINT ["/go/bin/k8s-ups-drainer"]