package native

// PiUpdateDeferralReason is an in-memory event reason, not persisted user state.
// A rejected admission is not proof that a task is running.
type PiUpdateDeferralReason string

const (
	PiUpdateDeferralBusyTimeout       PiUpdateDeferralReason = "busy_timeout"
	PiUpdateDeferralNoIdleVerifier    PiUpdateDeferralReason = "no_idle_verifier"
	PiUpdateDeferralAdmissionRejected PiUpdateDeferralReason = "admission_rejected"
)

func piUpdateDeferredEvent(reason PiUpdateDeferralReason, latest, before, detail string, installed PiInstallation) PiUpdateEvent {
	return PiUpdateEvent{State: "deferred", Deferral: reason, Latest: latest, Before: before, Detail: detail, Installation: installed}
}
