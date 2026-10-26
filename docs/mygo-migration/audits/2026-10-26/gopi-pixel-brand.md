# User-approved GOPI / pixel PI identity (bounded visual change)

Base commit `5dbc486e33d6761fbd11ea85378794cc8ff1fa3b`, tree `05882441c24ad879aac0b94b68633f31b5cad9f0` (same public source tree). The user rejected the ordinary bold-text candidate and generated PNGs; those are not included. This is one repo-native vector proposal, submitted for review before integration/publication.

## Artwork and scope

- The original embedded `logo.png` / `logo-dark.png` and original Orbit/Alpha app icon were inspected. No exact source font or editable wordmark vector was found. New GOPI paths are original square, heavy-stem lettering inspired by that visual direction, not a claim to reproduce an identified font.
- `crates/orbit-pi/assets/icons/gopi-wordmark.svg` contains four explicit glyph paths, square O/P counters and a narrow I. It uses existing SVG loading/native rendering, currentColor and one low-opacity edge stroke; no text/font, bitmap, filter, external image or new dependency.
- Sidebar uses its original 76px outer width, x/y, and separate light/dark original height slots. Contained aspect scaling keeps the original cap-height scale rather than stretching four letters across seven letters' width. The old yellow GO badge is removed. No controls below the wordmark are moved.
- Go-only `assets/icons/go/orbit-go.svg` uses the same P/I geometry in a stepped charcoal pixel tile with silver ink. It replaces the former cyan OP/yellow-GO design. Existing source filename is retained for packaging compatibility; this is not a package or repository rename.
- Existing `scripts/make-go-icon.py` deterministically regenerates the 512 PNG and all 11 ICNS PNG slots using installed librsvg/cairo. Only its manifest description changes. No Python cropping, old-PNG repaint, image generation, service, or renderer change.
- App icon and About intentionally use the regenerated Go-only asset. Original Rust/Alpha/marketing icons, shared original wordmarks, CFBundleIdentifier, window/package names, configuration/data directories and version are unchanged. No release, main merge or recovery-test authorization is implied.

## Bounded evidence

Actual mounted CPU captures: `evidence/orbit-gopi-pixel-evidence-2057/{sidebar,about}-{light,dark}.png`. Source-rendered actual-size icon previews: `icon-{16,32,128,1024}.png` in that directory. They were viewed at actual rendered sizes; PI remains distinguishable, counters/spacing remain open and no yellow badge remains. These are CPU/native component previews, not OS/Finder/Dock screenshots.

Focused race tests retain the existing sidebar width/top and New Task geometry checks, setup/sidebar ownership checks, About icon footprint and separate asset identity checks. They add pixel assertions for the silver P and I stems and dark counter/gap, verify the embedded vector dimensions and absence of the retired badge. Exact JSON run/pass names and final times are in the delivered manifest. UI vet and the existing five `GoIconTests` cover all slot dimensions, deterministic byte reproduction, tamper rejection and temporary bundle resource/provenance bytes without changing BundleID. No new full suite or recovery tests were run.

This visual scope ends after review; further work returns to functional parity rather than indefinite logo iteration.
