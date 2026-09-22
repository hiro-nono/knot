package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const (
	chatCompletionsPath = "/chat/completions"
	contentTypeJSON     = "application/json"
	authorizationPrefix = "Bearer "
)

// Client はOrcaRouter経由でAIプロバイダーにチャットリクエストを送信する。
type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

// NewClient はClientを生成する。
func NewClient(baseURL string, apiKey string, httpClient *http.Client) *Client {
	return &Client{
		baseURL:    baseURL,
		apiKey:     apiKey,
		httpClient: httpClient,
	}
}

// Usage はトークン使用量を表す。
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// ChatResponse はOrcaRouterのレスポンスから抽出した応答内容を表す。
type ChatResponse struct {
	Content      string
	FinishReason string
	Usage        Usage
}

// Role はチャットメッセージの発言者を表す。
const (
	RoleSystem    = "system"
	RoleUser      = "user"
	RoleAssistant = "assistant"
)

// Message はOrcaRouterとの間で送受信する1件のチャットメッセージを表す。
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatCompletionRequest struct {
	Model          string          `json:"model"`
	Messages       []Message       `json:"messages"`
	ResponseFormat *responseFormat `json:"response_format,omitempty"`
}

type chatCompletionChoice struct {
	Message      Message `json:"message"`
	FinishReason string  `json:"finish_reason"`
}

type chatCompletionResponse struct {
	Choices []chatCompletionChoice `json:"choices"`
	Usage   Usage                  `json:"usage"`
}

// ChatCompletion はmessagesをmodelに送信し、AIの応答を取得する。
func (c *Client) ChatCompletion(ctx context.Context, model string, messages []Message) (*ChatResponse, error) {
	req, err := c.buildRequest(ctx, model, messages)
	if err != nil {
		return nil, err
	}

	respBody, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	return parseChatCompletionResponse(respBody)
}

func (c *Client) buildRequest(ctx context.Context, model string, messages []Message) (*http.Request, error) {
	body, err := json.Marshal(chatCompletionRequest{Model: model, Messages: messages})
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	return c.newRequest(ctx, body)
}

func (c *Client) newRequest(ctx context.Context, body []byte) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+chatCompletionsPath, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Content-Type", contentTypeJSON)
	req.Header.Set("Authorization", authorizationPrefix+c.apiKey)

	return req, nil
}

func (c *Client) doRequest(req *http.Request) ([]byte, error) {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("orcarouter returned status %d: %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

func parseChatCompletionResponse(body []byte) (*ChatResponse, error) {
	var parsed chatCompletionResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}
	if len(parsed.Choices) == 0 {
		return nil, fmt.Errorf("orcarouter response has no choices")
	}

	choice := parsed.Choices[0]
	return &ChatResponse{
		Content:      choice.Message.Content,
		FinishReason: choice.FinishReason,
		Usage:        parsed.Usage,
	}, nil
}
