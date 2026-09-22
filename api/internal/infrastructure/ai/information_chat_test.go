package ai

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestBuildPreferenceChatSystemPrompt(t *testing.T) {
	sot := SourceOfTruth{Title: "旅行のお知らせ"}

	got, err := BuildPreferenceChatSystemPrompt(sot, map[string]string{"language": "ja"})
	if err != nil {
		t.Fatalf("BuildPreferenceChatSystemPrompt() error = %v", err)
	}

	for _, want := range []string{"旅行のお知らせ", "language", "is_persistent", "confirmation_question"} {
		if !strings.Contains(got, want) {
			t.Errorf("BuildPreferenceChatSystemPrompt() does not contain %q:\n%s", want, got)
		}
	}
}

func TestClient_ChatCompletionPreferenceChat_Persistent(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		content := `{
			"display": {"title": "お知らせ", "body": "かんたんな ほんぶんです"},
			"is_persistent": true,
			"preference_key": "reading_level",
			"preference_value": "easy",
			"confirmation_question": null
		}`
		_, _ = w.Write([]byte(`{"choices": [{"message": {"role": "assistant", "content": ` + strconvQuote(content) + `}, "finish_reason": "stop"}], "usage": {}}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "sk-orca-test", server.Client())

	got, err := client.ChatCompletionPreferenceChat(context.Background(), "opus", BuildMessages("system", "もっと簡単にして"))
	if err != nil {
		t.Fatalf("ChatCompletionPreferenceChat() error = %v", err)
	}

	if got.IsPersistent == nil || !*got.IsPersistent {
		t.Errorf("IsPersistent = %v, want true", got.IsPersistent)
	}
	if got.PreferenceKey == nil || *got.PreferenceKey != "reading_level" {
		t.Errorf("PreferenceKey = %v, want \"reading_level\"", got.PreferenceKey)
	}
	if got.ConfirmationQuestion != nil {
		t.Errorf("ConfirmationQuestion = %v, want nil", got.ConfirmationQuestion)
	}
}

func TestClient_ChatCompletionPreferenceChat_NeedsConfirmation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		content := `{
			"display": {"title": "お知らせ", "body": "本文です"},
			"is_persistent": null,
			"preference_key": null,
			"preference_value": null,
			"confirmation_question": "今後も常にこの設定にしますか？"
		}`
		_, _ = w.Write([]byte(`{"choices": [{"message": {"role": "assistant", "content": ` + strconvQuote(content) + `}, "finish_reason": "stop"}], "usage": {}}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "sk-orca-test", server.Client())

	got, err := client.ChatCompletionPreferenceChat(context.Background(), "opus", BuildMessages("system", "簡単にして"))
	if err != nil {
		t.Fatalf("ChatCompletionPreferenceChat() error = %v", err)
	}

	if got.IsPersistent != nil {
		t.Errorf("IsPersistent = %v, want nil", got.IsPersistent)
	}
	if got.ConfirmationQuestion == nil {
		t.Error("ConfirmationQuestion = nil, want a question")
	}
}
