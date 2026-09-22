package ai

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestBuildComparisonSystemPrompt(t *testing.T) {
	sot := SourceOfTruth{Title: "旅行のお知らせ"}

	got, err := BuildComparisonSystemPrompt(sot, map[string]string{"language": "ja"}, "reading_level")
	if err != nil {
		t.Fatalf("BuildComparisonSystemPrompt() error = %v", err)
	}

	for _, want := range []string{"旅行のお知らせ", "language", "reading_level", "pattern_a", "pattern_b"} {
		if !strings.Contains(got, want) {
			t.Errorf("BuildComparisonSystemPrompt() does not contain %q:\n%s", want, got)
		}
	}
}

func TestClient_ChatCompletionComparison_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		content := `{
			"pattern_a": {"value": "easy", "title": "お知らせ", "body": "<p>かんたん</p>"},
			"pattern_b": {"value": "detailed", "title": "お知らせ", "body": "<p>くわしい説明</p>"}
		}`
		_, _ = w.Write([]byte(`{"choices": [{"message": {"role": "assistant", "content": ` + strconvQuote(content) + `}, "finish_reason": "stop"}], "usage": {}}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "sk-orca-test", server.Client())

	got, err := client.ChatCompletionComparison(context.Background(), "opus", BuildMessages("system", "user"))
	if err != nil {
		t.Fatalf("ChatCompletionComparison() error = %v", err)
	}

	if got.PatternA.Value != "easy" || got.PatternB.Value != "detailed" {
		t.Errorf("got = %+v, unexpected", got)
	}
}
