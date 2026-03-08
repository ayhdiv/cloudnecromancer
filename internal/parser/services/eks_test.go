package services

import (
	"testing"

	"github.com/pfrederiksen/cloudnecromancer/internal/parser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEKSParser(t *testing.T) {
	p := &eksParser{}

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
			fixture:      "EKSCreateCluster",
			wantAction:   parser.ActionCreate,
			wantResType:  "cluster",
			wantResID:    "prod-k8s",
			wantService:  "eks",
			wantAttrKeys: []string{"version", "roleArn", "vpcId"},
		},
		{
			name:         "CreateNodegroup creates nodegroup",
			fixture:      "CreateNodegroup",
			wantAction:   parser.ActionCreate,
			wantResType:  "nodegroup",
			wantResID:    "arn:aws:eks:us-east-1:123456789012:nodegroup/prod-k8s/worker-nodes/a1b2c3d4",
			wantService:  "eks",
			wantAttrKeys: []string{"nodegroupName", "clusterName", "instanceTypes", "scalingConfig"},
		},
		{
			name:        "DeleteNodegroup deletes nodegroup",
			fixture:     "DeleteNodegroup",
			wantAction:  parser.ActionDelete,
			wantResType: "nodegroup",
			wantResID:   "arn:aws:eks:us-east-1:123456789012:nodegroup/prod-k8s/worker-nodes/a1b2c3d4",
			wantService: "eks",
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

func TestEKSParserCreateClusterAttributes(t *testing.T) {
	p := &eksParser{}
	event := loadFixture(t, "EKSCreateCluster")
	delta, err := p.Parse(event)
	require.NoError(t, err)

	assert.Equal(t, "1.28", delta.Attributes["version"])
	assert.Equal(t, "arn:aws:iam::123456789012:role/EKSClusterRole", delta.Attributes["roleArn"])
	assert.Equal(t, "vpc-0123456789abcdef0", delta.Attributes["vpcId"])
}

func TestEKSParserSupportedEvents(t *testing.T) {
	p := &eksParser{}
	assert.Equal(t, "eks", p.Service())
	events := p.SupportedEvents()
	// CreateCluster and DeleteCluster are omitted due to collision with ECS
	assert.Contains(t, events, "CreateNodegroup")
	assert.Contains(t, events, "DeleteNodegroup")
	assert.Contains(t, events, "UpdateClusterVersion")
	assert.Len(t, events, 3)
}

func TestEKSParserInvalidEvent(t *testing.T) {
	p := &eksParser{}
	_, err := p.Parse(map[string]any{})
	assert.Error(t, err)
}
