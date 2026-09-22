package ai

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestBuildInformationChatSystemPrompt(t *testing.T) {
	sot := SourceOfTruth{Title: "旅行のお知らせ"}

	got, err := BuildInformationChatSystemPrompt(sot, map[string]string{"language": "ja"})
	if err != nil {
		t.Fatalf("BuildInformationChatSystemPrompt() error = %v", err)
	}

	for _, want := range []string{"旅行のお知らせ", "language", "answer"} {
		if !strings.Contains(got, want) {
			t.Errorf("BuildInformationChatSystemPrompt() does not contain %q:\n%s", want, got)
		}
	}
}

func TestClient_ChatCompletionInformationChat_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		content := `{"answer": "集合時間は10時です。"}`
		_, _ = w.Write([]byte(`{"choices": [{"message": {"role": "assistant", "content": ` + strconvQuote(content) + `}, "finish_reason": "stop"}], "usage": {}}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "sk-orca-test", server.Client())

	got, err := client.ChatCompletionInformationChat(context.Background(), "opus", BuildMessages("system", "集合時間は何時ですか？"))
	if err != nil {
		t.Fatalf("ChatCompletionInformationChat() error = %v", err)
	}

	if got.Answer != "集合時間は10時です。" {
		t.Errorf("Answer = %q, want %q", got.Answer, "集合時間は10時です。")
	}
}
