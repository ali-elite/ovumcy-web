package services

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestPartnerAdviceServiceDefaultsToCurrentFlashModel(t *testing.T) {
	service := NewPartnerAdviceService(" key ")

	if service.apiKey != "key" {
		t.Fatalf("expected api key to be trimmed, got %q", service.apiKey)
	}
	if service.model != defaultPartnerAdviceModel {
		t.Fatalf("expected default model %q, got %q", defaultPartnerAdviceModel, service.model)
	}
}

func TestPartnerAdviceServiceWithModelAcceptsOptionalModelsPrefix(t *testing.T) {
	service := NewPartnerAdviceService("key").WithModel(" models/gemini-flash-latest ")

	if service.model != "gemini-flash-latest" {
		t.Fatalf("expected normalized model, got %q", service.model)
	}
}

func TestPartnerAdviceServiceRequestsLongCustomizedAdviceWithoutThinking(t *testing.T) {
	transport := roundTripFunc(func(request *http.Request) (*http.Response, error) {
		body, err := io.ReadAll(request.Body)
		if err != nil {
			t.Fatalf("read request body: %v", err)
		}

		payload := map[string]any{}
		if err := json.Unmarshal(body, &payload); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		config, ok := payload["generationConfig"].(map[string]any)
		if !ok {
			t.Fatalf("expected generationConfig in payload %#v", payload)
		}
		if got := config["maxOutputTokens"]; got != float64(900) {
			t.Fatalf("expected maxOutputTokens 900, got %#v", got)
		}
		thinkingConfig, ok := config["thinkingConfig"].(map[string]any)
		if !ok {
			t.Fatalf("expected thinkingConfig in payload %#v", config)
		}
		if got := thinkingConfig["thinkingBudget"]; got != float64(0) {
			t.Fatalf("expected thinkingBudget 0, got %#v", got)
		}
		contents := payload["contents"].([]any)
		parts := contents[0].(map[string]any)["parts"].([]any)
		prompt := parts[0].(map[string]any)["text"].(string)
		for _, expected := range []string{"luteal", "Cramps", "age 20-35", "around 700 tokens"} {
			if !strings.Contains(prompt, expected) {
				t.Fatalf("expected prompt to contain %q, got %s", expected, prompt)
			}
		}

		return &http.Response{
			StatusCode: http.StatusOK,
			Body: io.NopCloser(strings.NewReader(`{
				"candidates": [
					{"content": {"parts": [{"text": "Offer warmth and patience."}]}}
				]
			}`)),
			Header: make(http.Header),
		}, nil
	})
	service := NewPartnerAdviceService("key")
	service.client = &http.Client{Transport: transport}

	advice, err := service.GetAdvice(context.Background(), PartnerAdviceContext{
		Phase:     "luteal",
		CycleDay:  22,
		AgeGroup:  "age 20-35",
		UsageGoal: "cycle health",
		TodayMood: "2/5",
		TodaySymptoms: []string{
			"Cramps",
		},
	}, "en", false)
	if err != nil {
		t.Fatalf("GetAdvice returned error: %v", err)
	}
	if advice != "Offer warmth and patience." {
		t.Fatalf("unexpected advice %q", advice)
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (fn roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return fn(request)
}
