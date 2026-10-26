package workspace

import "unicode/utf8"

// Source #88 commit_message::Budget. Byte estimates are not tokenizer guarantees.
const commitDefaultTotal = 256 * 1024
const commitMinimumTotal = 24 * 1024
const commitMaximumTotal = 2 * 1024 * 1024
const commitTruncation = "\n… [truncated: prompt budget exhausted]"

type commitBudget struct{ total, diff, original int }

func commitTotal(window int) int {
	if window <= 0 {
		return commitDefaultTotal
	}
	// Saturate before multiplication (also safe on 32-bit targets).
	if window >= commitMaximumTotal {
		return commitMaximumTotal
	}
	return min(commitMaximumTotal, max(commitMinimumTotal, window*3/2))
}
func newCommitBudget(window, reserved, files int) commitBudget {
	total := commitTotal(window)
	available := max(commitMinimumTotal/2, total-reserved)
	b := commitBudget{total: total, diff: available / 2, original: available * 3 / 10}
	if files > 0 && b.original/files < 1024 {
		b.diff += b.original
		b.original = 0
	}
	return b
}
func commitBound(s string, limit int) string {
	if limit <= 0 {
		return ""
	}
	if len(s) <= limit {
		return s
	}
	marker := commitTruncation
	if len(marker) > limit {
		marker = ""
	}
	n := limit - len(marker)
	for n > 0 && !utf8.RuneStart(s[n]) {
		n--
	}
	return s[:n] + marker
}
func (b *commitBudget) takeOriginal(s string, remainingFiles int) string {
	n := min(48*1024, b.original/max(1, remainingFiles))
	if n < 1024 {
		return ""
	}
	result := commitBound(s, n)
	b.original -= len(result)
	return result
}
func (b *commitBudget) takeDiff(s string, remainingFiles int) string {
	n := min(b.diff, min(64*1024, max(1024, b.diff/max(1, remainingFiles))))
	if n < 1024 {
		return ""
	}
	result := commitBound(s, n)
	b.diff -= len(result)
	return result
}
