# Native UI source-first acceptance supplement

Frozen Rust: `299a32bc87fe3d4851bdb902d9c03dcf9bf99cac`. Cloud-only. This supplements the139-row master checklist. Scoped tests never mark an entire source row complete; every case keeps `full_row_parity_verified: false`. Matched visual and actual platform verdicts remain separate.

## Reproduce

Use the pinned Go1.27.3 toolchain through `scripts/go-native.sh test -race ./internal/ui ./cmd/orbit -count=1`. The canonical wrapper verifies upstream hashes and applies the isolated composer range-edit compatibility; cmd refuses unpatched startup. Tests use disposable homes and explicit fixtures, never authenticated model inference. Opt-in actual Ghostty tests require `ORBIT_TERMINAL_TEST=1` plus the pinned library. Skips are not passing evidence.

Native CPU captures: `ORBIT_CAPTURE_DIR=/absolute/output scripts/go-native.sh test ./internal/ui -run '^Test(NativeCapture|NativeUsageCapture)$' -count=1`. These are not matched Rust pixel comparisons.

## Highest-priority remaining failures

- C11: selection survives across mounted rows, but pinned MyGo clears it when a virtualized endpoint unmounts; Rust retains source selection snapshots. C17/C19/C20 specialized approval/questionnaire/widgets are not complete.
- P08 favorites/provider-filter model catalog and full appearance/localization remain incomplete. Settings source search semantics were corrected: label/synonym results open sections, not exact-row highlighting.
- Usage comparison/table/filter acceptance is being reconstructed separately; do not infer full UI parity from backend export/accounting tests.
- S12/S13 full sidebar marks/icons/sticky reveal remain incomplete. Linux desktop captures exist, but current Rust/Go viewport mismatch prevents a valid pixel comparison; no image is resized to conceal it. OS IME/accessibility/native lifecycle on all targets remains open.

## Cases

### UI-S01 · Window creation and parked lifetime (S01)

**Functional verdict:** `partial`. **Visual:** not compared to a matched Rust runtime. **Platform:** native CPU only unless explicitly noted.

Source: `crates/orbit-pi/src/main_window.rs:77::open`; `crates/orbit-pi/src/main_window.rs:136::spawn_heartbeat`; `crates/orbit-pi/src/app/view.rs:166::render`; `crates/orbit-pi/src/platform.rs:1507::titlebar_options`.

1. Start with empty isolated preferences.
2. Inspect native window minimum and initial maximize; close/reopen on macOS while a fake Pi session streams.

Expected: Source minimum 960×640; restore85% display width and90%height capped1440×920; one parked controller/process survives close.

Executed test coverage: `TestSourceWindowRestoreBounds`, `TestHiddenSidebarReservesNativeTrafficLights`, `TestSourceNativeMenuDoesNotAllocatePersistentWindowStrip`.

Remaining: Source minimum/restore bounds and native traffic-light clearance pass. Native menus now auto-hide on Linux/Windows while retaining edit roles and shortcuts. Actual initial-maximize/Alt-F10/Dock lifetime on each OS remains separate acceptance.

### UI-S02 · Geometry, fractional scaling and restored dimensions (S02)

**Functional verdict:** `partial`. **Visual:** not compared to a matched Rust runtime. **Platform:** native CPU only unless explicitly noted.

Source: `crates/orbit-pi/src/layout.rs:29::Layout`; `crates/orbit-pi/src/app/view.rs:166::render`; `crates/orbit-pi/src/app.rs:87::SIDEBAR_MIN_W`.

1. At1470×922 change scale1/1.25/1.5/2.
2. Drag sidebar/review/explorer/terminal widths; relaunch.
3. Repeat font12/14/18 and density80/100/120.

Expected: Logical rectangles stay stable under device scale; source clamps and persisted layout survive relaunch.

Executed test coverage: `TestFractionalScaleKeepsLogicalComposerGeometry`, `TestMinimumWindowModalsKeepActionsReachable`.

Remaining: DIP composer and modal bounds pass, including800×560 stress below source minimum. Full panel drag/relaunch/font-density matrix and matched pixels remain open.

### UI-S03 · Filtered command/model keyboard selection (S03, S04, S05)

**Functional verdict:** `pass_scoped`. **Visual:** not compared to a matched Rust runtime. **Platform:** native CPU only unless explicitly noted.

Source: `crates/orbit-pi/src/commands.rs:224::COMMANDS`; `crates/orbit-pi/src/commands.rs:794::BINDINGS`; `crates/orbit-pi/src/main.rs:357::bind_keys`; `crates/orbit-pi/src/app/pickers.rs:415::run_command`; `crates/orbit-pi/src/command_palette.rs:313::CommandPalette`; `crates/orbit-pi/src/app/pickers.rs:288::toggle_command_palette`; `crates/orbit-pi/src/app/pickers.rs:387::run_palette_command`; `crates/orbit-pi/src/main.rs:357::bind_keys`; `crates/orbit-pi/src/app/view.rs:3775::on_focus_next`; `crates/orbit-pi/src/app/view.rs:3804::focus_traversal_blocked`; `crates/orbit-pi/src/app/composer_ops.rs:537::close_add_menu`.

1. Open palette, move down six entries, filter to one and press Enter.
2. Filter to zero then restore and press Enter.
3. Open prompt from palette; Escape back to composer.

Expected: Filtered selection resets/clamps; zero results do not submit; stable modal keys restore focus.

Executed test coverage: `TestFilteredPickersResetAndClampKeyboardSelection`, `TestPickerKeyboardCommit`, `TestModalTransitionRestoresComposerFocus`.

Remaining: These exact native CPU flows pass, not every command/context or OS accelerator.

### UI-S04 · Native menu and modal precedence (S03, C16, C20)

**Functional verdict:** `partial`. **Visual:** not compared to a matched Rust runtime. **Platform:** native CPU only unless explicitly noted.

Source: `crates/orbit-pi/src/commands.rs:224::COMMANDS`; `crates/orbit-pi/src/commands.rs:794::BINDINGS`; `crates/orbit-pi/src/main.rs:357::bind_keys`; `crates/orbit-pi/src/app/pickers.rs:415::run_command`; `crates/orbit-pi/src/dialog.rs:171::Dialog`; `crates/orbit-pi/src/app/dialogs.rs:10::handle_extension_ui_request`; `crates/orbit-pi/src/app/dialogs.rs:205::open_dialog`; `crates/orbit-pi/src/app/dialogs.rs:236::respond_to_dialog`; `crates/orbit-pi/src/widgets.rs:?::source`; `crates/orbit-pi/src/custom_ui.rs:?::source`; `crates/orbit-pi/src/app/dialogs.rs:76::apply_extension_widget`; `crates/orbit-pi/src/app/dialogs.rs:114::apply_custom_frame`.

1. Open confirm, prompt, custom terminal, secret, image, model or palette.
2. Dispatch New task/Settings/Terminal through native menu.
3. For custom UI press Escape, then explicit Cancel.

Expected: No action mutates behind modal. Custom Escape sends ESC to Pi rather than dismissing; explicit Cancel cancels.

Executed test coverage: `TestNativeMenuCannotMutateBehindModal`, `TestCustomDialogDoesNotDismissEscape`, `TestLiveCustomDialogForwardsEscape`, `TestNativeMenuRespectsExternalPageModals`, `TestHasModalTracksReviewAndActualScreenshotPreview`.

Remaining: Native menu guard includes module-owned Git/Worktrees/MCP/Resources/report review/preview and Usage filter overlays. CPU routing tests pass; actual platform accelerator ordering remains separate.

### UI-S05 · Session switch, workspace routing and drafts (S06, S07, S08)

**Functional verdict:** `partial`. **Visual:** not compared to a matched Rust runtime. **Platform:** native CPU only unless explicitly noted.

Source: `crates/orbit-pi/src/sessions.rs:29::SessionInfo`; `crates/orbit-pi/src/sessions.rs:229::SessionWatcher`; `crates/orbit-pi/src/app/events.rs:18::tick`; `crates/orbit-pi/src/app/session.rs:581::on_new_session`; `crates/orbit-pi/src/app/session.rs:61::can_reuse_session`; `crates/orbit-pi/src/app/pickers.rs:118::start_task_in_folder`; `crates/orbit-pi/src/workspace_picker.rs:?::source`; `crates/orbit-pi/src/workspace_logo.rs:?::source`; `crates/orbit-pi/src/app.rs:150::ParkedSession`; `crates/orbit-pi/src/app/session.rs:299::park_active_session`; `crates/orbit-pi/src/app/runtime.rs:569::reap_idle_parked`; `crates/orbit-pi/src/app/events.rs:541::drain_parked_session`.

1. Enter a draft with attachments, switch to another saved task then back.
2. Switch to a task in another repository.
3. Observe Git/Files/Worktrees/terminal new-tab directory without starting an extra session.

Expected: Per-session draft is retained; domain actions target selected repository; loading sessions do not overwrite new active choice.

Executed test coverage: `TestDraftSurvivesSessionSwitch`, `TestInactiveFileLoadCannotStealSelectionAfterWorkspaceSwitch`.

Remaining: Draft/inactive file generation tests pass. Whole native controller repository-switch interaction with fake Pi and all domain mutations still needs a dedicated end-to-end fixture.

### UI-S06 · Sidebar sessions and workspace management (S09, S10, S11, S12, S13, S14)

**Functional verdict:** `partial`. **Visual:** not compared to a matched Rust runtime. **Platform:** native CPU only unless explicitly noted.

Source: `crates/orbit-pi/src/app/sidebar.rs:2509::open_session_slot`; `crates/orbit-pi/src/app/sidebar.rs:2521::cycle_session`; `crates/orbit-pi/src/app/sidebar.rs:2618::on_sidebar_confirm`; `crates/orbit-pi/src/app/session.rs:613::clone_session`; `crates/orbit-pi/src/app/settings.rs:5836::rename_session`; `crates/orbit-pi/src/sessions.rs:?::source`; `crates/orbit-pi/src/app/events.rs:18::tick`; `crates/orbit-pi/src/app/sidebar.rs:2188::on_menu_clone_session`; `crates/orbit-pi/src/pins.rs:?::source`; `crates/orbit-pi/src/archive.rs:?::source`; `crates/orbit-pi/src/app/sidebar.rs:2231::on_menu_delete_confirm`; `crates/orbit-pi/src/app/sidebar.rs:2214::on_menu_toggle_archive`; `crates/orbit-pi/src/app/sidebar.rs:176::build_sidebar_rows`; `crates/orbit-pi/src/app/sidebar.rs:370::build_flat_sidebar_rows`; `crates/orbit-pi/src/app.rs:2248::WorkspaceSort`; `crates/orbit-pi/src/app.rs:2339::SidebarGroupBy`; `crates/orbit-pi/src/app/sidebar.rs:2383::on_workspace_remove`; `crates/orbit-pi/src/app/sidebar.rs:2373::on_workspace_open_picker`; `crates/orbit-pi/src/workspace_mark.rs:13::WorkspaceMark`; `crates/orbit-pi/src/app.rs:?::save_workspaces`; `crates/orbit-pi/src/app/session.rs:2702::on_refresh`; `crates/orbit-pi/src/app/events.rs:1237::sync_workspace_watcher`; `crates/orbit-pi/src/app/refresh_tests.rs:?::source`.

1. Use numbered/next/previous session commands.
2. Rename, pin, archive, restore, clone and recoverably delete.
3. Change workspace sort/group/mark/icon and refresh fan-out.

Expected: Source ordering, active visibility, collapse persistence, all metadata operations and contextual menu semantics match.

Executed test coverage: `TestEmptySidebarDoesNotInventCwdProject`.

Remaining: Fresh cwd no longer manufactures a sidebar project; saved workspaces and actual sessions still populate groups. Source empty-state geometry is mounted. Full marks/icons/sticky active reveal/sort-state persistence remains incomplete.

### UI-C01 · Multiline and Chinese composition (C01)

**Functional verdict:** `pass_scoped`. **Visual:** not compared to a matched Rust runtime. **Platform:** native CPU only unless explicitly noted.

Source: `crates/orbit-pi/src/composer.rs:55::ComposerInput`; `crates/orbit-pi/src/main.rs:357::bind_keys`; `crates/orbit-pi/src/app/session.rs:88::submit_composer_as`.

1. Type line one, Shift+Enter, line two.
2. Start IME composition; Enter commits composition only.
3. Commit Chinese text then Enter.

Expected: No prompt during composition; native text editing and multiline text remain intact.

Executed test coverage: `TestIMECompositionDoesNotSubmit`, `TestComposerRepeatedEnterShiftAlternateAndRejection`.

Remaining: Headless composition event and multiline submission pass. Real Chinese/Japanese/Korean OS IME candidates, clipboard and accessibility remain unexecuted.

### UI-C02 · Delivery policy, repeated send and rejected retry (C02, C03)

**Functional verdict:** `partial`. **Visual:** not compared to a matched Rust runtime. **Platform:** native CPU only unless explicitly noted.

Source: `crates/orbit-pi/src/composer_send.rs:10::SendMode`; `crates/orbit-pi/src/app/session.rs:129::submit_as`; `crates/orbit-pi/src/theme.rs:324::UiPrefs`; `crates/orbit-pi/src/app/view.rs:4532::queue_bar`; `crates/orbit-pi/src/app/view.rs:4719::on_clear_queue`; `crates/orbit-pi/src/app/events.rs:?::source`; `crates/orbit-pi/src/app/session.rs:284::reset_queue`; `crates/orbit-pi/src/app/settings.rs:5580::set_follow_up_mode`.

1. Set follow_up; Alt+Enter sends steer.
2. Press Enter again with empty input.
3. Reject Submit during maintenance/disconnect, then retry.

Expected: One action per accepted nonempty submission; rejected draft remains; alternate mode is exact opposite.

Executed test coverage: `TestComposerRepeatedEnterShiftAlternateAndRejection`, `TestMaintenanceGateKeepsDraftAndBlocksNewProcess`.

Remaining: Exact tested sends/rejection pass. Source individual queue editing/removal and one/all queued delivery controls are incomplete in host UI.

### UI-C03 · Slash and file-mention autocomplete (C04)

**Functional verdict:** `partial`. **Visual:** not compared to a matched Rust runtime. **Platform:** native CPU only unless explicitly noted.

Source: `crates/orbit-pi/src/app/composer_ops.rs:9::autocomplete_entries`; `crates/orbit-pi/src/app/composer_ops.rs:113::commit_entry`; `crates/orbit-pi/src/app/composer_ops.rs:471::attach_file`; `crates/orbit-pi/src/mentions.rs:?::source`.

1. Type / then filter a command.
2. Type @ then select a nested file by keyboard.
3. Escape cancels; accepting replaces only trigger range.

Expected: Source composer_ops autocomplete ordering, keyboard ownership and trigger replacement.

Executed test coverage: `TestNativeComposerCompletionPrecedesSendAndAbort`.

Remaining: Native source-derived slash/skill/command/file completion is mounted; real command metadata and off-thread bounded filesystem scan, focus/caret/Enter/AltEnter/steer/Escape paths have tests. Canonical hash-checked MyGo compatibility supplies undoable caret-aware replacement; cmd refuses unpatched startup. Actual OS IME candidate panel and full source visual comparison remain unverified.

### UI-C04 · Attachments, drag/paste and partial failure (C05)

**Functional verdict:** `partial`. **Visual:** not compared to a matched Rust runtime. **Platform:** native CPU only unless explicitly noted.

Source: `crates/orbit-pi/src/app/composer_ops.rs:435::attach_image`; `crates/orbit-pi/src/app/composer_ops.rs:405::on_file_drop`; `crates/orbit-pi/src/app/events.rs:1108::drain_pasted_images`; `crates/orbit-pi/src/attachments.rs:55::compress_image`; `crates/orbit-pi/src/app/view.rs:4733::lightbox_layer`.

1. Select multiple valid images plus unreadable/oversized files.
2. Paste text and image; drag files.
3. Remove one; switch tasks; send and inspect RPC image bytes.

Expected: Bounded read/compression, actual per-file errors, no lost draft, source paste priority and max-count behavior.

Executed test coverage: No host test is claimed for this complete scenario..

Remaining: Attachment Prepare and transport domain tests exist; host native add/drop/clipboard/partial-failure end-to-end tests and source text-file mention semantics remain open.

### UI-C05 · Model/thinking/favorites (C06, P08)

**Functional verdict:** `partial`. **Visual:** not compared to a matched Rust runtime. **Platform:** native CPU only unless explicitly noted.

Source: `crates/orbit-pi/src/model_selector.rs:159::ModelSelector`; `crates/orbit-pi/src/model_selector_match.rs:?::source`; `crates/orbit-pi/src/app/pickers.rs:761::sync_model_selector`; `crates/orbit-pi/src/app/session.rs:52::can_choose_thinking`; `crates/orbit-pi/src/favorites.rs:?::source`; `crates/orbit-pi/src/recent_models.rs:?::source`; `crates/orbit-pi/src/app/settings.rs:1433::model_rows`; `crates/orbit-pi/src/app/settings.rs:1369::models_refresh`; `crates/orbit-pi/src/app/settings.rs:1082::model_toolbar`; `crates/orbit-pi/src/favorites.rs:?::source`.

1. Filter model picker to one/zero and recover.
2. Switch provider and select thinking level.
3. Favorite model, enable favorites-only and relaunch.

Expected: Actual catalog/capabilities; no unsupported thinking choice; favorites/recent persistence and source model cards.

Executed test coverage: `TestFilteredPickersResetAndClampKeyboardSelection`.

Remaining: Keyboard picker works. Source favorites-only/provider filter/detail catalog page remains incomplete; existing simple model list is not full P08.

### UI-C06 · Long streamed transcript reading position (C07, C08)

**Functional verdict:** `partial`. **Visual:** not compared to a matched Rust runtime. **Platform:** native CPU only unless explicitly noted.

Source: `crates/orbit-pi/src/transcript.rs:1081::load_from`; `crates/orbit-pi/src/transcript.rs:1181::apply_event`; `crates/orbit-pi/src/app/events.rs:645::apply_messages_snapshot`; `crates/orbit-pi/src/message_scroller.rs:?::source`; `crates/orbit-pi/src/transcript.rs:2278::jump_turn`; `crates/orbit-pi/src/transcript_view.rs:1396::render_navigation_rail`; `crates/orbit-pi/src/transcript.rs:877::persist_hint`.

1. Render120 variable-height messages.
2. Scroll up in30-DIP steps, append/grow streamed assistant message.
3. Jump to latest, grow again.

Expected: Stable keys retain reader position; follow resumes only at end; variable-height rows do not jump to bottom unexpectedly.

Executed test coverage: `TestTranscriptVariableHeightFollowAndReaderPosition`.

Remaining: Native variable-height follow/read-position test passes. Logical assistant turns now project across tool-result messages. Source navigation rail and full long-session performance/OS acceptance remain open.

### UI-C07 · Rich Markdown and code (C09)

**Functional verdict:** `partial`. **Visual:** not compared to a matched Rust runtime. **Platform:** native CPU only unless explicitly noted.

Source: `crates/orbit-pi/src/transcript_view.rs:5839::parse_blocks`; `crates/orbit-pi/src/transcript_view.rs:5201::parse_inline`; `crates/orbit-pi/src/transcript_view.rs:6177::render_block`; `crates/orbit-pi/src/transcript_view.rs:6537::render_code_block`; `crates/orbit-pi/src/highlight.rs:?::source`.

1. Render alerts, tables, lists, fenced code, links, images and malformed streamed fences.
2. Resize with code wrapping.
3. Change light/dark theme.

Expected: Source semantic Markdown and color-only syntax with stable font metrics.

Executed test coverage: `TestMarkdownAlertAndGeometry`.

Remaining: GFM alert/geometry sample passes; highlight tokenizer has1204 source-oracle cases. Whole source markdown corpus, per-message media/actions and exact spacing remain unverified.

### UI-C08 · Transcript selection, tools and turn summaries (C10, C11, C12)

**Functional verdict:** `partial`. **Visual:** not compared to a matched Rust runtime. **Platform:** native CPU only unless explicitly noted.

Source: `crates/orbit-pi/src/transcript_view.rs:302::TextSelection`; `crates/orbit-pi/src/transcript_view.rs:912::text_selection_menu`; `crates/orbit-pi/src/transcript.rs:1076::selected_text`; `crates/orbit-pi/src/transcript.rs:2271::mark_copied`; `crates/orbit-pi/src/transcript_view.rs:2387::render_activity_group`; `crates/orbit-pi/src/transcript_view.rs:3551::render_tool_detail`; `crates/orbit-pi/src/transcript_view.rs:3512::render_tool_error_strip`; `crates/orbit-pi/src/transcript_view.rs:3973::build_edit_diff`; `crates/orbit-pi/src/transcript_view.rs:7391::mcp_tool_parts`; `crates/orbit-pi/src/transcript_view.rs:4820::render_working_indicator`; `crates/orbit-pi/src/transcript_view.rs:4369::render_message_footer`; `crates/orbit-pi/src/transcript_view.rs:1811::render_summary_card`; `crates/orbit-pi/src/transcript_view.rs:7053::render_changed_files`.

1. Select across multiple Markdown blocks/messages.
2. Open Bash/read/edit/write/search tool cards and errors.
3. Inspect finished turn duration/tokens/cost.

Expected: Source structured tool presentation, cross-block selection, context menu and turn metrics.

Executed test coverage: `TestLogicalTranscriptKeepsAssistantProseAndToolResult`, `TestHostCrossMessageMarkdownAndAlertSelection`.

Remaining: Typed tools, logical turns, observed-live elapsed time, source tail/usage footer and shared user/Markdown/alert selection are mounted and native-tested. C11 remains a concrete framework gap: selection clears when a virtualized endpoint unmounts; Rust retains SelectedBlock snapshots. Do not equate mounted-row selection with full parity.

### UI-C09 · Transcript search (C13)

**Functional verdict:** `partial`. **Visual:** not compared to a matched Rust runtime. **Platform:** native CPU only unless explicitly noted.

Source: `crates/orbit-pi/src/app/search.rs:12::TranscriptSearch`; `crates/orbit-pi/src/app/search.rs:71::sync_transcript_search`; `crates/orbit-pi/src/transcript.rs:2388::find_messages`; `crates/orbit-pi/src/main.rs:357::bind_keys`.

1. Search text present in assistant and tool output.
2. Next/previous wraps; clear/close then search again.
3. Stream another matching row while search is active.

Expected: Find stays distinct from command palette; active hit scroll and highlight/count are correct.

Executed test coverage: `TestTranscriptFindKeyboardWrapTrimAndFocus`.

Remaining: Native trim/case/tool text, Enter/ShiftEnter wrap, Escape focus/no-abort and active wash tests pass. Search now uses logical turn text. Source floating320px bar geometry and full live-match evolution remain open.

### UI-C10 · Stop/retry/compact and context freshness (C14, C15)

**Functional verdict:** `partial`. **Visual:** not compared to a matched Rust runtime. **Platform:** native CPU only unless explicitly noted.

Source: `crates/orbit-pi/src/app/session.rs:417::on_abort`; `crates/orbit-pi/src/app/view.rs:3882::run_status_strip`; `crates/orbit-pi/src/app/runtime.rs:50::on_command_failure`; `crates/orbit-pi/src/app/settings.rs:5811::compact_now`; `crates/orbit-pi/src/app/settings.rs:5828::abort_retry`; `crates/orbit-pi/src/app/events.rs:?::source`; `crates/orbit-pi/src/context_meter.rs:633::ContextMeterData`; `crates/orbit-pi/src/context_meter.rs:120::usage_slices`; `crates/orbit-pi/src/app/runtime.rs:123::refresh_context_stats`; `crates/orbit-pi/src/app/view.rs:3195::context_button`.

1. Stop active run; retry a failed run.
2. Compact successfully and inspect missing context before refreshed stats.
3. Repeat failed/aborted compaction.

Expected: No cumulative-total substitute for context; missing is unknown, not0%; successful compaction invalidates old context.

Executed test coverage: `TestRuntimeUsageUsesPiSessionAndContextSeparately`, `TestRuntimeUsageMissingOrPostCompactionIsNotZero`, `TestLatestAssistantUsageDoesNotReuseStalePreviousOutput`.

Remaining: Pure adapters and agent compaction reducer tests pass. Full native retry/compaction indicators and every button/state transition remain open.

### UI-C11 · Extension special-purpose surfaces (C16, C17, C18, C19, C20)

**Functional verdict:** `failed_missing_cases`. **Visual:** not compared to a matched Rust runtime. **Platform:** native CPU only unless explicitly noted.

Source: `crates/orbit-pi/src/dialog.rs:171::Dialog`; `crates/orbit-pi/src/app/dialogs.rs:10::handle_extension_ui_request`; `crates/orbit-pi/src/app/dialogs.rs:205::open_dialog`; `crates/orbit-pi/src/app/dialogs.rs:236::respond_to_dialog`; `crates/orbit-pi/src/dialog.rs:127::ApprovalRequest`; `crates/orbit-pi/src/app/dialogs.rs:260::open_approval`; `crates/orbit-pi/src/app/view.rs:4379::approval_bar`; `crates/orbit-pi/src/access.rs:?::source`; `contrib/orbit-guard-extension/policy.js:?::source`; `crates/orbit-pi/src/workflow.rs:?::source`; `crates/orbit-pi/src/app/composer_ops.rs:891::set_workflow_mode`; `contrib/orbit-workflow-extension/index.js:?::source`; `contrib/orbit-workflow-extension/policy.js:?::source`; `crates/orbit-pi/src/ask.rs:70::AskPrompt`; `crates/orbit-pi/src/ask.rs:140::AskReplayState`; `crates/orbit-pi/src/app/ask.rs:?::source`; `crates/orbit-pi/src/app/view.rs:3982::ask_panel`; `crates/orbit-pi/src/transcript_view.rs:2928::render_ask_card`; `crates/orbit-pi/src/widgets.rs:?::source`; `crates/orbit-pi/src/custom_ui.rs:?::source`; `crates/orbit-pi/src/app/dialogs.rs:76::apply_extension_widget`; `crates/orbit-pi/src/app/dialogs.rs:114::apply_custom_frame`.

1. Exercise select/confirm/input/editor and custom ANSI render.
2. Exercise Allow once/Always/Deny guard, Plan/Build/Ask and multi-question ask_user_question.
3. Set/replace/clear widgets and editor text.

Expected: Exact IDs/response types, scoped keys, no fake unsupported capability and source specialized panels.

Executed test coverage: `TestCustomDialogDoesNotDismissEscape`, `TestLiveCustomDialogForwardsEscape`.

Remaining: General modal and ANSI custom bridge exist. Structured questionnaire, specialized inline approval/replay, widget placement and unavailable workflow/access presentation are not fully wired. No full extension parity claim.

### UI-P01 · Settings categories, search and scroll (P01)

**Functional verdict:** `partial`. **Visual:** not compared to a matched Rust runtime. **Platform:** native CPU only unless explicitly noted.

Source: `crates/orbit-pi/src/app.rs:2679::SettingsSection`; `crates/orbit-pi/src/app/settings.rs:50::SETTINGS_SECTIONS`; `crates/orbit-pi/src/app/settings.rs:116::SETTINGS_SEARCH_INDEX`.

1. Visit all14 source sections.
2. Scroll a long page, switch away for8 frames, return.
3. Search a row label/synonym and activate result.

Expected: Source section order and grouped rows; retained page scroll; translated setting-label/synonym search opens its section and clears query.

Executed test coverage: `TestSettingsNavigationComplete`, `TestSettingsPageScrollRetained`, `TestSettingsSourceSearchIndexAndSectionNavigation`, `TestSettingsSourceHeaderAndGeneralRows`.

Remaining: All14 source sections remain; nav groups/28px rows, fixed16px title,960px outer/912px inner column and source55-row setting-label/synonym search are implemented. Source search clears query and opens a section, not exact-row scrolling. Retained scroll, source-header/search, and actual notification-switch pointer/Space persistence pass under the canonical overlay (-race focused suite3.452s). General CPU capture1470×922 was inspected at05:45; source GPUI phi line-height,1.3 heading, left-aligned grouped nav and native36×20 switch geometry are now explicit. Full Settings cards, notification OS permission state/sound preview, matched pixels and OS flows remain unverified.

### UI-P02 · Runtime diagnostics and agent settings (P02, P03, P04)

**Functional verdict:** `partial`. **Visual:** not compared to a matched Rust runtime. **Platform:** native CPU only unless explicitly noted.

Source: `crates/orbit-pi/src/app/settings.rs:4931::runtime_rows`; `crates/orbit-pi/src/app/runtime.rs:488::runtime_start`; `crates/orbit-pi/src/app/runtime.rs:518::runtime_stop`; `crates/orbit-pi/src/app/runtime.rs:532::runtime_restart`; `crates/orbit-pi/src/onboarding.rs:?::source`; `crates/orbit-pi/src/pi_install.rs:?::source`; `crates/orbit-pi/src/app/view.rs:2664::render_onboarding`; `crates/orbit-pi/src/app/view.rs:2903::render_pi_migration`; `crates/orbit-pi/src/session_defaults.rs:?::source`; `crates/orbit-pi/src/auto_title.rs:?::source`; `crates/orbit-pi/src/app/settings.rs:5196::agent_rows`; `crates/orbit-pi/src/app/runtime.rs:217::apply_default_model`; `contrib/orbit-title-extension/index.js:?::source`.

1. Inspect active/parked PID/executable/uptime/stdout/stderr.
2. Start/Stop/Restart; dependency install failure/cancel.
3. Change defaults, auto-retry/compaction/title model; restart/resume.

Expected: Only real process fields, safe busy gates and source fresh-session-only defaults.

Executed test coverage: `TestMaintenanceGateKeepsDraftAndBlocksNewProcess`, `TestNativeRuntimeUsesRealSnapshotAndStoppedUnknowns`, `TestNativeAgentRetryIsUnknownUntilConfirmed`.

Remaining: Authoritative PID/executable/stderr/parked-processes/alive-only uptime and Start/Stop/Restart are mounted; retry/compaction unknown states are not invented and ACKs govern values. Source Runtime/Agent focused tests pass. Runtime patch enable/disable control and complete install/update/onboarding OS flows remain open.

### UI-P03 · Provider catalog, typed editor and credentials (P05, P06)

**Functional verdict:** `pass_scoped`. **Visual:** not compared to a matched Rust runtime. **Platform:** native CPU only unless explicitly noted.

Source: `crates/orbit-pi/src/providers.rs:?::source`; `crates/orbit-pi/src/app/settings.rs:2391::provider_rows`; `crates/orbit-pi/src/app/settings.rs:7639::ensure_provider_metadata`; `crates/orbit-pi/src/app/settings.rs:7665::provider_refresh`; `crates/orbit-pi/src/providers.rs:?::source`; `crates/orbit-pi/src/app/settings.rs:7772::provider_key_save`; `crates/orbit-pi/src/app/settings.rs:8027::provider_save`; `crates/orbit-pi/src/app/settings.rs:8103::provider_remove`.

1. Inspect catalog including fallback diagnostic and OAuth-only provider.
2. Save invalid custom provider; correct it and save into temporary home.
3. Force disk failure; cancel masked key dialog; resize800×560.

Expected: Capability-specific controls, validated fields, no painted key, failed save keeps draft, existing key preserved when blank.

Executed test coverage: `TestNativeProviderEditorValidationSecretAndMinimumSize`, `TestNativeProviderSaveFailureKeepsDraft`, `TestProviderCredentialCapabilities`.

Remaining: These native/disk fixtures pass. Full provider card geometry, source account detail/quota embed and real credential flows remain unverified.

### UI-P04 · OAuth cancellation and quota details (P07, P09)

**Functional verdict:** `partial`. **Visual:** not compared to a matched Rust runtime. **Platform:** native CPU only unless explicitly noted.

Source: `crates/orbit-pi/src/auth.rs:140::AuthManager`; `crates/orbit-pi/src/app/runtime.rs:395::handle_auth_effects`; `crates/orbit-pi/src/app/settings.rs:7805::provider_oauth_login`; `crates/orbit-pi/src/platform.rs:957::open_terminal_command`; `crates/orbit-pi/src/quota.rs:?::source`; `crates/orbit-pi/src/quota_hidden.rs:?::source`; `crates/orbit-pi/src/app/session.rs:?::source`; `crates/orbit-pi/src/app/runtime.rs:317::poll_quota_entries`; `crates/orbit-pi/src/app/mcp_ui.rs:2447::QuotaPopupTab`.

1. Start sign-in, cancel, then receive stale completion.
2. Use active-provider5h/peak/balance/global fallbacks and failed reports.
3. Inspect refresh/reset and hidden-provider preferences.

Expected: No stale sign-in overwrites later state; quota priority/unknown handling source-equivalent.

Executed test coverage: `TestQuotaHeadlinePriorityAndErrorExclusion`, `TestQuotaFractionsAndWindowRecognition`, `TestQuotaResetHintAndAmount`, `TestQuotaReportsAreSortedWithoutChangingInput`.

Remaining: Pure quota rules pass; live OAuth, full quota popup grouping/hidden-provider/MCP tabs and OS browser-return lifecycle remain unverified or incomplete.

### UI-P05 · Skills and plugin resources (P10, P11, P12, P13)

**Functional verdict:** `partial`. **Visual:** not compared to a matched Rust runtime. **Platform:** native CPU only unless explicitly noted.

Source: `crates/orbit-pi/src/skills.rs:88::discover`; `crates/orbit-pi/src/skills.rs:52::Skill`; `crates/orbit-pi/src/app/skills_ui.rs:32::render_skills_page`; `crates/orbit-pi/src/app/skills_ui.rs:722::apply_skill_action`; `crates/orbit-pi/src/skills.rs:338::set_enabled`; `crates/orbit-pi/src/skills.rs:440::delete`; `crates/orbit-pi/src/app/skills_ui.rs:772::delete_selected_skill`; `crates/orbit-pi/src/plugins.rs:99::discover`; `crates/orbit-pi/src/plugins.rs:135::install`; `crates/orbit-pi/src/plugins.rs:147::remove`; `crates/orbit-pi/src/plugins.rs:157::update`; `crates/orbit-pi/src/app/settings.rs:7478::plugin_run`; `crates/orbit-pi/src/plugins.rs:165::find_updates`; `crates/orbit-pi/src/app/plugin_update_ui.rs:19::start_plugin_update_check`; `crates/orbit-pi/src/app/plugin_update_ui.rs:52::skip_plugin_updates`; `crates/orbit-pi/src/app/plugin_update_ui.rs:69::apply_plugin_updates`.

1. Search/select Skills, inspect frontmatter-stripped detail, enable/disable/delete with cancellation.
2. Install by global/project scope; double-click operation, cancel, change workspace while pending.
3. Check/skip/clear/update actual package offers.

Expected: Source master/detail, scope/mutation serialization, stale completion rejection and precise persistence.

Executed test coverage: No host test is claimed for this complete scenario..

Remaining: Dedicated resourceui source master-detail and CRUD/update/skip lifecycle is mounted, including Offline flag, default-app open, Notify completion and global update notice below toasts.24 scoped package tests/race pass. Host/platform full flow and matched pixels remain unverified; see resource module report.

### UI-P06 · Typed MCP lifecycle (P14, P15, P16, P17)

**Functional verdict:** `partial`. **Visual:** not compared to a matched Rust runtime. **Platform:** native CPU only unless explicitly noted.

Source: `crates/orbit-pi/src/mcp.rs:?::source`; `crates/orbit-pi/src/mcp/manager.rs:74::McpManager`; `crates/orbit-pi/src/app/mcp_ui.rs:1135::mcp_rows`; `crates/orbit-pi/src/app/mcp_ui.rs:1272::mcp_handle_key`; `crates/orbit-pi/src/mcp.rs:160::McpTransport`; `crates/orbit-pi/src/mcp.rs:72::McpExposure`; `crates/orbit-pi/src/mcp/config.rs:?::source`; `crates/orbit-pi/src/app/mcp_ui.rs:523::mcp_editor_save`; `crates/orbit-pi/src/mcp/secrets.rs:?::source`; `crates/orbit-pi/src/mcp/config.rs:?::merge_secret_lines`; `crates/orbit-pi/src/mcp/manager.rs:?::source`; `crates/orbit-pi/src/app/mcp_ui.rs:797::mcp_sign_in`; `crates/orbit-pi/src/app/mcp_ui.rs:866::mcp_sign_out`; `crates/orbit-pi/src/app/mcp_ui.rs:90::tick_mcp`; `crates/orbit-pi/src/app/mcp_ui.rs:277::mcp_after_settle`; `crates/orbit-pi/src/app/mcp_ui.rs:301::mcp_force_apply`; `crates/orbit-pi/src/app.rs:150::ParkedSession`.

1. Create/edit stdio andHTTP servers acrossglobal/project scope.
2. Validate timeout, exposure and secret references; cancel OAuth/remove.
3. Apply while task busy; verify deferred session-preserving reload.

Expected: Source typed fields, redacted status, scope-aware metadata/secrets and truthful requested-vs-applied state.

Executed test coverage: No host test is claimed for this complete scenario..

Remaining: mcpui/workbench scoped race tests and19-case checklist pass. Host mount/deferred reload exists. Real OAuth and authoritative applied/readiness callback are not verified; Applied is deliberately not called on async restart acceptance.

### UI-P07 · Appearance, localization and help (P18, P19, P20)

**Functional verdict:** `partial`. **Visual:** not compared to a matched Rust runtime. **Platform:** native CPU only unless explicitly noted.

Source: `crates/orbit-pi/src/theme.rs:324::UiPrefs`; `crates/orbit-pi/src/theme.rs:482::FontPrefs`; `crates/orbit-pi/src/theme/appearance.rs:31::AppearancePrefs`; `crates/orbit-pi/src/app/settings.rs:5926::appearance_rows`; `crates/orbit-pi/src/assets.rs:?::source`; `crates/orbit-pi/src/backdrop.rs:?::source`; `crates/orbit-pi/src/dither.rs:?::source`; `crates/orbit-pi/src/app/settings.rs:6382::background_controls`; `crates/orbit-pi/src/app/settings.rs:6569::background_choose`; `crates/orbit-pi/src/app/view.rs:2271::new_task_backdrop`; `crates/orbit-pi/src/i18n.rs:?::source`; `crates/orbit-pi/src/main.rs:626::set_app_menus`; `crates/orbit-pi/src/theme.rs:1468::set_ui_prefs`; `crates/orbit-pi/src/app/settings.rs:4672::shortcut_rows`; `crates/orbit-pi/src/commands.rs:?::source`.

1. Change every theme/font/density/motion choice and relaunch.
2. Visit all sections in each of10 locales.
3. Configure animated/photo backdrop and inspect keyboard reference.

Expected: Source palette/tokens and persistent preferences, complete localization and backdrop.

Executed test coverage: `TestTokensAndCommandRegistry`, `TestFractionalScaleKeepsLogicalComposerGeometry`, `TestNativeBackdropBoundsAndFeatureVisibility`, `TestBackgroundChooseCoalesceFadeReuseFailureAndClear`, `TestBackdropPaintsBehindOpaqueComposer`.

Remaining: Source procedural/photo background is mounted below header only on empty task, with source choose/reset/blur/cell/fade controls and worker-only PhotoStore. Native tests cover geometry, pointer pass-through, opaque foreground occlusion, coalescing/cache reuse and failure retention. Runtime/About use truthful versions. Full locale coverage/whole Appearance content/OS dialogs/matched visual acceptance remain open.

### UI-U01 · Actual usage accounting and unavailable data (U01)

**Functional verdict:** `pass_scoped`. **Visual:** not compared to a matched Rust runtime. **Platform:** native CPU only unless explicitly noted.

Source: `crates/orbit-pi/src/usage/collect.rs:?::source`; `crates/orbit-pi/src/usage/model.rs:?::source`; `crates/orbit-pi/src/usage/aggregate.rs:?::source`.

1. Scan explicit temporary Pi JSONL fixtures, including unbilled provider failures, pending tool calls and compaction.
2. Filter model/provider and request-only flags.
3. Compare cost coverage and absent duration/cache against source.

Expected: No fabricatedzero metrics, no pending-tool failure, real call-start attribution and source turn semantics.

Executed test coverage: `TestUsageProjectionPreservesScopeCalendarAndMissingMetrics`, `TestSettingsBackendProjectionUsesModelPricingCoverage`, `TestSettingsBackendNativeUnbilledErrorsRemainVisible`, `TestLatencyPercentileRequiresSourceMinimum`.

Remaining: Focused domain/projection cases pass; this does not close the complete analytics corpus or all UI displays.

### UI-U02 · Usage Simple/Details loading/error/empty (U02)

**Functional verdict:** `partial`. **Visual:** not compared to a matched Rust runtime. **Platform:** native CPU only unless explicitly noted.

Source: `crates/orbit-pi/src/usage/view.rs:?::source`; `crates/orbit-pi/src/usage/chart.rs:?::source`; `crates/orbit-pi/src/usage/heatmap.rs:?::source`; `crates/orbit-pi/src/usage/table.rs:?::source`; `crates/orbit-pi/src/usage/page.rs:876::UsagePage`.

1. Render loading then actual data; change mode.
2. Fail refresh after prior good data; render only unbilled errors.
3. Inspect fixed-date labelled fixture screenshots.

Expected: Source modes and honest pending/partial/error/empty states, no invented live metrics.

Executed test coverage: `TestUsageNativeLoadedEmptyErrorAndModal`, `TestNativeUsageCapture`.

Remaining: Native cards/heatmap/chart/breakdowns/session/failure rows render. Source comparison deltas, signals, detailed cache/latency panels and exact layout remain missing; screenshots are fixtures, not Rust comparisons.

### UI-U03 · Date and scoped drill-down (U03)

**Functional verdict:** `partial`. **Visual:** not compared to a matched Rust runtime. **Platform:** native CPU only unless explicitly noted.

Source: `crates/orbit-pi/src/usage/filters.rs:?::source`; `crates/orbit-pi/src/usage/model.rs:675::UsageFilter`; `crates/orbit-pi/src/usage/page.rs:1580::set_filter`; `crates/orbit-pi/src/usage/page.rs:1758::scope_to_day`.

1. Select multiple workspaces/providers/models.
2. Select calendar day outside prior date range; bucket/session drill-down; clear focus.
3. Use reversed custom dates; choose range with a quiet middle day.

Expected: Calendar independent of date filter, valid scope remains, date normalization and contiguous zero-activity axis.

Executed test coverage: `TestSettingsBackendTrendDoesNotCompressMissingDays`, `TestUsageProjectionPreservesScopeCalendarAndMissingMetrics`, `TestUsageNativeMinimumWindowControls`.

Remaining: Multi-select menus, scope fields, calendar normalization and contiguous4k guarded buckets implemented. Searchable native multi-select popup and disambiguating identical model IDs acrossproviders remain incomplete.

### UI-U04 · Tables, detail dialogs and export (U04)

**Functional verdict:** `partial`. **Visual:** not compared to a matched Rust runtime. **Platform:** native CPU only unless explicitly noted.

Source: `crates/orbit-pi/src/usage/page.rs:79::query_sessions`; `crates/orbit-pi/src/usage/page.rs:2457::export`; `crates/orbit-pi/src/usage/page.rs:2528::Prefs`; `crates/orbit-pi/src/usage/table.rs:?::source`.

1. Sort equal-metric rows both ways; paginate; open source session.
2. Open failure detail; Escape and native menu guard.
3. Choose CSV/JSON, cancel native Save, then export to disposablefile.

Expected: Stable ties, no phantom detail metrics, actual session path, source export schema and unknown fields preserved.

Executed test coverage: `TestUsageSortTiesAreStable`, `TestUsageNativeLoadedEmptyErrorAndModal`, `TestSettingsBackendUsagePreferencesPreserveUnknownAndDefaultDescending`.

Remaining: Stable sorting/modal/prefs cases pass; JSON/CSV backend implemented and host chooser wired. Independent table states, column visibility/resize/context menus and actual native Save interaction remain open.

### UI-U05 · Maintenance interrupts and retention (U05, U06, U07, U10)

**Functional verdict:** `partial`. **Visual:** not compared to a matched Rust runtime. **Platform:** native CPU only unless explicitly noted.

Source: `crates/orbit-pi/src/notifications.rs:?::source`; `crates/orbit-pi/src/toast.rs:?::source`; `crates/orbit-pi/src/app/events.rs:1193::post_notification`; `crates/orbit-pi/src/app/toast_ui.rs:?::source`; `crates/orbit-pi/src/app/session.rs:1215::activate_session_from_notification`; `crates/orbit-pi/src/session_retention.rs:?::source`; `crates/orbit-pi/src/app/retention_ui.rs:?::source`; `crates/orbit-pi/src/app/settings.rs:5618::privacy_rows`; `crates/orbit-pi/src/temp_files.rs:?::source`; `crates/orbit-pi/src/app/temp_ui.rs:?::source`; `crates/orbit-pi/src/app/settings.rs:5618::privacy_rows`; `crates/orbit-pi/src/app/pi_update_ui.rs:30::check_pi_update_on_launch`; `crates/orbit-pi/src/app/pi_update_ui.rs:139::finish_pi_update`; `crates/orbit-pi/src/pi_install.rs:?::source`; `crates/orbit-pi/src/pi_update.rs:?::source`.

1. Hold runtime maintenance gate then send/open/change workspace.
2. Run retention on temporary sessions, including active/parked ones.
3. Cancel temp cleanup/install confirmations and inspect errors.

Expected: No prompt loss/new process during exclusive update; live sessions protected; explicit cleanup and honest progress.

Executed test coverage: `TestMaintenanceGateKeepsDraftAndBlocksNewProcess`, `TestRetentionRunsOffThreadAndReloadsArchiveSet`.

Remaining: Guard and temporary retention tests pass. Whole native update/install/temp dialogs and real OS notification lifecycle remain platform acceptance gates.

### UI-W01 · Review refresh minimum feedback (W39)

**Functional verdict:** `pass_scoped`. **Visual:** not compared to a matched Rust runtime. **Platform:** native CPU only unless explicitly noted.

Source: `crates/orbit-pi/src/sidepane.rs:749::refresh_from_button`.

1. Click Refresh against a fast disposable repository.
2. Complete content reload, then click again before the first feedback deadline.

Expected: Independent click feedback lasts at least650ms; repeated clicks extend it without fabricating loading data.

Executed test coverage: `TestReviewRefreshKeepsSource650msFeedback`.

Remaining: Initial acceptance failed at25.821182ms. Corrected with independent deadline and actual rendering predicate; default native regression now passes. Full Review animation/pixels remain open.

