FROM golang:alpine

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o image-processor ./cmd/image-processor/main.go

EXPOSE 8080