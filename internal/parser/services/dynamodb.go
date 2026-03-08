package services

import (
	"fmt"

	"github.com/pfrederiksen/cloudnecromancer/internal/parser"
)

func init() {
	parser.Register(&dynamodbParser{})
}

type dynamodbParser struct{}

func (p *dynamodbParser) Service() string { return "dynamodb" }

func (p *dynamodbParser) SupportedEvents() []string {
	return []string{
		"CreateTable",
		"DeleteTable",
		"UpdateTable",
		"CreateGlobalTable",
		"UpdateGlobalTable",
	}
}

func (p *dynamodbParser) Parse(event map[string]any) (*parser.ResourceDelta, error) {
	eventID, eventTime, eventName, err := parseEvent(event)
	if err != nil {
		return nil, fmt.Errorf("dynamodb parser: %w", err)
	}

	reqParams := getMap(event, "requestParameters")
	respElems := getMap(event, "responseElements")

	delta := &parser.ResourceDelta{
		EventID:   eventID,
		EventTime: eventTime,
		Service:   "dynamodb",
	}

	switch eventName {
	case "CreateTable":
		delta.Action = parser.ActionCreate
		delta.ResourceType = "table"
		delta.ResourceID = getString(reqParams, "tableName")
		delta.Attributes = make(map[string]any)
		if v := getString(reqParams, "tableName"); v != "" {
			delta.Attributes["tableName"] = v
		}
		if v := getSlice(reqParams, "keySchema"); v != nil {
			delta.Attributes["keySchema"] = v
		}
		if v := getSlice(reqParams, "attributeDefinitions"); v != nil {
			delta.Attributes["attributeDefinitions"] = v
		}
		if v := getString(reqParams, "billingMode"); v != "" {
			delta.Attributes["billingMode"] = v
		}
		tableDesc := getMap(respElems, "tableDescription")
		if v := getString(tableDesc, "tableArn"); v != "" {
			delta.Attributes["tableArn"] = v
		}

	case "DeleteTable":
		delta.Action = parser.ActionDelete
		delta.ResourceType = "table"
		delta.ResourceID = getString(reqParams, "tableName")

	case "UpdateTable":
		delta.Action = parser.ActionUpdate
		delta.ResourceType = "table"
		delta.ResourceID = getString(reqParams, "tableName")

	case "CreateGlobalTable":
		delta.Action = parser.ActionCreate
		delta.ResourceType = "global_table"
		delta.ResourceID = getString(reqParams, "globalTableName")

	case "UpdateGlobalTable":
		delta.Action = parser.ActionUpdate
		delta.ResourceType = "global_table"
		delta.ResourceID = getString(reqParams, "globalTableName")

	default:
		return nil, fmt.Errorf("dynamodb parser: unsupported event %s", eventName)
	}

	return delta, nil
}
