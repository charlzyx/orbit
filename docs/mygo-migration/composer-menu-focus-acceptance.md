# Composer add-menu return focus

Rust `composer_ops.rs:537–543` restores composer focus after add-menu dismissal;
`586–606` does so after every action, not only after a successful attachment.
Mounted native-menu tests proved Go lost subsequent typing after Cancel and
Attach image / Attach files / Paste image selections.

The isolated correction requests composer focus in the existing public Menu
builder. MyGo invokes that builder only when opening a menu or adopting its
chosen item; there is no public cancellation callback. Setting the return target
on open therefore also handles dismissal without an action. The owner-keyed
menu still rejects choices from obsolete tasks, and an absent composer does
not execute this builder. No toolkit or OS focus API is changed.

Source keyboard routing check: MyGo `ui/window.go:785` delegates to the platform
PopupMenu; Darwin `internal/darwin/menu.go:191–229` creates a real NSMenu and
uses `popUpMenuPositioningItem:atLocation:inView:`. The application Handle.Focus
changes its internal editor return target, not that native menu's tracking.
Arrow/Enter/Escape handling during actual platform menu tracking is therefore
unchanged by this one-line edit. Physical OS keyboard/menu interaction is NOT
accepted by the mounted tester and remains a separate native validation gate.

Tests cover real public menu cancel and three action choices followed by typed
text, plus a stale menu choice after task generation change while an unrelated
page input is focused. Existing mention/caret/IME/image-picker tests are rerun.
