# Native invisible-content accessibility probe

`visibility_access_test.go.txt` is an additive test for the pinned MyGo
`ui` package (9c69cf829ece196c5b3a3601f1c7bbe70526376f). Copy the verified
source to a disposable writable directory, copy this fixture to
`ui/orbit_visibility_access_test.go`, then run:

```
go test -race ./ui -run '^TestOrbitRetainedInvisibleInputIsolation$' -count=1
```

Do not edit the module cache. The app's canonical wrapper and production
sources do not change. The observed run on 2026-10-25 at08:24UTC passed in
1.359seconds. The fixture inspects the actual CPU accessibility tree, hidden
input/copy-paste/key isolation, idle frame requests and undo on reappearance.
It does not execute an OS screen reader or establish matched screenshot
parity.
