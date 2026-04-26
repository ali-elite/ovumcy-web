package services

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

var (
	ErrPartnerAdviceNoAPIKey = errors.New("no API key provided")
	ErrPartnerAdviceFailed   = errors.New("failed to fetch partner advice")
)

type PartnerAdviceService struct {
	apiKey string
	model  string
	client *http.Client
	cache  sync.Map
}

type cacheEntry struct {
	advice    string
	expiresAt time.Time
}

const defaultPartnerAdviceModel = "gemini-flash-latest"

type PartnerAdviceContext struct {
	Phase                 string   `json:"phase,omitempty"`
	CycleDay              int      `json:"cycle_day,omitempty"`
	MedianCycleLength     int      `json:"median_cycle_length,omitempty"`
	CompletedCycleCount   int      `json:"completed_cycle_count,omitempty"`
	AveragePeriodLength   float64  `json:"average_period_length,omitempty"`
	LastPeriodLength      int      `json:"last_period_length,omitempty"`
	NextPeriodEstimate    string   `json:"next_period_estimate,omitempty"`
	FertilityWindow       string   `json:"fertility_window,omitempty"`
	OvulationEstimate     string   `json:"ovulation_estimate,omitempty"`
	AgeGroup              string   `json:"age_group,omitempty"`
	UsageGoal             string   `json:"usage_goal,omitempty"`
	IrregularCycle        bool     `json:"irregular_cycle,omitempty"`
	UnpredictableCycle    bool     `json:"unpredictable_cycle,omitempty"`
	TodayPeriodLogged     bool     `json:"today_period_logged,omitempty"`
	TodayFlow             string   `json:"today_flow,omitempty"`
	TodayMood             string   `json:"today_mood,omitempty"`
	TodaySymptoms         []string `json:"today_symptoms,omitempty"`
	TodayCervicalMucus    string   `json:"today_cervical_mucus,omitempty"`
	TodayBBT              string   `json:"today_bbt,omitempty"`
	TodayCycleFactors     []string `json:"today_cycle_factors,omitempty"`
	PartnerSupportContext string   `json:"partner_support_context,omitempty"`
}

func NewPartnerAdviceService(apiKey string) *PartnerAdviceService {
	return &PartnerAdviceService{
		apiKey: strings.TrimSpace(apiKey),
		model:  defaultPartnerAdviceModel,
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

func (s *PartnerAdviceService) WithModel(model string) *PartnerAdviceService {
	if s == nil {
		return s
	}
	if normalized := normalizePartnerAdviceModel(model); normalized != "" {
		s.model = normalized
	}
	return s
}

func (s *PartnerAdviceService) GetAdvice(ctx context.Context, adviceContext PartnerAdviceContext, language string, skipCache bool) (string, error) {
	if s.apiKey == "" {
		return "", ErrPartnerAdviceNoAPIKey
	}

	if adviceContext.Phase == "" {
		adviceContext.Phase = "unknown"
	}

	cacheKey := fmt.Sprintf("%s:%s:%s", adviceContext.Phase, language, adviceContext.cacheFingerprint())

	if skipCache {
		s.cache.Delete(cacheKey)
	}

	if !skipCache {
		if entry, ok := s.cache.Load(cacheKey); ok {
			ce := entry.(cacheEntry)
			if time.Now().Before(ce.expiresAt) {
				return ce.advice, nil
			}
			s.cache.Delete(cacheKey)
		}
	}

	// Generate prompt
	prompt := adviceContext.prompt(language)

	// Build request payload for the Gemini generateContent API.
	reqPayload := map[string]interface{}{
		"contents": []interface{}{
			map[string]interface{}{
				"parts": []interface{}{
					map[string]interface{}{
						"text": prompt,
					},
				},
			},
		},
		"generationConfig": map[string]interface{}{
			"maxOutputTokens": 900,
			"temperature":     0.85,
			"thinkingConfig": map[string]interface{}{
				"thinkingBudget": 0,
			},
		},
	}

	bodyBytes, err := json.Marshal(reqPayload)
	if err != nil {
		return "", fmt.Errorf("%w: failed to marshal payload: %v", ErrPartnerAdviceFailed, err)
	}

	endpoint := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent", url.PathEscape(normalizePartnerAdviceModel(s.model)))
	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewReader(bodyBytes))
	if err != nil {
		return "", fmt.Errorf("%w: failed to create request: %v", ErrPartnerAdviceFailed, err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-goog-api-key", s.apiKey)

	resp, err := s.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("%w: request failed: %v", ErrPartnerAdviceFailed, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyStr, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("%w: API returned %d: %s", ErrPartnerAdviceFailed, resp.StatusCode, string(bodyStr))
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
		return "", fmt.Errorf("%w: failed to decode response: %v", ErrPartnerAdviceFailed, err)
	}

	if len(result.Candidates) == 0 || len(result.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("%w: no content generated", ErrPartnerAdviceFailed)
	}

	advice := result.Candidates[0].Content.Parts[0].Text

	// Cache the result for 24 hours
	s.cache.Store(cacheKey, cacheEntry{
		advice:    advice,
		expiresAt: time.Now().Add(24 * time.Hour),
	})

	return advice, nil
}

func normalizePartnerAdviceModel(model string) string {
	model = strings.TrimSpace(model)
	model = strings.TrimPrefix(model, "models/")
	return model
}

func (context PartnerAdviceContext) cacheFingerprint() string {
	payload, err := json.Marshal(context)
	if err != nil {
		return ""
	}
	return string(payload)
}

func (context PartnerAdviceContext) prompt(language string) string {
	payload, _ := json.MarshalIndent(context, "", "  ")
	return fmt.Sprintf(`You are an empathetic AI assistant inside a menstrual cycle tracking app.

Write advice for the partner, not for the cycle owner. Use the private health context below, but do not mention names, emails, account identifiers, exact notes, or any identity details. The context has already been sanitized.

Health and cycle context:
%s

Instructions:
- Respond in language code: %s.
- Make the answer more customized to the current cycle phase, symptoms, mood, cycle day, fertility context, usage goal, age group, and any logged signals.
- Output around 700 tokens. A practical range is 500-650 words depending on the language.
- Give warm, specific partner actions: emotional support, practical help, communication prompts, comfort ideas, and what to avoid.
- If fertility or trying/avoiding pregnancy context appears, keep advice supportive and non-alarming.
- Do not diagnose, predict pregnancy, or present medical certainty. Include a gentle safety note only if symptoms sound intense or unusual.
- Keep the tone calm, modern, and personal. Avoid generic filler and avoid saying you know the person's private identity.`, string(payload), language)
}
