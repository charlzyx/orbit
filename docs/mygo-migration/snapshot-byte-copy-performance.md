# Snapshot byte-copy performance correction

This is a narrow performance candidate, not proof that macOS typing is fixed.
The released `.20261027.1` runtime and prior `.20261026.5` have byte-identical
`internal/agent/manager.go` (Git blob `6c9ab7d943dd7ec1924fa7f2aef321724023564e`).
Visible `Controller.View` invokes `tickHost → syncAt → Manager.Snapshot` on the
UI owner for each frame, including input and 90 ms idle / 16 ms active heartbeats.
Snapshot recursively copied every raw JSON byte via reflection while holding
Manager.mu. The candidate uses `reflect.Copy` only for uint8 element slices;
nil/empty/type identity and independent storage remain unchanged. No caching,
clock changes, replay behavior, or UI ownership changes are introduced.

## Evidence

Non-race Linux amd64, Go 1.27.3, pinned canonical wrapper, AMD EPYC 9V74.
Synthetic raw protocol messages enter the actual reducer, then actual Snapshot.
Alternating user/assistant each carries approximately 1 KiB plain text. Timings
exclude fixture setup. These are CPU microbenchmarks, not macOS frame rates.

| Messages | Before | After | Bytes/op (unchanged) |
|---:|---:|---:|---:|
| 0 | 5.38 us | 4.31 us | 4,880 |
| 10 | 457 us | 19.6 us | 38,776 |
| 100 | 4.48 ms | 188 us | 340,172 |
| 1,000 | 40.1 ms | 1.85 ms | 3,363,432 |
| 10,000 | 429 ms | 20.9 ms | 33,532,840 |

The 1,000-message baseline CPU profile attributes 98.38% cumulative time to
`cloneValue`; reflect Index/Set/assignTo are the dominant work. Allocations are
not reduced; mutable snapshot isolation is retained.

An additional empty-message scenario includes synthetic model catalog and
sidebar rows. Three repetitions: 100 models + 100 sessions, median 285 us →
62.3 us; 1,000 + 1,000, median 2.88 ms → 648 us. Allocation counts remain 432
and 4,032 respectively. This alone does not explain empty-composer CPU saturation.

`transcriptui.Project` independently still costs approximately 7.85 ms and
45,015 allocations for 1,000 messages; that broader projection is unchanged.
`syncUsageStore` has no synchronous disk I/O and only scans session metadata
when the Usage page is active. Long-history projection remains a separate issue.

Validation: 11 Snapshot/concurrency/real-child regression names, each run/pass
three times with race detector, zero skip/fail. New tests cover nil, empty,
named uint8/slice, json.RawMessage, nested map/pointer isolation in both mutation
directions. Existing partial JSON, runtime lifecycle, concurrent Tick/Snapshot,
dialog and redaction tests remain enabled.

Reproduce timing with canonical wrapper and `-run '^$' -bench
'BenchmarkDiagnostic(ManagerSnapshot|EmptyWithCatalogAndSessions)' -benchmem`.
No user sessions, credentials, actual model calls, or external publication used.
