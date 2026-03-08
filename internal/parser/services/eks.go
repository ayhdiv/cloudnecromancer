package services

import (
	"fmt"

	"github.com/pfrederiksen/cloudnecromancer/internal/parser"
)

func init() {
	parser.Register(&eksParser{})
}

type eksParser struct{}

func (p *eksParser) Service() string { return "eks" }

func (p *eksParser) SupportedEvents() []string {
	return []string{
		// CreateCluster and DeleteCluster are intentionally omitted from
		// registration because ECS also uses the same event names. The EKS
		// parser can still handle them via direct Parse() calls when the
		// eventSource is known.
		"CreateNodegroup",
		"DeleteNodegroup",
		"UpdateClusterVersion",
	}
}

func (p *eksParser) Parse(event map[string]any) (*parser.ResourceDelta, error) {
	eventID, eventTime, eventName, err := parseEvent(event)
	if err != nil {
		return nil, fmt.Errorf("eks parser: %w", err)
	}

	reqParams := getMap(event, "requestParameters")
	respElems := getMap(event, "responseElements")

	delta := &parser.ResourceDelta{
		EventID:   eventID,
		EventTime: eventTime,
		Service:   "eks",
	}

	switch eventName {
	case "CreateCluster":
		delta.Action = parser.ActionCreate
		delta.ResourceType = "cluster"
		cluster := getMap(respElems, "cluster")
		delta.ResourceID = getString(cluster, "name")
		delta.Attributes = make(map[string]any)
		if v := getString(cluster, "version"); v != "" {
			delta.Attributes["version"] = v
		}
		if v := getString(cluster, "roleArn"); v != "" {
			delta.Attributes["roleArn"] = v
		}
		vpcConfig := getMap(cluster, "resourcesVpcConfig")
		if v := getString(vpcConfig, "vpcId"); v != "" {
			delta.Attributes["vpcId"] = v
		}

	case "DeleteCluster":
		delta.Action = parser.ActionDelete
		delta.ResourceType = "cluster"
		cluster := getMap(respElems, "cluster")
		delta.ResourceID = getString(cluster, "name")
		if delta.ResourceID == "" {
			delta.ResourceID = getString(reqParams, "name")
		}

	case "CreateNodegroup":
		delta.Action = parser.ActionCreate
		delta.ResourceType = "nodegroup"
		ng := getMap(respElems, "nodegroup")
		delta.ResourceID = getString(ng, "nodegroupArn")
		delta.Attributes = make(map[string]any)
		if v := getString(ng, "nodegroupName"); v != "" {
			delta.Attributes["nodegroupName"] = v
		}
		if v := getString(ng, "clusterName"); v != "" {
			delta.Attributes["clusterName"] = v
		}
		if instanceTypes := getSlice(ng, "instanceTypes"); len(instanceTypes) > 0 {
			delta.Attributes["instanceTypes"] = instanceTypes
		}
		if scalingConfig := getMap(ng, "scalingConfig"); scalingConfig != nil {
			delta.Attributes["scalingConfig"] = scalingConfig
		}

	case "DeleteNodegroup":
		delta.Action = parser.ActionDelete
		delta.ResourceType = "nodegroup"
		ng := getMap(respElems, "nodegroup")
		delta.ResourceID = getString(ng, "nodegroupArn")
		if delta.ResourceID == "" {
			delta.ResourceID = getString(reqParams, "nodegroupName")
		}

	case "UpdateClusterVersion":
		delta.Action = parser.ActionUpdate
		delta.ResourceType = "cluster"
		update := getMap(respElems, "update")
		delta.ResourceID = getString(reqParams, "name")
		delta.Attributes = make(map[string]any)
		if v := getString(update, "id"); v != "" {
			delta.Attributes["updateId"] = v
		}

	default:
		return nil, fmt.Errorf("eks parser: unsupported event %s", eventName)
	}

	return delta, nil
}
