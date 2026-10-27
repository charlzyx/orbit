package bugreportui

import (
	"fmt"
	"math"
	"strings"
	"sync"

	orbitassets "github.com/charlzyx/orbit"
	"github.com/charlzyx/orbit/internal/native"
	"github.com/egoist/mygo/ui"
)

// GPUI's default TextStyle is phi(), not MyGo's default text line height.
const sourceLineHeight float32 = 1.618034
const contentMaxWidth float32 = 960

type appearance struct {
	Theme
	native *ui.Theme
}

func (s *State) appearance(c *ui.Context) appearance {
	a := appearance{native: c.Theme()}
	if s.Options.Theme != nil {
		a.Theme = s.Options.Theme()
	}
	if a.Px == nil {
		a.Px = func(n float32) float32 { return n * a.native.FontSize / 14 }
	}
	if a.Space == nil {
		a.Space = func(n int) float32 { return a.Px(float32(n)) }
	}
	if a.SurfaceSecondary == (ui.Color{}) {
		a.SurfaceSecondary = a.native.Surface
	}
	if a.Elevated == (ui.Color{}) {
		a.Elevated = a.native.Surface
	}
	if a.TextSecondary == (ui.Color{}) {
		a.TextSecondary = a.native.TextMuted
	}
	if a.Action == (ui.Color{}) {
		a.Action = a.native.Accent
	}
	if a.ActionHover == (ui.Color{}) {
		a.ActionHover = a.Action
	}
	if a.ActionPressed == (ui.Color{}) {
		a.ActionPressed = a.Action
	}
	if a.ActionText == (ui.Color{}) {
		a.ActionText = a.native.AccentText
	}
	return a
}

var iconOnce sync.Once
var reportIcons map[string]*ui.SVG

func icon(c *ui.Context, name string, size float32, color ui.Color) {
	iconOnce.Do(func() {
		reportIcons = map[string]*ui.SVG{}
		for _, name := range []string{"github", "spark", "image", "x", "settings"} {
			if data, err := orbitassets.Asset("icons/" + name + ".svg"); err == nil {
				if svg, err := ui.ParseSVG(data); err == nil {
					reportIcons[name] = svg
				}
			}
		}
	})
	if svg := reportIcons[name]; svg != nil {
		ui.Icon(c, svg).Size(size, size).Shrink(0).TextColor(color)
	}
}

func (s *State) button(c *ui.Context, key, label, glyph string, primary, disabled bool) ui.Element {
	a := s.appearance(c)
	b := ui.ButtonBase(c.Key(key)).Label(label).Disabled(disabled).Height(a.Px(28)).Shrink(0).Padding(0, a.Space(8)).Gap(a.Space(4)).Radius(a.Px(8)).Background(a.Elevated).Border(1, a.native.Border).AlignItems(ui.Center)
	color, fill := a.native.Text, a.Elevated
	if primary {
		fill = a.Action
		color = a.ActionText
		b.BorderColor(ui.Transparent)
	}
	if b.Hovered() && !disabled {
		if primary {
			fill = a.ActionHover
		} else {
			fill = a.native.SurfaceHover
		}
	}
	if b.Pressed() && !disabled {
		if primary {
			fill = a.ActionPressed
		} else {
			b.Opacity(.7)
		}
	}
	if b.Focused() {
		b.BorderColor(a.native.Focus)
	}
	if disabled {
		b.Background(a.native.Text.Alpha(.05)).Opacity(.5)
	} else {
		// Source tokens.rs::raised_fill and button_shadow: lift only the HSL
		// lightness, with short contact + ambient shadows outside the face.
		top, bottom := liftLightness(fill, .02), liftLightness(fill, -.012)
		if fill.A < 242 {
			top = top.Alpha(1.12)
			bottom = bottom.Alpha(.94)
		}
		b.Gradient(top, bottom, 180)
		contact := float32(.10)
		if a.native.Dark {
			contact = .32
		}
		b.Shadow(0, 1, 1, 0, ui.RGBA(0, 0, 0, contact*.45)).Shadow(0, 1, 3, -1, ui.RGBA(0, 0, 0, contact*.25))
	}
	b.LineHeight(1)
	b.Children(func() {
		if glyph != "" {
			icon(c, glyph, a.Px(14), color)
		}
		ui.Text(c, label).FontSize(a.Px(14)).FontWeight(500).TextColor(color).SingleLine()
	})
	return b
}

// liftLightness adjusts HSL lightness while retaining hue and saturation.
func liftLightness(c ui.Color, delta float32) ui.Color {
	r, g, b := float32(c.R)/255, float32(c.G)/255, float32(c.B)/255
	lo, hi := min(r, g, b), max(r, g, b)
	light := (hi + lo) / 2
	next := max(float32(0), min(float32(1), light+delta))
	ratio := float32(0)
	if hi > lo {
		ratio = (1 - float32(math.Abs(float64(2*next-1)))) / (1 - float32(math.Abs(float64(2*light-1))))
	}
	channel := func(v float32) uint8 {
		return uint8(max(float32(0), min(float32(255), ((v-light)*ratio+next)*255)) + .5)
	}
	return ui.RGBA(channel(r), channel(g), channel(b), float32(c.A)/255)
}

// View owns the same fixed settings header and independently scrolling body as
// settings.rs. The native host mounts it outside its generic page scroll.
func View(c *ui.Context, s *State) ui.Element {
	s.Tick()
	c = c.Key("bug-report")
	a := s.appearance(c)
	root := ui.Column(c).MinWidth(0).LineHeight(sourceLineHeight).Children(func() {
		ui.Column(c.Key("header")).WidthPercent(100).MaxWidth(contentMaxWidth).AlignSelf(ui.Center).Shrink(0).Children(func() {
			ui.Column(c).AlignSelf(ui.Stretch).Padding(a.Space(20), a.Space(24), a.Space(12)).Gap(a.Space(4)).Children(func() {
				ui.Text(c, s.label("bug_report.heading", "Report a bug")).Label("Settings page title").FontSize(a.Px(16)).LineHeight(1.3).Height(a.Px(16) * 1.3).FontWeight(500)
				// Source's automatic-attachment promise is intentionally not shown:
				// this fork retains explicit diagnostic and identity consent.
				ui.Text(c, s.label("bug_report.review_description", "File an issue on GitHub after reviewing what will be shared.")).FontSize(a.Px(12)).Height(a.Px(12) * sourceLineHeight).TextColor(a.native.TextMuted)
			})
			ui.Box(c).AlignSelf(ui.Stretch).Height(1).MarginX(a.Space(24)).Background(a.native.Border)
		})
		ui.Scroll(c.Key("body")).TrackScroll(&s.scroll).Grow(1).FillHeight().MinHeight(0).AlignSelf(ui.Stretch).Children(func() {
			ui.Column(c).WidthPercent(100).MaxWidth(contentMaxWidth).AlignSelf(ui.Center).Padding(a.Space(20), a.Space(24), a.Space(24)).Gap(a.Space(20)).Children(func() {
				s.form(c)
				if s.Error != nil {
					ui.Text(c, s.errorMessage(s.Error)).Label("Report error").AlignSelf(ui.Stretch).Padding(a.Space(8), a.Space(12)).Radius(a.Px(8)).Background(a.native.Danger.Alpha(.10)).TextColor(a.native.Danger).FontSize(a.Px(12)).Selectable()
				}
				if s.Status != "" {
					ui.Text(c, s.displayMessage(s.Status)).FontSize(a.Px(12)).TextColor(a.native.TextMuted).Selectable()
				}
				s.footer(c)
			})
		})
	})
	if s.reviewOpen {
		s.review(c)
	}
	if s.optionsOpen {
		s.privacy(c)
	}
	if s.preview >= 0 && s.preview < len(s.shots) {
		s.previewView(c)
	}
	if (s.HasModal() || s.Busy) && (root.Shortcut(0, ui.KeyEscape) || c.Shortcut(0, ui.KeyEscape)) {
		s.Escape()
	}
	return root
}

func (s *State) form(c *ui.Context) {
	a := s.appearance(c)
	ui.Column(c.Key("section")).AlignSelf(ui.Stretch).Gap(a.Space(8)).Children(func() {
		ui.Column(c).AlignSelf(ui.Stretch).PaddingX(a.Space(4)).Gap(a.Space(2)).Children(func() {
			ui.Text(c, strings.ToUpper(s.label("bug_report.heading", "Report a bug"))).FontSize(a.Px(10)).Height(a.Px(10) * sourceLineHeight).FontWeight(500).TextColor(a.native.TextMuted)
			ui.Text(c, s.label("bug_report.review_intro", "Describe the problem and review your issue before sharing it, with only the details you choose.")).FontSize(a.Px(12)).Height(a.Px(12) * sourceLineHeight).TextColor(a.native.TextMuted)
		})
		ui.Column(c.Key("board")).Label("Report form board").AlignSelf(ui.Stretch).Background(a.SurfaceSecondary).Border(1, a.native.Border).Radius(a.Px(12)).Children(func() {
			ui.Row(c.Key("issue-type-row")).AlignSelf(ui.Stretch).Padding(a.Space(12), a.Space(16)).Gap(a.Space(12)).AlignItems(ui.Center).Justify(ui.SpaceBetween).Children(func() {
				ui.Text(c, s.label("bug_report.issue_type", "Issue type")).FontSize(a.Px(14)).FontWeight(500)
				s.kindSelector(c)
			})
			s.divider(c)
			s.field(c, "context", s.label("bug_report.context", "Context"), &s.Notes, s.label("bug_report.context_placeholder", "Describe what you were doing and what went wrong. The model drafts the title and description from this."), 96, func() {
				ui.Row(c).Gap(a.Space(8)).AlignItems(ui.Center).Children(func() {
					if s.Options.ModelLabel != nil {
						if model := s.Options.ModelLabel(); model != "" && model != "…" {
							ui.Text(c, strings.ReplaceAll(s.label("bug_report.using_model", "Using %{model}"), "%{model}", model)).FontSize(a.Px(10)).TextColor(a.native.TextMuted).SingleLine()
						}
					}
					label := s.label("bug_report.generate", "Generate draft")
					if s.Busy && s.operation == "Drafting from your notes" {
						label = s.label("bug_report.generating", "Generating…")
					}
					if s.button(c, "generate", label, "spark", false, s.Busy).Clicked() {
						s.Generate()
					}
				})
			})
			s.divider(c)
			s.field(c, "title", s.label("bug_report.title", "Title"), &s.Title, s.label("bug_report.title_placeholder", "A one-line summary of the problem"), 0, nil)
			s.divider(c)
			what, steps := s.fieldLabels()
			s.field(c, "what", what, &s.What, s.label("bug_report.what_placeholder", "Describe the bug and what you expected instead."), 120, nil)
			s.divider(c)
			s.field(c, "steps", steps, &s.Steps, s.label("bug_report.steps_placeholder", "One step per line, if you have them."), 84, nil)
			s.divider(c)
			s.screenshots(c)
		})
	})
}

func (s *State) fieldLabels() (string, string) {
	switch s.Kind {
	case native.ReportFeature:
		return s.label("bug_report.what_problem", "What problem does this solve?"), s.label("bug_report.steps_solution", "Proposed solution")
	case native.ReportOther:
		return s.label("bug_report.what_details", "Details"), s.label("bug_report.steps_details", "Additional details")
	default:
		return s.label("bug_report.what", "What happened?"), s.label("bug_report.steps", "Steps to reproduce")
	}
}

func (s *State) kindSelector(c *ui.Context) {
	a := s.appearance(c)
	ui.Row(c.Key("kind-selector")).Label(s.label("bug_report.issue_type", "Issue type")).Shrink(0).Children(func() {
		for index, item := range []struct {
			kind       native.ReportKind
			key, label string
		}{{native.ReportBug, "type_bug", "Bug report"}, {native.ReportFeature, "type_feature", "Feature request"}, {native.ReportOther, "type_other", "Other"}} {
			label := s.label("bug_report."+item.key, item.label)
			b := ui.ButtonBase(c.Key(string(item.kind))).Label(label).Disabled(s.Busy).Height(a.Px(28)).LineHeight(1).PaddingX(a.Space(8)).Border(1, a.native.Border).Background(a.Elevated).AlignItems(ui.Center)
			if index == 0 {
				b.Radius(a.Px(8), 0, 0, a.Px(8))
			} else {
				b.BorderWidth(1, 1, 1, 0)
				if index == 2 {
					b.Radius(0, a.Px(8), a.Px(8), 0)
				}
			}
			color := a.TextSecondary
			if s.Kind == item.kind {
				b.Background(a.native.SurfacePressed)
				color = a.native.Text
			} else if b.Hovered() {
				b.Background(a.native.SurfaceHover)
			}
			if b.Focused() {
				b.BorderColor(a.native.Focus)
			}
			if s.Busy {
				b.Opacity(.5)
			}
			b.Children(func() { ui.Text(c, label).FontSize(a.Px(14)).TextColor(color).SingleLine() })
			if b.Clicked() {
				s.SelectKind(item.kind)
			}
		}
	})
}

func (s *State) divider(c *ui.Context) {
	a := s.appearance(c)
	ui.Box(c).AlignSelf(ui.Stretch).Height(1).MarginX(a.Space(16)).Background(a.native.Border)
}

func (s *State) field(c *ui.Context, key, label string, value *string, placeholder string, height float32, action func()) {
	a := s.appearance(c)
	ui.Column(c.Key("field-"+key)).AlignSelf(ui.Stretch).Padding(a.Space(12), a.Space(16)).Gap(a.Space(6)).Children(func() {
		ui.Row(c).AlignSelf(ui.Stretch).Gap(a.Space(8)).AlignItems(ui.Center).Justify(ui.SpaceBetween).Children(func() {
			ui.Text(c, label).FontSize(a.Px(12)).Height(a.Px(12) * sourceLineHeight).FontWeight(500)
			if action != nil {
				action()
			}
		})
		var input ui.Element
		if height == 0 {
			input = ui.TextInputBase(c.Key("input"), value).Height(max(a.Px(32), a.Px(14)*sourceLineHeight+a.Px(12)+2))
		} else {
			input = ui.TextAreaBase(c.Key("input"), value).Height(max(height, a.Px(32)))
		}
		input.Label(label).Placeholder(placeholder).AlignSelf(ui.Stretch).FontSize(a.Px(14)).LineHeight(sourceLineHeight).Padding(a.Px(6), a.Px(8)).Border(1, a.native.Border).Radius(a.Px(8)).Background(a.SurfaceSecondary).Disabled(s.Busy)
		if input.Focused() {
			input.BorderColor(a.native.Focus)
		}
	})
}

func (s *State) screenshots(c *ui.Context) {
	a := s.appearance(c)
	ui.Column(c.Key("screenshots")).AlignSelf(ui.Stretch).Padding(a.Space(12), a.Space(16)).Gap(a.Space(6)).Children(func() {
		ui.Row(c).AlignSelf(ui.Stretch).AlignItems(ui.Center).Justify(ui.SpaceBetween).Gap(a.Space(8)).Children(func() {
			ui.Text(c, s.label("bug_report.screenshots", "Screenshots")).FontSize(a.Px(12)).Height(a.Px(12) * sourceLineHeight).FontWeight(500)
			if s.button(c, "add-screenshot", s.label("bug_report.add_screenshot", "Add screenshot"), "image", false, s.Busy || len(s.shots) >= native.MaxReportScreenshots).Clicked() {
				s.pickScreenshots()
			}
		})
		if len(s.shots) > 0 {
			ui.Row(c).AlignSelf(ui.Stretch).Wrap().Gap(a.Space(6)).Children(func() {
				for index, shot := range s.shots {
					ui.Row(c.Key(fmt.Sprintf("shot-%d", index))).Height(a.Px(28)).Gap(a.Space(4)).PaddingX(a.Space(6)).Background(a.Elevated).Border(1, a.native.Border).Radius(a.Px(8)).AlignItems(ui.Center).Children(func() {
						preview := ui.ButtonBase(c.Key("preview")).Disabled(s.Busy).Label("Preview / redact " + shot.Image.Name).Tooltip(strings.ReplaceAll(s.label("go.bug.preview", "Preview / redact %{name}"), "%{name}", shot.Image.Name)).Height(a.Px(26)).Gap(a.Space(4)).Children(func() {
							if shot.Bitmap != nil {
								ui.Image(c, shot.Bitmap).Size(a.Px(16), a.Px(16)).Radius(a.Px(4))
							}
							ui.Text(c, shot.Image.Name).MaxWidth(160).SingleLine().FontSize(a.Px(14)).TextColor(a.TextSecondary)
						})
						if preview.Clicked() {
							s.preview = index
						}
						remove := ui.ButtonBase(c.Key("remove")).Disabled(s.Busy).Label("Remove "+shot.Image.Name).Tooltip(strings.ReplaceAll(s.label("go.bug.remove", "Remove %{name}"), "%{name}", shot.Image.Name)).Size(a.Px(18), a.Px(24)).Center().Children(func() { icon(c, "x", a.Px(14), a.native.TextMuted) })
						if remove.Clicked() {
							s.RemoveScreenshot(index)
						}
					})
				}
			})
		}
		ui.Text(c, s.label("bug_report.screenshot_hint", "Screenshots upload through the `gh` CLI; if it isn't signed in, GitHub's web form opens for you to drag them in.")).FontSize(a.Px(10)).Height(a.Px(10) * sourceLineHeight).TextColor(a.native.TextMuted)
	})
}

func (s *State) footer(c *ui.Context) {
	a := s.appearance(c)
	ui.Row(c.Key("footer")).AlignSelf(ui.Stretch).AlignItems(ui.Center).Justify(ui.SpaceBetween).Gap(a.Space(12)).Children(func() {
		ui.Row(c).Grow(1).MinWidth(0).AlignItems(ui.Center).Gap(a.Space(6)).Children(func() {
			icon(c, "github", a.Px(14), a.native.TextMuted)
			note := s.label("go.bug.copy_0", "Diagnostics are off. Review before sharing.")
			if s.IncludeDiagnostics {
				note = s.label("go.bug.copy_1", "Selected diagnostics will be included in your review.")
			}
			if s.Busy {
				note = s.displayMessage(s.Activity) + "…"
			}
			ui.Text(c, note).FontSize(a.Px(12)).TextColor(a.native.TextMuted)
			if s.Busy {
				if s.button(c, "cancel", s.label("common.cancel", "Cancel"), "", false, s.cancelling).Clicked() {
					s.Cancel()
				}
			} else {
				if ui.ButtonBase(c.Key("privacy-options")).Label("Report privacy options").Tooltip(s.label("go.bug.copy_2", "Choose diagnostics, redaction, and window capture")).Height(a.Px(28)).PaddingX(a.Space(6)).Children(func() {
					ui.Text(c, s.label("go.bug.copy_3", "Privacy…")).FontSize(a.Px(12)).TextColor(a.TextSecondary)
				}).Clicked() {
					s.optionsOpen = true
				}
			}
		})
		ui.Row(c).Shrink(0).AlignItems(ui.Center).Gap(a.Space(8)).Children(func() {
			if s.button(c, "clear", s.label("bug_report.clear", "Clear"), "", false, s.Busy).Clicked() {
				s.Clear()
			}
			label := s.label("bug_report.submit", "Create issue on GitHub")
			if s.Busy && s.operation == "Submitting reviewed report" {
				label = s.label("bug_report.creating", "Creating…")
			}
			if s.button(c, "prepare", label, "github", true, s.Busy).Tooltip(s.label("go.bug.copy_4", "Review the report before confirming public submission")).Clicked() {
				s.Prepare()
			}
		})
	})
}

func dialog(c *ui.Context, open *bool, fn func()) {
	ui.DialogBase(c, open, func(back, panel ui.Element) {
		back.Background(ui.RGBA(0, 0, 0, .45))
		panel.Width(720).MaxWidthPercent(96).MaxHeight(760).MaxHeightPercent(95).Padding(18).Gap(12).Radius(10).Background(c.Theme().Surface).Border(1, c.Theme().Border).Children(fn)
	})
}
func (s *State) review(c *ui.Context) {
	dialog(c.Key("review"), &s.reviewOpen, func() {
		ui.Text(c, s.label("go.bug.copy_5", "Review public GitHub report")).Font(c.Theme().Font).FontSize(18).LineHeight(1.3).Height(24).Shrink(0).FontWeight(500).TextColor(c.Theme().Text)
		ui.Text(c, strings.ReplaceAll(s.label("go.bug.destination", "Destination: %{url}"), "%{url}", "https://github.com/"+s.prepared.Repository+"/issues")).Selectable()
		ui.Text(c, s.prepared.Title).FontWeight(500).TextColor(c.Theme().Text)
		ui.Scroll(c).Height(300).AlignSelf(ui.Stretch).Children(func() { ui.Text(c, s.prepared.BrowserBody).Selectable() })
		ui.Text(c, strings.ReplaceAll(s.label("go.bug.shared_count", "%{count} explicitly selected screenshots will be shared. Nothing is sent until you confirm."), "%{count}", fmt.Sprint(len(s.prepared.Screenshots))))
		ui.Checkbox(c, &s.reviewed, s.label("go.bug.copy_6", "I reviewed the report and screenshots and want to submit them publicly"))
		ui.Row(c).Gap(8).Justify(ui.End).Children(func() {
			if s.button(c, "edit", s.label("go.bug.copy_7", "Back to editing"), "", false, false).Clicked() {
				s.reviewOpen = false
				s.reviewed = false
			}
			if s.button(c, "copy", s.label("go.bug.copy_8", "Copy reviewed report"), "", false, false).Clicked() {
				c.Services().WriteClipboard(s.prepared.Title + "\n\n" + s.prepared.BrowserBody)
			}
			if s.button(c, "submit", s.label("go.bug.copy_9", "Submit reviewed report"), "", true, !s.reviewed || s.Busy).Clicked() {
				s.Submit()
			}
		})
	})
}

func (s *State) privacy(c *ui.Context) {
	dialog(c.Key("privacy"), &s.optionsOpen, func() {
		ui.Text(c, s.label("go.bug.copy_10", "Report privacy options")).Font(c.Theme().Font).FontSize(18).LineHeight(1.3).Height(24).Shrink(0).FontWeight(500).TextColor(c.Theme().Text)
		ui.Text(c, s.label("go.bug.copy_11", "These choices affect only this report. No logs, credentials or session files are collected.")).TextColor(c.Theme().TextMuted)
		ui.Checkbox(c, &s.RedactSensitive, s.label("go.bug.copy_12", "Redact recognizable secrets and my home path in report text")).Disabled(s.Busy)
		ui.Text(c, s.label("go.bug.copy_13", "Redaction cannot detect every secret. Review text and screenshots yourself. Only Context is sent to the selected model when you choose Generate draft.")).FontSize(12).TextColor(c.Theme().TextMuted)
		ui.Row(c).Gap(8).Children(func() {
			if ui.Checkbox(c, &s.IncludeDiagnostics, s.label("go.bug.copy_14", "Include build and runtime diagnostics")).Disabled(s.Busy).Changed() && s.IncludeDiagnostics {
				s.RefreshDiagnostics()
			}
			if s.button(c, "refresh", s.label("go.bug.copy_15", "Refresh diagnostics"), "", false, s.Busy || !s.IncludeDiagnostics).Clicked() {
				s.RefreshDiagnostics()
			}
		})
		if s.IncludeDiagnostics {
			if ui.Checkbox(c, &s.IncludeIdentity, s.label("go.bug.copy_16", "Include my GitHub username (contacts GitHub)")).Disabled(s.Busy).Changed() {
				s.RefreshDiagnostics()
			}
			ui.Scroll(c).Height(160).AlignSelf(ui.Stretch).Children(func() { ui.Text(c, s.Diagnostics.Markdown()).Selectable().FontSize(12) })
		}
		ui.Text(c, s.label("go.bug.copy_17", "Only images you select or explicitly capture are included. Image metadata is removed; use Preview to redact pixels.")).FontSize(12).TextColor(c.Theme().TextMuted)
		if s.Options.Capture != nil {
			ui.Checkbox(c, &s.CaptureConsent, s.label("go.bug.copy_18", "I want to capture the current Orbit window, including any visible private content")).Disabled(s.Busy)
			if s.button(c, "capture", s.label("go.bug.copy_19", "Capture this window"), "image", false, s.Busy || !s.CaptureConsent || len(s.shots) >= native.MaxReportScreenshots).Clicked() {
				s.optionsOpen = false
				s.Capture()
			}
		}
		if s.Error != nil {
			ui.Text(c, s.errorMessage(s.Error)).TextColor(c.Theme().Danger).Selectable()
		}
		ui.Row(c).AlignSelf(ui.Stretch).Justify(ui.End).Gap(8).Children(func() {
			if s.Busy {
				ui.Text(c, s.displayMessage(s.Activity)+"…").TextColor(c.Theme().TextMuted)
				if s.button(c, "cancel", s.label("common.cancel", "Cancel"), "", false, s.cancelling).Clicked() {
					s.Cancel()
				}
			}
			if s.button(c, "done", s.label("common.done", "Done"), "", false, false).Clicked() {
				s.optionsOpen = false
				s.CaptureConsent = false
			}
		})
	})
}

func (s *State) previewView(c *ui.Context) {
	open := true
	index := s.preview
	shot := &s.shots[index]
	dialog(c.Key("screenshot-preview"), &open, func() {
		ui.Text(c, shot.Image.Name).FontWeight(500).TextColor(c.Theme().Text)
		ui.Image(c, shot.Bitmap).Height(320).MaxWidth(680)
		ui.Text(c, strings.NewReplacer("%{width}", fmt.Sprint(shot.Image.Width), "%{height}", fmt.Sprint(shot.Image.Height)).Replace(s.label("go.bug.pixels", "Original pixels: %{width} × %{height}. Opaque black boxes remove selected pixels from the attachment only."))).FontSize(12)
		ui.Text(c, s.label("go.bug.copy_20", "Redaction boxes: x,y,width,height; separate boxes with semicolons")).FontSize(12)
		ui.TextInput(c, &shot.Redactions).Label("Redaction boxes: x,y,width,height; x,y,width,height").Placeholder("20,40,200,50").AlignSelf(ui.Stretch).Disabled(s.Busy)
		if s.Error != nil {
			ui.Text(c, s.errorMessage(s.Error)).TextColor(c.Theme().Danger).Selectable()
		}
		ui.Row(c).Gap(8).Children(func() {
			if s.button(c, "redact", s.label("go.bug.copy_21", "Apply redaction"), "", false, s.Busy || strings.TrimSpace(shot.Redactions) == "").Clicked() {
				s.redact(index)
			}
			if s.button(c, "done", s.label("common.done", "Done"), "", false, false).Clicked() {
				s.preview = -1
			}
		})
	})
	if !open {
		s.preview = -1
	}
}
