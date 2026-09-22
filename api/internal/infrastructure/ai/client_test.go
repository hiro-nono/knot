package ai

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestClient_ChatCompletion_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != chatCompletionsPath {
			t.Errorf("path = %s, want %s", r.URL.Path, chatCompletionsPath)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer sk-orca-test" {
			t.Errorf("Authorization header = %q, want %q", got, "Bearer sk-orca-test")
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"choices": [{"message": {"role": "assistant", "content": "hello"}, "finish_reason": "stop"}],
			"usage": {"prompt_tokens": 1, "completion_tokens": 2, "total_tokens": 3}
		}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "sk-orca-test", server.Client())

	got, err := client.ChatCompletion(context.Background(), "orcarouter/auto", BuildMessages("system prompt", "user input"))
	if err != nil {
		t.Fatalf("ChatCompletion() error = %v", err)
	}

	if got.Content != "hello" {
		t.Errorf("Content = %q, want %q", got.Content, "hello")
	}
	if got.FinishReason != "stop" {
		t.Errorf("FinishReason = %q, want %q", got.FinishReason, "stop")
	}
	if got.Usage.TotalTokens != 3 {
		t.Errorf("Usage.TotalTokens = %d, want 3", got.Usage.TotalTokens)
	}
}

func TestClient_ChatCompletion_NonOKStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error": "invalid api key"}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "invalid-key", server.Client())

	_, err := client.ChatCompletion(context.Background(), "orcarouter/auto", BuildMessages("system", "user"))
	if err == nil {
		t.Fatal("ChatCompletion() error = nil, want error")
	}
	if !strings.Contains(err.Error(), "401") {
		t.Errorf("error = %v, want it to mention status 401", err)
	}
}

func TestClient_ChatCompletion_NoChoices(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"choices": [], "usage": {}}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "sk-orca-test", server.Client())

	_, err := client.ChatCompletion(context.Background(), "orcarouter/auto", BuildMessages("system", "user"))
	if err == nil {
		t.Fatal("ChatCompletion() error = nil, want error")
	}
}
