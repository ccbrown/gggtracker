FROM golang:1.20-alpine

WORKDIR /go/src/github.com/ccbrown/gggtracker

ADD . .
RUN go vet .
RUN go test -v ./... 
RUN go build .

ENTRYPOINT ["./gggtracker"]
