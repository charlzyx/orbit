package orbitui

import (
	orbitassets "github.com/charlzyx/orbit"
	"github.com/egoist/mygo/ui"
	"testing"
)

func TestSourceSidebarNewTaskUsesSelectedChineseCatalog(t *testing.T) {
	a := fixture(t)
	var err error
	a.catalog, err = orbitassets.Locale("zh-CN")
	if err != nil {
		t.Fatal(err)
	}
	tt := ui.NewTester(a.View, 1470, 922)
	tt.Frame()
	label := a.Tr("menu.new_task")
	if label == "New Task" || label == "menu.new_task" || !tt.HasText(label) {
		t.Fatal("source sidebar label fell back to English", label, tt.Texts())
	}
}
