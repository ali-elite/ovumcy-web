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

func (s *PartnerAdviceService) GetAdvice(ctx context.Context, phase string, language string, skipCache bool) (string, error) {
	if s.apiKey == "" {
		return "", ErrPartnerAdviceNoAPIKey
	}

	if phase == "" {
		phase = "unknown"
	}

	cacheKey := fmt.Sprintf("%s:%s", phase, language)

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
	prompt := fmt.Sprintf("You are an empathetic AI assistant for a menstrual cycle tracking app. The user's partner is currently in the '%s' phase of their cycle. Give 1 to 2 short, practical, and empathetic tips on how the partner can support them right now. Keep it very concise (max 3 sentences total). Respond in the following language code: %s.", phase, language)

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
			"maxOutputTokens": 256,
			"temperature":     0.7,
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
