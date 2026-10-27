package agent

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"testing"
)

var perfSnapshotSink Snapshot

func BenchmarkDiagnosticManagerSnapshot(b *testing.B) {
	for _, n := range []int{0, 10, 100, 1000, 10000} {
		b.Run(fmt.Sprint(n), func(b *testing.B) {
			st := newState("/synthetic/workspace", "/synthetic/session.jsonl")
			for i := 0; i < n; i++ {
				role := "user"
				if i%2 == 1 {
					role = "assistant"
				}
				raw, _ := json.Marshal(map[string]any{"id": fmt.Sprint(i), "role": role, "content": []map[string]string{{"type": "text", "text": strings.Repeat("Synthetic code discussion. ", 40)}}, "timestamp": "2026-10-27T00:00:00Z"})
				st.appendMessage(raw, false)
			}
			m := &Manager{active: "benchmark", sessions: map[string]*liveSession{"benchmark": {state: st}}}
			if len(m.Snapshot().Messages) != n {
				b.Fatal("fixture count")
			}
			b.ReportAllocs()
			b.ResetTimer()
			for b.Loop() {
				perfSnapshotSink = m.Snapshot()
			}
		})
	}
}

type perfNamedBytes []byte

func TestSnapshotByteSliceCloneIsolation(b *testing.T) {
	type nested struct {
		Raw   json.RawMessage
		Data  []byte
		Named perfNamedBytes
		Nil   []byte
		Empty []byte
		Map   map[string][]byte
		Ptr   *json.RawMessage
	}
	ptr := json.RawMessage(`{"ok":true}`)
	original := nested{Raw: json.RawMessage(`{"text":"hello"}`), Data: []byte{1, 2}, Named: perfNamedBytes{3, 4}, Empty: []byte{}, Map: map[string][]byte{"key": {5, 6}}, Ptr: &ptr}
	cloned := cloneValue(reflect.ValueOf(original)).Interface().(nested)
	if !reflect.DeepEqual(original, cloned) || cloned.Nil != nil || cloned.Empty == nil {
		b.Fatal("semantic shape changed")
	}
	cloned.Raw[0] = 'X'
	cloned.Data[0] = 8
	cloned.Named[0] = 9
	cloned.Map["key"][0] = 10
	(*cloned.Ptr)[0] = 'Y'
	if original.Raw[0] != '{' || original.Data[0] != 1 || original.Named[0] != 3 || original.Map["key"][0] != 5 || (*original.Ptr)[0] != '{' {
		b.Fatal("alias retained")
	}
}

func TestSnapshotByteSliceCloneShapes(t *testing.T) {
	type octet uint8
	type named []octet
	for _, input := range []any{[]byte(nil), []byte{}, []byte{1, 2}, json.RawMessage(nil), json.RawMessage{}, json.RawMessage(`{\"partial\":`), perfNamedBytes(nil), perfNamedBytes{}, perfNamedBytes{3, 4}, named{5, 6}} {
		v := reflect.ValueOf(input)
		got := cloneValue(v)
		if got.Type() != v.Type() || got.IsNil() != v.IsNil() || !reflect.DeepEqual(got.Interface(), input) {
			t.Fatalf("shape/type changed for %T", input)
		}
		if v.Len() > 0 {
			before := v.Index(0).Uint()
			got.Index(0).SetUint(before + 1)
			if v.Index(0).Uint() != before {
				t.Fatalf("alias for %T", input)
			}
			v.Index(0).SetUint(before + 2)
			if got.Index(0).Uint() != before+1 {
				t.Fatalf("reverse alias for %T", input)
			}
		}
	}
}

func BenchmarkDiagnosticEmptyWithCatalogAndSessions(b *testing.B) {
	for _, n := range []int{100, 1000} {
		b.Run(fmt.Sprint(n), func(b *testing.B) {
			st := newState("/synthetic/workspace", "/synthetic/session.jsonl")
			m := &Manager{active: "benchmark", sessions: map[string]*liveSession{"benchmark": {state: st}}}
			for i := 0; i < n; i++ {
				raw, _ := json.Marshal(map[string]any{"id": fmt.Sprintf("model-%d", i), "provider": "synthetic", "name": "Synthetic Model", "contextWindow": 200000, "input": []string{"text", "image"}, "reasoning": true})
				m.sessions["benchmark"].state.snapshot.Models = append(m.sessions["benchmark"].state.snapshot.Models, ModelChoice{ID: fmt.Sprint(i), Provider: "synthetic", Name: "Synthetic Model", Raw: raw})
				m.diskRows = append(m.diskRows, SessionRow{ID: fmt.Sprint(i), Path: fmt.Sprintf("/synthetic/session-%d.jsonl", i), Workspace: "/synthetic/workspace", Summary: strings.Repeat("Synthetic session summary. ", 8)})
			}
			if len(m.Snapshot().Models) != n || len(m.Snapshot().Sessions) != n {
				b.Fatal("fixture count")
			}
			b.ReportAllocs()
			b.ResetTimer()
			for b.Loop() {
				perfSnapshotSink = m.Snapshot()
			}
		})
	}
}
