package model

import (
	"time"
	"uuid"
)

// InformationAccessType はinformationsテーブルのaccess_typeカラムの値を表す。
type InformationAccessType string

const (
	InformationAccessTypePublic     InformationAccessType = "public"
	InformationAccessTypeRestricted InformationAccessType = "restricted"
)

// InformationResponsePolicy はinformationsテーブルのresponse_policyカラムの値を表す。
type InformationResponsePolicy string

const (
	InformationResponsePolicyAnonymous     InformationResponsePolicy = "anonymous"
	InformationResponsePolicyAuthenticated InformationResponsePolicy = "authenticated"
)

// Information はinformationsテーブルの行を表す。
// 発信するユーザーが作成する最上位のレコードである。
type Information struct {
	ID              uuid.UUID
	AccountID       uuid.UUID
	CreatedByUserID uuid.UUID
	Title           string
	AccessType      InformationAccessType
	ResponsePolicy  InformationResponsePolicy
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// Recipient はrecipientsテーブルの行を表す。
// restrictedなInformationを閲覧できるUserを1件保持する。
type Recipient struct {
	ID            uuid.UUID
	InformationID uuid.UUID
	UserID        uuid.UUID
	CreatedAt     time.Time
}

// SourceType はCLAUDE.mdで定義されたSource of Truthの分類を表す。
type SourceType string

const (
	SourceTypeIdentity    SourceType = "identity"
	SourceTypeFact        SourceType = "fact"
	SourceTypeSchedule    SourceType = "schedule"
	SourceTypeCondition   SourceType = "condition"
	SourceTypeInteraction SourceType = "interaction"
)

// SourceStatus はSourceの値がどの程度確定しているかを表す。
type SourceStatus string

const (
	SourceStatusConfirmed SourceStatus = "confirmed"
	SourceStatusUndecided SourceStatus = "undecided"
	SourceStatusUnknown   SourceStatus = "unknown"
)

// SourceInteractionType は受信者がSourceに対してどう応答すべきかを表す。
// 受信者の応答を必要としないSourceの場合はnilになる。
type SourceInteractionType string

const (
	SourceInteractionTypeRadio SourceInteractionType = "radio"
	SourceInteractionTypeCheck SourceInteractionType = "check"
	SourceInteractionTypeText  SourceInteractionType = "text"
)

// Source はsourcesテーブルの行を表す。
// Informationに属し、構造化された情報を1件保持する。
type Source struct {
	ID              uuid.UUID
	InformationID   uuid.UUID
	Type            SourceType
	Key             string
	Value           string
	Status          SourceStatus
	InteractionType *SourceInteractionType
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// Option はoptionsテーブルの行を表す。
// Sourceに属し、選択肢を1件保持する。
type Option struct {
	ID        uuid.UUID
	SourceID  uuid.UUID
	Value     string
	SortOrder int
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Response はresponsesテーブルの行を表す。
// Informationに対するUserの返答1件を保持する(配下のResponseItemが個々の回答を持つ)。
// UserIDはnullable(匿名回答)。
type Response struct {
	ID            uuid.UUID
	InformationID uuid.UUID
	UserID        *uuid.UUID
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// ResponseItem はresponse_itemsテーブルの行を表す。
// Responseに属し、1つのSourceに対する回答を1件保持する。
// OptionID・Valueはともにnullableで、どちらか一方のみが設定される
// (check・radioはOptionID、textはValue)。
type ResponseItem struct {
	ID         uuid.UUID
	ResponseID uuid.UUID
	SourceID   uuid.UUID
	OptionID   *uuid.UUID
	Value      *string
	CreatedAt  time.Time
}
