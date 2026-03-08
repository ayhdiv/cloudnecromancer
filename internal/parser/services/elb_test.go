package services

import (
	"testing"

	"github.com/pfrederiksen/cloudnecromancer/internal/parser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestELBParser(t *testing.T) {
	p := &elbParser{}

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
			name:         "CreateLoadBalancer creates load_balancer",
			fixture:      "CreateLoadBalancer",
			wantAction:   parser.ActionCreate,
			wantResType:  "load_balancer",
			wantResID:    "arn:aws:elasticloadbalancing:us-east-1:123456789012:loadbalancer/app/my-app-alb/50dc6c495c0c9188",
			wantService:  "elasticloadbalancing",
			wantAttrKeys: []string{"loadBalancerName", "dNSName", "scheme", "vpcId", "type"},
		},
		{
			name:        "DeleteLoadBalancer deletes load_balancer",
			fixture:     "DeleteLoadBalancer",
			wantAction:  parser.ActionDelete,
			wantResType: "load_balancer",
			wantResID:   "arn:aws:elasticloadbalancing:us-east-1:123456789012:loadbalancer/app/my-app-alb/50dc6c495c0c9188",
			wantService: "elasticloadbalancing",
		},
		{
			name:         "CreateTargetGroup creates target_group",
			fixture:      "CreateTargetGroup",
			wantAction:   parser.ActionCreate,
			wantResType:  "target_group",
			wantResID:    "arn:aws:elasticloadbalancing:us-east-1:123456789012:targetgroup/my-app-targets/73e2d6bc24d8a067",
			wantService:  "elasticloadbalancing",
			wantAttrKeys: []string{"targetGroupName", "protocol", "port", "vpcId"},
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

func TestELBParserCreateLoadBalancerAttributes(t *testing.T) {
	p := &elbParser{}
	event := loadFixture(t, "CreateLoadBalancer")
	delta, err := p.Parse(event)
	require.NoError(t, err)

	assert.Equal(t, "my-app-alb", delta.Attributes["loadBalancerName"])
	assert.Equal(t, "my-app-alb-123456789.us-east-1.elb.amazonaws.com", delta.Attributes["dNSName"])
	assert.Equal(t, "internet-facing", delta.Attributes["scheme"])
	assert.Equal(t, "vpc-0123456789abcdef0", delta.Attributes["vpcId"])
	assert.Equal(t, "application", delta.Attributes["type"])
}

func TestELBParserSupportedEvents(t *testing.T) {
	p := &elbParser{}
	assert.Equal(t, "elasticloadbalancing", p.Service())
	events := p.SupportedEvents()
	assert.Contains(t, events, "CreateLoadBalancer")
	assert.Contains(t, events, "DeleteLoadBalancer")
	assert.Contains(t, events, "CreateTargetGroup")
	assert.Contains(t, events, "RegisterTargets")
	assert.Len(t, events, 8)
}

func TestELBParserInvalidEvent(t *testing.T) {
	p := &elbParser{}
	_, err := p.Parse(map[string]any{})
	assert.Error(t, err)
}
