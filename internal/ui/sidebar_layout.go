package orbitui

import "github.com/egoist/mygo/ui"

// The source sidebar owns its right border. Its six-DIP handle is fully
// inside that width, while the root drag listener uses the pointer's X.
const sidebarResizeGripWidth float32 = 6

func (a *App) sidebarResize(c *ui.Context) {
	line := ui.Box(c.Key("sidebar-divider")).Absolute().Left(a.SidebarWidth - 1).Top(0).Bottom(0).Width(1).Background(a.Theme.Border).Focusable().Role(ui.RoleSplitter).Label("Divider")
	handle := ui.Box(c.Key("sidebar-resize")).Absolute().Left(a.SidebarWidth - sidebarResizeGripWidth).Top(0).Bottom(0).Width(sidebarResizeGripWidth).Cursor(ui.CursorResizeEW).Label("Resize sidebar")
	if handle.Hovered() {
		handle.Background(a.Theme.Accent.Alpha(.4))
	}
	if dx, _, held := handle.Dragged(); held {
		line.Focus()
		if dx != 0 {
			x, _, _ := handle.PointerPosition()
			previous := a.SidebarWidth
			a.setSidebarDragWidth(c, handle.Bounds().X+x)
			if previous != a.SidebarWidth {
				a.persistSidebarWidth()
			}
			c.AnimationFrame()
		}
	}
	// Preserve the existing native keyboard-accessible divider behavior.
	for _, k := range []struct {
		key   ui.Key
		delta int
	}{{ui.KeyLeft, -10}, {ui.KeyRight, 10}} {
		if line.Shortcut(0, k.key) {
			a.setSidebarDragWidth(c, a.SidebarWidth+a.Theme.Space(k.delta))
			a.persistSidebarWidth()
		}
	}
}
func (a *App) setSidebarDragWidth(c *ui.Context, width float32) {
	w, _ := c.Size()
	a.SidebarWidth = max(SidebarMinWidth, min(width, max(SidebarMinWidth, w-400)))
	a.Layout.SidebarWidth = float64(a.SidebarWidth)
}
func (a *App) persistSidebarWidth() {
	a.save(a.Options.Store.SaveLayout(a.Layout))
}
