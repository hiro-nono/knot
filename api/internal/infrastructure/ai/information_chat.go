package ai

import (
	"context"
	"encoding/json"
	"fmt"
)

// InformationChatResponse は、資料の情報(Source of Truth)についての
// 受信者からの質問に対するAIの回答を表す。
//
// この機能は表示(見せ方)を調整するものではなく、あくまでSource of Truthに
// 含まれる情報について回答するQ&Aである。Preferenceの読み書きはここでは行わない
// (Preferenceの更新はA/B比較の選択結果からのみ行う。usecase.DisplayUsecase.Chatを参照)。
type InformationChatResponse struct {
	Answer string `json:"answer"`
}

const informationChatSchemaName = "information_chat_response"

func informationChatResponseSchema() map[string]any {
	return map[string]any{
		"type":                 "object",
		"additionalProperties": false,
		"required":             []string{"answer"},
		"properties": map[string]any{
			"answer": map[string]any{
				"type":        "string",
				"description": "受信者の質問に対する、Source of Truthの事実のみに基づく回答文(自然文)。",
			},
		},
	}
}

func informationChatResponseFormat() *responseFormat {
	return &responseFormat{
		Type: "json_schema",
		JSONSchema: &jsonSchemaSpec{
			Name:   informationChatSchemaName,
			Strict: true,
			Schema: informationChatResponseSchema(),
		},
	}
}

// ChatCompletionInformationChat はmessagesをmodelに送信し、
// 資料の情報についての質問への回答(InformationChatResponse)を取得する。
func (c *Client) ChatCompletionInformationChat(ctx context.Context, model string, messages []Message) (*InformationChatResponse, error) {
	body, err := json.Marshal(chatCompletionRequest{
		Model:          model,
		Messages:       messages,
		ResponseFormat: informationChatResponseFormat(),
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

	var result InformationChatResponse
	if err := json.Unmarshal([]byte(chatResp.Content), &result); err != nil {
		return nil, fmt.Errorf("unmarshal information chat response: %w", err)
	}

	return &result, nil
}

// BuildInformationChatSystemPrompt はSource of Truthと受信者の現在のPreferenceから、
// 資料の情報についての質問に答えるためのシステムプロンプトを組み立てる。
//
// これは表示の見せ方を調整するChatではなく、資料の中身(事実)についてのQ&Aである。
// Preferenceは回答の言葉遣い・言語などを合わせるための参考情報として渡すのみで、
// この機能がPreferenceを更新することはない(更新はA/B比較の選択結果からのみ行う)。
func BuildInformationChatSystemPrompt(sot SourceOfTruth, preference map[string]string) (string, error) {
	sotJSON, err := json.Marshal(sot)
	if err != nil {
		return "", fmt.Errorf("marshal source of truth: %w", err)
	}

	preferenceJSON, err := json.Marshal(preference)
	if err != nil {
		return "", fmt.Errorf("marshal preference: %w", err)
	}

	return fmt.Sprintf(`あなたは、確定済みの情報(Source of Truth)について、受信者からの質問に回答するアシスタントです。
これは表示の見た目や言い回しを変えてほしいという要望ではなく、資料の中身についての質問です。

以下のSource of Truthに含まれる情報(事実)のみに基づいて回答してください。
Source of Truthに記載の無いことを憶測や捏造で答えてはいけません。
記載が無い場合は、その旨を正直に伝えてください。

受信者のPreferenceは、回答の言葉遣い・言語・平易さなどを合わせるための参考情報です。
このPreference自体を変更する指示ではないため、あなたがPreferenceを更新することはありません。

answerに回答文をJSONで出力してください。

# Source of Truth
%s

# 受信者のPreference(参考情報)
%s`, string(sotJSON), string(preferenceJSON)), nil
}
