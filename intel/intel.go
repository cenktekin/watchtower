package intel

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
	"watchtower/feeds"
	"watchtower/weather"
)

type Provider string

const (
	ProviderGroq       Provider = "groq"
	ProviderOpenAI     Provider = "openai"
	ProviderDeepSeek   Provider = "deepseek"
	ProviderGemini     Provider = "gemini"
	ProviderClaude     Provider = "claude"
	ProviderLocal      Provider = "local"
	ProviderOllama     Provider = "ollama"
	ProviderOpenRouter Provider = "openrouter"
)

var providerDefaults = map[Provider]struct {
	endpoint     string
	defaultModel string
	authHeader   string
	authPrefix   string
}{
	ProviderGroq: {
		endpoint:     "https://api.groq.com/openai/v1/chat/completions",
		defaultModel: "llama-3.1-8b-instant",
		authHeader:   "Authorization",
		authPrefix:   "Bearer ",
	},
	ProviderOpenAI: {
		endpoint:     "https://api.openai.com/v1/chat/completions",
		defaultModel: "gpt-4o-mini",
		authHeader:   "Authorization",
		authPrefix:   "Bearer ",
	},
	ProviderDeepSeek: {
		endpoint:     "https://api.deepseek.com/v1/chat/completions",
		defaultModel: "deepseek-chat",
		authHeader:   "Authorization",
		authPrefix:   "Bearer ",
	},
	ProviderGemini: {
		endpoint:     "https://generativelanguage.googleapis.com/v1beta/models",
		defaultModel: "gemini-1.5-flash",
		authHeader:   "X-Goog-Api-Key",
		authPrefix:   "",
	},
	ProviderClaude: {
		endpoint:     "https://api.anthropic.com/v1/messages",
		defaultModel: "claude-3-haiku-20240307",
		authHeader:   "x-api-key",
		authPrefix:   "",
	},
	ProviderLocal: {
		endpoint:     "http://localhost:11434/v1/chat/completions",
		defaultModel: "llama3",
		authHeader:   "Authorization",
		authPrefix:   "Bearer ",
	},
	ProviderOllama: {
		endpoint:     "http://localhost:11434/v1/chat/completions",
		defaultModel: "llama3",
		authHeader:   "",
		authPrefix:   "",
	},
	ProviderOpenRouter: {
		endpoint:     "https://openrouter.ai/api/v1/chat/completions",
		defaultModel: "meta-llama/llama-3-8b-instruct:free",
		authHeader:   "Authorization",
		authPrefix:   "Bearer ",
	},
}

type LLMConfig struct {
	Provider Provider
	APIKey   string
	Model    string
}

func (c LLMConfig) Endpoint() string {
	p := providerDefaults[c.Provider]
	if c.Model == "" {
		return p.endpoint
	}
	if c.Provider == ProviderGemini {
		return p.endpoint + "/" + c.Model + ":generateContent"
	}
	return p.endpoint
}

func (c LLMConfig) ModelName() string {
	if c.Model != "" {
		return c.Model
	}
	return providerDefaults[c.Provider].defaultModel
}

func (c LLMConfig) AuthHeader() string {
	return providerDefaults[c.Provider].authHeader
}

func (c LLMConfig) AuthValue() string {
	p := providerDefaults[c.Provider]
	if p.authHeader == "" {
		return ""
	}
	return p.authPrefix + c.APIKey
}

// CountryRisk holds a risk score for one country
type CountryRisk struct {
	Country string
	Score   int    // 0–100
	Reason  string // one short phrase
}

// Brief holds an AI-generated intelligence summary
type Brief struct {
	Summary      string
	KeyThreats   []string
	CountryRisks []CountryRisk
	GeneratedAt  time.Time
	Model        string
}

// LocalBrief holds an AI-generated summary of local news and weather
type LocalBrief struct {
	Summary     string
	GeneratedAt time.Time
	Model       string
}

var httpClient = &http.Client{Timeout: 120 * time.Second}

// GenerateBrief calls the configured LLM to synthesize a brief, summary, and country risk scores
func GenerateBrief(ctx context.Context, cfg LLMConfig, items []feeds.NewsItem) (*Brief, error) {
	// Skip API key check for local providers (Ollama, Local)
	if cfg.APIKey == "" && cfg.Provider != ProviderOllama && cfg.Provider != ProviderLocal {
		return &Brief{
			Summary:     "LLM_API_KEY ayarlanmamış. AI özetlerini etkinleştirmek için ~/.config/watchtower/config.yaml dosyasına ekleyin.",
			GeneratedAt: time.Now(),
			Model:       "none",
		}, nil
	}

	if len(items) == 0 {
		return &Brief{
			Summary:     "Özetlenecek haber bulunamadı.",
			GeneratedAt: time.Now(),
		}, nil
	}

	// Build headline list (top 40 by severity)
	limit := 40
	if len(items) < limit {
		limit = len(items)
	}
	var sb strings.Builder
	for i, item := range items[:limit] {
		sb.WriteString(fmt.Sprintf("%d. [%s] %s (%s)\n",
			i+1, item.ThreatLevel.String(), item.Title, item.Source))
	}

	prompt := fmt.Sprintf(`Sen bir saha istihbarat subayısın, gazeteci değil.
Görevin, sahadaki operatörler ve tüccarlar için fiziksel ve ekonomik riskleri değerlendirmek.
Türkçe yanıt ver.

ÖZET:
<3-4 cümle. En tehlikeli gelişmeyle başla. Gereksiz detay yok.>

TEHDİTLER:
• <spesifik olay, konum, etki>
• <spesifik olay, konum, etki>
• <spesifik olay, konum, etki>
• <spesifik olay, konum, etki>
• <spesifik olay, konum, etki>

ÜLKE_RİSKLERİ:
<ÜlkeAdı>|<0-100>|<ne yapmalı: örn. "limanlardan kaçın", "gecikme beklenir", "operasyon güvenli">
<ÜlkeAdı>|<0-100>|<aksiyon ifadesi>
<ÜlkeAdı>|<0-100>|<aksiyon ifadesi>
<ÜlkeAdı>|<0-100>|<aksiyon ifadesi>
<ÜlkeAdı>|<0-100>|<aksiyon ifadesi>
<ÜlkeAdı>|<0-100>|<aksiyon ifadesi>
<ÜlkeAdı>|<0-100>|<aksiyon ifadesi>
<ÜlkeAdı>|<0-100>|<aksiyon ifadesi>

RİSK ÇAPALARI (rehber olarak kullan):
- 100: Aktif savaş, sınırlar kapalı, ticaret yok
- 75: Silahlı çatışma, tedarik zinciri koptu
- 50: Sıkıyönetim, grevler, ticaret aksadı
- 25: Siyasi huzursuzluk, işler normal
- 0: Tamamen stabil

KURALLAR:
- Diplomatik dil YASAK ("endişe", "gerginlik", "gelişmeler" yok)
- Somut olaylar yaz: bomba, grev, darbe, elektrik kesintisi, iflas, işgal
- Risk puanı = "Bugün buraya sevkiyat yapar mıyım?"
- Sebep = 2-4 kelime, aksiyon odaklı, gereksiz detay yok
- Markdown yok, giriş yok, açıklama yok

MANŞETLER:
%s`, sb.String())

	if cfg.Provider == ProviderClaude {
		return generateClaudeBrief(ctx, cfg, prompt)
	}
	if cfg.Provider == ProviderGemini {
		return generateGeminiBrief(ctx, cfg, prompt)
	}
	return generateOpenAICompatibleBrief(ctx, cfg, prompt)
}

// GenerateLocalBrief calls the configured LLM to synthesize a local news and weather summary
func GenerateLocalBrief(ctx context.Context, cfg LLMConfig, city string, items []feeds.NewsItem, cond *weather.Conditions, forecast []weather.DayForecast) (*LocalBrief, error) {
	// Skip API key check for local providers (Ollama, Local)
	if cfg.APIKey == "" && cfg.Provider != ProviderOllama && cfg.Provider != ProviderLocal {
		return &LocalBrief{
			Summary:     "LLM_API_KEY ayarlanmamış. AI özetlerini etkinleştirmek için ~/.config/watchtower/config.yaml dosyasına ekleyin.",
			GeneratedAt: time.Now(),
			Model:       "none",
		}, nil
	}

	var sb strings.Builder

	// Build local news headline list (top 20)
	sb.WriteString("YEREL HABER MANŞETLERİ:\n")
	limit := 20
	if len(items) > limit {
		items = items[:limit]
	}
	for i, item := range items {
		sb.WriteString(fmt.Sprintf("%d. %s (%s)\n", i+1, item.Title, item.Source))
	}

	// Add current weather
	sb.WriteString("\nMEVCUT HAVA DURUMU:\n")
	if cond != nil {
		sb.WriteString(fmt.Sprintf("Konum: %s\n", cond.City))
		sb.WriteString(fmt.Sprintf("Sıcaklık: %.1f°C (hissedilen %.1f°C)\n", cond.TempC, cond.FeelsLikeC))
		sb.WriteString(fmt.Sprintf("Durum: %s %s\n", cond.Icon, cond.Description))
		sb.WriteString(fmt.Sprintf("Nem: %d%%, Rüzgar: %.0f km/h, UV: %.0f\n", cond.Humidity, cond.WindSpeedKmh, cond.UVIndex))
	}

	// Add forecast
	sb.WriteString("\nHAVA TAHMİNİ (önümüzdeki 5 gün):\n")
	for i, f := range forecast {
		if i >= 5 {
			break
		}
		sb.WriteString(fmt.Sprintf("- %s: %s %s, High: %.0f°C, Low: %.0f°C, Rain: %.1fmm\n",
			f.Date.Format("Mon Jan 02"), f.Icon, f.Desc, f.MaxTempC, f.MinTempC, f.RainMM))
	}

	prompt := fmt.Sprintf(`Sen yerel haber ve hava durumu analizcisisin. %s için bu bilgiyi 2-3 cümlede özetle.
Odaklan:
1. Herhangi bir dikkat çekici yerel haber
2. Mevcut hava durumu ve önümüzdeki günler için hava durumu değerlendirmesi
3. Haberlerin kısa özeti

Tam olarak bu formatta yanıt ver, ekstra metin yok:

SUMMARY:
<2-3 cümlelik özet>

Kurallar:
- Kısa ve pratik tut
- Markdown formatlama yok
- En önemli bilgiyle başla
- Asla 'DATA'yı olduğu gibi geri gönderme, her zaman açıkla

DATA:
%s`, city, sb.String())

	if cfg.Provider == ProviderClaude {
		return generateClaudeLocalBrief(ctx, cfg, prompt)
	}
	if cfg.Provider == ProviderGemini {
		return generateGeminiLocalBrief(ctx, cfg, prompt)
	}
	return generateOpenAICompatibleLocalBrief(ctx, cfg, prompt)
}

func generateOpenAICompatibleBrief(ctx context.Context, cfg LLMConfig, prompt string) (*Brief, error) {
	body := map[string]interface{}{
		"model":       cfg.ModelName(),
		"temperature": 0,
		"max_tokens":  700,
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
	}

	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", cfg.Endpoint(), bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, err
	}
	if cfg.AuthHeader() != "" {
		req.Header.Set(cfg.AuthHeader(), cfg.AuthValue())
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%s request failed: %w", cfg.Provider, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("%s HTTP %d", cfg.Provider, resp.StatusCode)
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Model string `json:"model"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding %s response: %w", cfg.Provider, err)
	}
	if len(result.Choices) == 0 {
		return nil, fmt.Errorf("no response from %s", cfg.Provider)
	}

	summary, threats, risks := parseBriefResponse(result.Choices[0].Message.Content)

	return &Brief{
		Summary:      summary,
		KeyThreats:   threats,
		CountryRisks: risks,
		GeneratedAt:  time.Now(),
		Model:        result.Model,
	}, nil
}

func generateOpenAICompatibleLocalBrief(ctx context.Context, cfg LLMConfig, prompt string) (*LocalBrief, error) {
	body := map[string]interface{}{
		"model":       cfg.ModelName(),
		"temperature": 0,
		"max_tokens":  300,
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
	}

	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", cfg.Endpoint(), bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, err
	}
	if cfg.AuthHeader() != "" {
		req.Header.Set(cfg.AuthHeader(), cfg.AuthValue())
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%s request failed: %w", cfg.Provider, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("%s HTTP %d", cfg.Provider, resp.StatusCode)
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Model string `json:"model"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding %s response: %w", cfg.Provider, err)
	}
	if len(result.Choices) == 0 {
		return nil, fmt.Errorf("no response from %s", cfg.Provider)
	}

	summary := parseLocalBriefResponse(result.Choices[0].Message.Content)

	return &LocalBrief{
		Summary:     summary,
		GeneratedAt: time.Now(),
		Model:       result.Model,
	}, nil
}

func generateClaudeBrief(ctx context.Context, cfg LLMConfig, prompt string) (*Brief, error) {
	body := map[string]interface{}{
		"model":       cfg.ModelName(),
		"max_tokens":  700,
		"temperature": 0,
		"system":      "Sen bir jeopolitik istihbarat analizcisisin.",
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
	}

	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", cfg.Endpoint(), bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, err
	}
	if cfg.AuthHeader() != "" {
		req.Header.Set(cfg.AuthHeader(), cfg.AuthValue())
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("claude request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("claude HTTP %d", resp.StatusCode)
	}

	var result struct {
		Content []struct {
			Text string `json:"text"`
		} `json:"content"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding claude response: %w", err)
	}
	if len(result.Content) == 0 {
		return nil, fmt.Errorf("no response from claude")
	}

	summary, threats, risks := parseBriefResponse(result.Content[0].Text)

	return &Brief{
		Summary:      summary,
		KeyThreats:   threats,
		CountryRisks: risks,
		GeneratedAt:  time.Now(),
		Model:        cfg.ModelName(),
	}, nil
}

func generateClaudeLocalBrief(ctx context.Context, cfg LLMConfig, prompt string) (*LocalBrief, error) {
	body := map[string]interface{}{
		"model":       cfg.ModelName(),
		"max_tokens":  300,
		"temperature": 0,
		"system":      "Sen bir yerel haber ve hava durumu analizcisisin.",
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
	}

	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", cfg.Endpoint(), bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, err
	}
	if cfg.AuthHeader() != "" {
		req.Header.Set(cfg.AuthHeader(), cfg.AuthValue())
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("claude request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("claude HTTP %d", resp.StatusCode)
	}

	var result struct {
		Content []struct {
			Text string `json:"text"`
		} `json:"content"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding claude response: %w", err)
	}
	if len(result.Content) == 0 {
		return nil, fmt.Errorf("no response from claude")
	}

	summary := parseLocalBriefResponse(result.Content[0].Text)

	return &LocalBrief{
		Summary:     summary,
		GeneratedAt: time.Now(),
		Model:       cfg.ModelName(),
	}, nil
}

func generateGeminiBrief(ctx context.Context, cfg LLMConfig, prompt string) (*Brief, error) {
	body := map[string]interface{}{
		"contents": []map[string]interface{}{
			{
				"parts": []map[string]string{
					{"text": prompt},
				},
			},
		},
		"generationConfig": map[string]interface{}{
			"temperature":     0,
			"maxOutputTokens": 700,
		},
	}

	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", cfg.Endpoint(), bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, err
	}
	if cfg.AuthHeader() != "" {
		req.Header.Set(cfg.AuthHeader(), cfg.AuthValue())
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("gemini request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("gemini HTTP %d", resp.StatusCode)
	}

	var result struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding gemini response: %w", err)
	}
	if len(result.Candidates) == 0 || len(result.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("no response from gemini")
	}

	summary, threats, risks := parseBriefResponse(result.Candidates[0].Content.Parts[0].Text)

	return &Brief{
		Summary:      summary,
		KeyThreats:   threats,
		CountryRisks: risks,
		GeneratedAt:  time.Now(),
		Model:        cfg.ModelName(),
	}, nil
}

func generateGeminiLocalBrief(ctx context.Context, cfg LLMConfig, prompt string) (*LocalBrief, error) {
	body := map[string]interface{}{
		"contents": []map[string]interface{}{
			{
				"parts": []map[string]string{
					{"text": prompt},
				},
			},
		},
		"generationConfig": map[string]interface{}{
			"temperature":     0,
			"maxOutputTokens": 300,
		},
	}

	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", cfg.Endpoint(), bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, err
	}
	if cfg.AuthHeader() != "" {
		req.Header.Set(cfg.AuthHeader(), cfg.AuthValue())
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("gemini request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("gemini HTTP %d", resp.StatusCode)
	}

	var result struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding gemini response: %w", err)
	}
	if len(result.Candidates) == 0 || len(result.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("no response from gemini")
	}

	summary := parseLocalBriefResponse(result.Candidates[0].Content.Parts[0].Text)

	return &LocalBrief{
		Summary:     summary,
		GeneratedAt: time.Now(),
		Model:       cfg.ModelName(),
	}, nil
}

func parseLocalBriefResponse(content string) string {
	lines := strings.Split(content, "\n")
	inSummary := false
	var summaryLines []string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "SUMMARY:") {
			inSummary = true
			continue
		}
		if inSummary && trimmed != "" {
			summaryLines = append(summaryLines, trimmed)
		}
	}

	if len(summaryLines) > 0 {
		return strings.Join(summaryLines, " ")
	}
	return strings.TrimSpace(content)
}

func parseBriefResponse(content string) (string, []string, []CountryRisk) {
	var summary string
	var threats []string
	var risks []CountryRisk

	// Split into sections
	sections := map[string]string{}
	current := ""
	var buf strings.Builder

	for _, line := range strings.Split(content, "\n") {
		trimmed := strings.TrimSpace(line)
		switch trimmed {
		case "ÖZET:", "SUMMARY:":
			if current != "" {
				sections[current] = strings.TrimSpace(buf.String())
			}
			current = "SUMMARY"
			buf.Reset()
		case "TEHDİTLER:", "THREATS:":
			if current != "" {
				sections[current] = strings.TrimSpace(buf.String())
			}
			current = "THREATS"
			buf.Reset()
		case "ÜLKE_RİSKLERİ:", "COUNTRY_RISKS:":
			if current != "" {
				sections[current] = strings.TrimSpace(buf.String())
			}
			current = "COUNTRY_RISKS"
			buf.Reset()
		default:
			if current != "" {
				buf.WriteString(line + "\n")
			}
		}
	}
	if current != "" {
		sections[current] = strings.TrimSpace(buf.String())
	}

	// Parse SUMMARY
	summary = sections["SUMMARY"]

	// Parse THREATS
	for _, line := range strings.Split(sections["THREATS"], "\n") {
		line = strings.TrimSpace(line)
		line = strings.TrimPrefix(line, "•")
		line = strings.TrimPrefix(line, "-")
		line = strings.TrimPrefix(line, "*")
		line = strings.TrimSpace(line)
		if line != "" {
			threats = append(threats, line)
		}
	}

	// Parse COUNTRY_RISKS  format: Country|score|reason
	for _, line := range strings.Split(sections["COUNTRY_RISKS"], "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "|", 3)
		if len(parts) < 2 {
			continue
		}
		country := strings.TrimSpace(parts[0])
		score, err := strconv.Atoi(strings.TrimSpace(parts[1]))
		if err != nil || country == "" {
			continue
		}
		reason := ""
		if len(parts) == 3 {
			reason = strings.TrimSpace(parts[2])
		}
		risks = append(risks, CountryRisk{
			Country: country,
			Score:   clamp(score, 0, 100),
			Reason:  reason,
		})
	}

	return summary, threats, risks
}

func clamp(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
