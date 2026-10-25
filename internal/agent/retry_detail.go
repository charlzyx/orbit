package agent

import (
	"encoding/json"
	"github.com/charlzyx/orbit/internal/pirpc"
)

func retryDetail(event pirpc.Event) RetryDetail {
	d := RetryDetail{Active: true, Attempt: 1}
	var attempt uint64
	if json.Unmarshal(event.Fields["attempt"], &attempt) == nil && string(event.Fields["attempt"]) != "null" {
		d.Attempt = attempt
	}
	var max uint64
	if json.Unmarshal(event.Fields["maxAttempts"], &max) == nil && string(event.Fields["maxAttempts"]) != "null" {
		d.Max, d.HasMax = max, true
	}
	var message string
	if json.Unmarshal(event.Fields["errorMessage"], &message) == nil && string(event.Fields["errorMessage"]) != "null" {
		d.Error, d.ErrorProvided = message, true
	}
	return d
}
