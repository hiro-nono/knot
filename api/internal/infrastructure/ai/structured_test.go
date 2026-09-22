package ai

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClient_ChatCompletionStructured_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body chatCompletionRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		if body.ResponseFormat == nil || body.ResponseFormat.Type != "json_schema" {
			t.Fatalf("response_format = %+v, want json_schema", body.ResponseFormat)
		}

		content := `{
			"title": "旅行のお知らせ",
			"sources": [
				{
					"type": "schedule",
					"key": "departure_date",
					"value": "2026-10-01",
					"status": "confirmed",
					"question": null,
					"interaction_type": null,
					"options": []
				},
				{
					"type": "schedule",
					"key": "departure_time",
					"value": "",
					"status": "undecided",
					"question": "出発時刻は何時頃を予定していますか？",
					"interaction_type": null,
					"options": []
				}
			]
		}`

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"choices": [{"message": {"role": "assistant", "content": ` + strconvQuote(content) + `}, "finish_reason": "stop"}], "usage": {}}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "sk-orca-test", server.Client())

	got, err := client.ChatCompletionStructured(context.Background(), "orcarouter/auto", BuildMessages("system", "user"))
	if err != nil {
		t.Fatalf("ChatCompletionStructured() error = %v", err)
	}

	if got.Title != "旅行のお知らせ" {
		t.Errorf("Title = %q, want %q", got.Title, "旅行のお知らせ")
	}
	if len(got.Sources) != 2 {
		t.Fatalf("len(Sources) = %d, want 2", len(got.Sources))
	}
	if got.Sources[0].Type != "schedule" || got.Sources[0].Key != "departure_date" {
		t.Errorf("Sources[0] = %+v, unexpected", got.Sources[0])
	}
	if got.Sources[0].InteractionType != nil {
		t.Errorf("Sources[0].InteractionType = %v, want nil", got.Sources[0].InteractionType)
	}
	if got.Sources[0].Question != nil {
		t.Errorf("Sources[0].Question = %v, want nil (confirmed)", got.Sources[0].Question)
	}

	if got.Sources[1].Status != "undecided" {
		t.Errorf("Sources[1].Status = %q, want %q", got.Sources[1].Status, "undecided")
	}
	if got.Sources[1].Question == nil || *got.Sources[1].Question != "出発時刻は何時頃を予定していますか？" {
		t.Errorf("Sources[1].Question = %v, want a clarifying question", got.Sources[1].Question)
	}
}

func TestClient_ChatCompletionStructured_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"choices": [{"message": {"role": "assistant", "content": "not json"}, "finish_reason": "stop"}], "usage": {}}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "sk-orca-test", server.Client())

	_, err := client.ChatCompletionStructured(context.Background(), "orcarouter/auto", BuildMessages("system", "user"))
	if err == nil {
		t.Fatal("ChatCompletionStructured() error = nil, want error")
	}
}

func strconvQuote(s string) string {
	b, _ := json.Marshal(s)
	return string(b)
}
