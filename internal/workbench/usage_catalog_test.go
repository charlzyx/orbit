package workbench

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestUsageModelCatalogSourceOrderCacheAndNoUsageMessages(t *testing.T) {
	root := t.TempDir()
	a := filepath.Join(root, "group", "a.jsonl")
	b := filepath.Join(root, "group", "b.jsonl")
	put(t, a, `{"type":"session","id":"catalog","cwd":"/catalog","timestamp":"2026-10-01T00:00:00Z"}
{"type":"message","timestamp":"2026-10-03T00:00:00Z","message":{"role":"assistant","provider":"p","model":"zebra","content":[]}}
{"type":"message","timestamp":"2026-10-01T00:00:01Z","message":{"role":"assistant","provider":"p","model":"alpha","content":[]}}
{"type":"message","timestamp":"2026-10-01T00:00:02Z","message":{"role":"assistant","content":[]}}
`)
	put(t, b, `{"type":"session","id":"billed","cwd":"/billed","timestamp":"2026-10-01T00:00:00Z"}
{"type":"message","timestamp":"2026-10-01T00:00:03Z","message":{"role":"assistant","provider":"p","model":"alpha","usage":{"input":10}}}
{"type":"message","timestamp":"2026-10-01T00:00:04Z","message":{"role":"assistant","provider":"p","model":"zebra","usage":{"input":10}}}
{"type":"message","timestamp":"2026-10-01T00:00:05Z","message":{"role":"assistant","provider":"other","model":"alpha","content":[]}}
`)
	scanner := NewUsageScanner(root)
	want := []UsageModelRef{{"p", "zebra"}, {"p", "alpha"}, {"", ""}, {"other", "alpha"}}
	for n := 0; n < 3; n++ {
		index, err := scanner.Scan(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(index.ModelCatalog, want) || len(index.Records) != 2 || index.Summary(UsageFilter{}).Tokens.Total != 20 || index.Summary(UsageFilter{}).Cost != 0 || index.Summary(UsageFilter{}).PricedRequests != 0 {
			t.Fatalf("scan %d: %+v", n, index)
		}
		if n > 0 && index.Parsed != 0 {
			t.Fatal("cached catalog was reparsed", index.Parsed)
		}
		raw, err := json.Marshal(index)
		if err != nil {
			t.Fatal(err)
		}
		var fields map[string]json.RawMessage
		if err = json.Unmarshal(raw, &fields); err != nil {
			t.Fatal(err)
		}
		if _, ok := fields["ModelCatalog"]; ok {
			t.Fatal("transient catalog leaked into schema")
		}
		// Top-model ties follow first assistant occurrence across all files, not
		// lexicographic names or the chronologically sorted billable records.
		var out bytes.Buffer
		if err = index.ExportJSON(&out, UsageFilter{}, UsageExportOptions{Preset: "all", Location: time.UTC}); err != nil {
			t.Fatal(err)
		}
		var exported struct {
			Sessions []struct {
				TopModel string `json:"top_model"`
			} `json:"sessions"`
		}
		if err = json.Unmarshal(out.Bytes(), &exported); err != nil {
			t.Fatal(err)
		}
		if len(exported.Sessions) != 1 || exported.Sessions[0].TopModel != "alpha · P" {
			t.Fatal(out.String())
		}
		// Returned metadata cannot alias cached file catalogs.
		index.ModelCatalog[0] = UsageModelRef{"tampered", "external"}
	}
	if err := os.Remove(a); err != nil {
		t.Fatal(err)
	}
	index, err := scanner.Scan(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(index.ModelCatalog, []UsageModelRef{{"p", "alpha"}, {"p", "zebra"}, {"other", "alpha"}}) || index.Parsed != 0 {
		t.Fatal("removed-file models/order survived cached merge", index.ModelCatalog, index.Parsed)
	}
	// A valid empty-id file contributes no source models at all.
	put(t, filepath.Join(root, "group", "empty-id.jsonl"), `{"type":"session","id":""}
{"type":"message","message":{"role":"assistant","provider":"ignored","model":"ignored"}}
`)
	put(t, filepath.Join(root, "group", "unreadable.jsonl"), `{"type":"message","message":{"role":"assistant","provider":"ignored","model":"ignored"}}`)
	index, err = scanner.Scan(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	for _, ref := range index.ModelCatalog {
		if strings.Contains(ref.Provider, "ignored") {
			t.Fatal(index.ModelCatalog)
		}
	}
	if index.UnreadableFiles != 1 || index.Summary(UsageFilter{}).Requests != 2 {
		t.Fatal(index)
	}
}

func TestUsageModelCatalogLegacyFallbackIsDeterministicAndTransient(t *testing.T) {
	index := UsageIndex{Records: []UsageRecord{{Provider: "b", Model: "same"}, {Provider: "a", Model: "same"}, {Provider: "b", Model: "same"}}, Errors: []UsageError{{Provider: "", Model: ""}}, Tools: []ToolRun{{Provider: "t", Model: "tool"}}}
	want := []UsageModelRef{{"b", "same"}, {"a", "same"}, {"", ""}, {"t", "tool"}}
	if !reflect.DeepEqual(index.UsageModels(), want) {
		t.Fatal(index.UsageModels())
	}
	raw, _ := json.Marshal(index)
	var restored UsageIndex
	if err := json.Unmarshal(raw, &restored); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(restored.UsageModels(), want) {
		t.Fatal("legacy schema compatibility lost", restored.UsageModels())
	}
}

func TestUsageExportLocalizedSessionFallbackTitleTiesAndSeparateSearchFields(t *testing.T) {
	at := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
	index := UsageIndex{Records: []UsageRecord{
		{At: at, Session: "1", SessionPath: "/z", SessionTitle: "Zulu", Workspace: "/private/root/workspace", Provider: "p", Model: "m", Tokens: Tokens{Total: 10}},
		{At: at, Session: "9", SessionPath: "/a", SessionTitle: "Alpha", Workspace: "/private/root/workspace", Provider: "p", Model: "m", Tokens: Tokens{Total: 10}},
		{At: at, Session: "汉字🙂abcdefghi", SessionPath: "/empty", Workspace: "/private/root/workspace", Provider: "p", Model: "m", Tokens: Tokens{Total: 1}},
	}}
	export := func(options UsageExportOptions) []map[string]any {
		t.Helper()
		var b bytes.Buffer
		options.Preset = "all"
		options.Location = time.UTC
		if err := index.ExportJSON(&b, UsageFilter{}, options); err != nil {
			t.Fatal(err)
		}
		var doc map[string]json.RawMessage
		if err := json.Unmarshal(b.Bytes(), &doc); err != nil {
			t.Fatal(err)
		}
		if _, ok := doc["SessionTitles"]; ok {
			t.Fatal("display options leaked into schema")
		}
		var rows []map[string]any
		if err := json.Unmarshal(doc["sessions"], &rows); err != nil {
			t.Fatal(err)
		}
		return rows
	}
	rows := export(UsageExportOptions{})
	if len(rows) != 3 || rows[0]["title"] != "Alpha" || rows[1]["title"] != "Zulu" || rows[2]["title"] != "Session 汉字🙂abcde" {
		t.Fatal(rows)
	}
	titles := map[string]string{"/empty": "会话 汉字🙂abcde"}
	rows = export(UsageExportOptions{SessionTitles: titles, SessionSearch: "会话"})
	if len(rows) != 1 || rows[0]["title"] != "会话 汉字🙂abcde" {
		t.Fatal(rows)
	}
	for _, query := range []string{"Alpha workspace", "private/root", "Zulu workspace"} {
		if rows := export(UsageExportOptions{SessionSearch: query}); len(rows) != 0 {
			t.Fatal("source JSON search joined fields or exposed full path", query, rows)
		}
	}
	if rows := export(UsageExportOptions{SessionSearch: "WORKSPACE"}); len(rows) != 3 {
		t.Fatal(rows)
	}
	var csv bytes.Buffer
	if err := index.ExportCSV(&csv, UsageFilter{Session: "汉字🙂abcdefghi"}); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(csv.String(), "Session 汉字") || strings.Contains(csv.String(), "会话") {
		t.Fatal("CSV must retain raw empty title, unlike JSON", csv.String())
	}
}
