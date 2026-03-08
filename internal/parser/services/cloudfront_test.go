package services

import (
	"testing"

	"github.com/pfrederiksen/cloudnecromancer/internal/parser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCloudFrontParser(t *testing.T) {
	p := &cloudfrontParser{}

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
			name:         "CreateDistribution creates distribution",
			fixture:      "CreateDistribution",
			wantAction:   parser.ActionCreate,
			wantResType:  "distribution",
			wantResID:    "E1A2B3C4D5E6F7",
			wantService:  "cloudfront",
			wantAttrKeys: []string{"domainName", "status"},
		},
		{
			name:        "DeleteDistribution deletes distribution",
			fixture:     "DeleteDistribution",
			wantAction:  parser.ActionDelete,
			wantResType: "distribution",
			wantResID:   "E1A2B3C4D5E6F7",
			wantService: "cloudfront",
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

func TestCloudFrontParserCreateDistributionAttributes(t *testing.T) {
	p := &cloudfrontParser{}
	event := loadFixture(t, "CreateDistribution")
	delta, err := p.Parse(event)
	require.NoError(t, err)

	assert.Equal(t, "d111111abcdef8.cloudfront.net", delta.Attributes["domainName"])
	assert.Equal(t, "InProgress", delta.Attributes["status"])
}

func TestCloudFrontParserSupportedEvents(t *testing.T) {
	p := &cloudfrontParser{}
	assert.Equal(t, "cloudfront", p.Service())
	events := p.SupportedEvents()
	assert.Contains(t, events, "CreateDistribution")
	assert.Contains(t, events, "DeleteDistribution")
	assert.Contains(t, events, "UpdateDistribution")
	assert.Contains(t, events, "CreateOriginAccessControl")
	assert.Contains(t, events, "DeleteOriginAccessControl")
	assert.Len(t, events, 5)
}

func TestCloudFrontParserInvalidEvent(t *testing.T) {
	p := &cloudfrontParser{}
	_, err := p.Parse(map[string]any{})
	assert.Error(t, err)
}
