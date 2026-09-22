package usecase

import (
	"time"

	"knot-api/internal/domain"
	"knot-api/internal/infrastructure/ai"
)

// Message は対話の1ターン(発言者と本文)を表す。
// controllerとのHTTP入出力にもそのまま使う。
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// StructuredOption は選択肢1件を表す。
type StructuredOption struct {
	Value     string `json:"value"`
	SortOrder int    `json:"sort_order"`
}

// StructuredSource はAIが構造化した情報1件を表す。
type StructuredSource struct {
	Type            string             `json:"type"`
	Key             string             `json:"key"`
	Value           string             `json:"value"`
	Status          string             `json:"status"`
	Question        *string            `json:"question,omitempty"`
	InteractionType *string            `json:"interaction_type,omitempty"`
	Options         []StructuredOption `json:"options,omitempty"`
}

// StructuredInformation はAIが構造化したInformationのタイトルとSource一覧を表す。
type StructuredInformation struct {
	Title   string             `json:"title"`
	Sources []StructuredSource `json:"sources"`
}

// InformationSummaryView は自分が作成したInformationの一覧表示用の表現。
// Source等の詳細は含まず、一覧から個別のInformationへ辿るための最小限の情報のみを返す。
type InformationSummaryView struct {
	ID             string    `json:"id"`
	Title          string    `json:"title"`
	AccessType     string    `json:"access_type"`
	ResponsePolicy string    `json:"response_policy"`
	CreatedAt      time.Time `json:"created_at"`
}

func newInformationSummaryView(information *domain.Information) *InformationSummaryView {
	return &InformationSummaryView{
		ID:             information.ID().String(),
		Title:          information.Title(),
		AccessType:     string(information.AccessType()),
		ResponsePolicy: string(information.ResponsePolicy()),
		CreatedAt:      information.CreatedAt(),
	}
}

func toAIMessages(messages []Message) []ai.Message {
	out := make([]ai.Message, 0, len(messages))
	for _, m := range messages {
		out = append(out, ai.Message{Role: m.Role, Content: m.Content})
	}
	return out
}

func fromAIMessages(messages []ai.Message) []Message {
	out := make([]Message, 0, len(messages))
	for _, m := range messages {
		out = append(out, Message{Role: m.Role, Content: m.Content})
	}
	return out
}

func fromAIStructured(s *ai.StructuredInformation) *StructuredInformation {
	if s == nil {
		return nil
	}

	sources := make([]StructuredSource, 0, len(s.Sources))
	for _, src := range s.Sources {
		options := make([]StructuredOption, 0, len(src.Options))
		for _, o := range src.Options {
			options = append(options, StructuredOption{Value: o.Value, SortOrder: o.SortOrder})
		}

		sources = append(sources, StructuredSource{
			Type:            src.Type,
			Key:             src.Key,
			Value:           src.Value,
			Status:          src.Status,
			Question:        src.Question,
			InteractionType: src.InteractionType,
			Options:         options,
		})
	}

	return &StructuredInformation{Title: s.Title, Sources: sources}
}
