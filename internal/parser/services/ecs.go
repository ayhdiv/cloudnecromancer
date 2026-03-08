package services

import (
	"fmt"

	"github.com/pfrederiksen/cloudnecromancer/internal/parser"
)

func init() {
	parser.Register(&ecsParser{})
}

type ecsParser struct{}

func (p *ecsParser) Service() string { return "ecs" }

func (p *ecsParser) SupportedEvents() []string {
	return []string{
		"CreateCluster",
		"DeleteCluster",
		"CreateService",
		"DeleteService",
		"UpdateService",
		"RegisterTaskDefinition",
		"DeregisterTaskDefinition",
	}
}

func (p *ecsParser) Parse(event map[string]any) (*parser.ResourceDelta, error) {
	eventID, eventTime, eventName, err := parseEvent(event)
	if err != nil {
		return nil, fmt.Errorf("ecs parser: %w", err)
	}

	reqParams := getMap(event, "requestParameters")
	respElems := getMap(event, "responseElements")

	delta := &parser.ResourceDelta{
		EventID:   eventID,
		EventTime: eventTime,
		Service:   "ecs",
	}

	switch eventName {
	case "CreateCluster":
		delta.Action = parser.ActionCreate
		delta.ResourceType = "cluster"
		cluster := getMap(respElems, "cluster")
		delta.ResourceID = getString(cluster, "clusterArn")
		delta.Attributes = make(map[string]any)
		if v := getString(cluster, "clusterName"); v != "" {
			delta.Attributes["clusterName"] = v
		}

	case "DeleteCluster":
		delta.Action = parser.ActionDelete
		delta.ResourceType = "cluster"
		cluster := getMap(respElems, "cluster")
		delta.ResourceID = getString(cluster, "clusterArn")
		if delta.ResourceID == "" {
			delta.ResourceID = getString(reqParams, "cluster")
		}

	case "CreateService":
		delta.Action = parser.ActionCreate
		delta.ResourceType = "service"
		svc := getMap(respElems, "service")
		delta.ResourceID = getString(svc, "serviceArn")
		delta.Attributes = make(map[string]any)
		if v := getString(svc, "serviceName"); v != "" {
			delta.Attributes["serviceName"] = v
		}
		if v := getString(svc, "clusterArn"); v != "" {
			delta.Attributes["clusterArn"] = v
		}
		if v := getString(svc, "taskDefinition"); v != "" {
			delta.Attributes["taskDefinition"] = v
		}
		if v, ok := svc["desiredCount"]; ok {
			delta.Attributes["desiredCount"] = v
		}
		if v := getString(svc, "launchType"); v != "" {
			delta.Attributes["launchType"] = v
		}

	case "DeleteService":
		delta.Action = parser.ActionDelete
		delta.ResourceType = "service"
		svc := getMap(respElems, "service")
		delta.ResourceID = getString(svc, "serviceArn")
		if delta.ResourceID == "" {
			delta.ResourceID = getString(reqParams, "service")
		}

	case "UpdateService":
		delta.Action = parser.ActionUpdate
		delta.ResourceType = "service"
		svc := getMap(respElems, "service")
		delta.ResourceID = getString(svc, "serviceArn")
		delta.Attributes = make(map[string]any)
		if v, ok := svc["desiredCount"]; ok {
			delta.Attributes["desiredCount"] = v
		}
		if v := getString(svc, "taskDefinition"); v != "" {
			delta.Attributes["taskDefinition"] = v
		}

	case "RegisterTaskDefinition":
		delta.Action = parser.ActionCreate
		delta.ResourceType = "task_definition"
		td := getMap(respElems, "taskDefinition")
		delta.ResourceID = getString(td, "taskDefinitionArn")
		delta.Attributes = make(map[string]any)
		if v := getString(td, "family"); v != "" {
			delta.Attributes["family"] = v
		}
		if v := getString(td, "cpu"); v != "" {
			delta.Attributes["cpu"] = v
		}
		if v := getString(td, "memory"); v != "" {
			delta.Attributes["memory"] = v
		}

	case "DeregisterTaskDefinition":
		delta.Action = parser.ActionDelete
		delta.ResourceType = "task_definition"
		td := getMap(respElems, "taskDefinition")
		delta.ResourceID = getString(td, "taskDefinitionArn")
		if delta.ResourceID == "" {
			delta.ResourceID = getString(reqParams, "taskDefinition")
		}

	default:
		return nil, fmt.Errorf("ecs parser: unsupported event %s", eventName)
	}

	return delta, nil
}
