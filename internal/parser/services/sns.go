package services

import (
	"fmt"

	"github.com/pfrederiksen/cloudnecromancer/internal/parser"
)

func init() {
	parser.Register(&snsParser{})
}

type snsParser struct{}

func (p *snsParser) Service() string { return "sns" }

func (p *snsParser) SupportedEvents() []string {
	return []string{
		"CreateTopic",
		"DeleteTopic",
		"Subscribe",
		"Unsubscribe",
		"SetTopicAttributes",
	}
}

func (p *snsParser) Parse(event map[string]any) (*parser.ResourceDelta, error) {
	eventID, eventTime, eventName, err := parseEvent(event)
	if err != nil {
		return nil, fmt.Errorf("sns parser: %w", err)
	}

	reqParams := getMap(event, "requestParameters")
	respElems := getMap(event, "responseElements")

	delta := &parser.ResourceDelta{
		EventID:   eventID,
		EventTime: eventTime,
		Service:   "sns",
	}

	switch eventName {
	case "CreateTopic":
		delta.Action = parser.ActionCreate
		delta.ResourceType = "topic"
		delta.Attributes = make(map[string]any)
		if v := getString(respElems, "topicArn"); v != "" {
			delta.ResourceID = v
			delta.Attributes["topicArn"] = v
		}
		if v := getString(reqParams, "name"); v != "" {
			delta.Attributes["name"] = v
		}

	case "DeleteTopic":
		delta.Action = parser.ActionDelete
		delta.ResourceType = "topic"
		delta.ResourceID = getString(reqParams, "topicArn")

	case "Subscribe":
		delta.Action = parser.ActionCreate
		delta.ResourceType = "subscription"
		delta.Attributes = make(map[string]any)
		if v := getString(respElems, "subscriptionArn"); v != "" {
			delta.ResourceID = v
			delta.Attributes["subscriptionArn"] = v
		}
		if v := getString(reqParams, "topicArn"); v != "" {
			delta.Attributes["topicArn"] = v
		}
		if v := getString(reqParams, "protocol"); v != "" {
			delta.Attributes["protocol"] = v
		}
		if v := getString(reqParams, "endpoint"); v != "" {
			delta.Attributes["endpoint"] = v
		}

	case "Unsubscribe":
		delta.Action = parser.ActionDelete
		delta.ResourceType = "subscription"
		delta.ResourceID = getString(reqParams, "subscriptionArn")

	case "SetTopicAttributes":
		delta.Action = parser.ActionUpdate
		delta.ResourceType = "topic"
		delta.ResourceID = getString(reqParams, "topicArn")

	default:
		return nil, fmt.Errorf("sns parser: unsupported event %s", eventName)
	}

	return delta, nil
}
