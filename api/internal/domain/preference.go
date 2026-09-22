package domain

import (
	"errors"
	"time"

	"uuid"
)

// Preference は受信者(User)ごとの表示最適化設定を表すEntity。
// キーごとの値(例: "language" -> "ja", "reading_level" -> "easy")を保持する。
//
// キーは自由記述のstringであり、固定のenumにはしない
// (PredefinedPreferenceKeysで代表的なキーの再利用を促すのみ)。
type Preference struct {
	id        uuid.UUID
	userID    uuid.UUID
	items     map[string]string
	createdAt time.Time
	updatedAt time.Time
}

// NewPreference は受信者(User)の新しいPreferenceを生成する。
func NewPreference(userID uuid.UUID) (*Preference, error) {
	if userID == uuid.Nil() {
		return nil, errors.New("userID は必須です")
	}

	now := time.Now()
	return &Preference{
		id:        uuid.New(),
		userID:    userID,
		items:     map[string]string{},
		createdAt: now,
		updatedAt: now,
	}, nil
}

// ReconstructPreference は永続化されたデータからPreferenceを復元する。
// 呼び出し元(repository)がすでに正当なデータであることを保証するため、
// バリデーションは行わない。
func ReconstructPreference(id uuid.UUID, userID uuid.UUID, items map[string]string, createdAt, updatedAt time.Time) *Preference {
	if items == nil {
		items = map[string]string{}
	}

	return &Preference{
		id:        id,
		userID:    userID,
		items:     items,
		createdAt: createdAt,
		updatedAt: updatedAt,
	}
}

func (p *Preference) ID() uuid.UUID {
	return p.id
}

func (p *Preference) UserID() uuid.UUID {
	return p.userID
}

// Get はkeyに対応する値を返す。存在しない場合はok=falseを返す。
func (p *Preference) Get(key string) (string, bool) {
	v, ok := p.items[key]
	return v, ok
}

// Items は保持している全項目のコピーを返す。
func (p *Preference) Items() map[string]string {
	out := make(map[string]string, len(p.items))
	for k, v := range p.items {
		out[k] = v
	}
	return out
}

// Set はkeyの値を更新(またはまだ無ければ新規作成)する。
func (p *Preference) Set(key string, value string) error {
	if key == "" {
		return errors.New("key は必須です")
	}

	p.items[key] = value
	p.updatedAt = time.Now()
	return nil
}

func (p *Preference) CreatedAt() time.Time {
	return p.createdAt
}

func (p *Preference) UpdatedAt() time.Time {
	return p.updatedAt
}
