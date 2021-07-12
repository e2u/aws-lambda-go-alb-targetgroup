package core

import (
	"github.com/aws/aws-lambda-go/events"
)

func getRequest(path string, method string) events.ALBTargetGroupRequest {
	return events.ALBTargetGroupRequest{
		Path:       path,
		HTTPMethod: method,
	}
}

func getRequestContext() events.ALBTargetGroupRequestContext {
	return events.ALBTargetGroupRequestContext{
		ELB: events.ELBContext{
			TargetGroupArn: "arn:test",
		},
	}
}
