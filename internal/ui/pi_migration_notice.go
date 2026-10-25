package orbitui

import orbitassets "github.com/charlzyx/orbit"

// notifyPiMigrationRequired preserves the source notification, not a navigation
// action: the source setup page may still be hidden when all dependencies exist.
func (a *App) notifyPiMigrationRequired(latest string) {
	a.Notify(orbitassets.Translate(a.catalog, a.english, "pi_update.migration_required", map[string]string{"version": latest}), true)
}
