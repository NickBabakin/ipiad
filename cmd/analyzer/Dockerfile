FROM golang:1.18 AS builder

WORKDIR /usr/src/app

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .
RUN go build -o analyzer ./cmd/analyzer

FROM ubuntu:22.04  
COPY --from=builder /usr/src/app/analyzer /usr/local/bin/
CMD ["analyzer"]