package services

import (
	"testing"

	"github.com/pfrederiksen/cloudnecromancer/internal/parser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestECSParser(t *testing.T) {
	p := &ecsParser{}

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
			name:         "CreateCluster creates cluster",
			fixture:      "ECSCreateCluster",
			wantAction:   parser.ActionCreate,
			wantResType:  "cluster",
			wantResID:    "arn:aws:ecs:us-east-1:123456789012:cluster/production-cluster",
			wantService:  "ecs",
			wantAttrKeys: []string{"clusterName"},
		},
		{
			name:        "DeleteCluster deletes cluster",
			fixture:     "ECSDeleteCluster",
			wantAction:  parser.ActionDelete,
			wantResType: "cluster",
			wantResID:   "arn:aws:ecs:us-east-1:123456789012:cluster/production-cluster",
			wantService: "ecs",
		},
		{
			name:         "RegisterTaskDefinition creates task_definition",
			fixture:      "RegisterTaskDefinition",
			wantAction:   parser.ActionCreate,
			wantResType:  "task_definition",
			wantResID:    "arn:aws:ecs:us-east-1:123456789012:task-definition/my-web-app:1",
			wantService:  "ecs",
			wantAttrKeys: []string{"family", "cpu", "memory"},
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

func TestECSParserRegisterTaskDefinitionAttributes(t *testing.T) {
	p := &ecsParser{}
	event := loadFixture(t, "RegisterTaskDefinition")
	delta, err := p.Parse(event)
	require.NoError(t, err)

	assert.Equal(t, "my-web-app", delta.Attributes["family"])
	assert.Equal(t, "256", delta.Attributes["cpu"])
	assert.Equal(t, "512", delta.Attributes["memory"])
}

func TestECSParserSupportedEvents(t *testing.T) {
	p := &ecsParser{}
	assert.Equal(t, "ecs", p.Service())
	events := p.SupportedEvents()
	assert.Contains(t, events, "CreateCluster")
	assert.Contains(t, events, "DeleteCluster")
	assert.Contains(t, events, "RegisterTaskDefinition")
	assert.Len(t, events, 7)
}

func TestECSParserInvalidEvent(t *testing.T) {
	p := &ecsParser{}
	_, err := p.Parse(map[string]any{})
	assert.Error(t, err)
}
