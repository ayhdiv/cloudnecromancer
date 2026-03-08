package services

import (
	"testing"

	"github.com/pfrederiksen/cloudnecromancer/internal/parser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestECRParser(t *testing.T) {
	p := &ecrParser{}

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
			name:         "CreateRepository creates repository",
			fixture:      "CreateRepository",
			wantAction:   parser.ActionCreate,
			wantResType:  "repository",
			wantResID:    "my-app/backend",
			wantService:  "ecr",
			wantAttrKeys: []string{"repositoryName", "repositoryArn", "repositoryUri"},
		},
		{
			name:        "DeleteRepository deletes repository",
			fixture:     "DeleteRepository",
			wantAction:  parser.ActionDelete,
			wantResType: "repository",
			wantResID:   "my-app/backend",
			wantService: "ecr",
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

func TestECRParserCreateRepositoryAttributes(t *testing.T) {
	p := &ecrParser{}
	event := loadFixture(t, "CreateRepository")
	delta, err := p.Parse(event)
	require.NoError(t, err)

	assert.Equal(t, "my-app/backend", delta.Attributes["repositoryName"])
	assert.Equal(t, "arn:aws:ecr:us-east-1:123456789012:repository/my-app/backend", delta.Attributes["repositoryArn"])
	assert.Equal(t, "123456789012.dkr.ecr.us-east-1.amazonaws.com/my-app/backend", delta.Attributes["repositoryUri"])
}

func TestECRParserSupportedEvents(t *testing.T) {
	p := &ecrParser{}
	assert.Equal(t, "ecr", p.Service())
	events := p.SupportedEvents()
	assert.Contains(t, events, "CreateRepository")
	assert.Contains(t, events, "DeleteRepository")
	assert.Contains(t, events, "PutLifecyclePolicy")
	assert.Contains(t, events, "PutImageScanningConfiguration")
	assert.Len(t, events, 4)
}

func TestECRParserInvalidEvent(t *testing.T) {
	p := &ecrParser{}
	_, err := p.Parse(map[string]any{})
	assert.Error(t, err)
}
