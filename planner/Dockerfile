# syntax=docker/dockerfile:1.7

FROM golang:1.26-alpine AS builder
WORKDIR /app

RUN apk add --no-cache gcc musl-dev

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

COPY planner/ ./planner/

ENV CGO_ENABLED=1
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go build -o /bin/tmd ./planner/cmd/tmd && \
    go build -o /bin/tm  ./planner/cmd/tm

FROM alpine:3.19
WORKDIR /app

COPY --from=builder /bin/tmd /usr/local/bin/tmd
COPY --from=builder /bin/tm  /usr/local/bin/tm

CMD ["tmd"]