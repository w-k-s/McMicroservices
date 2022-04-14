FROM golang:1.17 as builder

COPY . /go/src/github.com/w-k-s/McMicroservices/kitchen-service

WORKDIR /go/src/github.com/w-k-s/McMicroservices/kitchen-service

RUN go get ./...

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o app github.com/w-k-s/McMicroservices/kitchen-service/cmd/server

FROM alpine:latest

RUN apk --no-cache add ca-certificates
WORKDIR /root/

RUN mkdir -p /root/.kitchen/migrations
COPY --from=builder /go/src/github.com/w-k-s/McMicroservices/kitchen-service/app .
COPY --from=builder /go/src/github.com/w-k-s/McMicroservices/kitchen-service/migrations .kitchen/migrations

ENTRYPOINT ["./app"]