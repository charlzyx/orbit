package workbench

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"path/filepath"
	"testing"
	"time"
)

func TestUsageProviderScopedModelUnionFromRealStoreAndExports(t *testing.T) {
	root := t.TempDir()
	for n, ref := range []UsageModelRef{{"a", "shared"}, {"b", "shared"}, {"a", "other"}, {"b", "other"}} {
		put(t, filepath.Join(root, "group", fmt.Sprint(n)+".jsonl"), fmt.Sprintf(`{"type":"session","id":"s%d","cwd":"/w","timestamp":"2026-10-24T00:00:00Z"}
{"type":"message","timestamp":"2026-10-24T00:00:01Z","message":{"role":"user","content":"hello"}}
{"type":"message","timestamp":"2026-10-24T00:00:02Z","message":{"role":"assistant","provider":%q,"model":%q,"usage":{"input":10},"content":[{"type":"toolCall","id":"call","name":"read"}]}}
{"type":"message","timestamp":"2026-10-24T00:00:03Z","message":{"role":"toolResult","toolCallId":"call","isError":true}}
`, n, ref.Provider, ref.Model))
	}
	index, err := NewUsageScanner(root).Scan(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	filter := UsageFilter{ModelRefs: []UsageModelRef{{"a", "shared"}, {"b", "other"}}}
	for repeat := 0; repeat < 3; repeat++ {
		records := index.FilteredRecords(filter)
		if len(records) != 2 || len(index.FilteredTools(filter)) != 2 || len(index.FilteredErrors(filter)) != 2 {
			t.Fatalf("pair union became a cross product: %+v", records)
		}
		if got := index.Summary(filter); got.Requests != 2 || got.ToolRuns != 2 || got.Turns != 4 {
			t.Fatalf("request filter incorrectly narrowed prompts: %+v", got)
		}
		var b bytes.Buffer
		if err = index.ExportCSV(&b, filter); err != nil {
			t.Fatal(err)
		}
		rows, err := csv.NewReader(&b).ReadAll()
		if err != nil || len(rows) != 3 {
			t.Fatalf("CSV %v %v", rows, err)
		}
		for _, row := range rows[1:] {
			if !((row[4] == "A" && row[5] == "shared") || (row[4] == "B" && row[5] == "other")) {
				t.Fatalf("unrelated CSV row: %v", row)
			}
		}
		b.Reset()
		if err = index.ExportJSON(&b, filter, UsageExportOptions{Preset: "all", Location: time.UTC}); err != nil {
			t.Fatal(err)
		}
		var out map[string]any
		if err = json.Unmarshal(b.Bytes(), &out); err != nil {
			t.Fatal(err)
		}
		if out["summary"].(map[string]any)["requests"] != float64(2) {
			t.Fatal(out["summary"])
		}
	}
	if got := index.Summary(UsageFilter{Models: []string{"shared"}}); got.Requests != 2 {
		t.Fatal("legacy raw model filter changed", got)
	}
	filter.Provider = "a"
	if index.Summary(filter).Requests != 1 {
		t.Fatal("provider intersection ignored")
	}
}

// Frozen collect.rs deliberately treats zero, missing and malformed reasoning
// as absent. Reasoning is a subset of output, never added to total tokens.
func TestUsageReasoningPresenceMatchesFrozenCollector(t *testing.T) {
	for _, tc := range []struct {
		name, field string
		tokens      uint64
		present     bool
	}{
		{"missing", "", 0, false}, {"zero", `,"reasoning":0`, 0, false},
		{"positive", `,"reasoning":7`, 7, true}, {"negative", `,"reasoning":-1`, 0, false},
		{"fraction", `,"reasoning":1.5`, 0, false}, {"null", `,"reasoning":null`, 0, false},
		{"string", `,"reasoning":"7"`, 0, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			put(t, filepath.Join(root, "p", "s.jsonl"), `{"type":"session","id":"s","cwd":"/w","timestamp":"2026-10-24T00:00:00Z"}
{"type":"message","timestamp":"2026-10-24T00:00:01Z","message":{"role":"assistant","provider":"a","model":"m","usage":{"input":10,"output":20`+tc.field+`}}}
`)
			index, err := NewUsageScanner(root).Scan(context.Background())
			if err != nil {
				t.Fatal(err)
			}
			if len(index.Records) != 1 {
				t.Fatal(index)
			}
			r := index.Records[0]
			if r.HasReasoning != tc.present || r.Tokens.Reasoning != tc.tokens || r.Tokens.Total != 30 {
				t.Fatalf("source reasoning presence/accounting: %+v", r)
			}
		})
	}
}
