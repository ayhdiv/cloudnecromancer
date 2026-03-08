package services

import (
	"fmt"

	"github.com/pfrederiksen/cloudnecromancer/internal/parser"
)

func init() {
	parser.Register(&ecrParser{})
}

type ecrParser struct{}

func (p *ecrParser) Service() string { return "ecr" }

func (p *ecrParser) SupportedEvents() []string {
	return []string{
		"CreateRepository",
		"DeleteRepository",
		"PutLifecyclePolicy",
		"PutImageScanningConfiguration",
	}
}

func (p *ecrParser) Parse(event map[string]any) (*parser.ResourceDelta, error) {
	eventID, eventTime, eventName, err := parseEvent(event)
	if err != nil {
		return nil, fmt.Errorf("ecr parser: %w", err)
	}

	reqParams := getMap(event, "requestParameters")
	respElems := getMap(event, "responseElements")

	delta := &parser.ResourceDelta{
		EventID:   eventID,
		EventTime: eventTime,
		Service:   "ecr",
	}

	switch eventName {
	case "CreateRepository":
		delta.Action = parser.ActionCreate
		delta.ResourceType = "repository"
		delta.Attributes = make(map[string]any)
		repo := getMap(respElems, "repository")
		if v := getString(repo, "repositoryName"); v != "" {
			delta.ResourceID = v
			delta.Attributes["repositoryName"] = v
		}
		if v := getString(repo, "repositoryArn"); v != "" {
			delta.Attributes["repositoryArn"] = v
		}
		if v := getString(repo, "repositoryUri"); v != "" {
			delta.Attributes["repositoryUri"] = v
		}

	case "DeleteRepository":
		delta.Action = parser.ActionDelete
		delta.ResourceType = "repository"
		delta.ResourceID = getString(reqParams, "repositoryName")

	case "PutLifecyclePolicy":
		delta.Action = parser.ActionUpdate
		delta.ResourceType = "repository"
		delta.ResourceID = getString(reqParams, "repositoryName")

	case "PutImageScanningConfiguration":
		delta.Action = parser.ActionUpdate
		delta.ResourceType = "repository"
		delta.ResourceID = getString(reqParams, "repositoryName")

	default:
		return nil, fmt.Errorf("ecr parser: unsupported event %s", eventName)
	}

	return delta, nil
}
