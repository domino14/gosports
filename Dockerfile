FROM golang:latest
ENV GOPATH=/go
ADD . /go/src/github.com/domino14/gosports
WORKDIR /go/src/github.com/domino14/gosports
RUN go get
RUN go build