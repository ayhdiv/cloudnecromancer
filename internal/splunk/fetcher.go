package splunk

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/schollz/progressbar/v3"

	internalaws "github.com/pfrederiksen/cloudnecromancer/internal/aws"
)

// safeSPLIdentifier validates Splunk index and sourcetype names.
// These must contain only alphanumeric characters, underscores, hyphens,
// colons, and dots to prevent SPL injection.
var safeSPLIdentifier = regexp.MustCompile(`^[a-zA-Z0-9_:.\-]+$`)

// FetchConfig holds configuration for fetching CloudTrail events from Splunk.
type FetchConfig struct {
	// Index is the Splunk index containing CloudTrail events.
	// Default: "aws_cloudtrail".
	Index string

	// Sourcetype filters on the Splunk sourcetype.
	// Default: "aws:cloudtrail".
	Sourcetype string

	// AccountID filters events to a specific AWS account.
	AccountID string

	// Regions filters events to specific AWS regions.
	// If empty, all regions are returned.
	Regions []string

	// SearchOverride replaces the auto-generated SPL query entirely.
	// When set, Index/Sourcetype/AccountID/Regions are ignored.
	// WARNING: This runs arbitrary SPL with the authenticated token's permissions.
	SearchOverride string

	// ShowProgress enables a progress spinner.
	ShowProgress bool
}

// Fetcher retrieves CloudTrail events from Splunk.
type Fetcher struct {
	client *Client
	config FetchConfig
}

// NewFetcher creates a Splunk-backed event fetcher.
// Returns an error if Index or Sourcetype contain unsafe characters.
func NewFetcher(client *Client, config FetchConfig) (*Fetcher, error) {
	if config.Index == "" {
		config.Index = "aws_cloudtrail"
	}
	if config.Sourcetype == "" {
		config.Sourcetype = "aws:cloudtrail"
	}

	// Validate index and sourcetype to prevent SPL injection
	if !safeSPLIdentifier.MatchString(config.Index) {
		return nil, fmt.Errorf("invalid splunk index name %q: must match %s", config.Index, safeSPLIdentifier.String())
	}
	if !safeSPLIdentifier.MatchString(config.Sourcetype) {
		return nil, fmt.Errorf("invalid splunk sourcetype %q: must match %s", config.Sourcetype, safeSPLIdentifier.String())
	}

	return &Fetcher{client: client, config: config}, nil
}

// FetchEvents retrieves CloudTrail events from Splunk for the given time range.
// It returns events in the same format as the CloudTrail API fetcher.
func (f *Fetcher) FetchEvents(startTime, endTime time.Time) ([]internalaws.RawEvent, error) {
	search := f.buildSearch()

	var bar *progressbar.ProgressBar
	if f.config.ShowProgress {
		bar = progressbar.NewOptions(-1,
			progressbar.OptionSetDescription("  [splunk]"),
			progressbar.OptionSetWidth(30),
			progressbar.OptionSpinnerType(14),
			progressbar.OptionSetWriter(os.Stderr),
		)
	}

	var events []internalaws.RawEvent
	seen := make(map[string]struct{})

	err := f.client.ExportSearch(search, startTime, endTime, func(row SearchResult) error {
		ev, err := parseSearchResult(row)
		if err != nil {
			// Skip unparseable results rather than failing the whole fetch
			return nil
		}

		if _, dup := seen[ev.EventID]; dup {
			return nil
		}
		seen[ev.EventID] = struct{}{}
		events = append(events, ev)

		if bar != nil {
			_ = bar.Add(1)
		}
		return nil
	})

	if bar != nil {
		_ = bar.Finish()
	}

	if err != nil {
		return nil, fmt.Errorf("splunk search: %w", err)
	}

	return events, nil
}

// splunkQuote wraps a value in double quotes with Splunk-compatible escaping.
// Splunk treats double-quoted strings as literals. Backslashes and double
// quotes within the value are escaped.
func splunkQuote(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `"`, `\"`)
	return `"` + s + `"`
}

// buildSearch generates the SPL query for CloudTrail events.
func (f *Fetcher) buildSearch() string {
	if f.config.SearchOverride != "" {
		return f.config.SearchOverride
	}

	// Build SPL search. The "search" prefix is required by the export endpoint.
	// Index and Sourcetype are validated in NewFetcher, so safe to interpolate.
	spl := fmt.Sprintf("search index=%s sourcetype=%s", f.config.Index, f.config.Sourcetype)

	if f.config.AccountID != "" {
		spl += " recipientAccountId=" + splunkQuote(f.config.AccountID)
	}

	if len(f.config.Regions) > 0 {
		if len(f.config.Regions) == 1 {
			spl += " awsRegion=" + splunkQuote(f.config.Regions[0])
		} else {
			parts := make([]string, len(f.config.Regions))
			for i, r := range f.config.Regions {
				parts[i] = "awsRegion=" + splunkQuote(r)
			}
			spl += " (" + strings.Join(parts, " OR ") + ")"
		}
	}

	// Return the full raw JSON event — _raw contains the original CloudTrail JSON
	spl += " | fields _raw, _time, eventID, eventName, eventSource, awsRegion"

	return spl
}

// parseSearchResult converts a Splunk search result into a RawEvent.
func parseSearchResult(row SearchResult) (internalaws.RawEvent, error) {
	r := row.Result

	ev := internalaws.RawEvent{}

	// _raw contains the full CloudTrail JSON event
	rawJSON, _ := r["_raw"].(string)
	if rawJSON == "" {
		return ev, fmt.Errorf("missing _raw field")
	}
	ev.RawJSON = rawJSON

	// Try to extract fields from the result row first (Splunk field extraction),
	// then fall back to parsing _raw JSON.
	if id, ok := r["eventID"].(string); ok && id != "" {
		ev.EventID = id
	}
	if name, ok := r["eventName"].(string); ok && name != "" {
		ev.EventName = name
	}
	if source, ok := r["eventSource"].(string); ok && source != "" {
		ev.EventSource = source
	}
	if region, ok := r["awsRegion"].(string); ok && region != "" {
		ev.Region = region
	}

	// Parse _time for the event timestamp
	if timeStr, ok := r["_time"].(string); ok && timeStr != "" {
		if t, err := time.Parse(time.RFC3339, timeStr); err == nil {
			ev.EventTime = t
		}
	}

	// If key fields are missing from Splunk result row, parse from _raw
	if ev.EventID == "" || ev.EventName == "" || ev.EventSource == "" {
		var parsed map[string]any
		if err := json.Unmarshal([]byte(rawJSON), &parsed); err != nil {
			return ev, fmt.Errorf("parse _raw: %w", err)
		}
		if ev.EventID == "" {
			ev.EventID, _ = parsed["eventID"].(string)
		}
		if ev.EventName == "" {
			ev.EventName, _ = parsed["eventName"].(string)
		}
		if ev.EventSource == "" {
			ev.EventSource, _ = parsed["eventSource"].(string)
		}
		if ev.Region == "" {
			ev.Region, _ = parsed["awsRegion"].(string)
		}
		if ev.EventTime.IsZero() {
			if timeStr, ok := parsed["eventTime"].(string); ok {
				ev.EventTime, _ = time.Parse(time.RFC3339, timeStr)
			}
		}
	}

	if ev.EventID == "" {
		return ev, fmt.Errorf("could not extract eventID")
	}

	return ev, nil
}
