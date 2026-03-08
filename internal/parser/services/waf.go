package services

import (
	"fmt"

	"github.com/pfrederiksen/cloudnecromancer/internal/parser"
)

func init() {
	parser.Register(&wafParser{})
}

type wafParser struct{}

func (p *wafParser) Service() string { return "waf" }

func (p *wafParser) SupportedEvents() []string {
	return []string{
		"CreateWebACL",
		"DeleteWebACL",
		"UpdateWebACL",
		"CreateRuleGroup",
		"DeleteRuleGroup",
		// CreateIPSet and DeleteIPSet are intentionally omitted from registration
		// because GuardDuty also uses the same event names. The WAF parser can
		// still handle them via direct Parse() calls when the eventSource is known.
	}
}

func (p *wafParser) Parse(event map[string]any) (*parser.ResourceDelta, error) {
	eventID, eventTime, eventName, err := parseEvent(event)
	if err != nil {
		return nil, fmt.Errorf("waf parser: %w", err)
	}

	reqParams := getMap(event, "requestParameters")
	respElems := getMap(event, "responseElements")

	delta := &parser.ResourceDelta{
		EventID:   eventID,
		EventTime: eventTime,
		Service:   "waf",
	}

	switch eventName {
	case "CreateWebACL":
		delta.Action = parser.ActionCreate
		delta.ResourceType = "web_acl"
		delta.Attributes = make(map[string]any)
		summary := getMap(respElems, "summary")
		delta.ResourceID = getString(summary, "id")
		if v := getString(summary, "name"); v != "" {
			delta.Attributes["name"] = v
		}
		if v := getString(summary, "aRN"); v != "" {
			delta.Attributes["aRN"] = v
		}
		if v := getString(reqParams, "scope"); v != "" {
			delta.Attributes["scope"] = v
		}

	case "DeleteWebACL":
		delta.Action = parser.ActionDelete
		delta.ResourceType = "web_acl"
		delta.Attributes = make(map[string]any)
		delta.ResourceID = getString(reqParams, "id")
		if v := getString(reqParams, "name"); v != "" {
			delta.Attributes["name"] = v
		}

	case "UpdateWebACL":
		delta.Action = parser.ActionUpdate
		delta.ResourceType = "web_acl"
		delta.ResourceID = getString(reqParams, "id")

	case "CreateRuleGroup":
		delta.Action = parser.ActionCreate
		delta.ResourceType = "rule_group"
		delta.Attributes = make(map[string]any)
		summary := getMap(respElems, "summary")
		delta.ResourceID = getString(summary, "id")
		if v := getString(summary, "name"); v != "" {
			delta.Attributes["name"] = v
		}

	case "DeleteRuleGroup":
		delta.Action = parser.ActionDelete
		delta.ResourceType = "rule_group"
		delta.ResourceID = getString(reqParams, "id")

	case "CreateIPSet":
		delta.Action = parser.ActionCreate
		delta.ResourceType = "ip_set"
		delta.Attributes = make(map[string]any)
		summary := getMap(respElems, "summary")
		delta.ResourceID = getString(summary, "id")
		if v := getString(summary, "name"); v != "" {
			delta.Attributes["name"] = v
		}

	case "DeleteIPSet":
		delta.Action = parser.ActionDelete
		delta.ResourceType = "ip_set"
		delta.ResourceID = getString(reqParams, "id")

	default:
		return nil, fmt.Errorf("waf parser: unsupported event %s", eventName)
	}

	return delta, nil
}
