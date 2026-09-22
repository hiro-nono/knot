package domain

import (
	"errors"
	"time"

	"uuid"
)

// User は認証されたユーザー自身のプロフィール情報(姓・名・言語)を表すEntity。
// UserとAccount(Organization)の所属関係・権限はMembershipが管理するため、
// Userはaccount_idを直接持たない。1人のUserは複数のAccountにMembership経由で
// 所属できる。
type User struct {
	id        uuid.UUID
	lastName  string
	firstName string
	language  string
	createdAt time.Time
	updatedAt time.Time
}

// NewUser は新しいUserを生成する。
func NewUser(lastName, firstName, language string) (*User, error) {
	if lastName == "" {
		return nil, errors.New("lastName は必須です")
	}
	if firstName == "" {
		return nil, errors.New("firstName は必須です")
	}
	if language == "" {
		return nil, errors.New("language は必須です")
	}

	now := time.Now()
	return &User{
		id:        uuid.New(),
		lastName:  lastName,
		firstName: firstName,
		language:  language,
		createdAt: now,
		updatedAt: now,
	}, nil
}

// ReconstructUser は永続化されたデータからUserを復元する。
// バリデーションは行わない。
func ReconstructUser(id uuid.UUID, lastName, firstName, language string, createdAt, updatedAt time.Time) *User {
	return &User{
		id:        id,
		lastName:  lastName,
		firstName: firstName,
		language:  language,
		createdAt: createdAt,
		updatedAt: updatedAt,
	}
}

func (u *User) ID() uuid.UUID {
	return u.id
}

func (u *User) LastName() string {
	return u.lastName
}

func (u *User) FirstName() string {
	return u.firstName
}

func (u *User) Language() string {
	return u.language
}

func (u *User) CreatedAt() time.Time {
	return u.createdAt
}

func (u *User) UpdatedAt() time.Time {
	return u.updatedAt
}

// UpdateProfile は姓・名・言語を更新する(本人による編集)。
func (u *User) UpdateProfile(lastName, firstName, language string) error {
	if lastName == "" {
		return errors.New("lastName は必須です")
	}
	if firstName == "" {
		return errors.New("firstName は必須です")
	}
	if language == "" {
		return errors.New("language は必須です")
	}

	u.lastName = lastName
	u.firstName = firstName
	u.language = language
	u.updatedAt = time.Now()
	return nil
}
