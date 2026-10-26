# Own-menu tracking feasibility only

Unwired helper: no Orbit import, workflow, entrypoint or native execution. Implementation is now based on verified Provider source ee25813f9a2065ae8632d4c15b24e3c99e7bf469, tree0476116687acf0838d52a0246210270000530530. Original feasibility was built against e5c9fc; those results are not final native evidence.

## Evidence and intended mechanism

Pinned MyGo `internal/darwin/menu.go:191–228` builds the popup and synchronously calls public `popUpMenuPositioningItem:atLocation:inView:`. Waiting for a normal main-thread work item while that call is active may not prove native-menu routing. This candidate instead registers an independent observer (not a delegate replacement) for public NSMenu tracking/action notifications. Only a previously armed exact owner/window, one finite in-content target, bounded deadline, and exact four visible menu titles authorizes one Escape key-down/up queued to NSApplication from the begin notification. The end notification and zero action notifications are separately required; construction/posting is not success.

Apple documentation: [begin tracking](https://developer.apple.com/documentation/appkit/nsmenu/didbegintrackingnotification), [end tracking](https://developer.apple.com/documentation/appkit/nsmenu/didendtrackingnotification). The latter documents notification even without a selected action; notification object is the root popup/main menu. The real popup must be the expected root in the owned process. Constants are loaded as public exported AppKit symbols, not guessed strings.

No `cancelTracking`, `performAction`, delegate replacement, global event tap/injection, manual event-loop pump, private toolkit import or permission change. Native callbacks/methods require the existing App main thread. Observer must be registered before input, closed at lifetime end, and never rearmed to retry. The singleton map retains at most one live observer and removes it on Close. Current-process identity is supplied by a future read-only actual-App observer; no caller yet exists. The separate test-only overlay adds app readiness and bounded unique AX lookup; the process-proof CLI reads real start time and executable. None has executed on macOS. Begin additionally requires the exact fresh Add click receipt; any other popup first fails closed.

## Limits / result

Pure policy tests cover exact one Escape + actual end requirement; duplicate/zero target, NaN/Inf/outside geometry, PID/start/window mismatch, hidden/non-key window, early/late event, wrong labels/menu, action dispatch, duplicate begin and closed lifetime. These tests do not execute AppKit. Darwin ARM64 compilation establishes type/link feasibility only; actual notification delivery, correct popup root, Escape consumption by native tracking loop and focus return remain NOT_RUN. If those fail, result stays BLOCKED; no alternative cancellation selector is permitted.

This layer has not changed production, package build or release workflows. The later four-case clipboard/menu harness remains a separate reviewed implementation and admission task.
