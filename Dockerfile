## Build
FROM golang:1.20 AS build
WORKDIR /app
COPY go.mod ./
COPY go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o lb ./cmd/loadbalancer

## Deploy
FROM alpine:3.17.2
RUN apk add --no-cache ca-certificates && update-ca-certificates
WORKDIR /root
COPY --from=build /app/lb lb
ENTRYPOINT [ "/root/lb", "--backends", "http://localhost:3031,http://localhost:3032,http://localhost:3033,http://localhost:3034"]