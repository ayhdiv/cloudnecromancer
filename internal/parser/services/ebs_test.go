package services

import (
	"testing"

	"github.com/pfrederiksen/cloudnecromancer/internal/parser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEBSParser(t *testing.T) {
	p := &ebsParser{}

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
			name:         "CreateVolume creates volume",
			fixture:      "CreateVolume",
			wantAction:   parser.ActionCreate,
			wantResType:  "volume",
			wantResID:    "vol-0a1b2c3d4e5f6g7h8",
			wantService:  "ebs",
			wantAttrKeys: []string{"size", "volumeType", "availabilityZone", "encrypted"},
		},
		{
			name:        "DeleteVolume deletes volume",
			fixture:     "DeleteVolume",
			wantAction:  parser.ActionDelete,
			wantResType: "volume",
			wantResID:   "vol-0a1b2c3d4e5f6g7h8",
			wantService: "ebs",
		},
		{
			name:         "CreateSnapshot creates snapshot",
			fixture:      "CreateSnapshot",
			wantAction:   parser.ActionCreate,
			wantResType:  "snapshot",
			wantResID:    "snap-0abc123def456789a",
			wantService:  "ebs",
			wantAttrKeys: []string{"volumeId"},
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

func TestEBSParserCreateVolumeAttributes(t *testing.T) {
	p := &ebsParser{}
	event := loadFixture(t, "CreateVolume")
	delta, err := p.Parse(event)
	require.NoError(t, err)

	assert.Equal(t, "100", delta.Attributes["size"])
	assert.Equal(t, "gp3", delta.Attributes["volumeType"])
	assert.Equal(t, "us-east-1a", delta.Attributes["availabilityZone"])
	assert.Equal(t, "true", delta.Attributes["encrypted"])
}

func TestEBSParserSupportedEvents(t *testing.T) {
	p := &ebsParser{}
	assert.Equal(t, "ebs", p.Service())
	events := p.SupportedEvents()
	assert.Contains(t, events, "CreateVolume")
	assert.Contains(t, events, "DeleteVolume")
	assert.Contains(t, events, "CreateSnapshot")
	assert.Contains(t, events, "DeleteSnapshot")
	assert.Contains(t, events, "ModifyVolume")
	assert.Len(t, events, 5)
}

func TestEBSParserInvalidEvent(t *testing.T) {
	p := &ebsParser{}
	_, err := p.Parse(map[string]any{})
	assert.Error(t, err)
}
