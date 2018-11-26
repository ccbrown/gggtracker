FROM golang:1.10-alpine

WORKDIR /go/src/github.com/ccbrown/gggtracker

RUN apk add --no-cache git
RUN wget -O - https://raw.githubusercontent.com/golang/dep/master/install.sh | sh

ADD . .
RUN dep ensure
RUN go generate ./...
RUN go vet . && go test -v ./... 
RUN go build .

ENTRYPOINT ["./gggtracker"]
