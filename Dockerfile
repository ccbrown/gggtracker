FROM golang:1.20-alpine

WORKDIR /go/src/github.com/ccbrown/gggtracker

RUN apk add --no-cache git g++

ADD . .
RUN go generate ./...
RUN go vet .
RUN go test -v ./... 
RUN go build .

ENTRYPOINT ["./gggtracker"]
