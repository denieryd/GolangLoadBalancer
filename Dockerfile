FROM golang:1.20 AS builder
WORKDIR /app
COPY go.mod ./
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o lb ./cmd/loadbalancer

FROM alpine:latest
RUN apk add --no-cache ca-certificates && update-ca-certificates
WORKDIR /root
COPY --from=builder /app/lb .
ENTRYPOINT [ "/root/lb", "--backends", "http://localhost:3031,http://localhost:3032,http://localhost:3033,http://localhost:3034"]