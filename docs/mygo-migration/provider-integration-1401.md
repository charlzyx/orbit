# Provider completion integration, final v2

Exact public base3fe50d35a65f56f40cd73979531cb80296a9e611. Successful same-workspace provider save effects survive dismissal, while editor identity still protects replacement drafts/errors. Existing active-only delayed config apply remains; no promise every parked runtime reloads.

The first full chain87063 FAILED two new error-fixture cases with `save not admitted`; its log is retained at orbit-provider-gate-1321/full.jsonl. Exact diagnostic repeats showed actual directory-read errors had already been applied during subsequent real Click frames, not unadmitted saves. The test-only fix establishes an unbuffered completion hold before actual Click, then restores the original channel only after receiving the real completion. Original3s deadline/real filesystem/error/new-draft assertions remain. Timeout stays FAIL, no unsafe queue close/restore or cancellation claim. Production providers.go SHA2569bffc06f8539d04a677de9d6013a3216e22d44fe1d7ad98d4a78361b019c5e9f is unchanged from the first reviewed candidate.

Final v2 serial full chain48864 EXIT0, all1855 non-document hashes unchanged.266 exact package/test identities run/pass once, retaining252 and14 provider names. Root3496 passing events/19 enumerated conditional skips are not feature counts. Composer/selection19 and menu22top/32events zero skip; terminal115events/4conditional skips. Wrapper, vet, Darwin ARM64, publisher40/icon5/legacy21 all pass.

Only temporary fixture models.json and synthetic values were used. No real credentials/network/model call, new marker/package/release, or native clipboard/menu input. The separately designed cloudMac composer harness is not part of this source batch.
