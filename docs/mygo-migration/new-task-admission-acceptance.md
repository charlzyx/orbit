# New Task: commit local draft changes only after host admission

Independent additive on local frozen P03 baseline `d54db58`; no installer UI.

Real entry-point audit found that `App.command(NewSession)` saved then cleared
Draft/attachments and marked pendingNew before dispatching. Controller maintenance
correctly rejected the operation, but its "draft is kept" notice was false. The
old maintenance test called Controller.Action directly, missing all UI mutations.
A mounted RED reproduced loss through sidebar, keyboard shortcut, native-menu
command, and command palette (`/tmp/orbit-newtask-admission-red.log`, 4.561s).

The action now carries a private positive admission receipt, initially false.
Controller acknowledges it only after Manager.NewSession returns nil. App dispatches
first and commits draft-map/save/clear/navigation only after that acknowledgement.
No optimistic clear/rollback and no global-maintenance-only shortcut is used.
Missing/no-op controller or an actual manager error cannot discard the draft.

Acknowledgement means the host accepted the task intent; it does NOT assert that
the asynchronous new_session response or process bootstrap succeeded. Manager
transition/lease/generation handling remains unchanged. Direct Controller actions
without a receipt retain their previous behavior. A stale/reentrant callback that
already presents a different opaque runtime/session, path/workspace, text or
attachment IDs cannot clear that newer presentation. Late acknowledgement does
not schedule a delayed clear; the flag is atomic so an incorrectly asynchronous
custom callback cannot introduce a data race.

Rust `app/session.rs:581–611` routes explicit NewSession through its runtime/task
transition, and the shared menu/keyboard/sidebar path is still preserved. This
increment adds the missing local admission boundary needed for Go's maintenance
refusal, rather than changing remote RPC success/cancel semantics.

Tests exercise all four real UI entry paths while maintenance is held; absent
controller, no-op and closed-manager rejection; positive host acknowledgement;
stale task/text/attachment presentation; late acknowledgement; and the actual
controller with an intentionally nonexistent fixture launcher. The latter proves
that admission is distinct from remote connection success. No installer, network,
authentication or real-user configuration is exercised.
