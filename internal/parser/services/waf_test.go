package services

import (
	"testing"

	"github.com/pfrederiksen/cloudnecromancer/internal/parser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWAFParser(t *testing.T) {
	p := &wafParser{}

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
			name:         "CreateWebACL creates web_acl",
			fixture:      "CreateWebACL",
			wantAction:   parser.ActionCreate,
			wantResType:  "web_acl",
			wantResID:    "acl-1234567890abcdef0",
			wantService:  "waf",
			wantAttrKeys: []string{"name", "aRN", "scope"},
		},
		{
			name:         "DeleteWebACL deletes web_acl",
			fixture:      "DeleteWebACL",
			wantAction:   parser.ActionDelete,
			wantResType:  "web_acl",
			wantResID:    "acl-1234567890abcdef0",
			wantService:  "waf",
			wantAttrKeys: []string{"name"},
		},
		{
			name:         "CreateRuleGroup creates rule_group",
			fixture:      "CreateRuleGroup",
			wantAction:   parser.ActionCreate,
			wantResType:  "rule_group",
			wantResID:    "rg-abcdef1234567890",
			wantService:  "waf",
			wantAttrKeys: []string{"name"},
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

func TestWAFParserCreateWebACLAttributes(t *testing.T) {
	p := &wafParser{}
	event := loadFixture(t, "CreateWebACL")
	delta, err := p.Parse(event)
	require.NoError(t, err)

	assert.Equal(t, "prod-web-acl", delta.Attributes["name"])
	assert.Equal(t, "arn:aws:wafv2:us-east-1:123456789012:regional/webacl/prod-web-acl/acl-1234567890abcdef0", delta.Attributes["aRN"])
	assert.Equal(t, "REGIONAL", delta.Attributes["scope"])
}

func TestWAFParserSupportedEvents(t *testing.T) {
	p := &wafParser{}
	assert.Equal(t, "waf", p.Service())
	events := p.SupportedEvents()
	assert.Contains(t, events, "CreateWebACL")
	assert.Contains(t, events, "DeleteWebACL")
	assert.Contains(t, events, "UpdateWebACL")
	assert.Contains(t, events, "CreateRuleGroup")
	assert.Contains(t, events, "DeleteRuleGroup")
	assert.Len(t, events, 5)
}

func TestWAFParserInvalidEvent(t *testing.T) {
	p := &wafParser{}
	_, err := p.Parse(map[string]any{})
	assert.Error(t, err)
}
