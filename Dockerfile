FROM golang:1.20 AS builder
WORKDIR /app
COPY go.mod ./
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o lb ./cmd/loadbalancer

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root
COPY --from=builder /app/lb .
ENTRYPOINT [ "/root/lb" ]