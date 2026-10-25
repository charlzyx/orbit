package orbitui

import (
	"context"
	"errors"
	"os"
	"time"

	"github.com/charlzyx/orbit/internal/agent"
	"github.com/charlzyx/orbit/internal/modelui"
	"github.com/charlzyx/orbit/internal/native"
	"github.com/charlzyx/orbit/internal/preferences"
	"github.com/charlzyx/orbit/internal/workbench"
	"github.com/egoist/mygo/ui"
)

func (a *App) initModels() {
	a.Models = modelui.New(a.Options.Store, modelui.Options{
		Theme: func() modelui.Theme {
			t := a.Theme
			alpha := float32(.16)
			if t.Dark {
				alpha = .14
			}
			return modelui.Theme{Elevated: t.Elevated, SurfaceSecondary: t.SurfaceSecondary, Secondary: t.TextSecondary, Active: t.Active, Hover: t.Hover, BorderStrong: t.Text.Alpha(alpha), AccentText: t.AccentText, Overlay: t.Text.Alpha(.05), CodeFont: a.Fonts.CodeFamily, ReduceMotion: a.Prefs.ReduceMotion, CodePx: func(n float32) float32 { return n * t.CodeSize / 13 }, Px: t.Px, Space: t.Space}
		},
		Text: a.text, Notify: a.Notify,
		OnSelectModel: func(provider, id string) error {
			a.Dispatch(Action{Kind: "SetModel", Value: id, Data: map[string]any{"provider": provider}})
			return nil
		},
		OnSelectThinking: func(level string) error {
			a.Dispatch(Action{Kind: "SetThinking", Value: level})
			return nil
		},
		OnRefresh: func() error {
			if a.Options.RefreshModels == nil {
				return errors.New("Model catalog refresh is not connected")
			}
			return a.Options.RefreshModels()
		},
		OnDefaultsChanged: func(v preferences.SessionDefault) { a.sessionDefaultPrefs = v },
		OnTitleChanged:    func(v preferences.AutoTitle) { a.autoTitlePrefs = v },
		OnDismiss: func(mouse bool) {
			a.modelOpen = false
			a.thinkingOpen = false

			a.composer.Focus()
		},
	})
}
func (a *App) syncModelCatalog() {
	if a.Models == nil {
		return
	}
	models := a.Model.ModelCatalog
	if models == nil {
		for _, m := range a.Model.Models {
			models = append(models, agent.ModelChoice{ID: m.ID, Provider: m.Provider, Name: m.Name, Reasoning: m.Reasoning})
		}
	}
	current := a.Model.CurrentModel
	if current.ID == "" {
		current.ID = a.Model.ModelID
		current.Provider = a.Model.ModelProvider
	}
	a.Models.SetSnapshot(modelui.Snapshot{Models: models, Current: current, Levels: a.Model.ThinkingLevels, CurrentLevel: a.Model.Thinking, Connected: a.Model.Connected, Loading: a.Model.Status == "loading", Running: a.Model.Running, Error: a.Model.Error})
	a.Models.SetTitlePreferences(a.autoTitlePrefs)
}

// refreshModelCatalog owns the complete source action, not merely a catalog
// query. The maintenance lock excludes new prompts/session switches throughout
// the update and the accepted restart, including parked processes.
func (x *Controller) refreshModelCatalog() error {
	a := x.UI
	if x.catalogRefreshing {
		return errors.New("The model catalog is already updating")
	}
	release, err := x.beginRuntimeOperation()
	if err != nil {
		return err
	}
	x.catalogRefreshing = true
	if a.Models != nil {
		a.Models.SetRefreshState(true, "")
	}
	binary, home, piDir, dir := x.Agent.PiBinary(), x.Store.UserHome, x.Store.PiDir, x.workspace
	svc := workbench.New(x.Store, dir)
	a.Async(func() func() {
		defer release()
		env, runErr := svc.SecretEnv()
		if runErr == nil {
			ctx, cancel := context.WithTimeout(x.background, 2*time.Minute)
			defer cancel()
			_, runErr = native.RefreshPiModelCatalogWithConfig(ctx, native.ModelCatalogOptions{Binary: binary, Home: home, PiDir: piDir, Workspace: dir, Environment: append(os.Environ(), env...)})
		}
		if runErr == nil {
			if x.Agent.Snapshot().ActiveSession.Workspace == "" {
				runErr = x.Agent.Open(dir, "")
			} else {
				runErr = x.Agent.Restart()
			}
		}
		detail := ""
		if runErr != nil {
			detail = svc.Redact(runErr.Error())
		}
		return func() {
			x.catalogRefreshing = false
			if a.Models != nil {
				a.Models.SetRefreshState(false, detail)
			}
			if detail != "" {
				a.Notify(detail, true)
			} else {
				a.Notify(a.text("settings.models_updated", "Model catalog updated"), false)
			}
		}
	})
	return nil
}

func (a *App) openModelPicker(kind modelui.PickerKind, now time.Time) {
	if a.Models == nil || a.Model.Running {
		return
	}
	if a.Models.PickerIsOpen() && a.Models.PickerKind() == kind {
		a.Models.DismissPicker(false)
		return
	}
	if !a.modelDismissedAt.IsZero() && now.Sub(a.modelDismissedAt) < 200*time.Millisecond {
		return
	}
	a.workspaceOpen = false
	a.autocomplete.Dismiss()
	a.syncModelCatalog()
	if a.Models.PickerOpen(kind) {
		a.queryPickerCatalog()
		a.modelOpen = kind == modelui.ModelPicker
		a.thinkingOpen = kind == modelui.ThinkingPicker
	}
}
func (a *App) modelPickerPopup(c *ui.Context, anchor ui.Element, kind modelui.PickerKind) {
	if a.Models == nil {
		return
	}
	requested := a.modelOpen && kind == modelui.ModelPicker || a.thinkingOpen && kind == modelui.ThinkingPicker
	if requested && !a.Models.PickerIsOpen() {
		if a.Model.Running {
			a.modelOpen = false
			a.thinkingOpen = false
			return
		}
		a.syncModelCatalog()
		if !a.Models.PickerOpen(kind) {
			a.modelOpen = false
			a.thinkingOpen = false
			return
		}
		a.queryPickerCatalog()
	}
	if !a.Models.PickerIsOpen() || a.Models.PickerKind() != kind {
		return
	}
	open := true
	popup := ui.PopoverBase(c.Key("composer-model-picker:"+string(kind)), anchor, &open, func(panel ui.Element) {
		panel.AttachTo(anchor, ui.AnchorTopLeft, ui.AnchorBottomLeft).Margin(0, 0, 5, 0).Column().Width(a.Models.PickerWidth()).MaxWidth(a.Models.PickerWidth()).Label("Composer " + string(kind) + " picker")
		a.Models.PickerView(c)
	})
	if popup.Dismissed() {
		if popup.PressedOutside() {
			a.modelDismissedAt = c.Now()
		}
		a.Models.DismissPicker(popup.PressedOutside())
	}
}

func (a *App) queryPickerCatalog() {
	if a.Options.QueryModelCatalog != nil {
		a.Options.QueryModelCatalog()
	}
}
