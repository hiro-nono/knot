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

// htmlBodyFormatInstructions は、DisplayContent.Body(表示本文)をAIに
// 生成させる際、Markdownではなく安全なHTML断片 + Tailwind CSSで
// 出力させるための指示文。
//
// 受信者側のレンダラーはこの出力をサニタイズしたうえで
// dangerouslySetInnerHTMLとして描画する想定のため、scriptやイベント
// ハンドラ属性などを含めないことを明示する。
const htmlBodyFormatInstructions = `bodyは、Markdownではなく、TailwindCSSのユーティリティクラスで
スタイリングされたHTML断片として出力してください。
- <div><p><span><h1〜h6><ul><ol><li><strong><em><a><table><thead><tbody><tr><th><td><br><hr>
  など、意味のある範囲で自由にHTML要素・レイアウト(flex/gridなど)を使って構いません。
- 見た目はすべてTailwindCSSのクラス名で指定してください(class="...")。インラインstyle属性、
  <style>タグ、<script>タグ、on〜(onclickなど)のイベントハンドラ属性は絶対に含めないでください。
- <html><head><body>などページ全体を表すタグは不要です。bodyの中身となる断片のみを出力してください。
- 外部リソース(画像・フォント・iframe等)の読み込みは行わないでください。`

func displayContentSchema() map[string]any {
	return map[string]any{
		"type":                 "object",
		"additionalProperties": false,
		"required":             []string{"title", "body"},
		"properties": map[string]any{
			"title": map[string]any{"type": "string"},
			"body": map[string]any{
				"type":        "string",
				"description": "TailwindCSSのクラスを使った安全なHTML断片(Markdown不可)。script/style/on*属性は含めない。",
			},
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
// 「見せ方」だけをPreferenceに合わせて最適化させる。bodyはMarkdownではなく
// TailwindCSSクラスで装飾したHTML断片として出力させる(htmlBodyFormatInstructions)。
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

専門用語・略語(例: API, JWT, 認証基盤, HS256, ES256 など)は、事実の一部としてそのまま保持してください。
ただし読みやすさ(reading_level)がeasy・simple寄り、または受信者が非専門家と推測される場合は、
その用語を省略・置換せず残したまま、直後に一般の人にもわかる短い言い換えや補足説明を
(例: 「JWT(ログイン状態を安全に保つための仕組み)」のように)添えてください。
これは事実の追加ではなく、理解を助けるための見せ方の工夫として扱ってください。

titleとbody(表示本文)をJSONで出力してください。

%s

# Source of Truth
%s

# 受信者のPreference
%s`, htmlBodyFormatInstructions, string(sotJSON), string(preferenceJSON)), nil
}
