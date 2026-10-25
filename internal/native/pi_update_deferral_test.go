package native

import "testing"

func TestPiUpdateDeferralEventKeepsTypedReasonAndFacts(t *testing.T) {
	installed := PiInstallation{Kind: PiManaged, Bin: "/fixture/pi", Version: "0.60.1", OnPath: true}
	for _, reason := range []PiUpdateDeferralReason{PiUpdateDeferralBusyTimeout, PiUpdateDeferralNoIdleVerifier, PiUpdateDeferralAdmissionRejected} {
		event := piUpdateDeferredEvent(reason, "0.60.3", "0.60.1", "internal fixture detail", installed)
		if event.State != "deferred" || event.Deferral != reason || event.Latest != "0.60.3" || event.Before != "0.60.1" || event.Detail != "internal fixture detail" || event.Installation != installed {
			t.Fatalf("deferral event lost existing facts: %+v", event)
		}
	}
}
