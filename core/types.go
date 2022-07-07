package core

import (
	"fmt"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
)

func GatewayTimeout() events.ALBTargetGroupResponse {
	return events.ALBTargetGroupResponse{StatusCode: http.StatusGatewayTimeout}
}

func NewLoggedError(format string, a ...interface{}) error {
	err := fmt.Errorf(format, a...)
	fmt.Println(err.Error())
	return err
}
