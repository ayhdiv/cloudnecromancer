package services

import (
	"fmt"

	"github.com/pfrederiksen/cloudnecromancer/internal/parser"
)

func init() {
	parser.Register(&sqsParser{})
}

type sqsParser struct{}

func (p *sqsParser) Service() string { return "sqs" }

func (p *sqsParser) SupportedEvents() []string {
	return []string{
		"CreateQueue",
		"DeleteQueue",
		"SetQueueAttributes",
		"PurgeQueue",
	}
}

func (p *sqsParser) Parse(event map[string]any) (*parser.ResourceDelta, error) {
	eventID, eventTime, eventName, err := parseEvent(event)
	if err != nil {
		return nil, fmt.Errorf("sqs parser: %w", err)
	}

	reqParams := getMap(event, "requestParameters")
	respElems := getMap(event, "responseElements")

	delta := &parser.ResourceDelta{
		EventID:   eventID,
		EventTime: eventTime,
		Service:   "sqs",
	}

	switch eventName {
	case "CreateQueue":
		delta.Action = parser.ActionCreate
		delta.ResourceType = "queue"
		delta.Attributes = make(map[string]any)
		if v := getString(reqParams, "queueName"); v != "" {
			delta.ResourceID = v
			delta.Attributes["queueName"] = v
		}
		if v := getString(respElems, "queueUrl"); v != "" {
			delta.Attributes["queueUrl"] = v
		}

	case "DeleteQueue":
		delta.Action = parser.ActionDelete
		delta.ResourceType = "queue"
		delta.ResourceID = getString(reqParams, "queueUrl")

	case "SetQueueAttributes":
		delta.Action = parser.ActionUpdate
		delta.ResourceType = "queue"
		delta.ResourceID = getString(reqParams, "queueUrl")

	case "PurgeQueue":
		delta.Action = parser.ActionUpdate
		delta.ResourceType = "queue"
		delta.ResourceID = getString(reqParams, "queueUrl")

	default:
		return nil, fmt.Errorf("sqs parser: unsupported event %s", eventName)
	}

	return delta, nil
}
