-- 初期スキーマ。これ以前の段階的なマイグレーション(memberships導入→撤去→再導入等の
-- 試行錯誤を含む)を、この時点で確定したDB設計として1本に集約したもの。
-- テーブル構成・関係はREADME.mdのER図と対応する。

-- ==============================
-- accounts / users / memberships
-- ==============================

CREATE TYPE account_type AS ENUM ('personal', 'organization');
CREATE TYPE account_role AS ENUM ('admin', 'user');
CREATE TYPE account_status AS ENUM ('active', 'frozen', 'suspended', 'withdrawn', 'banned');

-- Account はOrganization自体と、その認証状態を管理する。
-- nameはOrganizationの名称で、personalの場合は必ずNULL、organizationの場合は必須。
CREATE TABLE accounts (
    id UUID PRIMARY KEY,
    provider_id TEXT NOT NULL UNIQUE,
    account_type account_type NOT NULL,
    name TEXT,
    role account_role NOT NULL,
    status account_status NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT accounts_name_matches_account_type CHECK (
        (account_type = 'personal' AND name IS NULL) OR
        (account_type = 'organization' AND name IS NOT NULL AND name <> '')
    )
);

CREATE INDEX idx_accounts_status ON accounts(status);

-- User は認証されたユーザー自身のプロフィール情報を保持する。
-- Accountとの所属関係・権限はMembershipが管理するため、account_idは持たない。
CREATE TABLE users (
    id UUID PRIMARY KEY,
    last_name TEXT NOT NULL,
    first_name TEXT NOT NULL,
    language TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TYPE membership_role AS ENUM ('owner', 'admin', 'member');
CREATE TYPE membership_status AS ENUM ('active', 'removed');

-- Membership はUserがどのAccount(Organization)に所属し、その中でどんな権限(role)・
-- 状態(status)を持つかを表す。同一account_id・user_idの組み合わせは常に1件のみ
-- 存在する(再追加時は既存行をreactivateする)。
CREATE TABLE memberships (
    id UUID PRIMARY KEY,
    account_id UUID NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role membership_role NOT NULL,
    status membership_status NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (account_id, user_id)
);

CREATE INDEX idx_memberships_account_id ON memberships(account_id);
CREATE INDEX idx_memberships_user_id ON memberships(user_id);

CREATE TYPE membership_event_type AS ENUM ('created', 'role_changed', 'removed', 'reactivated');

-- MembershipEvent はMembershipに対して何が起きたかを記録する(履歴、追記のみ)。
CREATE TABLE membership_events (
    id UUID PRIMARY KEY,
    membership_id UUID NOT NULL REFERENCES memberships(id) ON DELETE CASCADE,
    event_type membership_event_type NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_membership_events_membership_id ON membership_events(membership_id);

CREATE TYPE membership_removal_request_status AS ENUM ('pending', 'approved', 'rejected');

-- MembershipRemovalRequest は、AdminによるMembership除外の申請と、
-- Ownerによる承認待ち状態を表す。
CREATE TABLE membership_removal_requests (
    id UUID PRIMARY KEY,
    membership_id UUID NOT NULL REFERENCES memberships(id) ON DELETE CASCADE,
    requested_by_user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status membership_removal_request_status NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_membership_removal_requests_membership_id ON membership_removal_requests(membership_id);

-- 同一Membershipに対してpending状態のリクエストが複数同時に存在しないようにする。
CREATE UNIQUE INDEX idx_membership_removal_requests_pending_unique
    ON membership_removal_requests(membership_id)
    WHERE status = 'pending';

CREATE TYPE account_status_event_type AS ENUM (
    'created',
    'frozen',
    'suspended',
    'withdrawn',
    'banned',
    'unfrozen',
    'reactivated'
);

-- AccountStatusEvent はAccountのステータス更新イベントを記録する(履歴、追記のみ)。
CREATE TABLE account_status_events (
    id UUID PRIMARY KEY,
    account_id UUID NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    event_type account_status_event_type NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_account_status_events_account_id ON account_status_events(account_id);

-- ==============================
-- informations / sources / options
-- ==============================

CREATE TYPE information_access_type AS ENUM ('public', 'restricted');
CREATE TYPE information_response_policy AS ENUM ('anonymous', 'authenticated');

-- Information は発信者が作成する情報の集約ルート。
-- access_typeにより閲覧可否(public=リンクを知っていれば誰でも、
-- restricted=recipientsに登録されたUserのみ)を、response_policyにより
-- publicなInformationへの回答に認証が必須か(authenticated)不要か(anonymous)を
-- 制御する(restrictedの場合response_policyは参照されず、常に認証+Recipient必須)。
CREATE TABLE informations (
    id UUID PRIMARY KEY,
    account_id UUID NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    created_by_user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title TEXT NOT NULL,
    access_type information_access_type NOT NULL,
    response_policy information_response_policy NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_informations_account_id ON informations(account_id);

CREATE TYPE source_type AS ENUM ('identity', 'fact', 'schedule', 'condition', 'interaction');
CREATE TYPE source_status AS ENUM ('confirmed', 'undecided', 'unknown');
CREATE TYPE source_interaction_type AS ENUM ('radio', 'check', 'text');

-- Source はInformationに属し、構造化された情報を1件保持する。
CREATE TABLE sources (
    id UUID PRIMARY KEY,
    information_id UUID NOT NULL REFERENCES informations(id) ON DELETE CASCADE,
    type source_type NOT NULL,
    key TEXT NOT NULL,
    value TEXT NOT NULL,
    status source_status NOT NULL,
    interaction_type source_interaction_type,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_sources_information_id ON sources(information_id);

-- Option はSourceに属し、選択肢を1件保持する。
CREATE TABLE options (
    id UUID PRIMARY KEY,
    source_id UUID NOT NULL REFERENCES sources(id) ON DELETE CASCADE,
    value TEXT NOT NULL,
    sort_order INTEGER NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_options_source_id ON options(source_id);

-- ==============================
-- recipients / preferences / responses
-- ==============================

-- Recipient はInformation(access_type=restricted)を閲覧できるUserを表す
-- (Informationの閲覧権限のみを表し、Responseとは独立している)。
CREATE TABLE recipients (
    id UUID PRIMARY KEY,
    information_id UUID NOT NULL REFERENCES informations(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (information_id, user_id)
);

CREATE INDEX idx_recipients_information_id ON recipients(information_id);
CREATE INDEX idx_recipients_user_id ON recipients(user_id);

-- Preference はUserごとの表示最適化設定を保持する。
CREATE TABLE preferences (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL UNIQUE,
    items JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Response はInformationに対するUserの返答を表す(Recipientそのものには紐付けない)。
-- user_idはnullable。PUBLIC+ANONYMOUSなInformationはGoogle Formのように
-- 未ログインでも回答できるため。UNIQUE(information_id, user_id)は
-- PostgreSQLがNULL同士を区別する性質により、匿名回答は複数存在できる一方、
-- 認証済みUserの重複回答は引き続き防止される。
CREATE TABLE responses (
    id UUID PRIMARY KEY,
    information_id UUID NOT NULL REFERENCES informations(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (information_id, user_id)
);

CREATE INDEX idx_responses_information_id ON responses(information_id);

-- ResponseItem はResponseに属し、1つのSourceに対する回答を1件保持する。
-- option_id・valueはどちらか一方のみを設定する(check・radioはoption_id、textはvalue)。
CREATE TABLE response_items (
    id UUID PRIMARY KEY,
    response_id UUID NOT NULL REFERENCES responses(id) ON DELETE CASCADE,
    source_id UUID NOT NULL REFERENCES sources(id) ON DELETE CASCADE,
    option_id UUID REFERENCES options(id) ON DELETE CASCADE,
    value TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (response_id, source_id),
    CHECK (num_nonnulls(option_id, value) = 1)
);

CREATE INDEX idx_response_items_response_id ON response_items(response_id);
