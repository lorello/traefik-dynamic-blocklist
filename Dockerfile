# syntax=docker/dockerfile:1

FROM golang:alpine AS builder

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY *.go ./

RUN CGO_ENABLED=0 GOOS=linux go build -a -o traefik-dynamic-blocklist .
RUN find /go 

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /srv

COPY --from=builder /app/traefik-dynamic-blocklist ./

EXPOSE 8000

CMD [ "/srv/traefik-dynamic-blocklist" ]


