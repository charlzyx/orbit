# Composer Access / Workflow finite pickers

Scope: isolated follow-on to wave8 `0bf7cb6`; excludes separately approved palette titles and safe Rename. Pinned Rust: `299a32bc87fe3d4851bdb902d9c03dcf9bf99cac`.

## Source contract

- `crates/orbit-pi/src/app/pickers.rs:350–358,461–464`: palette dispatch toggles the existing composer picker; it does not navigate to Chat or open a free-text dialog.
- `app/composer_ops.rs:686–750,804–878`: finite current selection, wrapping keyboard navigation, selection/close restoring composer focus, competing composer menus closed.
- `app/composer_ops.rs:758–781`: Access is a global persisted policy; it is not scoped to the selected task.
- `app/composer_ops.rs:885–901`: Workflow belongs to current session identity; before hydration it is retained as pending rather than rejected for empty ID.
- `app/view.rs:424,791–897,1006,1428–1462,1706`: chips/pickers exist only in visible noncompact composer. Hidden or compact commands retain same-owner open intent, without mounting a global popup or changing page. The picker appears when composer becomes visible again. Workspace/runtime/session/path changes cancel intent.

## Implementation and admission

Public native PopoverBase anchors a finite menu to each original chip. It shares enum/apply logic with palette commands; current marker, hints, Up/Down wrap, Enter and Escape are supported. Native public chip labels remain unchanged. Add-menu identity advances when mode picker opens; held old choices cannot act on a newer picker. No toolkit/private API changes.

`CaptureWorkflowTargetFor` validates displayed runtime/session/path/workspace under the manager mutex, with no filesystem or readiness RPC. It returns the existing opaque key/generation. Nil/closed/stopped/replaced identity is rejected; a legitimate unhydrated task remains supported. Selection rechecks original target and the Controller queues the original target to unchanged `SetWorkflowModeFor`, never a newly captured current task. Other existing SetWorkflow callers retain their prior behavior.

Access remains global. UI validates the enum and dispatches; it no longer writes access.json before Controller admission. The existing Controller maintenance gate and Manager.SetAccessMode are the sole persistence path. This avoids the reproduced maintenance refusal that previously still changed the file. Invalid backend values remain rejected. No permission semantics or setter lease protocol changed.

## Evidence

- Actual mounted palette/hidden-page RED: `composer-mode-red-2117.log`.
- Actual temporary-store maintenance RED: `composer-mode-access-admission-red-2126.log`; file changed despite refused action before removal of the UI prewrite.
- The same Access log also contains a test-fixture mistake: reopening with empty session path starts a fresh runtime. The final fixture supplies a synthetic saved path and verifies the original parked runtime ID is reused and updated. This is not claimed as a product defect.
- Final focused JSON run/pass inventory, vet and darwin/arm64 compile are recorded in the adjacent delivered manifest. Intermediate 2128/2130 runs failed compilation due to an unused fixture path and are not green evidence.
- Mounted tests cover actual palette/chip pointer and keyboard behavior, Chinese text, compact/hidden/full-review retention, focus, new-owner/ABA/held clicks, competing Add/model menus, actual Controller stale display and delayed dispatch, exact-task Workflow persistence, temporary-store Access bytes unchanged on maintenance/invalid input, and global update of current plus original parked runtime.
- The 2132 JSON gate retained one intermittent pre-arm fixture failure (`Ask` not mounted on first render); the delayed-dispatch fixture now explicitly drains/renders its initial frame and synchronizes the displayed snapshot before arming. It still performs real mounted selection, then same-child NewSession, delayed original Action delivery and exact new-task nonmutation/error assertions. No production guard or timeout was loosened. A separate diagnostic edit briefly failed compilation (2135 log), also not green evidence.
- Agent tests cover legal empty-ID pending persistence after hydration, same-key new generation, mismatched display dimensions, Stop/Close/replacement and unknown backend values.

## Limits

All fixtures are disposable. No user settings, real permission changes, external endpoint, inference or auth were used. Native mounted synthetic tests and arm64 compile are not macOS OS accessibility, screenshot parity, real terminal, or full migration acceptance. A visible enum choice may briefly project locally before a rejected action refreshes the model; persistence remains guarded. Existing source English catalog residuals are unchanged. Full combined regression belongs to the coordinator's next isolated integration gate.
