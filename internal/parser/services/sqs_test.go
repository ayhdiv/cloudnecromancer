package services

import (
	"testing"

	"github.com/pfrederiksen/cloudnecromancer/internal/parser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSQSParser(t *testing.T) {
	p := &sqsParser{}

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
			name:         "CreateQueue creates queue",
			fixture:      "CreateQueue",
			wantAction:   parser.ActionCreate,
			wantResType:  "queue",
			wantResID:    "order-processing-queue",
			wantService:  "sqs",
			wantAttrKeys: []string{"queueName", "queueUrl"},
		},
		{
			name:        "DeleteQueue deletes queue",
			fixture:     "DeleteQueue",
			wantAction:  parser.ActionDelete,
			wantResType: "queue",
			wantResID:   "https://sqs.us-east-1.amazonaws.com/123456789012/order-processing-queue",
			wantService: "sqs",
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

func TestSQSParserCreateQueueAttributes(t *testing.T) {
	p := &sqsParser{}
	event := loadFixture(t, "CreateQueue")
	delta, err := p.Parse(event)
	require.NoError(t, err)

	assert.Equal(t, "order-processing-queue", delta.Attributes["queueName"])
	assert.Equal(t, "https://sqs.us-east-1.amazonaws.com/123456789012/order-processing-queue", delta.Attributes["queueUrl"])
}

func TestSQSParserSupportedEvents(t *testing.T) {
	p := &sqsParser{}
	assert.Equal(t, "sqs", p.Service())
	events := p.SupportedEvents()
	assert.Contains(t, events, "CreateQueue")
	assert.Contains(t, events, "DeleteQueue")
	assert.Contains(t, events, "SetQueueAttributes")
	assert.Contains(t, events, "PurgeQueue")
	assert.Len(t, events, 4)
}

func TestSQSParserInvalidEvent(t *testing.T) {
	p := &sqsParser{}
	_, err := p.Parse(map[string]any{})
	assert.Error(t, err)
}
