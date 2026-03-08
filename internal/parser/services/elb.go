package services

import (
	"fmt"

	"github.com/pfrederiksen/cloudnecromancer/internal/parser"
)

func init() {
	parser.Register(&elbParser{})
}

type elbParser struct{}

func (p *elbParser) Service() string { return "elasticloadbalancing" }

func (p *elbParser) SupportedEvents() []string {
	return []string{
		"CreateLoadBalancer",
		"DeleteLoadBalancer",
		"CreateTargetGroup",
		"DeleteTargetGroup",
		"CreateListener",
		"DeleteListener",
		"RegisterTargets",
		"DeregisterTargets",
	}
}

func (p *elbParser) Parse(event map[string]any) (*parser.ResourceDelta, error) {
	eventID, eventTime, eventName, err := parseEvent(event)
	if err != nil {
		return nil, fmt.Errorf("elb parser: %w", err)
	}

	reqParams := getMap(event, "requestParameters")
	respElems := getMap(event, "responseElements")

	delta := &parser.ResourceDelta{
		EventID:   eventID,
		EventTime: eventTime,
		Service:   "elasticloadbalancing",
	}

	switch eventName {
	case "CreateLoadBalancer":
		delta.Action = parser.ActionCreate
		delta.ResourceType = "load_balancer"
		delta.Attributes = make(map[string]any)
		lbs := getSlice(respElems, "loadBalancers")
		if len(lbs) > 0 {
			lb, _ := lbs[0].(map[string]any)
			if lb != nil {
				delta.ResourceID = getString(lb, "loadBalancerArn")
				if v := getString(lb, "loadBalancerName"); v != "" {
					delta.Attributes["loadBalancerName"] = v
				}
				if v := getString(lb, "dNSName"); v != "" {
					delta.Attributes["dNSName"] = v
				}
				if v := getString(lb, "scheme"); v != "" {
					delta.Attributes["scheme"] = v
				}
				if v := getString(lb, "vpcId"); v != "" {
					delta.Attributes["vpcId"] = v
				}
				if v := getString(lb, "type"); v != "" {
					delta.Attributes["type"] = v
				}
			}
		}

	case "DeleteLoadBalancer":
		delta.Action = parser.ActionDelete
		delta.ResourceType = "load_balancer"
		delta.ResourceID = getString(reqParams, "loadBalancerArn")

	case "CreateTargetGroup":
		delta.Action = parser.ActionCreate
		delta.ResourceType = "target_group"
		delta.Attributes = make(map[string]any)
		tgs := getSlice(respElems, "targetGroups")
		if len(tgs) > 0 {
			tg, _ := tgs[0].(map[string]any)
			if tg != nil {
				delta.ResourceID = getString(tg, "targetGroupArn")
				if v := getString(tg, "targetGroupName"); v != "" {
					delta.Attributes["targetGroupName"] = v
				}
				if v := getString(tg, "protocol"); v != "" {
					delta.Attributes["protocol"] = v
				}
				if v := getString(tg, "port"); v != "" {
					delta.Attributes["port"] = v
				}
				if v := getString(tg, "vpcId"); v != "" {
					delta.Attributes["vpcId"] = v
				}
			}
		}

	case "DeleteTargetGroup":
		delta.Action = parser.ActionDelete
		delta.ResourceType = "target_group"
		delta.ResourceID = getString(reqParams, "targetGroupArn")

	case "CreateListener":
		delta.Action = parser.ActionCreate
		delta.ResourceType = "listener"
		delta.Attributes = make(map[string]any)
		listeners := getSlice(respElems, "listeners")
		if len(listeners) > 0 {
			l, _ := listeners[0].(map[string]any)
			if l != nil {
				delta.ResourceID = getString(l, "listenerArn")
				if v := getString(l, "loadBalancerArn"); v != "" {
					delta.Attributes["loadBalancerArn"] = v
				}
				if v := getString(l, "port"); v != "" {
					delta.Attributes["port"] = v
				}
				if v := getString(l, "protocol"); v != "" {
					delta.Attributes["protocol"] = v
				}
			}
		}

	case "DeleteListener":
		delta.Action = parser.ActionDelete
		delta.ResourceType = "listener"
		delta.ResourceID = getString(reqParams, "listenerArn")

	case "RegisterTargets":
		delta.Action = parser.ActionUpdate
		delta.ResourceType = "target_group"
		delta.ResourceID = getString(reqParams, "targetGroupArn")

	case "DeregisterTargets":
		delta.Action = parser.ActionUpdate
		delta.ResourceType = "target_group"
		delta.ResourceID = getString(reqParams, "targetGroupArn")

	default:
		return nil, fmt.Errorf("elb parser: unsupported event %s", eventName)
	}

	return delta, nil
}
