package services

import (
	"testing"

	"github.com/pfrederiksen/cloudnecromancer/internal/parser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSNSParser(t *testing.T) {
	p := &snsParser{}

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
			name:         "CreateTopic creates topic",
			fixture:      "CreateTopic",
			wantAction:   parser.ActionCreate,
			wantResType:  "topic",
			wantResID:    "arn:aws:sns:us-east-1:123456789012:order-notifications",
			wantService:  "sns",
			wantAttrKeys: []string{"topicArn", "name"},
		},
		{
			name:        "DeleteTopic deletes topic",
			fixture:     "DeleteTopic",
			wantAction:  parser.ActionDelete,
			wantResType: "topic",
			wantResID:   "arn:aws:sns:us-east-1:123456789012:order-notifications",
			wantService: "sns",
		},
		{
			name:         "Subscribe creates subscription",
			fixture:      "Subscribe",
			wantAction:   parser.ActionCreate,
			wantResType:  "subscription",
			wantResID:    "arn:aws:sns:us-east-1:123456789012:order-notifications:a1b2c3d4-5678-90ab-cdef-EXAMPLE22222",
			wantService:  "sns",
			wantAttrKeys: []string{"subscriptionArn", "topicArn", "protocol", "endpoint"},
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

func TestSNSParserCreateTopicAttributes(t *testing.T) {
	p := &snsParser{}
	event := loadFixture(t, "CreateTopic")
	delta, err := p.Parse(event)
	require.NoError(t, err)

	assert.Equal(t, "arn:aws:sns:us-east-1:123456789012:order-notifications", delta.Attributes["topicArn"])
	assert.Equal(t, "order-notifications", delta.Attributes["name"])
}

func TestSNSParserSupportedEvents(t *testing.T) {
	p := &snsParser{}
	assert.Equal(t, "sns", p.Service())
	events := p.SupportedEvents()
	assert.Contains(t, events, "CreateTopic")
	assert.Contains(t, events, "DeleteTopic")
	assert.Contains(t, events, "Subscribe")
	assert.Contains(t, events, "Unsubscribe")
	assert.Contains(t, events, "SetTopicAttributes")
	assert.Len(t, events, 5)
}

func TestSNSParserInvalidEvent(t *testing.T) {
	p := &snsParser{}
	_, err := p.Parse(map[string]any{})
	assert.Error(t, err)
}
