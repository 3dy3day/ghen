package bot

import "github.com/aws/aws-lambda-go/events"

type Bot interface {
	Broadcast(string) error
	Reply(string, events.APIGatewayProxyRequest) error
}
