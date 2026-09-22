package ai

import (
	"context"
	"encoding/json"
	"fmt"
)

// PreferenceChatResponse は受信者からの表示調整の要望に対するAIの応答を表す。
//
// IsPersistentがnilの場合、AIは一時的な指示か今後も適用するPreferenceかを
// 判断できておらず、ConfirmationQuestionに確認質問が入る
// (この場合、Preferenceは更新しない)。
type PreferenceChatResponse struct {
	Display              DisplayContent `json:"display"`
	IsPersistent         *bool          `json:"is_persistent"`
	PreferenceKey        *string        `json:"preference_key"`
	PreferenceValue      *string        `json:"preference_value"`
	ConfirmationQuestion *string        `json:"confirmation_question"`
}

const preferenceChatSchemaName = "preference_chat_response"

func preferenceChatResponseSchema() map[string]any {
	return map[string]any{
		"type":                 "object",
		"additionalProperties": false,
		"required":             []string{"display", "is_persistent", "preference_key", "preference_value", "confirmation_question"},
		"properties": map[string]any{
			"display": displayContentSchema(),
			"is_persistent": map[string]any{
				"type": []string{"boolean", "null"},
			},
			"preference_key": map[string]any{
				"type": []string{"string", "null"},
			},
			"preference_value": map[string]any{
				"type": []string{"string", "null"},
			},
			"confirmation_question": map[string]any{
				"type": []string{"string", "null"},
			},
		},
	}
}

func preferenceChatResponseFormat() *responseFormat {
	return &responseFormat{
		Type: "json_schema",
		JSONSchema: &jsonSchemaSpec{
			Name:   preferenceChatSchemaName,
			Strict: true,
			Schema: preferenceChatResponseSchema(),
		},
	}
}

// ChatCompletionPreferenceChat はmessagesをmodelに送信し、
// 受信者からの表示調整の要望への応答(PreferenceChatResponse)を取得する。
func (c *Client) ChatCompletionPreferenceChat(ctx context.Context, model string, messages []Message) (*PreferenceChatResponse, error) {
	body, err := json.Marshal(chatCompletionRequest{
		Model:          model,
		Messages:       messages,
		ResponseFormat: preferenceChatResponseFormat(),
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

	var result PreferenceChatResponse
	if err := json.Unmarshal([]byte(chatResp.Content), &result); err != nil {
		return nil, fmt.Errorf("unmarshal preference chat response: %w", err)
	}

	return &result, nil
}

// BuildPreferenceChatSystemPrompt はSource of Truthと受信者の現在のPreferenceから、
// Chatでの表示調整の要望に応答するためのシステムプロンプトを組み立てる。
//
// 要望が今回限りの一時的な指示か、今後も適用するPreferenceかをAIに判断させ、
// 判断できない場合は確認質問を返させる。この確認・Preferenceの更新方法自体は
// usecase側の責務であり、このプロンプトはAIの応答内容(1回分)の生成方針のみを定める。
func BuildPreferenceChatSystemPrompt(sot SourceOfTruth, preference map[string]string) (string, error) {
	sotJSON, err := json.Marshal(sot)
	if err != nil {
		return "", fmt.Errorf("marshal source of truth: %w", err)
	}

	preferenceJSON, err := json.Marshal(preference)
	if err != nil {
		return "", fmt.Errorf("marshal preference: %w", err)
	}

	return fmt.Sprintf(`あなたは、確定済みの情報(Source of Truth)を受信者向けに表示するアシスタントです。

受信者からの表示に関する要望(例:「もっと簡単な表現にしてほしい」「ひらがなを多くしてほしい」
「別の言語で表示してほしい」)を解釈し、その要望を反映したdisplayを毎回生成してください。
Source of Truthに含まれる情報(事実)は一切変更・省略・追加してはいけません。

要望が今回限りの一時的な指示なのか、今後も継続して適用してほしいPreferenceなのかを判断してください。
- 今後も適用してほしいと明確に読み取れる場合: is_persistentをtrueにし、preference_keyと
  preference_valueに更新後の値を設定してください(例: preference_key="reading_level",
  preference_value="easy")。
- 今回限りだと明確に読み取れる場合: is_persistentをfalseにし、preference_keyと
  preference_valueはnullにしてください。
- どちらか判断できない場合: is_persistentをnullにし、confirmation_questionに
  「今後も常にこの設定にしますか？」といった確認の質問を設定してください。
is_persistentがtrueまたはfalseの場合、confirmation_questionはnullにしてください。

# Source of Truth
%s

# 受信者の現在のPreference
%s`, string(sotJSON), string(preferenceJSON)), nil
}
