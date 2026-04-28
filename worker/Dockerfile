FROM golang:1.24-alpine AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY worker/ ./worker/

RUN go build -o /bin/worker ./worker/

FROM alpine:3.19

COPY --from=builder /bin/worker /usr/local/bin/worker
