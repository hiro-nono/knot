package domain

import (
	"errors"
	"fmt"
	"time"
	"uuid"
)

// InformationAccessType はInformationへのアクセス方式を表す。
type InformationAccessType string

const (
	// InformationAccessTypePublic はリンクを知っている誰でも閲覧できることを表す。
	// 受信者を個別にrecipientsへ登録する必要はない。
	InformationAccessTypePublic InformationAccessType = "public"
	// InformationAccessTypeRestricted はrecipientsに登録されたUserのみが
	// 閲覧できることを表す。
	InformationAccessTypeRestricted InformationAccessType = "restricted"
)

func (t InformationAccessType) valid() bool {
	switch t {
	case InformationAccessTypePublic, InformationAccessTypeRestricted:
		return true
	default:
		return false
	}
}

// InformationResponsePolicy はInformationに対する返答(Response)を、
// 誰が行えるかを表す。accessTypeがpublicの場合にのみ意味を持ち、
// restrictedの場合は常にRecipient+認証必須となるため参照されない。
type InformationResponsePolicy string

const (
	// InformationResponsePolicyAnonymous はaccessTypeがpublicの場合、
	// 未ログインを含め誰でも回答できることを表す(user_idはnilを許容)。
	InformationResponsePolicyAnonymous InformationResponsePolicy = "anonymous"
	// InformationResponsePolicyAuthenticated はaccessTypeがpublicの場合、
	// ログイン済みのUserのみが回答できることを表す(user_id必須)。
	InformationResponsePolicyAuthenticated InformationResponsePolicy = "authenticated"
)

func (p InformationResponsePolicy) valid() bool {
	switch p {
	case InformationResponsePolicyAnonymous, InformationResponsePolicyAuthenticated:
		return true
	default:
		return false
	}
}

// Information は発信者が作成する情報の集約ルートを表すEntity。
// Source, Optionを子・孫として保持する。
//
// accountIDはこのInformationを所有するAccount、createdByUserIDは実際に
// 作成したUserを表す。accessTypeにより、リンクを知っていれば誰でも閲覧できるか
// (public)、recipientsに登録されたUserのみが閲覧できるか(restricted)が決まる。
// 実際のアクセス可否判定はrepository(recipients)を必要とするためusecase層が担う。
//
// responsePolicyは、Google Formのように「未ログインでも回答可能」にするか
// どうかを制御する回答設定。accessTypeとの組み合わせで以下のように解釈する:
//   - public + anonymous:      誰でも回答できる(user_id NULL可)
//   - public + authenticated:  ログイン済みのUserなら誰でも回答できる(user_id必須)
//   - restricted:               responsePolicyに関わらずRecipientかつ認証必須
type Information struct {
	id              uuid.UUID
	accountID       uuid.UUID
	createdByUserID uuid.UUID
	title           string
	accessType      InformationAccessType
	responsePolicy  InformationResponsePolicy
	sources         []Source
	createdAt       time.Time
	updatedAt       time.Time
}

// NewInformation は新しいInformationを生成する。
func NewInformation(accountID, createdByUserID uuid.UUID, title string, accessType InformationAccessType, responsePolicy InformationResponsePolicy) (*Information, error) {
	if accountID == uuid.Nil() {
		return nil, errors.New("accountID は必須です")
	}
	if createdByUserID == uuid.Nil() {
		return nil, errors.New("createdByUserID は必須です")
	}
	if title == "" {
		return nil, errors.New("title は必須です")
	}
	if !accessType.valid() {
		return nil, fmt.Errorf("access_type の値が不正です: %q", accessType)
	}
	if !responsePolicy.valid() {
		return nil, fmt.Errorf("response_policy の値が不正です: %q", responsePolicy)
	}

	now := time.Now()
	return &Information{
		id:              uuid.New(),
		accountID:       accountID,
		createdByUserID: createdByUserID,
		title:           title,
		accessType:      accessType,
		responsePolicy:  responsePolicy,
		sources:         []Source{},
		createdAt:       now,
		updatedAt:       now,
	}, nil
}

// ReconstructInformation は永続化されたデータからInformationを復元する。
// 呼び出し元(repository)がすでに正当なデータであることを保証するため、
// バリデーションは行わない。
func ReconstructInformation(id, accountID, createdByUserID uuid.UUID, title string, accessType InformationAccessType, responsePolicy InformationResponsePolicy, sources []Source, createdAt, updatedAt time.Time) *Information {
	if sources == nil {
		sources = []Source{}
	}

	return &Information{
		id:              id,
		accountID:       accountID,
		createdByUserID: createdByUserID,
		title:           title,
		accessType:      accessType,
		responsePolicy:  responsePolicy,
		sources:         sources,
		createdAt:       createdAt,
		updatedAt:       updatedAt,
	}
}

func (i *Information) ID() uuid.UUID {
	return i.id
}

func (i *Information) AccountID() uuid.UUID {
	return i.accountID
}

func (i *Information) CreatedByUserID() uuid.UUID {
	return i.createdByUserID
}

func (i *Information) Title() string {
	return i.title
}

func (i *Information) AccessType() InformationAccessType {
	return i.accessType
}

// IsPublic はaccessTypeがpublicかどうかを返す。
func (i *Information) IsPublic() bool {
	return i.accessType == InformationAccessTypePublic
}

func (i *Information) ResponsePolicy() InformationResponsePolicy {
	return i.responsePolicy
}

// RequiresAuthenticatedResponder は、返答(Response)に認証済みUser(user_id)が
// 必須かどうかを返す。restrictedは常にtrue、publicはresponsePolicyに従う。
func (i *Information) RequiresAuthenticatedResponder() bool {
	return i.accessType == InformationAccessTypeRestricted || i.responsePolicy == InformationResponsePolicyAuthenticated
}

// RequiresRecipientForResponse は、返答(Response)にRecipientとしての登録が
// 必須かどうかを返す。restrictedの場合のみtrue。
func (i *Information) RequiresRecipientForResponse() bool {
	return i.accessType == InformationAccessTypeRestricted
}

func (i *Information) Sources() []Source {
	return i.sources
}

func (i *Information) CreatedAt() time.Time {
	return i.createdAt
}

func (i *Information) UpdatedAt() time.Time {
	return i.updatedAt
}

// AddSource はこのInformationにSourceを追加する。
// 追加するSourceは、このInformationのIDを参照している必要がある。
func (i *Information) AddSource(source Source) error {
	if source.InformationID() != i.id {
		return errors.New("source は指定されたinformationに属していません")
	}

	i.sources = append(i.sources, source)
	i.updatedAt = time.Now()
	return nil
}
