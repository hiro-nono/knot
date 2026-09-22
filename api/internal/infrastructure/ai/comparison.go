package ai

import (
	"context"
	"encoding/json"
	"fmt"
)

// ComparisonPattern は、指定したPreferenceキーについてAIが決めた
// 1つの値(傾向)と、その値を採用した場合の表示を表す。
// Valueは受信者に見せるものではなく、選択結果をPreferenceとして
// 記録するためのメタデータである。
type ComparisonPattern struct {
	Value string `json:"value"`
	Title string `json:"title"`
	Body  string `json:"body"`
}

// ComparisonContent は、あるPreferenceキーについてAIが提案する、
// 意味のある対照的な2パターン(A/B)を表す。
type ComparisonContent struct {
	PatternA ComparisonPattern `json:"pattern_a"`
	PatternB ComparisonPattern `json:"pattern_b"`
}

const comparisonSchemaName = "comparison_content"

func comparisonPatternSchema() map[string]any {
	return map[string]any{
		"type":                 "object",
		"additionalProperties": false,
		"required":             []string{"value", "title", "body"},
		"properties": map[string]any{
			"value": map[string]any{
				"type":        "string",
				"description": "このパターンが採用したPreference値(傾向)。受信者には見せず、選択結果の記録にのみ使う。",
			},
			"title": map[string]any{"type": "string"},
			"body": map[string]any{
				"type":        "string",
				"description": "TailwindCSSのクラスを使った安全なHTML断片(Markdown不可)。script/style/on*属性は含めない。",
			},
		},
	}
}

func comparisonContentSchema() map[string]any {
	return map[string]any{
		"type":                 "object",
		"additionalProperties": false,
		"required":             []string{"pattern_a", "pattern_b"},
		"properties": map[string]any{
			"pattern_a": comparisonPatternSchema(),
			"pattern_b": comparisonPatternSchema(),
		},
	}
}

func comparisonResponseFormat() *responseFormat {
	return &responseFormat{
		Type: "json_schema",
		JSONSchema: &jsonSchemaSpec{
			Name:   comparisonSchemaName,
			Strict: true,
			Schema: comparisonContentSchema(),
		},
	}
}

// ChatCompletionComparison はmessagesをmodelに送信し、
// 対照的な2パターンの表示(ComparisonContent)を取得する。
func (c *Client) ChatCompletionComparison(ctx context.Context, model string, messages []Message) (*ComparisonContent, error) {
	body, err := json.Marshal(chatCompletionRequest{
		Model:          model,
		Messages:       messages,
		ResponseFormat: comparisonResponseFormat(),
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

	var content ComparisonContent
	if err := json.Unmarshal([]byte(chatResp.Content), &content); err != nil {
		return nil, fmt.Errorf("unmarshal comparison content: %w", err)
	}

	return &content, nil
}

// BuildComparisonSystemPrompt はSource of Truthと受信者の現在のPreferenceから、
// 指定したPreferenceキーについて対照的な2パターン(A/B)を生成させるための
// システムプロンプトを組み立てる。
//
// どの値(傾向)を比較するかはAI自身に決めさせる。ユーザーに値を自由入力させると
// 「何と何を比べればよいか」の負担が大きいため、そのキーにおいて受信者にとって
// 意味のある違いが出やすい、対照的な2つの値をAIに考えさせる設計にしている。
// Source of Truthの内容(事実)は変更・省略させない点はBuildDisplaySystemPromptと同じ。
func BuildComparisonSystemPrompt(sot SourceOfTruth, preference map[string]string, key string) (string, error) {
	sotJSON, err := json.Marshal(sot)
	if err != nil {
		return "", fmt.Errorf("marshal source of truth: %w", err)
	}

	preferenceJSON, err := json.Marshal(preference)
	if err != nil {
		return "", fmt.Errorf("marshal preference: %w", err)
	}

	return fmt.Sprintf(`あなたは、確定済みの情報(Source of Truth)の表示を、受信者に比較してもらうための
2パターン(A/B)生成するアシスタントです。

比較対象のPreferenceキーは「%s」です。このキーについて、受信者にとって違いが
はっきりわかる、対照的な2つの値(傾向)をあなた自身で考えてください
(例えばキーがreading_levelなら"easy"と"detailed"、キーがtoneなら"casual"と"formal"など)。
2つの値は明確に異なるものにしてください。

以下のSource of Truthに含まれる情報(事実)は一切変更・省略・追加してはいけません。
表現の平易さ、情報量、言語などの「見せ方」だけを、それぞれの値に合わせて最適化してください。
「%s」以外のPreferenceについては、受信者の現在のPreferenceに従ってください。

pattern_aとpattern_bそれぞれについて、value(採用した値)・title・body(表示本文)をJSONで出力してください。

%s

# Source of Truth
%s

# 受信者の現在のPreference
%s`, key, key, htmlBodyFormatInstructions, string(sotJSON), string(preferenceJSON)), nil
}
