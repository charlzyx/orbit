# General attachment picker: restore Rust single-selection interaction

Pinned Rust composer_ops.rs:471–496 calls AsyncFileDialog.pick_file (one file).
The only live Go native.BrowseAttachments caller is Controller's general composer
file dialog. There is no separate multi-selection dependency or string route.
Visible user-message searches for 多选/multiple/attachment found no requirement;
the attachment request on 2026-10-18 asked for drop, pasted images and attachment
upload, not a multi-select chooser. Existing repository references described the
Go implementation, not a product requirement overriding the Rust migration.

The native general chooser now uses single selection; owner completion applies
only its first result, matching Rust's single handle even if a backend returns
extra entries. Cancel, errors, singleton/task/window guards and generic-file
read-error path fallback are unchanged. Actual multi-file drop remains supported;
image-only chooser/filter and clipboard behavior are unchanged. This is restoring
source interaction, not labeling the earlier multi-select feature a vulnerability.

RED used native options and actual Controller completion with a mounted composer:
old Multiple=true and two returned paths inserted both. Cancel/error controls and
actual multi-file drop stay covered. No physical OS dialog was opened; mounted
fixtures do not establish native Mac confirmation or keyboard acceptance.
