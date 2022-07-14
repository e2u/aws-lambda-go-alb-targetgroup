package core

import (
	"bytes"
	"encoding/base64"
	"errors"
	"net/http"
	"strings"
	"unicode/utf8"

	"github.com/aws/aws-lambda-go/events"
)

const defaultStatusCode = -1
const contentTypeHeaderKey = "Content-Type"

type ResponseWriter struct {
	headers   http.Header
	body      bytes.Buffer
	status    int
	observers []chan<- bool
}

func NewResponseWriter() *ResponseWriter {
	return &ResponseWriter{
		headers:   make(http.Header),
		status:    defaultStatusCode,
		observers: make([]chan<- bool, 0),
	}
}

func (r *ResponseWriter) Header() http.Header {
	return r.headers
}

func (r *ResponseWriter) Write(body []byte) (int, error) {
	if r.status == -1 {
		r.status = http.StatusOK
	}

	if r.Header().Get(contentTypeHeaderKey) == "" {
		r.Header().Add(contentTypeHeaderKey, http.DetectContentType(body))
	}

	return (&r.body).Write(body)
}

func (r *ResponseWriter) WriteHeader(status int) {
	r.status = status
}

func (r *ResponseWriter) CloseNotify() <-chan bool {
	ch := make(chan bool)

	r.observers = append(r.observers, ch)

	return ch
}

func (r *ResponseWriter) notifyClosed() {
	for _, v := range r.observers {
		v <- true
	}
}

func (r *ResponseWriter) headersToMap() map[string]string {
	m := make(map[string]string)
	for k, vs := range r.headers {
		m[k] = strings.Join(vs, ",")
	}
	return m
}

func (r *ResponseWriter) GetResponse() (events.ALBTargetGroupResponse, error) {
	r.notifyClosed()

	if r.status == defaultStatusCode {
		return events.ALBTargetGroupResponse{}, errors.New("status code not set on response")
	}

	var output string
	isBase64 := false

	bb := (&r.body).Bytes()

	if utf8.Valid(bb) {
		output = string(bb)
	} else {
		output = base64.StdEncoding.EncodeToString(bb)
		isBase64 = true
	}

	return events.ALBTargetGroupResponse{
		StatusCode:        r.status,
		Headers:           r.headersToMap(),
		MultiValueHeaders: r.headers,
		Body:              output,
		IsBase64Encoded:   isBase64,
	}, nil
}
