# Composer mode admission feedback correction

Additive to frozen `a25f90551ec3eef66a41aaaee11defbb8ce937bbcb9e12a27b99d993c4071f7f`; the original patch and tree remain unchanged.

Independent review correctly identified that removing the pre-admission Access disk write was insufficient: local `a.access` / `a.workflow` still projected a choice before Controller admission. Actual mounted temporary-home tests reproduced both maintenance-rejected choices and an atomic Access write failure displaying the rejected value. Evidence: `composer-mode-admission-red-2147.jsonl` (10.440 s, expected failures).

The production correction removes exactly those two optimistic assignments. Closing the picker, focus restoration, validated enum dispatch, original opaque Workflow target, Controller admission and global Access semantics are unchanged. The existing `App.SetSnapshot` is the only path projecting a chosen backend mode. A successful operation may therefore retain the old chip until the next normal snapshot refresh; a refused operation no longer briefly advertises success. No new acknowledgment protocol or setter behavior is introduced.

Tests use only disposable homes and synthetic Pi children. They inspect real mounted chips immediately after maintenance rejection, after atomic Access write failure, and after existing snapshot synchronization. Positive tests require actual stored backend success before the new visible chip. Additional cases cover asynchronous stale Workflow-target rejection and legal pre-hydration empty-ID pending choice, including persistence on hydration. The filesystem failure is induced with a directory at the temporary Access destination, not permission or security weakening.

The final manifest records exact JSON run/pass hitsets and vet. This is focused synthetic evidence, not physical macOS, OS accessibility, or complete migration acceptance. It supersedes the original document's limitation allowing brief optimistic local mode projection.

The first combined gate (`composer-mode-admission-final-2150.jsonl`) exposed an old recording-only Chinese test that expected new English mode labels without supplying any backend snapshot. That fixture now first verifies the old Chinese labels survive recording-only dispatch, then explicitly supplies the confirmed snapshot before retaining its original English reload assertions. The real persistence/confirmation proof remains in the new Controller tests, not in the recording fixture. The failed gate is retained rather than reclassified as passed.
