FROM golang:alpine

WORKDIR /go/src/github.com/ccbrown/gggtracker
ADD . .

RUN apk add --no-cache git && go get -t ./...
RUN go vet . && go test -v ./... 
RUN go build .

ENTRYPOINT ["gggtracker"]
