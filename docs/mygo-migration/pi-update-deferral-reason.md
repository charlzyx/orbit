# Typed automatic-update deferral reason

The earlier status-notification delta deliberately left `deferred` silent because
it combined three different conditions. This small follow-up adds an in-memory
`PiUpdateDeferralReason` field and attaches it at the existing three event sites:

| Existing site | Reason | UI behavior |
| --- | --- | --- |
| Missing Busy verifier before the idle loop | no_idle_verifier | Still silent; no running-task claim |
| Existing 120-second deadline while Busy loop remains active | busy_timeout | Source informational available_next_launch catalog, latest version |
| Existing maintenance admission rejected after the loop | admission_rejected | Still silent; rejection is not proof a task is running |

No timer, idle check, release lookup, installer, plan, admission, retry or worker
control flow changes. The field is internal transient event data, not persisted
preferences or a wire/protocol schema. Existing internal detail text is retained
but never displayed by the new notice mapping. Unknown/empty reasons stay silent.
Pinned Rust `app/pi_update_ui.rs:101–108` supports the busy-timeout informational
notice; the other Go-specific reasons have no equivalent source running-task
claim. This does not infer that every admission rejection is caused by a task.

A pure event constructor fixture checks reason and existing version/installation
facts. Mounted mock-event fixtures first fail against the old silent busy result,
then cover source severity/copy, translated version, unknown/non-busy silence,
owner-queue application and stale-generation rejection. A source callsite audit
confirms all three existing deferred emit points use their distinct reason.
These tests do not execute the automatic worker, wait 120 seconds, contact a
release endpoint or run an updater. Final integration remains review-gated.
