package services

import (
	"fmt"

	"github.com/pfrederiksen/cloudnecromancer/internal/parser"
)

func init() {
	parser.Register(&cloudfrontParser{})
}

type cloudfrontParser struct{}

func (p *cloudfrontParser) Service() string { return "cloudfront" }

func (p *cloudfrontParser) SupportedEvents() []string {
	return []string{
		"CreateDistribution",
		"DeleteDistribution",
		"UpdateDistribution",
		"CreateOriginAccessControl",
		"DeleteOriginAccessControl",
	}
}

func (p *cloudfrontParser) Parse(event map[string]any) (*parser.ResourceDelta, error) {
	eventID, eventTime, eventName, err := parseEvent(event)
	if err != nil {
		return nil, fmt.Errorf("cloudfront parser: %w", err)
	}

	reqParams := getMap(event, "requestParameters")
	respElems := getMap(event, "responseElements")

	delta := &parser.ResourceDelta{
		EventID:   eventID,
		EventTime: eventTime,
		Service:   "cloudfront",
	}

	switch eventName {
	case "CreateDistribution":
		delta.Action = parser.ActionCreate
		delta.ResourceType = "distribution"
		delta.Attributes = make(map[string]any)
		dist := getMap(respElems, "distribution")
		delta.ResourceID = getString(dist, "id")
		if v := getString(dist, "domainName"); v != "" {
			delta.Attributes["domainName"] = v
		}
		if v := getString(dist, "status"); v != "" {
			delta.Attributes["status"] = v
		}

	case "DeleteDistribution":
		delta.Action = parser.ActionDelete
		delta.ResourceType = "distribution"
		delta.ResourceID = getString(reqParams, "id")

	case "UpdateDistribution":
		delta.Action = parser.ActionUpdate
		delta.ResourceType = "distribution"
		delta.ResourceID = getString(reqParams, "id")

	case "CreateOriginAccessControl":
		delta.Action = parser.ActionCreate
		delta.ResourceType = "origin_access_control"
		oac := getMap(respElems, "originAccessControl")
		delta.ResourceID = getString(oac, "id")

	case "DeleteOriginAccessControl":
		delta.Action = parser.ActionDelete
		delta.ResourceType = "origin_access_control"
		delta.ResourceID = getString(reqParams, "id")

	default:
		return nil, fmt.Errorf("cloudfront parser: unsupported event %s", eventName)
	}

	return delta, nil
}
