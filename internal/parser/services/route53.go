package services

import (
	"fmt"

	"github.com/pfrederiksen/cloudnecromancer/internal/parser"
)

func init() {
	parser.Register(&route53Parser{})
}

type route53Parser struct{}

func (p *route53Parser) Service() string { return "route53" }

func (p *route53Parser) SupportedEvents() []string {
	return []string{
		"CreateHostedZone",
		"DeleteHostedZone",
		"ChangeResourceRecordSets",
	}
}

func (p *route53Parser) Parse(event map[string]any) (*parser.ResourceDelta, error) {
	eventID, eventTime, eventName, err := parseEvent(event)
	if err != nil {
		return nil, fmt.Errorf("route53 parser: %w", err)
	}

	reqParams := getMap(event, "requestParameters")
	respElems := getMap(event, "responseElements")

	delta := &parser.ResourceDelta{
		EventID:   eventID,
		EventTime: eventTime,
		Service:   "route53",
	}

	switch eventName {
	case "CreateHostedZone":
		delta.Action = parser.ActionCreate
		delta.ResourceType = "hosted_zone"
		delta.Attributes = make(map[string]any)
		hostedZone := getMap(respElems, "hostedZone")
		if v := getString(hostedZone, "id"); v != "" {
			delta.ResourceID = v
			delta.Attributes["id"] = v
		}
		if v := getString(hostedZone, "name"); v != "" {
			delta.Attributes["name"] = v
		}
		if v := getString(hostedZone, "callerReference"); v != "" {
			delta.Attributes["callerReference"] = v
		}

	case "DeleteHostedZone":
		delta.Action = parser.ActionDelete
		delta.ResourceType = "hosted_zone"
		delta.ResourceID = getString(reqParams, "id")

	case "ChangeResourceRecordSets":
		delta.Action = parser.ActionUpdate
		delta.ResourceType = "hosted_zone"
		delta.ResourceID = getString(reqParams, "hostedZoneId")

	default:
		return nil, fmt.Errorf("route53 parser: unsupported event %s", eventName)
	}

	return delta, nil
}
