package services

import (
	"testing"

	"github.com/pfrederiksen/cloudnecromancer/internal/parser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRoute53Parser(t *testing.T) {
	p := &route53Parser{}

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
			name:         "CreateHostedZone creates hosted_zone",
			fixture:      "CreateHostedZone",
			wantAction:   parser.ActionCreate,
			wantResType:  "hosted_zone",
			wantResID:    "/hostedzone/Z1234567890ABC",
			wantService:  "route53",
			wantAttrKeys: []string{"id", "name", "callerReference"},
		},
		{
			name:        "DeleteHostedZone deletes hosted_zone",
			fixture:     "DeleteHostedZone",
			wantAction:  parser.ActionDelete,
			wantResType: "hosted_zone",
			wantResID:   "/hostedzone/Z1234567890ABC",
			wantService: "route53",
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

func TestRoute53ParserCreateHostedZoneAttributes(t *testing.T) {
	p := &route53Parser{}
	event := loadFixture(t, "CreateHostedZone")
	delta, err := p.Parse(event)
	require.NoError(t, err)

	assert.Equal(t, "/hostedzone/Z1234567890ABC", delta.Attributes["id"])
	assert.Equal(t, "example.com.", delta.Attributes["name"])
	assert.Equal(t, "ref-2025-11-15-001", delta.Attributes["callerReference"])
}

func TestRoute53ParserSupportedEvents(t *testing.T) {
	p := &route53Parser{}
	assert.Equal(t, "route53", p.Service())
	events := p.SupportedEvents()
	assert.Contains(t, events, "CreateHostedZone")
	assert.Contains(t, events, "DeleteHostedZone")
	assert.Contains(t, events, "ChangeResourceRecordSets")
	assert.Len(t, events, 3)
}

func TestRoute53ParserInvalidEvent(t *testing.T) {
	p := &route53Parser{}
	_, err := p.Parse(map[string]any{})
	assert.Error(t, err)
}
