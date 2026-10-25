package orbitui

import (
	"strings"
	"testing"

	orbitassets "github.com/charlzyx/orbit"
	"github.com/charlzyx/orbit/internal/native"
	"github.com/egoist/mygo/ui"
)

func TestPiUpdateDeferralOnlyConfirmedBusyGetsSourceNotice(t *testing.T) {
	for _, reason := range []native.PiUpdateDeferralReason{"", native.PiUpdateDeferralNoIdleVerifier, native.PiUpdateDeferralAdmissionRejected, "unknown", native.PiUpdateDeferralBusyTimeout} {
		t.Run(string(reason), func(t *testing.T) {
			a := fixture(t)
			actions := 0
			a.Options.OnAction = func(Action) { actions++ }
			a.notifyPiUpdateStatus(native.PiUpdateEvent{State: "deferred", Deferral: reason, Latest: "0.60.3", Detail: "private raw admission explanation"})
			items := a.toastStack.Items()
			if reason != native.PiUpdateDeferralBusyTimeout {
				if len(items) != 0 {
					t.Fatal("uncertain deferral invented a running task", items)
				}
				return
			}
			want := orbitassets.Translate(a.catalog, a.english, "pi_update.available_next_launch", map[string]string{"version": "0.60.3"})
			if len(items) != 1 || items[0].Body != want || items[0].Kind != native.ToastInfo {
				t.Fatalf("confirmed busy notice missing: %+v", items)
			}
			tt := ui.NewTester(a.View, 1470, 922)
			tt.Frame()
			texts := strings.Join(tt.Texts(), "\n")
			if !strings.Contains(texts, want) || strings.Contains(texts, "private raw admission explanation") || actions != 0 {
				t.Fatal("incorrect busy notification delivery")
			}
		})
	}
}

func TestPiUpdateDeferralOwnerDeliveryAndLocalizedVersion(t *testing.T) {
	x := workspaceDialogFixture(t)
	x.UI.catalog = map[string]string{"pi_update.available_next_launch": "版本 %{version} 将在下次启动更新；任务仍在运行"}
	owner := x.capturePiUpdateOwner()
	x.enqueue(func() {
		x.applyPiUpdateEvent(owner, native.PiUpdateEvent{State: "deferred", Deferral: native.PiUpdateDeferralBusyTimeout, Latest: "0.60.3"})
	})
	if len(x.UI.toastStack.Items()) != 0 {
		t.Fatal("event bypassed owner queue")
	}
	x.drainHostUpdates()
	items := x.UI.toastStack.Items()
	if len(items) != 1 || items[0].Body != "版本 0.60.3 将在下次启动更新；任务仍在运行" || items[0].Kind != native.ToastInfo {
		t.Fatalf("owner did not deliver translated busy notice: %+v", items)
	}
	x.invalidateSetupProbe()
	x.applyPiUpdateEvent(owner, native.PiUpdateEvent{State: "deferred", Deferral: native.PiUpdateDeferralBusyTimeout, Latest: "stale-version"})
	if len(x.UI.toastStack.Items()) != 1 {
		t.Fatal("typed deferral bypassed receipt ownership")
	}
}
