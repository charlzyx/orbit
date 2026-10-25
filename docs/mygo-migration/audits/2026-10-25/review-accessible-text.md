# Review accessible-text corrective delta

Follow-on to b7726e4. Removed the diagnostic ID Label from diff RichText: MyGo uses that label instead of the text's readable accessibility name. Native painting was unchanged, but test-only geometry IDs must not replace actual diff content.

Mounted regression was red on b7726e4 (`review-a11y-red.log`): `HasText("file0 line1")` failed and Texts contained the internal ID instead. After removing the override, geometry/pan tests find the real content string. Original semantic expectations were not weakened. New regression verifies the readable text in wrapped and unwrapped modes.

Canonical focused race test of accessible text + long-line/pan/width/scrollbar/sticky/push/held-press cases: PASS,54.562s (`/workspace/shared/orbit-toolchains-continued/review-a11y-race.log`). The parent owns a fresh full-UI/full-repository aggregate on its final integrated SHA; this focused delta is not that aggregate or pixel parity. No release/main writes.
