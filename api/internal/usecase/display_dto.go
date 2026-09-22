package usecase

import "knot-api/internal/domain"

// DisplayContent は受信者向けに最適化された表示内容を表す。
type DisplayContent struct {
	Title string `json:"title"`
	Body  string `json:"body"`
}

// OptionView はOptionを外部に返却するための表現。
type OptionView struct {
	ID        string `json:"id"`
	Value     string `json:"value"`
	SortOrder int    `json:"sort_order"`
}

// SourceView はSourceを外部に返却するための表現。
// 受信者がResponseを送信する際に必要なsource_id・interaction_type・
// option_idを提示するために使う(POST /informations/:id/responsesの入力)。
type SourceView struct {
	ID              string       `json:"id"`
	Type            string       `json:"type"`
	Key             string       `json:"key"`
	Value           string       `json:"value"`
	InteractionType *string      `json:"interaction_type,omitempty"`
	Options         []OptionView `json:"options,omitempty"`
}

func newSourceView(source *domain.Source, options []*domain.Option) SourceView {
	var interactionType *string
	if source.InteractionType() != nil {
		v := string(*source.InteractionType())
		interactionType = &v
	}

	optionViews := make([]OptionView, 0, len(options))
	for _, o := range options {
		optionViews = append(optionViews, OptionView{
			ID:        o.ID().String(),
			Value:     o.Value(),
			SortOrder: o.SortOrder(),
		})
	}

	return SourceView{
		ID:              source.ID().String(),
		Type:            string(source.Type()),
		Key:             source.Key(),
		Value:           source.Value(),
		InteractionType: interactionType,
		Options:         optionViews,
	}
}
