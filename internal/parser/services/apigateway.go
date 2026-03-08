package services

import (
	"fmt"

	"github.com/pfrederiksen/cloudnecromancer/internal/parser"
)

func init() {
	parser.Register(&apigatewayParser{})
}

type apigatewayParser struct{}

func (p *apigatewayParser) Service() string { return "apigateway" }

func (p *apigatewayParser) SupportedEvents() []string {
	return []string{
		"CreateRestApi",
		"DeleteRestApi",
		"CreateApi",
		"DeleteApi",
		"CreateStage",
		"DeleteStage",
	}
}

func (p *apigatewayParser) Parse(event map[string]any) (*parser.ResourceDelta, error) {
	eventID, eventTime, eventName, err := parseEvent(event)
	if err != nil {
		return nil, fmt.Errorf("apigateway parser: %w", err)
	}

	reqParams := getMap(event, "requestParameters")
	respElems := getMap(event, "responseElements")

	delta := &parser.ResourceDelta{
		EventID:   eventID,
		EventTime: eventTime,
		Service:   "apigateway",
	}

	switch eventName {
	case "CreateRestApi":
		delta.Action = parser.ActionCreate
		delta.ResourceType = "rest_api"
		delta.ResourceID = getString(respElems, "id")
		delta.Attributes = make(map[string]any)
		if v := getString(respElems, "name"); v != "" {
			delta.Attributes["name"] = v
		}
		if v := getString(respElems, "description"); v != "" {
			delta.Attributes["description"] = v
		}

	case "DeleteRestApi":
		delta.Action = parser.ActionDelete
		delta.ResourceType = "rest_api"
		delta.ResourceID = getString(reqParams, "restApiId")

	case "CreateApi":
		delta.Action = parser.ActionCreate
		delta.ResourceType = "http_api"
		delta.ResourceID = getString(respElems, "apiId")
		delta.Attributes = make(map[string]any)
		if v := getString(respElems, "name"); v != "" {
			delta.Attributes["name"] = v
		}
		if v := getString(respElems, "protocolType"); v != "" {
			delta.Attributes["protocolType"] = v
		}

	case "DeleteApi":
		delta.Action = parser.ActionDelete
		delta.ResourceType = "http_api"
		delta.ResourceID = getString(reqParams, "apiId")

	case "CreateStage":
		delta.Action = parser.ActionCreate
		delta.ResourceType = "stage"
		delta.ResourceID = getString(reqParams, "stageName")
		delta.Attributes = make(map[string]any)
		if v := getString(reqParams, "restApiId"); v != "" {
			delta.Attributes["restApiId"] = v
		}
		if v := getString(reqParams, "apiId"); v != "" {
			delta.Attributes["apiId"] = v
		}

	case "DeleteStage":
		delta.Action = parser.ActionDelete
		delta.ResourceType = "stage"
		delta.ResourceID = getString(reqParams, "stageName")

	default:
		return nil, fmt.Errorf("apigateway parser: unsupported event %s", eventName)
	}

	return delta, nil
}
