package event

import (
	"context"
	"encoding/json"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge/types"
)

const source = "opg.supervision.finance"

type EventBridgeClient interface {
	PutEvents(ctx context.Context, params *eventbridge.PutEventsInput, optFns ...func(*eventbridge.Options)) (*eventbridge.PutEventsOutput, error)
}

type Client struct {
	eventBusName string
	eventBridge  EventBridgeClient
}

func NewClient(cfg aws.Config, eventBusName string) *Client {
	return &Client{
		eventBridge:  eventbridge.NewFromConfig(cfg),
		eventBusName: eventBusName,
	}
}

func (c *Client) send(ctx context.Context, eventType string, detail any) error {
	v, err := json.Marshal(detail)
	if err != nil {
		return err
	}

	_, err = c.eventBridge.PutEvents(ctx, &eventbridge.PutEventsInput{
		Entries: []types.PutEventsRequestEntry{{
			EventBusName: aws.String(c.eventBusName),
			Source:       aws.String(source),
			DetailType:   aws.String(eventType),
			Detail:       aws.String(string(v)),
		}},
	})

	return err
}
