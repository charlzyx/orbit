# Test-only fresh extension receipt timestamp

`reset-a11y-final-full-1920.log` failed the unrelated existing
TestRetryCancellationFooterLocalizationAndSessionOwnership at line148 after
9.24s: local cancellation disappeared but new extension status was absent.
Assets passed; full UI failed; vet passed. Do not report that run as green.
Original test alone passed race3x in7.657s (`reset-retry-isolated-1930.log`).

Source cause: retry_feedback_test assigned ExtensionStatusAt=tt.Now(), a frozen
virtual frame clock captured earlier, while extension_panels.go:444–449 compares
extension status to time.Now() with a six-second TTL. Under full race load the
"newly received" fixture was already older than six real seconds. This is a
mixed-clock fixture defect, not evidence to change production expiration.

The independent additive changes only that synthetic new extension event's
receipt timestamp to time.Now() and asserts that it is strictly later than the
local retry cancellation timestamp. The local retryFeedbackNow=tt.Now clock,
virtual Advance, every original ownership/localization assertion and all
production clocks/TTLs are retained. No sleeps, larger timeouts or skipped
assertions are introduced. Final focused5x/full gates must be read separately.
