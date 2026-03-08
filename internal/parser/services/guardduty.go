package services

import (
	"fmt"

	"github.com/pfrederiksen/cloudnecromancer/internal/parser"
)

func init() {
	parser.Register(&guarddutyParser{})
}

type guarddutyParser struct{}

func (p *guarddutyParser) Service() string { return "guardduty" }

func (p *guarddutyParser) SupportedEvents() []string {
	return []string{
		"CreateDetector",
		"DeleteDetector",
		"UpdateDetector",
		"CreateIPSet",
		"DeleteIPSet",
		"CreateFilter",
		"DeleteFilter",
	}
}

func (p *guarddutyParser) Parse(event map[string]any) (*parser.ResourceDelta, error) {
	eventID, eventTime, eventName, err := parseEvent(event)
	if err != nil {
		return nil, fmt.Errorf("guardduty parser: %w", err)
	}

	reqParams := getMap(event, "requestParameters")
	respElems := getMap(event, "responseElements")

	delta := &parser.ResourceDelta{
		EventID:   eventID,
		EventTime: eventTime,
		Service:   "guardduty",
	}

	switch eventName {
	case "CreateDetector":
		delta.Action = parser.ActionCreate
		delta.ResourceType = "detector"
		delta.ResourceID = getString(respElems, "detectorId")

	case "DeleteDetector":
		delta.Action = parser.ActionDelete
		delta.ResourceType = "detector"
		delta.ResourceID = getString(reqParams, "detectorId")

	case "UpdateDetector":
		delta.Action = parser.ActionUpdate
		delta.ResourceType = "detector"
		delta.ResourceID = getString(reqParams, "detectorId")

	case "CreateIPSet":
		delta.Action = parser.ActionCreate
		delta.ResourceType = "ip_set"
		delta.ResourceID = getString(respElems, "ipSetId")

	case "DeleteIPSet":
		delta.Action = parser.ActionDelete
		delta.ResourceType = "ip_set"
		delta.ResourceID = getString(reqParams, "ipSetId")

	case "CreateFilter":
		delta.Action = parser.ActionCreate
		delta.ResourceType = "filter"
		delta.ResourceID = getString(reqParams, "name")

	case "DeleteFilter":
		delta.Action = parser.ActionDelete
		delta.ResourceType = "filter"
		delta.ResourceID = getString(reqParams, "name")

	default:
		return nil, fmt.Errorf("guardduty parser: unsupported event %s", eventName)
	}

	return delta, nil
}
