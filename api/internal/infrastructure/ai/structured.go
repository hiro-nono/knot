package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// StructuredInformation はAIの応答を構造化したもの。
// Informationのタイトルと、そこに属するSource一覧を表す。
type StructuredInformation struct {
	Title   string             `json:"title"`
	Sources []StructuredSource `json:"sources"`
}

// StructuredSource はdomain.Sourceの各フィールドに対応する構造化データ。
// Statusがconfirmed以外の場合、Questionに発信者へ確認するための質問が入る
// (confirmedの場合はnil)。この質問への回答をもとに実際にconfirmedとして
// 扱うかどうかはusecase側の責務であり、ここでは判断しない。
type StructuredSource struct {
	Type            string             `json:"type"`
	Key             string             `json:"key"`
	Value           string             `json:"value"`
	Status          string             `json:"status"`
	Question        *string            `json:"question"`
	InteractionType *string            `json:"interaction_type"`
	Options         []StructuredOption `json:"options"`
}

// StructuredOption はdomain.Optionの各フィールドに対応する構造化データ。
type StructuredOption struct {
	Value     string `json:"value"`
	SortOrder int    `json:"sort_order"`
}

type responseFormat struct {
	Type       string          `json:"type"`
	JSONSchema *jsonSchemaSpec `json:"json_schema,omitempty"`
}

type jsonSchemaSpec struct {
	Name   string `json:"name"`
	Strict bool   `json:"strict"`
	Schema any    `json:"schema"`
}

func structuredInformationResponseFormat() *responseFormat {
	return &responseFormat{
		Type: "json_schema",
		JSONSchema: &jsonSchemaSpec{
			Name:   structuredInformationSchemaName,
			Strict: true,
			Schema: structuredInformationSchema(),
		},
	}
}

// ChatCompletionStructured はmessagesをmodelに送信し、AIの応答を
// タイトルとSource一覧を持つ構造化JSON(StructuredInformation)として取得する。
func (c *Client) ChatCompletionStructured(ctx context.Context, model string, messages []Message) (*StructuredInformation, error) {
	req, err := c.buildStructuredRequest(ctx, model, messages)
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

	var structured StructuredInformation
	if err := json.Unmarshal([]byte(chatResp.Content), &structured); err != nil {
		return nil, fmt.Errorf("unmarshal structured content: %w", err)
	}

	return &structured, nil
}

func (c *Client) buildStructuredRequest(ctx context.Context, model string, messages []Message) (*http.Request, error) {
	body, err := json.Marshal(chatCompletionRequest{
		Model:          model,
		Messages:       messages,
		ResponseFormat: structuredInformationResponseFormat(),
	})
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	return c.newRequest(ctx, body)
}
