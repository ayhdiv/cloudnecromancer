package services

import (
	"testing"

	"github.com/pfrederiksen/cloudnecromancer/internal/parser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSSMParser(t *testing.T) {
	p := &ssmParser{}

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
			name:         "PutParameter creates parameter",
			fixture:      "PutParameter",
			wantAction:   parser.ActionCreate,
			wantResType:  "parameter",
			wantResID:    "/app/config/db-host",
			wantService:  "ssm",
			wantAttrKeys: []string{"type", "tier"},
		},
		{
			name:        "DeleteParameter deletes parameter",
			fixture:     "DeleteParameter",
			wantAction:  parser.ActionDelete,
			wantResType: "parameter",
			wantResID:   "/app/config/db-host",
			wantService: "ssm",
		},
		{
			name:         "DeleteParameters deletes multiple parameters",
			fixture:      "DeleteParameters",
			wantAction:   parser.ActionDelete,
			wantResType:  "parameter",
			wantResID:    "/app/config/old-param1",
			wantService:  "ssm",
			wantAttrKeys: []string{"names"},
		},
		{
			name:         "CreateDocument creates document",
			fixture:      "CreateDocument",
			wantAction:   parser.ActionCreate,
			wantResType:  "document",
			wantResID:    "InstallNginx",
			wantService:  "ssm",
			wantAttrKeys: []string{"documentType"},
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

func TestSSMParserPutParameterAttributes(t *testing.T) {
	p := &ssmParser{}
	event := loadFixture(t, "PutParameter")
	delta, err := p.Parse(event)
	require.NoError(t, err)

	assert.Equal(t, "SecureString", delta.Attributes["type"])
	assert.Equal(t, "Standard", delta.Attributes["tier"])
}

func TestSSMParserDeleteParametersMultiple(t *testing.T) {
	p := &ssmParser{}
	event := loadFixture(t, "DeleteParameters")
	delta, err := p.Parse(event)
	require.NoError(t, err)

	assert.Equal(t, parser.ActionDelete, delta.Action)
	assert.Equal(t, "/app/config/old-param1", delta.ResourceID)
	names, ok := delta.Attributes["names"].([]any)
	require.True(t, ok)
	assert.Len(t, names, 2)
}

func TestSSMParserSupportedEvents(t *testing.T) {
	p := &ssmParser{}
	assert.Equal(t, "ssm", p.Service())
	events := p.SupportedEvents()
	assert.Contains(t, events, "PutParameter")
	assert.Contains(t, events, "DeleteParameter")
	assert.Contains(t, events, "DeleteParameters")
	assert.Contains(t, events, "CreateDocument")
	assert.Contains(t, events, "DeleteDocument")
	assert.Contains(t, events, "UpdateDocument")
	assert.Contains(t, events, "CreateMaintenanceWindow")
	assert.Contains(t, events, "DeleteMaintenanceWindow")
	assert.Len(t, events, 8)
}

func TestSSMParserInvalidEvent(t *testing.T) {
	p := &ssmParser{}
	_, err := p.Parse(map[string]any{})
	assert.Error(t, err)
}
