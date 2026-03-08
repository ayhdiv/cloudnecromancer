package services

import (
	"testing"

	"github.com/pfrederiksen/cloudnecromancer/internal/parser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSecretsManagerParser(t *testing.T) {
	p := &secretsmanagerParser{}

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
			name:         "CreateSecret creates secret",
			fixture:      "CreateSecret",
			wantAction:   parser.ActionCreate,
			wantResType:  "secret",
			wantResID:    "arn:aws:secretsmanager:us-east-1:123456789012:secret:prod/database/credentials-a1b2c3",
			wantService:  "secretsmanager",
			wantAttrKeys: []string{"name"},
		},
		{
			name:        "DeleteSecret deletes secret",
			fixture:     "DeleteSecret",
			wantAction:  parser.ActionDelete,
			wantResType: "secret",
			wantResID:   "arn:aws:secretsmanager:us-east-1:123456789012:secret:prod/database/credentials-a1b2c3",
			wantService: "secretsmanager",
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

func TestSecretsManagerParserCreateSecretAttributes(t *testing.T) {
	p := &secretsmanagerParser{}
	event := loadFixture(t, "CreateSecret")
	delta, err := p.Parse(event)
	require.NoError(t, err)

	assert.Equal(t, "prod/database/credentials", delta.Attributes["name"])
}

func TestSecretsManagerParserSupportedEvents(t *testing.T) {
	p := &secretsmanagerParser{}
	assert.Equal(t, "secretsmanager", p.Service())
	events := p.SupportedEvents()
	assert.Contains(t, events, "CreateSecret")
	assert.Contains(t, events, "DeleteSecret")
	assert.Contains(t, events, "UpdateSecret")
	assert.Contains(t, events, "RotateSecret")
	assert.Contains(t, events, "PutSecretValue")
	assert.Contains(t, events, "RestoreSecret")
	assert.Len(t, events, 6)
}

func TestSecretsManagerParserInvalidEvent(t *testing.T) {
	p := &secretsmanagerParser{}
	_, err := p.Parse(map[string]any{})
	assert.Error(t, err)
}
