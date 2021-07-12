package core

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambdacontext"
)

const CustomHostVariable = "GO_API_HOST"

const DefaultServerAddress = "https://aws-serverless-go-api.com"

type RequestAccessor struct {
	stripBasePath string
}

func (r *RequestAccessor) StripBasePath(basePath string) string {
	if strings.Trim(basePath, " ") == "" {
		r.stripBasePath = ""
		return ""
	}

	newBasePath := basePath
	if !strings.HasPrefix(newBasePath, "/") {
		newBasePath = "/" + newBasePath
	}

	if strings.HasSuffix(newBasePath, "/") {
		newBasePath = newBasePath[:len(newBasePath)-1]
	}

	r.stripBasePath = newBasePath

	return newBasePath
}

func (r *RequestAccessor) ProxyEventToHTTPRequest(req events.ALBTargetGroupRequest) (*http.Request, error) {
	httpRequest, err := r.EventToRequest(req)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return addToHeader(httpRequest, req)
}

func (r *RequestAccessor) EventToRequestWithContext(ctx context.Context, req events.ALBTargetGroupRequest) (*http.Request, error) {
	httpRequest, err := r.EventToRequest(req)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return addToContext(ctx, httpRequest, req), nil
}

func (r *RequestAccessor) EventToRequest(req events.ALBTargetGroupRequest) (*http.Request, error) {
	decodedBody := []byte(req.Body)
	if req.IsBase64Encoded {
		base64Body, err := base64.StdEncoding.DecodeString(req.Body)
		if err != nil {
			return nil, err
		}
		decodedBody = base64Body
	}

	path := req.Path
	if r.stripBasePath != "" && len(r.stripBasePath) > 1 {
		if strings.HasPrefix(path, r.stripBasePath) {
			path = strings.Replace(path, r.stripBasePath, "", 1)
		}
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	serverAddress := DefaultServerAddress
	if customAddress, ok := os.LookupEnv(CustomHostVariable); ok {
		serverAddress = customAddress
	}
	path = serverAddress + path

	if len(req.MultiValueQueryStringParameters) > 0 {
		queryString := ""
		for q, l := range req.MultiValueQueryStringParameters {
			for _, v := range l {
				if queryString != "" {
					queryString += "&"
				}
				queryString += url.QueryEscape(q) + "=" + url.QueryEscape(v)
			}
		}
		path += "?" + queryString
	} else if len(req.QueryStringParameters) > 0 {
		queryString := ""
		for q := range req.QueryStringParameters {
			if queryString != "" {
				queryString += "&"
			}
			queryString += url.QueryEscape(q) + "=" + url.QueryEscape(req.QueryStringParameters[q])
		}
		path += "?" + queryString
	}

	httpRequest, err := http.NewRequest(
		strings.ToUpper(req.HTTPMethod),
		path,
		bytes.NewReader(decodedBody),
	)

	if err != nil {
		fmt.Printf("Could not convert request %s:%s to http.Request\n", req.HTTPMethod, req.Path)
		log.Println(err)
		return nil, err
	}
	for h := range req.Headers {
		httpRequest.Header.Add(h, req.Headers[h])
	}
	return httpRequest, nil
}

func addToHeader(req *http.Request, apiGwRequest events.ALBTargetGroupRequest) (*http.Request, error) {
	return req, nil
}

func addToContext(ctx context.Context, req *http.Request, apiGwRequest events.ALBTargetGroupRequest) *http.Request {
	lc, _ := lambdacontext.FromContext(ctx)
	rc := requestContext{lambdaContext: lc, gatewayProxyContext: apiGwRequest.RequestContext}
	ctx = context.WithValue(ctx, ctxKey{}, rc)
	return req.WithContext(ctx)
}

type ctxKey struct{}

type requestContext struct {
	lambdaContext       *lambdacontext.LambdaContext
	gatewayProxyContext events.ALBTargetGroupRequestContext
}