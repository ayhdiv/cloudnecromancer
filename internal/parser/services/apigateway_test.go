package services

import (
	"testing"

	"github.com/pfrederiksen/cloudnecromancer/internal/parser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAPIGatewayParser(t *testing.T) {
	p := &apigatewayParser{}

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
			name:         "CreateRestApi creates rest_api",
			fixture:      "CreateRestApi",
			wantAction:   parser.ActionCreate,
			wantResType:  "rest_api",
			wantResID:    "abc123def4",
			wantService:  "apigateway",
			wantAttrKeys: []string{"name", "description"},
		},
		{
			name:        "DeleteRestApi deletes rest_api",
			fixture:     "DeleteRestApi",
			wantAction:  parser.ActionDelete,
			wantResType: "rest_api",
			wantResID:   "abc123def4",
			wantService: "apigateway",
		},
		{
			name:         "CreateApi creates http_api",
			fixture:      "CreateApi",
			wantAction:   parser.ActionCreate,
			wantResType:  "http_api",
			wantResID:    "xyz789ghi0",
			wantService:  "apigateway",
			wantAttrKeys: []string{"name", "protocolType"},
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

func TestAPIGatewayParserCreateRestApiAttributes(t *testing.T) {
	p := &apigatewayParser{}
	event := loadFixture(t, "CreateRestApi")
	delta, err := p.Parse(event)
	require.NoError(t, err)

	assert.Equal(t, "my-rest-api", delta.Attributes["name"])
	assert.Equal(t, "My REST API for orders", delta.Attributes["description"])
}

func TestAPIGatewayParserCreateApiAttributes(t *testing.T) {
	p := &apigatewayParser{}
	event := loadFixture(t, "CreateApi")
	delta, err := p.Parse(event)
	require.NoError(t, err)

	assert.Equal(t, "my-http-api", delta.Attributes["name"])
	assert.Equal(t, "HTTP", delta.Attributes["protocolType"])
}

func TestAPIGatewayParserSupportedEvents(t *testing.T) {
	p := &apigatewayParser{}
	assert.Equal(t, "apigateway", p.Service())
	events := p.SupportedEvents()
	assert.Contains(t, events, "CreateRestApi")
	assert.Contains(t, events, "DeleteRestApi")
	assert.Contains(t, events, "CreateApi")
	assert.Contains(t, events, "DeleteApi")
	assert.Contains(t, events, "CreateStage")
	assert.Contains(t, events, "DeleteStage")
	assert.Len(t, events, 6)
}

func TestAPIGatewayParserInvalidEvent(t *testing.T) {
	p := &apigatewayParser{}
	_, err := p.Parse(map[string]any{})
	assert.Error(t, err)
}
