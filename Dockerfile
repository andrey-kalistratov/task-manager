FROM golang:1.24-alpine

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o worker-bin ./worker/main.go

COPY worker/config.json /app/config.json

CMD ["./worker-bin",  "-config", "/app/config.json"]
