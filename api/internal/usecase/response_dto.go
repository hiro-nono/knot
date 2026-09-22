package usecase

import (
	"time"

	"knot-api/internal/domain"
)

// ResponseItemView はResponseItemを外部に返却するための表現。
type ResponseItemView struct {
	SourceID string  `json:"source_id"`
	OptionID *string `json:"option_id,omitempty"`
	Value    *string `json:"value,omitempty"`
}

// ResponseView はResponseを外部に返却するための表現。
// UserIDは匿名回答の場合nil(未設定)になる。
type ResponseView struct {
	ID            string             `json:"id"`
	InformationID string             `json:"information_id"`
	UserID        *string            `json:"user_id,omitempty"`
	Items         []ResponseItemView `json:"items"`
	CreatedAt     time.Time          `json:"created_at"`
	UpdatedAt     time.Time          `json:"updated_at"`
}

func newResponseView(response *domain.Response) *ResponseView {
	items := response.Items()
	itemViews := make([]ResponseItemView, 0, len(items))
	for _, item := range items {
		var optionID *string
		if item.OptionID() != nil {
			v := item.OptionID().String()
			optionID = &v
		}

		itemViews = append(itemViews, ResponseItemView{
			SourceID: item.SourceID().String(),
			OptionID: optionID,
			Value:    item.Value(),
		})
	}

	var userID *string
	if response.UserID() != nil {
		v := response.UserID().String()
		userID = &v
	}

	return &ResponseView{
		ID:            response.ID().String(),
		InformationID: response.InformationID().String(),
		UserID:        userID,
		Items:         itemViews,
		CreatedAt:     response.CreatedAt(),
		UpdatedAt:     response.UpdatedAt(),
	}
}
