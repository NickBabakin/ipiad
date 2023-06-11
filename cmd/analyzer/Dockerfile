FROM golang:alpine
RUN apk update
COPY . /go/src
WORKDIR /go/src
CMD [ "go","run","analyzer.go" ]