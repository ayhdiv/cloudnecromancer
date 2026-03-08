package services

import (
	"fmt"

	"github.com/pfrederiksen/cloudnecromancer/internal/parser"
)

func init() {
	parser.Register(&elasticacheParser{})
}

type elasticacheParser struct{}

func (p *elasticacheParser) Service() string { return "elasticache" }

func (p *elasticacheParser) SupportedEvents() []string {
	return []string{
		"CreateCacheCluster",
		"DeleteCacheCluster",
		"ModifyCacheCluster",
		"CreateReplicationGroup",
		"DeleteReplicationGroup",
	}
}

func (p *elasticacheParser) Parse(event map[string]any) (*parser.ResourceDelta, error) {
	eventID, eventTime, eventName, err := parseEvent(event)
	if err != nil {
		return nil, fmt.Errorf("elasticache parser: %w", err)
	}

	reqParams := getMap(event, "requestParameters")

	delta := &parser.ResourceDelta{
		EventID:   eventID,
		EventTime: eventTime,
		Service:   "elasticache",
	}

	switch eventName {
	case "CreateCacheCluster":
		delta.Action = parser.ActionCreate
		delta.ResourceType = "cache_cluster"
		delta.ResourceID = getString(reqParams, "cacheClusterId")
		delta.Attributes = make(map[string]any)
		if v := getString(reqParams, "cacheClusterId"); v != "" {
			delta.Attributes["cacheClusterId"] = v
		}
		if v := getString(reqParams, "engine"); v != "" {
			delta.Attributes["engine"] = v
		}
		if v := getString(reqParams, "cacheNodeType"); v != "" {
			delta.Attributes["cacheNodeType"] = v
		}
		if v := getString(reqParams, "numCacheNodes"); v != "" {
			delta.Attributes["numCacheNodes"] = v
		}

	case "DeleteCacheCluster":
		delta.Action = parser.ActionDelete
		delta.ResourceType = "cache_cluster"
		delta.ResourceID = getString(reqParams, "cacheClusterId")

	case "ModifyCacheCluster":
		delta.Action = parser.ActionUpdate
		delta.ResourceType = "cache_cluster"
		delta.ResourceID = getString(reqParams, "cacheClusterId")

	case "CreateReplicationGroup":
		delta.Action = parser.ActionCreate
		delta.ResourceType = "replication_group"
		delta.ResourceID = getString(reqParams, "replicationGroupId")
		delta.Attributes = make(map[string]any)
		if v := getString(reqParams, "replicationGroupId"); v != "" {
			delta.Attributes["replicationGroupId"] = v
		}
		if v := getString(reqParams, "replicationGroupDescription"); v != "" {
			delta.Attributes["description"] = v
		}
		if v := getString(reqParams, "engine"); v != "" {
			delta.Attributes["engine"] = v
		}

	case "DeleteReplicationGroup":
		delta.Action = parser.ActionDelete
		delta.ResourceType = "replication_group"
		delta.ResourceID = getString(reqParams, "replicationGroupId")

	default:
		return nil, fmt.Errorf("elasticache parser: unsupported event %s", eventName)
	}

	return delta, nil
}
