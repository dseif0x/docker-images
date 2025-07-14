FROM golang:alpine AS builder
WORKDIR $GOPATH/src/mypackage/myapp/
COPY . .
RUN go get -d -v
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /go/bin/worker-email-sender


FROM gcr.io/distroless/static-debian12
WORKDIR /go/bin
COPY --from=builder /go/bin/worker-email-sender /go/bin/worker-email-sender
ENTRYPOINT ["/go/bin/worker-email-sender"]
