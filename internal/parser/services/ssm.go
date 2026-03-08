package services

import (
	"fmt"

	"github.com/pfrederiksen/cloudnecromancer/internal/parser"
)

func init() {
	parser.Register(&ssmParser{})
}

type ssmParser struct{}

func (p *ssmParser) Service() string { return "ssm" }

func (p *ssmParser) SupportedEvents() []string {
	return []string{
		"PutParameter",
		"DeleteParameter",
		"DeleteParameters",
		"CreateDocument",
		"DeleteDocument",
		"UpdateDocument",
		"CreateMaintenanceWindow",
		"DeleteMaintenanceWindow",
	}
}

func (p *ssmParser) Parse(event map[string]any) (*parser.ResourceDelta, error) {
	eventID, eventTime, eventName, err := parseEvent(event)
	if err != nil {
		return nil, fmt.Errorf("ssm parser: %w", err)
	}

	reqParams := getMap(event, "requestParameters")

	delta := &parser.ResourceDelta{
		EventID:   eventID,
		EventTime: eventTime,
		Service:   "ssm",
	}

	switch eventName {
	case "PutParameter":
		delta.Action = parser.ActionCreate
		delta.ResourceType = "parameter"
		delta.Attributes = make(map[string]any)
		delta.ResourceID = getString(reqParams, "name")
		if v := getString(reqParams, "type"); v != "" {
			delta.Attributes["type"] = v
		}
		if v := getString(reqParams, "tier"); v != "" {
			delta.Attributes["tier"] = v
		}

	case "DeleteParameter":
		delta.Action = parser.ActionDelete
		delta.ResourceType = "parameter"
		delta.ResourceID = getString(reqParams, "name")

	case "DeleteParameters":
		delta.Action = parser.ActionDelete
		delta.ResourceType = "parameter"
		// DeleteParameters takes a list of names; use the first as the resource ID
		// and store all names in attributes for completeness.
		names := getSlice(reqParams, "names")
		delta.Attributes = make(map[string]any)
		if len(names) > 0 {
			if first, ok := names[0].(string); ok {
				delta.ResourceID = first
			}
			delta.Attributes["names"] = names
		}

	case "CreateDocument":
		delta.Action = parser.ActionCreate
		delta.ResourceType = "document"
		delta.Attributes = make(map[string]any)
		delta.ResourceID = getString(reqParams, "name")
		if v := getString(reqParams, "documentType"); v != "" {
			delta.Attributes["documentType"] = v
		}

	case "DeleteDocument":
		delta.Action = parser.ActionDelete
		delta.ResourceType = "document"
		delta.ResourceID = getString(reqParams, "name")

	case "UpdateDocument":
		delta.Action = parser.ActionUpdate
		delta.ResourceType = "document"
		delta.ResourceID = getString(reqParams, "name")

	case "CreateMaintenanceWindow":
		delta.Action = parser.ActionCreate
		delta.ResourceType = "maintenance_window"
		respElems := getMap(event, "responseElements")
		delta.ResourceID = getString(respElems, "windowId")

	case "DeleteMaintenanceWindow":
		delta.Action = parser.ActionDelete
		delta.ResourceType = "maintenance_window"
		delta.ResourceID = getString(reqParams, "windowId")

	default:
		return nil, fmt.Errorf("ssm parser: unsupported event %s", eventName)
	}

	return delta, nil
}
