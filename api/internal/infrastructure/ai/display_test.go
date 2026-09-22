package ai

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestBuildDisplaySystemPrompt(t *testing.T) {
	sot := SourceOfTruth{
		Title: "旅行のお知らせ",
		Sources: []SourceOfTruthSource{
			{Type: "schedule", Key: "date", Value: "2026-10-01"},
		},
	}

	got, err := BuildDisplaySystemPrompt(sot, map[string]string{"reading_level": "easy"})
	if err != nil {
		t.Fatalf("BuildDisplaySystemPrompt() error = %v", err)
	}

	for _, want := range []string{"旅行のお知らせ", "2026-10-01", "reading_level", "easy"} {
		if !strings.Contains(got, want) {
			t.Errorf("BuildDisplaySystemPrompt() does not contain %q:\n%s", want, got)
		}
	}
}

func TestClient_ChatCompletionDisplay_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"choices": [{"message": {"role": "assistant", "content": "{\"title\":\"お知らせ\",\"body\":\"本文です\"}"}, "finish_reason": "stop"}],
			"usage": {}
		}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "sk-orca-test", server.Client())

	got, err := client.ChatCompletionDisplay(context.Background(), "opus", BuildMessages("system", "user"))
	if err != nil {
		t.Fatalf("ChatCompletionDisplay() error = %v", err)
	}
	if got.Title != "お知らせ" || got.Body != "本文です" {
		t.Errorf("got = %+v, unexpected", got)
	}
}

func TestClient_ChatCompletionDisplay_NoChoices(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"choices": [], "usage": {}}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "sk-orca-test", server.Client())

	_, err := client.ChatCompletionDisplay(context.Background(), "opus", BuildMessages("system", "user"))
	if err == nil {
		t.Fatal("ChatCompletionDisplay() error = nil, want error")
	}
}
