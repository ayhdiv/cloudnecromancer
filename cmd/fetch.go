package cmd

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudtrail"
	"github.com/spf13/cobra"

	internalaws "github.com/pfrederiksen/cloudnecromancer/internal/aws"
	"github.com/pfrederiksen/cloudnecromancer/internal/splunk"
	"github.com/pfrederiksen/cloudnecromancer/internal/store"
)

var accountIDPattern = regexp.MustCompile(`^\d{12}$`)

var (
	fetchAccountID string
	fetchRegion    string
	fetchRegions   string
	fetchStart     string
	fetchEnd       string

	// Source selection
	fetchSource string

	// Splunk options
	splunkURL        string
	splunkToken      string
	splunkIndex      string
	splunkSourcetype string
	splunkQuery      string
	splunkInsecure   bool
)

var fetchCmd = &cobra.Command{
	Use:   "fetch",
	Short: "Fetch CloudTrail events and store them in DuckDB",
	Long: `Fetch CloudTrail events from AWS CloudTrail API or Splunk.

Sources:
  cloudtrail  Query the CloudTrail LookupEvents API directly (default, 90-day limit)
  splunk      Query a Splunk instance containing indexed CloudTrail events (no time limit)

Splunk authentication uses a bearer token, which can be passed via --splunk-token
or the SPLUNK_TOKEN environment variable.`,
	RunE: runFetch,
}

func init() {
	fetchCmd.Flags().StringVar(&fetchAccountID, "account-id", "", "AWS account ID (required)")
	fetchCmd.Flags().StringVar(&fetchRegion, "region", "us-east-1", "Single region to fetch from")
	fetchCmd.Flags().StringVar(&fetchRegions, "regions", "", "Comma-separated list of regions (overrides --region)")
	fetchCmd.Flags().StringVar(&fetchStart, "start", "", "Start time in RFC3339 format (required)")
	fetchCmd.Flags().StringVar(&fetchEnd, "end", "", "End time in RFC3339 format (required)")

	// Source selection
	fetchCmd.Flags().StringVar(&fetchSource, "source", "cloudtrail", "Event source: cloudtrail, splunk")

	// Splunk options
	fetchCmd.Flags().StringVar(&splunkURL, "splunk-url", "", "Splunk REST API base URL (e.g. https://splunk.corp:8089)")
	fetchCmd.Flags().StringVar(&splunkToken, "splunk-token", "", "Splunk bearer token (or set SPLUNK_TOKEN env var)")
	fetchCmd.Flags().StringVar(&splunkIndex, "splunk-index", "aws_cloudtrail", "Splunk index containing CloudTrail events")
	fetchCmd.Flags().StringVar(&splunkSourcetype, "splunk-sourcetype", "aws:cloudtrail", "Splunk sourcetype for CloudTrail events")
	fetchCmd.Flags().StringVar(&splunkQuery, "splunk-query", "", "Override the generated SPL query entirely")
	fetchCmd.Flags().BoolVar(&splunkInsecure, "splunk-insecure", false, "Skip TLS certificate verification for Splunk")

	_ = fetchCmd.MarkFlagRequired("start")
	_ = fetchCmd.MarkFlagRequired("end")

	rootCmd.AddCommand(fetchCmd)
}

func runFetch(cmd *cobra.Command, args []string) error {
	startTime, err := time.Parse(time.RFC3339, fetchStart)
	if err != nil {
		return fmt.Errorf("invalid --start: %w", err)
	}
	endTime, err := time.Parse(time.RFC3339, fetchEnd)
	if err != nil {
		return fmt.Errorf("invalid --end: %w", err)
	}
	if !startTime.Before(endTime) {
		return fmt.Errorf("invalid time range: --start (%s) must be before --end (%s)", fetchStart, fetchEnd)
	}

	regions := []string{fetchRegion}
	if fetchRegions != "" {
		regions = strings.Split(fetchRegions, ",")
		for i := range regions {
			regions[i] = strings.TrimSpace(regions[i])
		}
	}

	var events []internalaws.RawEvent

	switch strings.ToLower(fetchSource) {
	case "cloudtrail":
		if fetchAccountID == "" {
			return fmt.Errorf("--account-id is required when using --source cloudtrail")
		}
		if !accountIDPattern.MatchString(fetchAccountID) {
			return fmt.Errorf("invalid --account-id: must be a 12-digit number, got %q", fetchAccountID)
		}
		var fetchErr error
		events, fetchErr = fetchFromCloudTrail(startTime, endTime, regions)
		if fetchErr != nil {
			return fetchErr
		}

	case "splunk":
		var fetchErr error
		events, fetchErr = fetchFromSplunk(startTime, endTime, regions)
		if fetchErr != nil {
			return fetchErr
		}

	default:
		return fmt.Errorf("unknown --source %q: must be 'cloudtrail' or 'splunk'", fetchSource)
	}

	// Store events
	db, err := store.NewStore(dbPath)
	if err != nil {
		return fmt.Errorf("open store: %w", err)
	}
	defer db.Close()

	accountID := fetchAccountID
	if accountID == "" {
		accountID = "unknown" // Splunk mode may not require explicit account ID
	}

	inserted, err := db.InsertEvents(events, accountID)
	if err != nil {
		return fmt.Errorf("insert events: %w", err)
	}

	// Summary
	serviceSet := make(map[string]struct{})
	for _, ev := range events {
		serviceSet[ev.EventSource] = struct{}{}
	}

	fmt.Fprintf(os.Stderr, "\nDone! %d events fetched, %d new events stored, %d services covered\n",
		len(events), inserted, len(serviceSet))
	if len(events) > 0 {
		minT, maxT := events[0].EventTime, events[0].EventTime
		for _, ev := range events[1:] {
			if ev.EventTime.Before(minT) {
				minT = ev.EventTime
			}
			if ev.EventTime.After(maxT) {
				maxT = ev.EventTime
			}
		}
		fmt.Fprintf(os.Stderr, "Date range: %s to %s\n", minT.Format(time.RFC3339), maxT.Format(time.RFC3339))
	}

	return nil
}

func fetchFromCloudTrail(startTime, endTime time.Time, regions []string) ([]internalaws.RawEvent, error) {
	ctx := context.Background()

	cfgOpts := []func(*config.LoadOptions) error{}
	if profile != "" {
		cfgOpts = append(cfgOpts, config.WithSharedConfigProfile(profile))
	}

	cfg, err := config.LoadDefaultConfig(ctx, cfgOpts...)
	if err != nil {
		return nil, fmt.Errorf("load AWS config: %w", err)
	}

	client := cloudtrail.NewFromConfig(cfg)
	fetcher := internalaws.NewFetcher(client, !quiet)

	fmt.Fprintf(os.Stderr, "Source: CloudTrail API\n")
	fmt.Fprintf(os.Stderr, "Fetching events from %d region(s): %s\n", len(regions), strings.Join(regions, ", "))
	fmt.Fprintf(os.Stderr, "Time range: %s to %s\n", startTime.Format(time.RFC3339), endTime.Format(time.RFC3339))

	events, err := fetcher.FetchEvents(ctx, startTime, endTime, regions)
	if err != nil {
		return nil, fmt.Errorf("fetch events: %w", err)
	}
	return events, nil
}

func fetchFromSplunk(startTime, endTime time.Time, regions []string) ([]internalaws.RawEvent, error) {
	// Resolve Splunk token: flag > env var
	// Prefer SPLUNK_TOKEN env var over --splunk-token flag to avoid
	// exposing the token in process listings (ps aux).
	token := splunkToken
	if token == "" {
		token = os.Getenv("SPLUNK_TOKEN")
	}
	if token == "" {
		return nil, fmt.Errorf("Splunk token required: set SPLUNK_TOKEN environment variable (preferred) or --splunk-token flag")
	}

	// Resolve Splunk URL: flag > env var
	splunkEndpoint := splunkURL
	if splunkEndpoint == "" {
		splunkEndpoint = os.Getenv("SPLUNK_URL")
	}
	if splunkEndpoint == "" {
		return nil, fmt.Errorf("Splunk URL required: set --splunk-url or SPLUNK_URL environment variable")
	}

	// Validate URL scheme — require HTTPS unless --splunk-insecure is set
	parsed, err := url.Parse(splunkEndpoint)
	if err != nil {
		return nil, fmt.Errorf("invalid --splunk-url: %w", err)
	}
	if parsed.Scheme == "http" && !splunkInsecure {
		return nil, fmt.Errorf("refusing to send credentials over HTTP; use https:// or add --splunk-insecure to override")
	}

	// Validate account ID if provided in Splunk mode
	if fetchAccountID != "" && !accountIDPattern.MatchString(fetchAccountID) {
		return nil, fmt.Errorf("invalid --account-id: must be a 12-digit number, got %q", fetchAccountID)
	}

	var clientOpts []splunk.ClientOption
	if splunkInsecure {
		clientOpts = append(clientOpts, splunk.WithInsecureSkipVerify())
	}

	client := splunk.NewClient(splunkEndpoint, token, clientOpts...)
	fetcher, err := splunk.NewFetcher(client, splunk.FetchConfig{
		Index:          splunkIndex,
		Sourcetype:     splunkSourcetype,
		AccountID:      fetchAccountID,
		Regions:        regions,
		SearchOverride: splunkQuery,
		ShowProgress:   !quiet,
	})
	if err != nil {
		return nil, err
	}

	// Redact userinfo from URL before logging
	safeURL := *parsed
	safeURL.User = nil
	fmt.Fprintf(os.Stderr, "Source: Splunk (%s)\n", safeURL.String())
	fmt.Fprintf(os.Stderr, "Index: %s, Sourcetype: %s\n", splunkIndex, splunkSourcetype)
	fmt.Fprintf(os.Stderr, "Time range: %s to %s\n", startTime.Format(time.RFC3339), endTime.Format(time.RFC3339))

	events, err := fetcher.FetchEvents(startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("fetch events: %w", err)
	}
	return events, nil
}
