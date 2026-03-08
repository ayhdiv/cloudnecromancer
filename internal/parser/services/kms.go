package services

import (
	"fmt"

	"github.com/pfrederiksen/cloudnecromancer/internal/parser"
)

func init() {
	parser.Register(&kmsParser{})
}

type kmsParser struct{}

func (p *kmsParser) Service() string { return "kms" }

func (p *kmsParser) SupportedEvents() []string {
	return []string{
		"CreateKey",
		"ScheduleKeyDeletion",
		"DisableKey",
		"EnableKey",
		"EnableKeyRotation",
		"DisableKeyRotation",
		"CreateAlias",
		"DeleteAlias",
	}
}

func (p *kmsParser) Parse(event map[string]any) (*parser.ResourceDelta, error) {
	eventID, eventTime, eventName, err := parseEvent(event)
	if err != nil {
		return nil, fmt.Errorf("kms parser: %w", err)
	}

	reqParams := getMap(event, "requestParameters")
	respElems := getMap(event, "responseElements")

	delta := &parser.ResourceDelta{
		EventID:   eventID,
		EventTime: eventTime,
		Service:   "kms",
	}

	switch eventName {
	case "CreateKey":
		delta.Action = parser.ActionCreate
		delta.ResourceType = "key"
		keyMeta := getMap(respElems, "keyMetadata")
		delta.ResourceID = getString(keyMeta, "keyId")
		delta.Attributes = make(map[string]any)
		if v := getString(keyMeta, "arn"); v != "" {
			delta.Attributes["keyArn"] = v
		}
		if v := getString(keyMeta, "keyUsage"); v != "" {
			delta.Attributes["keyUsage"] = v
		}
		if v := getString(keyMeta, "keySpec"); v != "" {
			delta.Attributes["keySpec"] = v
		}

	case "ScheduleKeyDeletion":
		delta.Action = parser.ActionDelete
		delta.ResourceType = "key"
		delta.ResourceID = getString(respElems, "keyId")
		if delta.ResourceID == "" {
			delta.ResourceID = getString(reqParams, "keyId")
		}

	case "DisableKey":
		delta.Action = parser.ActionUpdate
		delta.ResourceType = "key"
		delta.ResourceID = getString(reqParams, "keyId")

	case "EnableKey":
		delta.Action = parser.ActionUpdate
		delta.ResourceType = "key"
		delta.ResourceID = getString(reqParams, "keyId")

	case "EnableKeyRotation":
		delta.Action = parser.ActionUpdate
		delta.ResourceType = "key"
		delta.ResourceID = getString(reqParams, "keyId")

	case "DisableKeyRotation":
		delta.Action = parser.ActionUpdate
		delta.ResourceType = "key"
		delta.ResourceID = getString(reqParams, "keyId")

	case "CreateAlias":
		delta.Action = parser.ActionCreate
		delta.ResourceType = "alias"
		delta.ResourceID = getString(reqParams, "aliasName")
		delta.Attributes = make(map[string]any)
		if v := getString(reqParams, "targetKeyId"); v != "" {
			delta.Attributes["targetKeyId"] = v
		}

	case "DeleteAlias":
		delta.Action = parser.ActionDelete
		delta.ResourceType = "alias"
		delta.ResourceID = getString(reqParams, "aliasName")

	default:
		return nil, fmt.Errorf("kms parser: unsupported event %s", eventName)
	}

	return delta, nil
}
