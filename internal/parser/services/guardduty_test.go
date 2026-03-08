package services

import (
	"testing"

	"github.com/pfrederiksen/cloudnecromancer/internal/parser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGuardDutyParser(t *testing.T) {
	p := &guarddutyParser{}

	tests := []struct {
		name        string
		fixture     string
		wantAction  parser.Action
		wantResType string
		wantResID   string
		wantService string
	}{
		{
			name:        "CreateDetector creates detector",
			fixture:     "CreateDetector",
			wantAction:  parser.ActionCreate,
			wantResType: "detector",
			wantResID:   "d4b040365221be2b54a6264dc1b6d587",
			wantService: "guardduty",
		},
		{
			name:        "DeleteDetector deletes detector",
			fixture:     "DeleteDetector",
			wantAction:  parser.ActionDelete,
			wantResType: "detector",
			wantResID:   "d4b040365221be2b54a6264dc1b6d587",
			wantService: "guardduty",
		},
		{
			name:        "CreateFilter creates filter",
			fixture:     "CreateFilter",
			wantAction:  parser.ActionCreate,
			wantResType: "filter",
			wantResID:   "high-severity-findings",
			wantService: "guardduty",
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
		})
	}
}

func TestGuardDutyParserSupportedEvents(t *testing.T) {
	p := &guarddutyParser{}
	assert.Equal(t, "guardduty", p.Service())
	events := p.SupportedEvents()
	assert.Contains(t, events, "CreateDetector")
	assert.Contains(t, events, "DeleteDetector")
	assert.Contains(t, events, "UpdateDetector")
	assert.Contains(t, events, "CreateIPSet")
	assert.Contains(t, events, "DeleteIPSet")
	assert.Contains(t, events, "CreateFilter")
	assert.Contains(t, events, "DeleteFilter")
	assert.Len(t, events, 7)
}

func TestGuardDutyParserInvalidEvent(t *testing.T) {
	p := &guarddutyParser{}
	_, err := p.Parse(map[string]any{})
	assert.Error(t, err)
}
