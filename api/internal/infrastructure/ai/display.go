package ai

import (
	"context"
	"encoding/json"
	"fmt"
)

// SourceOfTruthSource は表示生成のためにAIへ渡す、確定済みSourceの1件を表す。
type SourceOfTruthSource struct {
	Type            string   `json:"type"`
	Key             string   `json:"key"`
	Value           string   `json:"value"`
	InteractionType *string  `json:"interaction_type,omitempty"`
	Options         []string `json:"options,omitempty"`
}

// SourceOfTruth は表示生成のためにAIへ渡す、確定済みInformationの全体を表す。
type SourceOfTruth struct {
	Title   string                `json:"title"`
	Sources []SourceOfTruthSource `json:"sources"`
}

// DisplayContent は受信者向けに最適化された表示内容を表す。
type DisplayContent struct {
	Title string `json:"title"`
	Body  string `json:"body"`
}

const displaySchemaName = "display_content"

func displayContentSchema() map[string]any {
	return map[string]any{
		"type":                 "object",
		"additionalProperties": false,
		"required":             []string{"title", "body"},
		"properties": map[string]any{
			"title": map[string]any{"type": "string"},
			"body":  map[string]any{"type": "string"},
		},
	}
}

func displayResponseFormat() *responseFormat {
	return &responseFormat{
		Type: "json_schema",
		JSONSchema: &jsonSchemaSpec{
			Name:   displaySchemaName,
			Strict: true,
			Schema: displayContentSchema(),
		},
	}
}

// ChatCompletionDisplay はmessagesをmodelに送信し、
// 受信者向けに最適化された表示内容(DisplayContent)を取得する。
func (c *Client) ChatCompletionDisplay(ctx context.Context, model string, messages []Message) (*DisplayContent, error) {
	body, err := json.Marshal(chatCompletionRequest{
		Model:          model,
		Messages:       messages,
		ResponseFormat: displayResponseFormat(),
	})
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	req, err := c.newRequest(ctx, body)
	if err != nil {
		return nil, err
	}

	respBody, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	chatResp, err := parseChatCompletionResponse(respBody)
	if err != nil {
		return nil, err
	}

	var content DisplayContent
	if err := json.Unmarshal([]byte(chatResp.Content), &content); err != nil {
		return nil, fmt.Errorf("unmarshal display content: %w", err)
	}

	return &content, nil
}

// BuildDisplaySystemPrompt はSource of Truthと受信者のPreferenceから、
// 表示生成用のシステムプロンプトを組み立てる。
//
// Source of Truthの内容(事実)は変更・省略させず、表現の平易さ・情報量・言語などの
// 「見せ方」だけをPreferenceに合わせて最適化させる。
func BuildDisplaySystemPrompt(sot SourceOfTruth, preference map[string]string) (string, error) {
	sotJSON, err := json.Marshal(sot)
	if err != nil {
		return "", fmt.Errorf("marshal source of truth: %w", err)
	}

	preferenceJSON, err := json.Marshal(preference)
	if err != nil {
		return "", fmt.Errorf("marshal preference: %w", err)
	}

	return fmt.Sprintf(`あなたは、確定済みの情報(Source of Truth)を受信者向けの表示に変換するアシスタントです。

以下のSource of Truthに含まれる情報(事実)は一切変更・省略・追加してはいけません。
表現の平易さ、情報量、言語などの「見せ方」だけを、受信者のPreferenceに合わせて最適化してください。
Preferenceに無い項目については、標準的でわかりやすい見せ方にしてください。

titleとbody(表示本文)をJSONで出力してください。

# Source of Truth
%s

# 受信者のPreference
%s`, string(sotJSON), string(preferenceJSON)), nil
}
