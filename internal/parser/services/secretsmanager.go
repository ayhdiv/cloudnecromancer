package services

import (
	"fmt"

	"github.com/pfrederiksen/cloudnecromancer/internal/parser"
)

func init() {
	parser.Register(&secretsmanagerParser{})
}

type secretsmanagerParser struct{}

func (p *secretsmanagerParser) Service() string { return "secretsmanager" }

func (p *secretsmanagerParser) SupportedEvents() []string {
	return []string{
		"CreateSecret",
		"DeleteSecret",
		"UpdateSecret",
		"RotateSecret",
		"PutSecretValue",
		"RestoreSecret",
	}
}

func (p *secretsmanagerParser) Parse(event map[string]any) (*parser.ResourceDelta, error) {
	eventID, eventTime, eventName, err := parseEvent(event)
	if err != nil {
		return nil, fmt.Errorf("secretsmanager parser: %w", err)
	}

	reqParams := getMap(event, "requestParameters")
	respElems := getMap(event, "responseElements")

	delta := &parser.ResourceDelta{
		EventID:   eventID,
		EventTime: eventTime,
		Service:   "secretsmanager",
	}

	switch eventName {
	case "CreateSecret":
		delta.Action = parser.ActionCreate
		delta.ResourceType = "secret"
		delta.ResourceID = getString(respElems, "aRN")
		delta.Attributes = make(map[string]any)
		if v := getString(respElems, "name"); v != "" {
			delta.Attributes["name"] = v
		}

	case "DeleteSecret":
		delta.Action = parser.ActionDelete
		delta.ResourceType = "secret"
		delta.ResourceID = getString(respElems, "aRN")
		if delta.ResourceID == "" {
			delta.ResourceID = getString(reqParams, "secretId")
		}

	case "UpdateSecret":
		delta.Action = parser.ActionUpdate
		delta.ResourceType = "secret"
		delta.ResourceID = getString(respElems, "aRN")
		if delta.ResourceID == "" {
			delta.ResourceID = getString(reqParams, "secretId")
		}

	case "RotateSecret":
		delta.Action = parser.ActionUpdate
		delta.ResourceType = "secret"
		delta.ResourceID = getString(respElems, "aRN")
		if delta.ResourceID == "" {
			delta.ResourceID = getString(reqParams, "secretId")
		}

	case "PutSecretValue":
		delta.Action = parser.ActionUpdate
		delta.ResourceType = "secret"
		delta.ResourceID = getString(respElems, "aRN")
		if delta.ResourceID == "" {
			delta.ResourceID = getString(reqParams, "secretId")
		}

	case "RestoreSecret":
		delta.Action = parser.ActionUpdate
		delta.ResourceType = "secret"
		delta.ResourceID = getString(respElems, "aRN")
		if delta.ResourceID == "" {
			delta.ResourceID = getString(reqParams, "secretId")
		}

	default:
		return nil, fmt.Errorf("secretsmanager parser: unsupported event %s", eventName)
	}

	return delta, nil
}
