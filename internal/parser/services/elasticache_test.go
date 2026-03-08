package services

import (
	"testing"

	"github.com/pfrederiksen/cloudnecromancer/internal/parser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestElastiCacheParser(t *testing.T) {
	p := &elasticacheParser{}

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
			name:         "CreateCacheCluster creates cache_cluster",
			fixture:      "CreateCacheCluster",
			wantAction:   parser.ActionCreate,
			wantResType:  "cache_cluster",
			wantResID:    "my-redis-cluster",
			wantService:  "elasticache",
			wantAttrKeys: []string{"cacheClusterId", "engine", "cacheNodeType", "numCacheNodes"},
		},
		{
			name:        "DeleteCacheCluster deletes cache_cluster",
			fixture:     "DeleteCacheCluster",
			wantAction:  parser.ActionDelete,
			wantResType: "cache_cluster",
			wantResID:   "my-redis-cluster",
			wantService: "elasticache",
		},
		{
			name:         "CreateReplicationGroup creates replication_group",
			fixture:      "CreateReplicationGroup",
			wantAction:   parser.ActionCreate,
			wantResType:  "replication_group",
			wantResID:    "my-redis-repl",
			wantService:  "elasticache",
			wantAttrKeys: []string{"replicationGroupId", "description", "engine"},
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

func TestElastiCacheParserCreateCacheClusterAttributes(t *testing.T) {
	p := &elasticacheParser{}
	event := loadFixture(t, "CreateCacheCluster")
	delta, err := p.Parse(event)
	require.NoError(t, err)

	assert.Equal(t, "my-redis-cluster", delta.Attributes["cacheClusterId"])
	assert.Equal(t, "redis", delta.Attributes["engine"])
	assert.Equal(t, "cache.t3.micro", delta.Attributes["cacheNodeType"])
	assert.Equal(t, "1", delta.Attributes["numCacheNodes"])
}

func TestElastiCacheParserCreateReplicationGroupAttributes(t *testing.T) {
	p := &elasticacheParser{}
	event := loadFixture(t, "CreateReplicationGroup")
	delta, err := p.Parse(event)
	require.NoError(t, err)

	assert.Equal(t, "my-redis-repl", delta.Attributes["replicationGroupId"])
	assert.Equal(t, "Primary Redis replication group", delta.Attributes["description"])
	assert.Equal(t, "redis", delta.Attributes["engine"])
}

func TestElastiCacheParserSupportedEvents(t *testing.T) {
	p := &elasticacheParser{}
	assert.Equal(t, "elasticache", p.Service())
	events := p.SupportedEvents()
	assert.Contains(t, events, "CreateCacheCluster")
	assert.Contains(t, events, "DeleteCacheCluster")
	assert.Contains(t, events, "ModifyCacheCluster")
	assert.Contains(t, events, "CreateReplicationGroup")
	assert.Contains(t, events, "DeleteReplicationGroup")
	assert.Len(t, events, 5)
}

func TestElastiCacheParserInvalidEvent(t *testing.T) {
	p := &elasticacheParser{}
	_, err := p.Parse(map[string]any{})
	assert.Error(t, err)
}
