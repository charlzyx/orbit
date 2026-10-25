package workspace

import (
	"encoding/json"
	"hash/fnv"
	"os"
	"path/filepath"
	"strings"
	"unicode"
	"unicode/utf8"
)

func validGitHubLogin(s string) bool {
	s = strings.TrimSpace(s)
	if s == "" || len(s) > 39 || strings.HasPrefix(s, "-") || strings.HasSuffix(s, "-") || strings.Contains(s, "--") {
		return false
	}
	for _, r := range s {
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-') {
			return false
		}
	}
	return true
}
func LoadAvatarOverrides(home string) map[string]string {
	result := map[string]string{}
	if home == "" {
		home, _ = os.UserHomeDir()
	}
	raw, e := os.ReadFile(filepath.Join(home, ".orbit-pi", "git-avatars.json"))
	if e != nil {
		return result
	}
	var values map[string]any
	if json.Unmarshal(raw, &values) != nil {
		return result
	}
	for email, value := range values {
		login, ok := value.(string)
		if ok && validGitHubLogin(login) {
			result[strings.ToLower(strings.TrimSpace(email))] = strings.TrimSpace(login)
		}
	}
	return result
}
func ResolveAvatarLogin(name, email string, overrides map[string]string) string {
	if login := overrides[strings.ToLower(strings.TrimSpace(email))]; validGitHubLogin(login) {
		return strings.TrimSpace(login)
	}
	local, domain, ok := strings.Cut(strings.TrimSpace(email), "@")
	if ok && strings.EqualFold(domain, "users.noreply.github.com") {
		login := local[strings.LastIndex(local, "+")+1:]
		if validGitHubLogin(login) {
			return login
		}
	}
	if validGitHubLogin(name) {
		return strings.TrimSpace(name)
	}
	return ""
}
func AvatarURL(name, email string, overrides map[string]string) string {
	login := ResolveAvatarLogin(name, email, overrides)
	if login == "" {
		return ""
	}
	return "https://github.com/" + login + ".png?size=48"
}
func AvatarInitials(name string) string {
	words := strings.Fields(name)
	if len(words) == 0 {
		return "?"
	}
	first, _ := utf8.DecodeRuneInString(words[0])
	out := string(unicode.ToUpper(first))
	if len(words) > 1 {
		last, _ := utf8.DecodeRuneInString(words[len(words)-1])
		return out + string(unicode.ToUpper(last))
	}
	runes := []rune(words[0])
	if len(runes) > 1 {
		out += string(unicode.ToUpper(runes[1]))
	}
	return out
}
func AvatarHue(name, email string) uint32 {
	key := strings.ToLower(strings.TrimSpace(email))
	if key == "" {
		key = name
	}
	hash := fnv.New32a()
	_, _ = hash.Write([]byte(key))
	return hash.Sum32() % 360
}
