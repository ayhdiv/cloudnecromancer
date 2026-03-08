package services

import (
	"testing"

	"github.com/pfrederiksen/cloudnecromancer/internal/parser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestKMSParser(t *testing.T) {
	p := &kmsParser{}

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
			name:         "CreateKey creates key",
			fixture:      "CreateKey",
			wantAction:   parser.ActionCreate,
			wantResType:  "key",
			wantResID:    "a1b2c3d4-5678-90ab-cdef-EXAMPLE11111",
			wantService:  "kms",
			wantAttrKeys: []string{"keyArn", "keyUsage", "keySpec"},
		},
		{
			name:        "ScheduleKeyDeletion deletes key",
			fixture:     "ScheduleKeyDeletion",
			wantAction:  parser.ActionDelete,
			wantResType: "key",
			wantResID:   "a1b2c3d4-5678-90ab-cdef-EXAMPLE11111",
			wantService: "kms",
		},
		{
			name:         "CreateAlias creates alias",
			fixture:      "CreateAlias",
			wantAction:   parser.ActionCreate,
			wantResType:  "alias",
			wantResID:    "alias/my-app-key",
			wantService:  "kms",
			wantAttrKeys: []string{"targetKeyId"},
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

func TestKMSParserCreateKeyAttributes(t *testing.T) {
	p := &kmsParser{}
	event := loadFixture(t, "CreateKey")
	delta, err := p.Parse(event)
	require.NoError(t, err)

	assert.Equal(t, "arn:aws:kms:us-east-1:123456789012:key/a1b2c3d4-5678-90ab-cdef-EXAMPLE11111", delta.Attributes["keyArn"])
	assert.Equal(t, "ENCRYPT_DECRYPT", delta.Attributes["keyUsage"])
	assert.Equal(t, "SYMMETRIC_DEFAULT", delta.Attributes["keySpec"])
}

func TestKMSParserSupportedEvents(t *testing.T) {
	p := &kmsParser{}
	assert.Equal(t, "kms", p.Service())
	events := p.SupportedEvents()
	assert.Contains(t, events, "CreateKey")
	assert.Contains(t, events, "ScheduleKeyDeletion")
	assert.Contains(t, events, "CreateAlias")
	assert.Contains(t, events, "DeleteAlias")
	assert.Len(t, events, 8)
}

func TestKMSParserInvalidEvent(t *testing.T) {
	p := &kmsParser{}
	_, err := p.Parse(map[string]any{})
	assert.Error(t, err)
}
