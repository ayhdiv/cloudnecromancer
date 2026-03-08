package export

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/pfrederiksen/cloudnecromancer/internal/engine"
)

// sanitizeCSVCell prevents CSV formula injection by prefixing dangerous
// values with a tab character. Spreadsheet applications interpret cells
// starting with =, +, -, @, \t, or \r as formulas.
func sanitizeCSVCell(s string) string {
	if len(s) == 0 {
		return s
	}
	switch s[0] {
	case '=', '+', '-', '@', '\t', '\r':
		return "\t" + s
	}
	return s
}

// CSVExporter writes a Snapshot as a Splunk-compatible lookup CSV.
type CSVExporter struct{}

// Export writes the snapshot as CSV with headers.
func (e *CSVExporter) Export(snapshot *engine.Snapshot, w io.Writer) error {
	cw := csv.NewWriter(w)
	defer cw.Flush()

	// Header
	header := []string{
		"resource_id", "resource_type", "service", "state", "region",
		"account_id", "created_at", "last_modified", "snapshot_timestamp", "attributes_json",
	}
	if err := cw.Write(header); err != nil {
		return fmt.Errorf("csv header: %w", err)
	}

	region := ""
	if len(snapshot.Regions) > 0 {
		region = snapshot.Regions[0]
	}

	for typeKey, resources := range snapshot.Resources {
		service, resourceType := splitTypeKey(typeKey)
		for _, res := range resources {
			attrsJSON, _ := json.Marshal(res.Attributes)
			row := []string{
				sanitizeCSVCell(res.ResourceID),
				sanitizeCSVCell(resourceType),
				sanitizeCSVCell(service),
				sanitizeCSVCell(res.State),
				sanitizeCSVCell(region),
				sanitizeCSVCell(snapshot.AccountID),
				res.CreatedAt.Format(time.RFC3339),
				res.LastModified.Format(time.RFC3339),
				snapshot.Timestamp.Format(time.RFC3339),
				sanitizeCSVCell(string(attrsJSON)),
			}
			if err := cw.Write(row); err != nil {
				return fmt.Errorf("csv row: %w", err)
			}
		}
	}

	return nil
}

func splitTypeKey(typeKey string) (service, resourceType string) {
	parts := strings.SplitN(typeKey, ":", 2)
	if len(parts) == 2 {
		return parts[0], parts[1]
	}
	return typeKey, ""
}
