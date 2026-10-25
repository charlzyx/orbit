package agent

import "testing"

func TestRetryDetailObservedLifecycle(t *testing.T) {
	s := newState("/one", "")
	apply(t, &s, `{"type":"auto_retry_start","attempt":2,"maxAttempts":5,"errorMessage":"rate limited"}`)
	d := s.snapshot.RetryDetail
	if !s.snapshot.Retrying || !d.Active || d.Attempt != 2 || !d.HasMax || d.Max != 5 || d.Error != "rate limited" || !d.ErrorProvided || s.snapshot.ConnectionError != "" {
		t.Fatal(s.snapshot)
	}
	frozen := s.snapshot
	apply(t, &s, `{"type":"auto_retry_start","attempt":3,"errorMessage":"unavailable"}`)
	if s.snapshot.RetryDetail.Attempt != 3 || s.snapshot.RetryDetail.HasMax || frozen.RetryDetail != d {
		t.Fatal("detail leaked across attempts/snapshot", s.snapshot, frozen)
	}
	apply(t, &s, `{"type":"auto_retry_end","success":true}`)
	if s.snapshot.Retrying || s.snapshot.RetryDetail.Active || s.snapshot.ConnectionError != "" {
		t.Fatal(s.snapshot)
	}
	for _, end := range []string{`{"type":"auto_retry_end"}`, `{"type":"agent_settled"}`, `{"type":"process_exited"}`, `{"type":"transport_error","error":"closed"}`} {
		apply(t, &s, `{"type":"auto_retry_start"}`)
		apply(t, &s, end)
		if s.snapshot.Retrying || s.snapshot.RetryDetail != (RetryDetail{}) {
			t.Fatal("stale retry detail", end, s.snapshot)
		}
	}
}

func TestRetryDetailMalformedAndFailure(t *testing.T) {
	for _, fields := range []string{``, `,"attempt":null,"maxAttempts":null,"errorMessage":null`, `,"attempt":-1,"maxAttempts":"5","errorMessage":3`, `,"attempt":1.5,"maxAttempts":-1`} {
		s := newState("/one", "")
		apply(t, &s, `{"type":"auto_retry_start"`+fields+`}`)
		if s.snapshot.RetryDetail != (RetryDetail{Active: true, Attempt: 1, Revision: 1}) {
			t.Fatal(fields, s.snapshot.RetryDetail)
		}
	}
	for _, tc := range []struct{ raw, want string }{
		{`{"type":"auto_retry_end","success":false,"finalError":"overloaded"}`, "Automatic retry failed: overloaded"},
		{`{"type":"auto_retry_end","success":false}`, "Automatic retry failed: pi exhausted its automatic retries"},
		{`{"type":"auto_retry_end","success":false,"finalError":42}`, "Automatic retry failed: pi exhausted its automatic retries"},
		{`{"type":"auto_retry_end","success":true}`, ""},
		{`{"type":"auto_retry_end"}`, ""},
	} {
		s := newState("/one", "")
		apply(t, &s, `{"type":"auto_retry_start"}`)
		apply(t, &s, tc.raw)
		if s.snapshot.ConnectionError != tc.want || s.snapshot.RetryDetail.Active {
			t.Fatal(tc, s.snapshot)
		}
	}
}

func TestRetryDetailRevisionChangesEvenForIdenticalEvent(t *testing.T) {
	s := newState("/one", "")
	event := `{"type":"auto_retry_start","attempt":1,"errorMessage":"same"}`
	apply(t, &s, event)
	first := s.snapshot.RetryDetail
	apply(t, &s, event)
	second := s.snapshot.RetryDetail
	if first.Revision == 0 || second.Revision <= first.Revision || first.Attempt != second.Attempt || first.Error != second.Error {
		t.Fatal(first, second)
	}
	apply(t, &s, `{"type":"auto_retry_end","success":true}`)
	apply(t, &s, event)
	if s.snapshot.RetryDetail.Revision <= second.Revision {
		t.Fatal("end reused cancellation revision", s.snapshot)
	}
}
