package agent

// A new sign-in may not display a URL or one-time device code from its previous
// session. Provider availability and credentials are deliberately unaffected.
func clearAuthChallenge(auth *AuthState) {
	auth.URL, auth.UserCode, auth.VerificationURI = "", "", ""
	auth.ExpiresAt = 0
}

// Transport failures are scoped to their own sign-in. A delayed error cannot
// replace a newer challenge or an already completed successful login.
func failAuthLogin(auth *AuthState, sessionID string, err error) {
	if err == nil || sessionID == "" || auth.SessionID != sessionID || (auth.Status != "pending" && auth.Status != "waiting") {
		return
	}
	auth.Status = "failed"
	auth.Error = err.Error()
	clearAuthChallenge(auth)
}
