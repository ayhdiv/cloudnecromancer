package services

import (
	"fmt"

	"github.com/pfrederiksen/cloudnecromancer/internal/parser"
)

func init() {
	parser.Register(&cloudwatchParser{})
}

type cloudwatchParser struct{}

func (p *cloudwatchParser) Service() string { return "logs" }

func (p *cloudwatchParser) SupportedEvents() []string {
	return []string{
		"CreateLogGroup",
		"DeleteLogGroup",
		"PutRetentionPolicy",
		"DeleteRetentionPolicy",
		"CreateLogStream",
		"DeleteLogStream",
	}
}

func (p *cloudwatchParser) Parse(event map[string]any) (*parser.ResourceDelta, error) {
	eventID, eventTime, eventName, err := parseEvent(event)
	if err != nil {
		return nil, fmt.Errorf("cloudwatch parser: %w", err)
	}

	reqParams := getMap(event, "requestParameters")

	delta := &parser.ResourceDelta{
		EventID:   eventID,
		EventTime: eventTime,
		Service:   "logs",
	}

	switch eventName {
	case "CreateLogGroup":
		delta.Action = parser.ActionCreate
		delta.ResourceType = "log_group"
		delta.ResourceID = getString(reqParams, "logGroupName")

	case "DeleteLogGroup":
		delta.Action = parser.ActionDelete
		delta.ResourceType = "log_group"
		delta.ResourceID = getString(reqParams, "logGroupName")

	case "PutRetentionPolicy":
		delta.Action = parser.ActionUpdate
		delta.ResourceType = "log_group"
		delta.ResourceID = getString(reqParams, "logGroupName")
		delta.Attributes = make(map[string]any)
		if v, ok := reqParams["retentionInDays"]; ok {
			delta.Attributes["retentionInDays"] = v
		}

	case "DeleteRetentionPolicy":
		delta.Action = parser.ActionUpdate
		delta.ResourceType = "log_group"
		delta.ResourceID = getString(reqParams, "logGroupName")

	case "CreateLogStream":
		delta.Action = parser.ActionCreate
		delta.ResourceType = "log_stream"
		delta.ResourceID = getString(reqParams, "logStreamName")
		delta.Attributes = make(map[string]any)
		if v := getString(reqParams, "logGroupName"); v != "" {
			delta.Attributes["logGroupName"] = v
		}

	case "DeleteLogStream":
		delta.Action = parser.ActionDelete
		delta.ResourceType = "log_stream"
		delta.ResourceID = getString(reqParams, "logStreamName")

	default:
		return nil, fmt.Errorf("cloudwatch parser: unsupported event %s", eventName)
	}

	return delta, nil
}
