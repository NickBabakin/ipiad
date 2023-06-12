FROM golang:1.18 AS builder

WORKDIR /usr/src/app

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .
RUN go build -o crawler ./cmd/crawler

FROM ubuntu:22.04  
COPY --from=builder /usr/src/app/crawler /usr/local/bin/
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
CMD ["crawler"]