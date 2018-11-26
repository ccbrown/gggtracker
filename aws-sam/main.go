package main

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws/external"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh/terminal"
	"golang.org/x/sys/unix"

	"github.com/ccbrown/gggtracker/server"
)

type ResponseWriter struct {
	header     http.Header
	statusCode int
	body       *bytes.Buffer
}

func (w *ResponseWriter) Header() http.Header {
	if w.header == nil {
		w.header = http.Header{}
	}
	return w.header
}

func (w *ResponseWriter) Write(b []byte) (int, error) {
	if w.body == nil {
		w.body = &bytes.Buffer{}
	}
	return w.body.Write(b)
}

func (w *ResponseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
}

func (w *ResponseWriter) APIGatewayProxyResponse() (*events.APIGatewayProxyResponse, error) {
	ret := &events.APIGatewayProxyResponse{}
	if w.statusCode != 0 {
		ret.StatusCode = w.statusCode
	} else {
		ret.StatusCode = http.StatusOK
	}
	for key, values := range w.header {
		if len(values) > 1 {
			return nil, fmt.Errorf("header has multiple values: " + key)
		} else if len(values) == 1 {
			if ret.Headers == nil {
				ret.Headers = map[string]string{}
			}
			ret.Headers[key] = values[0]
		}
	}
	if w.body != nil {
		ret.Body = base64.StdEncoding.EncodeToString(w.body.Bytes())
		ret.IsBase64Encoded = true
	}
	return ret, nil
}

func NewRequest(request *events.APIGatewayProxyRequest) (*http.Request, error) {
	resource, err := url.ParseRequestURI(request.Path)
	if err != nil {
		return nil, errors.Wrap(err, "unable to parse request URI")
	}
	if len(request.QueryStringParameters) > 0 {
		query := url.Values{}
		for key, value := range request.QueryStringParameters {
			query.Set(key, value)
		}
		resource.RawQuery = query.Encode()
	}
	req, err := http.ReadRequest(bufio.NewReader(strings.NewReader(request.HTTPMethod + " " + resource.RequestURI() + " HTTP/1.0\r\n\r\n")))
	if err != nil {
		return nil, errors.Wrap(err, "unable to create request")
	}

	req.Proto = "HTTP/1.1"
	req.ProtoMinor = 1

	if request.Body != "" {
		var body []byte
		if request.IsBase64Encoded {
			body, err = base64.StdEncoding.DecodeString(request.Body)
			if err != nil {
				return nil, errors.Wrap(err, "unable to decode base64 body")
			}
		} else {
			body = []byte(request.Body)
		}
		req.ContentLength = int64(len(body))
		req.Body = ioutil.NopCloser(bytes.NewReader(body))
	}

	for key, value := range request.Headers {
		req.Header.Set(key, value)
	}
	req.Host = req.Header.Get("Host")
	req.RemoteAddr = request.RequestContext.Identity.SourceIP

	return req, nil
}

func Handler(handler http.Handler) func(*events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	return func(request *events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
		req, err := NewRequest(request)
		if err != nil {
			return nil, err
		}
		resp := &ResponseWriter{}
		handler.ServeHTTP(resp, req)
		return resp.APIGatewayProxyResponse()
	}
}

func StartHTTPHandler() {
	config, err := external.LoadDefaultAWSConfig()
	if err != nil {
		logrus.Fatal(err)
	}
	db, err := server.NewDynamoDBDatabase(dynamodb.New(config), os.Getenv("GGGTRACKER_DYNAMODB_TABLE"))
	if err != nil {
		logrus.Fatal(err)
	}
	defer db.Close()

	e := server.New(db, os.Getenv("GGGTRACKER_GA"))
	lambda.Start(Handler(e))
}

func main() {
	if !terminal.IsTerminal(unix.Stdout) {
		logrus.SetFormatter(&logrus.JSONFormatter{})
	}

	flags := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	if err := flags.Parse(os.Args[1:]); err != nil {
		if err == flag.ErrHelp {
			// Exit with no error if --help was given. This is used to test the build.
			os.Exit(0)
		}
		logrus.Fatal(err)
	}

	StartHTTPHandler()
}
