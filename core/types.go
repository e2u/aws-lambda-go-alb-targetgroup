package core

import (
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"net/http"
)

func GatewayTimeout() events.ALBTargetGroupResponse {
	return events.ALBTargetGroupResponse{StatusCode: http.StatusGatewayTimeout}
}


func NewLoggedError(format string, a ...interface{}) error {
	err := fmt.Errorf(format, a...)
	fmt.Println(err.Error())
	return err
}