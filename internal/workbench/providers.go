package workbench

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/charlzyx/orbit/internal/runtimepath"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type Provider struct {
	ID, Name, BaseURL, API string
	HasAPIKey              bool
	ModelIDs               []string
}
type ProviderEdit struct {
	ID, Name, BaseURL, API string
	APIKey                 *string
	ModelIDs               []string
}

func (s *Service) Providers() ([]Provider, error) {
	m, e := readObject(filepath.Join(s.Preferences.PiDir, "models.json"))
	if e != nil {
		return nil, e
	}
	if m["providers"] == nil {
		return nil, nil
	}
	providers, ok := m["providers"].(map[string]any)
	if !ok {
		return nil, errors.New("models.json providers must be an object")
	}
	var out []Provider
	for id, v := range providers {
		m, ok := v.(map[string]any)
		if !ok {
			continue
		}
		p := Provider{ID: id, Name: stringValue(m["name"]), BaseURL: stringValue(m["baseUrl"]), API: stringValue(m["api"]), HasAPIKey: stringValue(m["apiKey"]) != ""}
		models, _ := m["models"].([]any)
		for _, v := range models {
			if model, ok := v.(map[string]any); ok {
				if id := stringValue(model["id"]); id != "" {
					p.ModelIDs = append(p.ModelIDs, id)
				}
			}
		}
		out = append(out, p)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out, nil
}
func (s *Service) SaveProvider(p ProviderEdit) error {
	p.ID = strings.TrimSpace(p.ID)
	if p.ID == "" {
		return errors.New("provider id is required")
	}
	return s.edit(filepath.Join(s.Preferences.PiDir, "models.json"), func(m map[string]any) error {
		providers, e := object(m, "providers")
		if e != nil {
			return e
		}
		entry, ok := providers[p.ID].(map[string]any)
		if !ok {
			entry = map[string]any{}
		}
		for k, v := range map[string]string{"name": p.Name, "baseUrl": p.BaseURL, "api": p.API} {
			if strings.TrimSpace(v) == "" {
				delete(entry, k)
			} else {
				entry[k] = strings.TrimSpace(v)
			}
		}
		if p.APIKey != nil && strings.TrimSpace(*p.APIKey) != "" {
			entry["apiKey"] = strings.TrimSpace(*p.APIKey)
		}
		if len(p.ModelIDs) > 0 {
			old, _ := entry["models"].([]any)
			models := make([]any, 0, len(p.ModelIDs))
			for _, id := range p.ModelIDs {
				id = strings.TrimSpace(id)
				if id == "" {
					continue
				}
				var found any
				for _, raw := range old {
					if model, ok := raw.(map[string]any); ok && stringValue(model["id"]) == id {
						found = model
						break
					}
				}
				if found == nil {
					found = map[string]any{"id": id}
				}
				models = append(models, found)
			}
			entry["models"] = models
		}
		if len(entry) == 0 {
			delete(providers, p.ID)
		} else {
			providers[p.ID] = entry
		}
		return nil
	})
}
func (s *Service) RemoveProvider(id string) error {
	path := filepath.Join(s.Preferences.PiDir, "models.json")
	s.mu.Lock()
	defer s.mu.Unlock()
	m, e := readObject(path)
	if e != nil {
		return e
	}
	providers, ok := m["providers"].(map[string]any)
	if !ok {
		return nil
	}
	delete(providers, id)
	return writeObject(path, m)
}

type CredentialStatus struct {
	Kind       string
	Configured bool
}

// Credentials exposes status only; raw credential values never enter UI snapshots.
func (s *Service) Credentials() (map[string]CredentialStatus, error) {
	m, e := readObject(filepath.Join(s.Preferences.PiDir, "auth.json"))
	if e != nil {
		return nil, e
	}
	out := map[string]CredentialStatus{}
	hasOllamaSession := false
	for id, v := range m {
		if entry, ok := v.(map[string]any); ok {
			kind := stringValue(entry["type"])
			switch kind {
			case "ollama_cloud_session":
				hasOllamaSession = true
				continue
			case "oauth", "api_key":
			default:
				if _, ok := entry["refresh"]; ok {
					kind = "oauth"
				} else if _, ok := entry["access"]; ok {
					kind = "oauth"
				} else if _, ok := entry["key"]; ok {
					kind = "api_key"
				} else {
					continue
				}
			}
			out[id] = CredentialStatus{Kind: kind, Configured: true}
		}
	}
	if hasOllamaSession {
		for _, id := range []string{"ollama", "ollama-cloud"} {
			if _, ok := out[id]; !ok {
				out[id] = CredentialStatus{Kind: "ollama_cloud_session", Configured: true}
			}
		}
	}
	return out, nil
}
func (s *Service) SaveAPIKey(id, key string) error {
	id = strings.TrimSpace(id)
	if strings.TrimSpace(id) == "" || strings.TrimSpace(key) == "" {
		return errors.New("provider and API key are required")
	}
	return s.edit(filepath.Join(s.Preferences.PiDir, "auth.json"), func(m map[string]any) error {
		m[id] = map[string]any{"type": "api_key", "key": strings.TrimSpace(key)}
		return nil
	})
}
func (s *Service) RemoveCredential(id string) error {
	return s.edit(filepath.Join(s.Preferences.PiDir, "auth.json"), func(m map[string]any) error { delete(m, id); return nil })
}

// ProviderMetadata describes Pi's installed provider capabilities, not whether
// the running session has loaded or authenticated the provider.
type ProviderMetadata struct {
	ID              string   `json:"id"`
	Name            string   `json:"name"`
	OAuth           bool     `json:"oauth"`
	APIKey          bool     `json:"apiKey"`
	EnvVars         []string `json:"envVars"`
	Note            string   `json:"-"`
	CatalogCount    int      `json:"-"`
	HasCatalogCount bool     `json:"-"`
}
type ProviderCatalog struct {
	Providers []ProviderMetadata
	Dynamic   bool
}

// ProviderCatalog introspects the installed Pi package without contacting any
// provider. A valid fallback is returned alongside the diagnostic on failure.
func (s *Service) ProviderCatalog(ctx context.Context) (ProviderCatalog, error) {
	catalog := ProviderCatalog{Providers: BuiltinProviders()}
	data, e := s.piAIDataDir()
	if e != nil {
		return catalog, e
	}
	counts := readProviderCatalogCounts(data)
	providersDir := filepath.Dir(data)
	dist := filepath.Dir(providersDir)
	out, e := s.resourceProbe(ctx, dist, "node", "--input-type=module", "-e", providerIntrospectionScript, filepath.Join(providersDir, "all.js"), filepath.Join(dist, "env-api-keys.js"))
	if e == nil {
		var providers []ProviderMetadata
		providers, e = parseDynamicProviders(out)
		if e == nil {
			catalog.Providers = providers
			catalog.Dynamic = true
		}
	}
	for n := range catalog.Providers {
		p := &catalog.Providers[n]
		p.CatalogCount, p.HasCatalogCount = counts[p.ID]
		p.Note = ProviderNote(p.ID)
	}
	return catalog, e
}
func (s *Service) piAIDataDir() (string, error) {
	binary := s.PiBinary
	if binary == "" {
		binary = "pi"
	}
	if !strings.ContainsAny(binary, "/\\") {
		binary = runtimepath.Resolve(binary)
	}
	if binary == "" {
		return "", errors.New("Pi executable was not found")
	}
	canonical, e := filepath.EvalSymlinks(binary)
	if e != nil {
		return "", e
	}
	canonical, e = filepath.Abs(canonical)
	if e != nil {
		return "", e
	}
	for dir := canonical; ; dir = filepath.Dir(dir) {
		candidate := filepath.Join(dir, "node_modules", "@earendil-works", "pi-ai", "dist", "providers", "data")
		if st, e := os.Stat(candidate); e == nil && st.IsDir() {
			return candidate, nil
		}
		if filepath.Dir(dir) == dir {
			break
		}
	}
	return "", errors.New("Pi provider package was not found; using the bundled provider metadata")
}

const providerIntrospectionScript = `
import { pathToFileURL } from 'node:url';
const { builtinProviders } = await import(pathToFileURL(process.argv[1]).href);
let findEnvKeys = () => undefined;
try { const mod = await import(pathToFileURL(process.argv[2]).href); if (typeof mod.findEnvKeys === 'function') findEnvKeys = mod.findEnvKeys; } catch {}
const anyEnv = new Proxy({}, { get: (_, k) => (typeof k === 'string' ? 'x' : undefined) });
console.log(JSON.stringify(builtinProviders().map(p => ({id:p.id,name:p.name??p.id,oauth:!!p.auth?.oauth,apiKey:!!p.auth?.apiKey,envVars:findEnvKeys(p.id,anyEnv)??[]}))));
`

func parseDynamicProviders(raw string) ([]ProviderMetadata, error) {
	var entries []map[string]json.RawMessage
	if e := json.Unmarshal([]byte(raw), &entries); e != nil {
		return nil, e
	}
	if len(entries) == 0 {
		return nil, errors.New("Pi returned an empty provider catalog")
	}
	out := make([]ProviderMetadata, 0, len(entries))
	for _, m := range entries {
		var p ProviderMetadata
		if e := json.Unmarshal(m["id"], &p.ID); e != nil || p.ID == "" {
			return nil, errors.New("Pi provider catalog entry has no id")
		}
		_ = json.Unmarshal(m["name"], &p.Name)
		if p.Name == "" {
			p.Name = p.ID
		}
		_ = json.Unmarshal(m["oauth"], &p.OAuth)
		_ = json.Unmarshal(m["apiKey"], &p.APIKey)
		var vars []any
		_ = json.Unmarshal(m["envVars"], &vars)
		for _, v := range vars {
			if value, ok := v.(string); ok {
				p.EnvVars = append(p.EnvVars, value)
			}
		}
		out = append(out, p)
	}
	return out, nil
}
func readProviderCatalogCounts(dir string) map[string]int {
	out := map[string]int{}
	entries, _ := os.ReadDir(dir)
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}
		m, e := readObject(filepath.Join(dir, entry.Name()))
		if e != nil {
			continue
		}
		count := 0
		for _, api := range m {
			if models, ok := api.(map[string]any); ok {
				count += len(models)
			}
		}
		if count > 0 {
			out[strings.TrimSuffix(entry.Name(), ".json")] = count
		}
	}
	return out
}
func ProviderNote(id string) string {
	for _, p := range builtinProviderMetadata {
		if p.ID == id {
			return p.Note
		}
	}
	return ""
}
func BuiltinProviders() []ProviderMetadata {
	out := append([]ProviderMetadata(nil), builtinProviderMetadata...)
	for n := range out {
		out[n].EnvVars = append([]string(nil), out[n].EnvVars...)
	}
	return out
}
func ValidateProviderEdit(p ProviderEdit, inCatalog bool) error {
	id := strings.TrimSpace(p.ID)
	if id == "" {
		return errors.New("provider id is required")
	}
	for n, c := range id {
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || (n > 0 && strings.ContainsRune("._-", c))) {
			return errors.New("provider id must start with a letter or digit and contain only letters, digits, dots, underscores, or hyphens")
		}
	}
	base := strings.TrimSpace(p.BaseURL)
	if base != "" && !strings.HasPrefix(base, "http://") && !strings.HasPrefix(base, "https://") {
		return errors.New("provider base URL must use http or https")
	}
	if base == "" && !inCatalog {
		return errors.New("custom provider base URL is required")
	}
	models := false
	for _, id := range p.ModelIDs {
		models = models || strings.TrimSpace(id) != ""
	}
	if !models && !inCatalog {
		return errors.New("add at least one custom model id")
	}
	return nil
}
func ProviderEditDefaults(p Provider, inCatalog bool) ProviderEdit {
	api := p.API
	if api == "" && !inCatalog {
		api = "openai-completions"
	}
	return ProviderEdit{ID: p.ID, Name: p.Name, BaseURL: p.BaseURL, API: api, ModelIDs: append([]string(nil), p.ModelIDs...)}
}

const ollamaSessionKey = "ollama-cloud-session"

func (s *Service) SaveOllamaSession(session string) error {
	session = strings.TrimSpace(session)
	if session == "" || !strings.Contains(session, "=") {
		return errors.New("paste the Ollama Cookie header containing a name=value pair")
	}
	return s.edit(filepath.Join(s.Preferences.PiDir, "auth.json"), func(m map[string]any) error {
		if legacy, ok := m["ollama"].(map[string]any); ok && legacy["type"] == "ollama_cloud_session" {
			delete(m, "ollama")
		}
		m[ollamaSessionKey] = map[string]any{"type": "ollama_cloud_session", "session": session}
		return nil
	})
}
func (s *Service) RemoveOllamaSession() error { return s.RemoveCredential(ollamaSessionKey) }

// Frozen Orbit baseline fallback; installed Pi introspection takes precedence.
var builtinProviderMetadata = []ProviderMetadata{
	{ID: "anthropic", Name: "Anthropic", APIKey: true, OAuth: true, EnvVars: []string{"ANTHROPIC_API_KEY"}, Note: ""},
	{ID: "openai", Name: "OpenAI", APIKey: true, OAuth: false, EnvVars: []string{"OPENAI_API_KEY"}, Note: ""},
	{ID: "openai-codex", Name: "ChatGPT (Codex)", APIKey: false, OAuth: true, EnvVars: []string{}, Note: "ChatGPT Plus/Pro subscription."},
	{ID: "github-copilot", Name: "GitHub Copilot", APIKey: true, OAuth: true, EnvVars: []string{"COPILOT_GITHUB_TOKEN"}, Note: ""},
	{ID: "xai", Name: "xAI", APIKey: true, OAuth: true, EnvVars: []string{"XAI_API_KEY"}, Note: ""},
	{ID: "openrouter", Name: "OpenRouter", APIKey: true, OAuth: true, EnvVars: []string{"OPENROUTER_API_KEY"}, Note: ""},
	{ID: "radius", Name: "Radius", APIKey: true, OAuth: true, EnvVars: []string{"RADIUS_API_KEY"}, Note: ""},
	{ID: "kimi-coding", Name: "Kimi For Coding", APIKey: true, OAuth: true, EnvVars: []string{"KIMI_API_KEY"}, Note: ""},
	{ID: "ant-ling", Name: "Ant Ling", APIKey: true, OAuth: false, EnvVars: []string{"ANT_LING_API_KEY"}, Note: ""},
	{ID: "azure-openai-responses", Name: "Azure OpenAI", APIKey: true, OAuth: false, EnvVars: []string{"AZURE_OPENAI_API_KEY"}, Note: "Also needs AZURE_OPENAI_BASE_URL or AZURE_OPENAI_RESOURCE_NAME."},
	{ID: "deepseek", Name: "DeepSeek", APIKey: true, OAuth: false, EnvVars: []string{"DEEPSEEK_API_KEY"}, Note: ""},
	{ID: "nvidia", Name: "NVIDIA NIM", APIKey: true, OAuth: false, EnvVars: []string{"NVIDIA_API_KEY"}, Note: ""},
	{ID: "google", Name: "Google Gemini", APIKey: true, OAuth: false, EnvVars: []string{"GEMINI_API_KEY"}, Note: ""},
	{ID: "google-vertex", Name: "Google Vertex", APIKey: true, OAuth: false, EnvVars: []string{"GOOGLE_CLOUD_API_KEY"}, Note: "Or Application Default Credentials plus GOOGLE_CLOUD_PROJECT and GOOGLE_CLOUD_LOCATION."},
	{ID: "amazon-bedrock", Name: "Amazon Bedrock", APIKey: true, OAuth: false, EnvVars: []string{"AWS_BEARER_TOKEN_BEDROCK"}, Note: "Or ambient AWS credentials (profile, IAM keys, SSO)."},
	{ID: "mistral", Name: "Mistral", APIKey: true, OAuth: false, EnvVars: []string{"MISTRAL_API_KEY"}, Note: ""},
	{ID: "groq", Name: "Groq", APIKey: true, OAuth: false, EnvVars: []string{"GROQ_API_KEY"}, Note: ""},
	{ID: "cerebras", Name: "Cerebras", APIKey: true, OAuth: false, EnvVars: []string{"CEREBRAS_API_KEY"}, Note: ""},
	{ID: "cloudflare-ai-gateway", Name: "Cloudflare AI Gateway", APIKey: true, OAuth: false, EnvVars: []string{"CLOUDFLARE_API_KEY"}, Note: "Also needs CLOUDFLARE_ACCOUNT_ID and CLOUDFLARE_GATEWAY_ID."},
	{ID: "cloudflare-workers-ai", Name: "Cloudflare Workers AI", APIKey: true, OAuth: false, EnvVars: []string{"CLOUDFLARE_API_KEY"}, Note: "Also needs CLOUDFLARE_ACCOUNT_ID."},
	{ID: "vercel-ai-gateway", Name: "Vercel AI Gateway", APIKey: true, OAuth: false, EnvVars: []string{"AI_GATEWAY_API_KEY"}, Note: ""},
	{ID: "zai", Name: "ZAI Coding Plan (Global)", APIKey: true, OAuth: false, EnvVars: []string{"ZAI_API_KEY"}, Note: ""},
	{ID: "zai-coding-cn", Name: "ZAI Coding Plan (China)", APIKey: true, OAuth: false, EnvVars: []string{"ZAI_CODING_CN_API_KEY"}, Note: ""},
	{ID: "opencode", Name: "OpenCode Zen", APIKey: true, OAuth: false, EnvVars: []string{"OPENCODE_API_KEY"}, Note: ""},
	{ID: "opencode-go", Name: "OpenCode Go", APIKey: true, OAuth: false, EnvVars: []string{"OPENCODE_API_KEY"}, Note: ""},
	{ID: "huggingface", Name: "Hugging Face", APIKey: true, OAuth: false, EnvVars: []string{"HF_TOKEN"}, Note: ""},
	{ID: "fireworks", Name: "Fireworks", APIKey: true, OAuth: false, EnvVars: []string{"FIREWORKS_API_KEY"}, Note: ""},
	{ID: "together", Name: "Together AI", APIKey: true, OAuth: false, EnvVars: []string{"TOGETHER_API_KEY"}, Note: ""},
	{ID: "baseten", Name: "Baseten", APIKey: true, OAuth: false, EnvVars: []string{"BASETEN_API_KEY"}, Note: ""},
	{ID: "minimax", Name: "MiniMax", APIKey: true, OAuth: false, EnvVars: []string{"MINIMAX_API_KEY"}, Note: ""},
	{ID: "minimax-cn", Name: "MiniMax (China)", APIKey: true, OAuth: false, EnvVars: []string{"MINIMAX_CN_API_KEY"}, Note: ""},
	{ID: "moonshotai", Name: "Moonshot AI", APIKey: true, OAuth: false, EnvVars: []string{"MOONSHOT_API_KEY"}, Note: ""},
	{ID: "moonshotai-cn", Name: "Moonshot AI (China)", APIKey: true, OAuth: false, EnvVars: []string{"MOONSHOT_API_KEY"}, Note: ""},
	{ID: "qwen-token-plan", Name: "Qwen Token Plan", APIKey: true, OAuth: false, EnvVars: []string{"QWEN_TOKEN_PLAN_API_KEY"}, Note: ""},
	{ID: "qwen-token-plan-individual", Name: "Qwen Token Plan (Individual)", APIKey: true, OAuth: false, EnvVars: []string{"QWEN_TOKEN_PLAN_API_KEY"}, Note: ""},
	{ID: "qwen-token-plan-cn", Name: "Qwen Token Plan (China)", APIKey: true, OAuth: false, EnvVars: []string{"QWEN_TOKEN_PLAN_CN_API_KEY"}, Note: ""},
	{ID: "xiaomi", Name: "Xiaomi MiMo", APIKey: true, OAuth: false, EnvVars: []string{"XIAOMI_API_KEY"}, Note: ""},
	{ID: "xiaomi-token-plan-cn", Name: "Xiaomi MiMo Token Plan (China)", APIKey: true, OAuth: false, EnvVars: []string{"XIAOMI_TOKEN_PLAN_CN_API_KEY"}, Note: ""},
	{ID: "xiaomi-token-plan-ams", Name: "Xiaomi MiMo Token Plan (Amsterdam)", APIKey: true, OAuth: false, EnvVars: []string{"XIAOMI_TOKEN_PLAN_AMS_API_KEY"}, Note: ""},
	{ID: "xiaomi-token-plan-sgp", Name: "Xiaomi MiMo Token Plan (Singapore)", APIKey: true, OAuth: false, EnvVars: []string{"XIAOMI_TOKEN_PLAN_SGP_API_KEY"}, Note: ""},
}
