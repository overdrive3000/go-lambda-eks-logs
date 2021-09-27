package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sns"
)

type eventData struct {
	ObjectRef resourceData `json:"objectRef,omitempty"`
}

type resourceData struct {
	Name      string `json:"name,omitempty"`
	Namespace string `json:"namespace,omitempty"`
}

func handler(ctx context.Context, event events.CloudwatchLogsEvent) error {

	data, err := event.AWSLogs.Parse()
	if err != nil {
		return err
	}

	var e eventData
	json.Unmarshal([]byte(data.LogEvents[0].Message), &e)

	topic := os.Getenv("SNS_TOPIC_ARN")
	sess := session.Must(session.NewSession())
	svc := sns.New(sess)
	_, err = svc.Publish(&sns.PublishInput{
		Message:  aws.String(fmt.Sprintf("Pod %v in namespace %v has been OOMKilled", e.ObjectRef.Name, e.ObjectRef.Namespace)),
		TopicArn: aws.String(topic),
	})
	if err != nil {
		return err
	}

	return nil
}

func main() {
	lambda.Start(handler)
}
