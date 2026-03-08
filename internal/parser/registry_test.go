package parser_test

import (
	"testing"

	"github.com/pfrederiksen/cloudnecromancer/internal/parser"
	// Import services to trigger init() registrations.
	_ "github.com/pfrederiksen/cloudnecromancer/internal/parser/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegisteredEventsContainsAllExpected(t *testing.T) {
	registered := parser.RegisteredEvents()

	// Spot-check events from each service category
	spotChecks := []string{
		// Original 5 services
		"RunInstances", "TerminateInstances", "CreateSecurityGroup",
		"CreateRole", "AttachRolePolicy",
		"CreateBucket", "PutBucketPolicy",
		"CreateFunction20150331", "DeleteFunction20150331",
		"CreateDBInstance", "ModifyDBInstance",
		// Tier 1: IR-critical
		"CreateLoadBalancer", "CreateTargetGroup",
		"CreateCluster", "RegisterTaskDefinition",
		"CreateNodegroup", "UpdateClusterVersion",
		"CreateKey", "ScheduleKeyDeletion",
		"CreateSecret", "RotateSecret",
		"CreateLogGroup", "PutRetentionPolicy",
		// Tier 2: Compliance
		"CreateTable", "UpdateTable",
		"CreateTopic", "Subscribe",
		"CreateQueue", "SetQueueAttributes",
		"CreateRestApi", "CreateApi",
		"CreateHostedZone", "ChangeResourceRecordSets",
		"CreateRepository", "PutImageScanningConfiguration",
		"CreateCacheCluster", "CreateReplicationGroup",
		// Tier 3: Security
		"CreateWebACL", "CreateRuleGroup",
		"CreateDetector", "UpdateDetector",
		"CreateDistribution", "CreateOriginAccessControl",
		"CreateVolume", "CreateSnapshot",
		"CreateDocument", "PutParameter",
	}

	for _, event := range spotChecks {
		assert.Contains(t, registered, event, "event %s should be registered", event)
	}

	assert.Len(t, registered, 133, "total registered events should match expected count")
}

func TestLookupReturnsCorrectParser(t *testing.T) {
	tests := []struct {
		eventName   string
		wantService string
	}{
		{"RunInstances", "ec2"},
		{"TerminateInstances", "ec2"},
		{"CreateRole", "iam"},
		{"CreateBucket", "s3"},
		{"CreateFunction20150331", "lambda"},
		{"CreateDBInstance", "rds"},
	}

	for _, tc := range tests {
		t.Run(tc.eventName, func(t *testing.T) {
			p, err := parser.Lookup(tc.eventName)
			require.NoError(t, err)
			assert.Equal(t, tc.wantService, p.Service())
		})
	}
}

func TestLookupUnknownEventReturnsError(t *testing.T) {
	_, err := parser.Lookup("NonExistentEvent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no parser registered")
}
