package agent

import (
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"

	"github.com/charlzyx/orbit/internal/pirpc"
)

// The origin travels in these responseful abort/clear/compact request IDs, rather than an
// ever-growing tombstone map. Late replies remain attributable after timeout.
const abortRequestPrefix = "orbit-abort-v1:"

type abortOrigin struct {
	sessionID    string
	reset, nonce uint64
}

func (o abortOrigin) command(kind string) pirpc.Command {
	c := pirpc.NewCommand(kind, nil)
	c.ID = fmt.Sprintf("%s%d:%d:%s:%s", abortRequestPrefix, o.reset, o.nonce, base64.RawURLEncoding.EncodeToString([]byte(o.sessionID)), kind)
	return c
}

func parseAbortOrigin(id, command string) (abortOrigin, bool) {
	if !strings.HasPrefix(id, abortRequestPrefix) {
		return abortOrigin{}, false
	}
	parts := strings.Split(strings.TrimPrefix(id, abortRequestPrefix), ":")
	if len(parts) != 4 || parts[3] != command || (command != pirpc.ClearQueue && command != pirpc.Abort && command != pirpc.Compact) {
		return abortOrigin{}, false
	}
	reset, err := strconv.ParseUint(parts[0], 10, 64)
	if err != nil || strconv.FormatUint(reset, 10) != parts[0] {
		return abortOrigin{}, false
	}
	nonce, err := strconv.ParseUint(parts[1], 10, 64)
	if err != nil || nonce == 0 || strconv.FormatUint(nonce, 10) != parts[1] {
		return abortOrigin{}, false
	}
	session, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil || base64.RawURLEncoding.EncodeToString(session) != parts[2] {
		return abortOrigin{}, false
	}
	return abortOrigin{sessionID: string(session), reset: reset, nonce: nonce}, true
}

func (o abortOrigin) matches(entry *liveSession) bool {
	return !entry.stopped && entry.workflowGeneration == o.reset && entry.state.snapshot.ActiveSession.ID == o.sessionID
}
