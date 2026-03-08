package services

import (
	"fmt"

	"github.com/pfrederiksen/cloudnecromancer/internal/parser"
)

func init() {
	parser.Register(&ebsParser{})
}

type ebsParser struct{}

func (p *ebsParser) Service() string { return "ebs" }

func (p *ebsParser) SupportedEvents() []string {
	return []string{
		"CreateVolume",
		"DeleteVolume",
		"CreateSnapshot",
		"DeleteSnapshot",
		"ModifyVolume",
	}
}

func (p *ebsParser) Parse(event map[string]any) (*parser.ResourceDelta, error) {
	eventID, eventTime, eventName, err := parseEvent(event)
	if err != nil {
		return nil, fmt.Errorf("ebs parser: %w", err)
	}

	reqParams := getMap(event, "requestParameters")
	respElems := getMap(event, "responseElements")

	delta := &parser.ResourceDelta{
		EventID:   eventID,
		EventTime: eventTime,
		Service:   "ebs",
	}

	switch eventName {
	case "CreateVolume":
		delta.Action = parser.ActionCreate
		delta.ResourceType = "volume"
		delta.Attributes = make(map[string]any)
		delta.ResourceID = getString(respElems, "volumeId")
		if v := getString(respElems, "size"); v != "" {
			delta.Attributes["size"] = v
		}
		if v := getString(respElems, "volumeType"); v != "" {
			delta.Attributes["volumeType"] = v
		}
		if v := getString(respElems, "availabilityZone"); v != "" {
			delta.Attributes["availabilityZone"] = v
		}
		if v := getString(respElems, "encrypted"); v != "" {
			delta.Attributes["encrypted"] = v
		}

	case "DeleteVolume":
		delta.Action = parser.ActionDelete
		delta.ResourceType = "volume"
		delta.ResourceID = getString(reqParams, "volumeId")

	case "CreateSnapshot":
		delta.Action = parser.ActionCreate
		delta.ResourceType = "snapshot"
		delta.Attributes = make(map[string]any)
		delta.ResourceID = getString(respElems, "snapshotId")
		if v := getString(respElems, "volumeId"); v != "" {
			delta.Attributes["volumeId"] = v
		}

	case "DeleteSnapshot":
		delta.Action = parser.ActionDelete
		delta.ResourceType = "snapshot"
		delta.ResourceID = getString(reqParams, "snapshotId")

	case "ModifyVolume":
		delta.Action = parser.ActionUpdate
		delta.ResourceType = "volume"
		delta.ResourceID = getString(reqParams, "volumeId")

	default:
		return nil, fmt.Errorf("ebs parser: unsupported event %s", eventName)
	}

	return delta, nil
}
