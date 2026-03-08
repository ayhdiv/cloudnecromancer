package services

import (
	"testing"

	"github.com/pfrederiksen/cloudnecromancer/internal/parser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDynamoDBParser(t *testing.T) {
	p := &dynamodbParser{}

	tests := []struct {
		name         string
		fixture      string
		wantAction   parser.Action
		wantResType  string
		wantResID    string
		wantService  string
		wantAttrKeys []string
	}{
		{
			name:         "CreateTable creates table",
			fixture:      "CreateTable",
			wantAction:   parser.ActionCreate,
			wantResType:  "table",
			wantResID:    "users-table",
			wantService:  "dynamodb",
			wantAttrKeys: []string{"tableName", "keySchema", "attributeDefinitions", "billingMode", "tableArn"},
		},
		{
			name:        "DeleteTable deletes table",
			fixture:     "DeleteTable",
			wantAction:  parser.ActionDelete,
			wantResType: "table",
			wantResID:   "users-table",
			wantService: "dynamodb",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			event := loadFixture(t, tc.fixture)
			delta, err := p.Parse(event)
			require.NoError(t, err)

			assert.Equal(t, tc.wantAction, delta.Action)
			assert.Equal(t, tc.wantResType, delta.ResourceType)
			assert.Equal(t, tc.wantResID, delta.ResourceID)
			assert.Equal(t, tc.wantService, delta.Service)
			assert.NotEmpty(t, delta.EventID)
			assert.False(t, delta.EventTime.IsZero())

			for _, key := range tc.wantAttrKeys {
				assert.Contains(t, delta.Attributes, key, "missing attribute %s", key)
			}
		})
	}
}

func TestDynamoDBParserCreateTableAttributes(t *testing.T) {
	p := &dynamodbParser{}
	event := loadFixture(t, "CreateTable")
	delta, err := p.Parse(event)
	require.NoError(t, err)

	assert.Equal(t, "users-table", delta.Attributes["tableName"])
	assert.Equal(t, "PAY_PER_REQUEST", delta.Attributes["billingMode"])
	assert.Equal(t, "arn:aws:dynamodb:us-east-1:123456789012:table/users-table", delta.Attributes["tableArn"])
}

func TestDynamoDBParserSupportedEvents(t *testing.T) {
	p := &dynamodbParser{}
	assert.Equal(t, "dynamodb", p.Service())
	events := p.SupportedEvents()
	assert.Contains(t, events, "CreateTable")
	assert.Contains(t, events, "DeleteTable")
	assert.Contains(t, events, "UpdateTable")
	assert.Contains(t, events, "CreateGlobalTable")
	assert.Contains(t, events, "UpdateGlobalTable")
	assert.Len(t, events, 5)
}

func TestDynamoDBParserInvalidEvent(t *testing.T) {
	p := &dynamodbParser{}
	_, err := p.Parse(map[string]any{})
	assert.Error(t, err)
}
