package splunk

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildSearch_Default(t *testing.T) {
	f, err := NewFetcher(nil, FetchConfig{
		AccountID: "123456789012",
	})
	require.NoError(t, err)
	spl := f.buildSearch()
	assert.Contains(t, spl, "index=aws_cloudtrail")
	assert.Contains(t, spl, "sourcetype=aws:cloudtrail")
	assert.Contains(t, spl, "recipientAccountId=")
	assert.Contains(t, spl, "123456789012")
}

func TestBuildSearch_WithRegions(t *testing.T) {
	f, err := NewFetcher(nil, FetchConfig{
		Regions: []string{"us-east-1", "eu-west-1"},
	})
	require.NoError(t, err)
	spl := f.buildSearch()
	assert.Contains(t, spl, "awsRegion=")
	assert.Contains(t, spl, "OR")
}

func TestBuildSearch_SingleRegion(t *testing.T) {
	f, err := NewFetcher(nil, FetchConfig{
		Regions: []string{"us-east-1"},
	})
	require.NoError(t, err)
	spl := f.buildSearch()
	assert.Contains(t, spl, `awsRegion="us-east-1"`)
	assert.NotContains(t, spl, "OR")
}

func TestBuildSearch_Override(t *testing.T) {
	f, err := NewFetcher(nil, FetchConfig{
		SearchOverride: "search index=custom | head 10",
	})
	require.NoError(t, err)
	assert.Equal(t, "search index=custom | head 10", f.buildSearch())
}

func TestBuildSearch_CustomIndex(t *testing.T) {
	f, err := NewFetcher(nil, FetchConfig{
		Index:      "my_ct_index",
		Sourcetype: "aws:cloudtrail:v2",
	})
	require.NoError(t, err)
	spl := f.buildSearch()
	assert.Contains(t, spl, "index=my_ct_index")
	assert.Contains(t, spl, "sourcetype=aws:cloudtrail:v2")
}

func TestNewFetcher_RejectsUnsafeIndex(t *testing.T) {
	tests := []struct {
		name  string
		index string
	}{
		{"pipe injection", "aws_cloudtrail | delete"},
		{"semicolon", "aws_cloudtrail; delete"},
		{"backtick", "aws`command`"},
		{"space", "aws cloudtrail"},
		{"bracket", "aws[0]"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := NewFetcher(nil, FetchConfig{Index: tc.index})
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "invalid splunk index")
		})
	}
}

func TestNewFetcher_RejectsUnsafeSourcetype(t *testing.T) {
	_, err := NewFetcher(nil, FetchConfig{Sourcetype: "aws:cloudtrail | delete"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid splunk sourcetype")
}

func TestSplunkQuote(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"simple", `"simple"`},
		{`has "quotes"`, `"has \"quotes\""`},
		{`back\slash`, `"back\\slash"`},
		{"pipe|injection", `"pipe|injection"`},
	}
	for _, tc := range tests {
		assert.Equal(t, tc.want, splunkQuote(tc.input))
	}
}

func TestParseSearchResult(t *testing.T) {
	rawEvent := map[string]any{
		"eventID":     "abc-123",
		"eventName":   "RunInstances",
		"eventSource": "ec2.amazonaws.com",
		"awsRegion":   "us-east-1",
		"eventTime":   "2026-02-15T03:00:00Z",
	}
	rawJSON, _ := json.Marshal(rawEvent)

	row := SearchResult{
		Result: map[string]any{
			"_raw":        string(rawJSON),
			"_time":       "2026-02-15T03:00:00Z",
			"eventID":     "abc-123",
			"eventName":   "RunInstances",
			"eventSource": "ec2.amazonaws.com",
			"awsRegion":   "us-east-1",
		},
	}

	ev, err := parseSearchResult(row)
	require.NoError(t, err)
	assert.Equal(t, "abc-123", ev.EventID)
	assert.Equal(t, "RunInstances", ev.EventName)
	assert.Equal(t, "ec2.amazonaws.com", ev.EventSource)
	assert.Equal(t, "us-east-1", ev.Region)
	assert.Equal(t, string(rawJSON), ev.RawJSON)
}

func TestParseSearchResult_FallbackToRaw(t *testing.T) {
	rawEvent := map[string]any{
		"eventID":     "abc-456",
		"eventName":   "CreateRole",
		"eventSource": "iam.amazonaws.com",
		"awsRegion":   "us-east-1",
		"eventTime":   "2026-02-15T04:00:00Z",
	}
	rawJSON, _ := json.Marshal(rawEvent)

	// Only _raw provided, no extracted fields
	row := SearchResult{
		Result: map[string]any{
			"_raw": string(rawJSON),
		},
	}

	ev, err := parseSearchResult(row)
	require.NoError(t, err)
	assert.Equal(t, "abc-456", ev.EventID)
	assert.Equal(t, "CreateRole", ev.EventName)
	assert.Equal(t, "iam.amazonaws.com", ev.EventSource)
}

func TestParseSearchResult_MissingRaw(t *testing.T) {
	row := SearchResult{
		Result: map[string]any{},
	}
	_, err := parseSearchResult(row)
	assert.Error(t, err)
}

func TestFetchEvents_Integration(t *testing.T) {
	events := []map[string]any{
		{
			"eventID":     "evt-001",
			"eventName":   "RunInstances",
			"eventSource": "ec2.amazonaws.com",
			"awsRegion":   "us-east-1",
			"eventTime":   "2026-02-15T03:00:00Z",
		},
		{
			"eventID":     "evt-002",
			"eventName":   "CreateRole",
			"eventSource": "iam.amazonaws.com",
			"awsRegion":   "us-east-1",
			"eventTime":   "2026-02-15T03:30:00Z",
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/services/search/jobs/export", r.URL.Path)
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))

		w.Header().Set("Content-Type", "application/json")
		enc := json.NewEncoder(w)
		for _, ev := range events {
			rawJSON, _ := json.Marshal(ev)
			result := map[string]any{
				"preview": false,
				"result": map[string]any{
					"_raw":        string(rawJSON),
					"_time":       ev["eventTime"],
					"eventID":     ev["eventID"],
					"eventName":   ev["eventName"],
					"eventSource": ev["eventSource"],
					"awsRegion":   ev["awsRegion"],
				},
			}
			_ = enc.Encode(result)
		}
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token")
	fetcher, err := NewFetcher(client, FetchConfig{
		AccountID: "123456789012",
	})
	require.NoError(t, err)

	start := time.Date(2026, 2, 15, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 2, 16, 0, 0, 0, 0, time.UTC)

	result, err := fetcher.FetchEvents(start, end)
	require.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, "evt-001", result[0].EventID)
	assert.Equal(t, "evt-002", result[1].EventID)
}

func TestFetchEvents_Dedup(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		enc := json.NewEncoder(w)
		rawJSON := `{"eventID":"dup-001","eventName":"RunInstances","eventSource":"ec2.amazonaws.com","awsRegion":"us-east-1"}`
		for range 3 {
			result := map[string]any{
				"preview": false,
				"result": map[string]any{
					"_raw":        rawJSON,
					"eventID":     "dup-001",
					"eventName":   "RunInstances",
					"eventSource": "ec2.amazonaws.com",
					"awsRegion":   "us-east-1",
					"_time":       "2026-02-15T03:00:00Z",
				},
			}
			_ = enc.Encode(result)
		}
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token")
	fetcher, err := NewFetcher(client, FetchConfig{})
	require.NoError(t, err)

	start := time.Date(2026, 2, 15, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 2, 16, 0, 0, 0, 0, time.UTC)

	result, err := fetcher.FetchEvents(start, end)
	require.NoError(t, err)
	assert.Len(t, result, 1, "duplicates should be removed")
}

func TestFetchEvents_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprint(w, "authentication failed")
	}))
	defer server.Close()

	client := NewClient(server.URL, "bad-token")
	fetcher, err := NewFetcher(client, FetchConfig{})
	require.NoError(t, err)

	start := time.Date(2026, 2, 15, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 2, 16, 0, 0, 0, 0, time.UTC)

	_, err = fetcher.FetchEvents(start, end)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "401")
	// Error should NOT contain the raw body ("authentication failed")
	assert.NotContains(t, err.Error(), "authentication failed")
}

func TestFetchEvents_SkipsPreviewResults(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		enc := json.NewEncoder(w)

		// Preview result — should be skipped
		_ = enc.Encode(map[string]any{
			"preview": true,
			"result": map[string]any{
				"_raw":    `{"eventID":"preview-001"}`,
				"eventID": "preview-001",
			},
		})

		// Real result
		rawJSON := `{"eventID":"real-001","eventName":"CreateBucket","eventSource":"s3.amazonaws.com","awsRegion":"us-east-1"}`
		_ = enc.Encode(map[string]any{
			"preview": false,
			"result": map[string]any{
				"_raw":        rawJSON,
				"eventID":     "real-001",
				"eventName":   "CreateBucket",
				"eventSource": "s3.amazonaws.com",
				"awsRegion":   "us-east-1",
				"_time":       "2026-02-15T03:00:00Z",
			},
		})
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token")
	fetcher, err := NewFetcher(client, FetchConfig{})
	require.NoError(t, err)

	start := time.Date(2026, 2, 15, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 2, 16, 0, 0, 0, 0, time.UTC)

	result, err := fetcher.FetchEvents(start, end)
	require.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, "real-001", result[0].EventID)
}
