# knot

送信者が作成した情報を、AIが受信者ごとに最適化して届けるコミュニケーションプラットフォーム。

- `api`: 発信者・受信者それぞれにAIエージェントを提供するバックエンド(Go)
- `app`: フロントエンド(Next.js)

## 必要なもの

- [Go](https://go.dev/) 1.27以上
- [Bun](https://bun.sh/) 1.3以上
- [Docker](https://www.docker.com/) / Docker Compose(Postgresの起動に使う)
- Supabaseプロジェクト(認証。JWT署名鍵は非対称鍵に設定しておくこと)
- OrcaRouterのAPIキー(AI呼び出し)

## 起動手順

### 1. 環境変数ファイルを用意する

```bash
cp api/.env.example api/.env
cp app/.env.example app/.env
```

### 2. 環境変数を埋める

`api/.env` に必要な値:

| 変数名 | 値の入手先 |
|---|---|
| `ORCAROUTER_BASE_URL` / `ORCAROUTER_API_KEY` | OrcaRouterのダッシュボード |
| `AI_MODEL` | OrcaRouterで利用可能なモデルID(例: `openai/gpt-4o-mini`) |
| `PORT` | 例: `8080` |
| `GIN_MODE` | 例: `debug` |
| `POSTGRES_HOST` / `POSTGRES_USER` / `POSTGRES_PASSWORD` / `POSTGRES_DB` / `POSTGRES_PORT` | 手順3のPostgresに合わせて設定(例は下記) |
| `SUPABASE_URL` | Supabaseダッシュボード `Settings > API` のProject URL(`app/.env`の`NEXT_PUBLIC_SUPABASE_URL`と同じ値) |
| `CORS_ALLOWED_ORIGINS` | 未設定の場合`http://localhost:3000`がデフォルトで許可される |

`app/.env` に必要な値:

| 変数名 | 値の入手先 |
|---|---|
| `NEXT_PUBLIC_SUPABASE_URL` | Supabaseダッシュボード `Settings > API` のProject URL |
| `NEXT_PUBLIC_SUPABASE_ANON_KEY` | Supabaseダッシュボード `Settings > API` のanon key |
| `NEXT_PUBLIC_API_BASE_URL` | ローカルでは`http://localhost:8080`(手順4のapiに合わせる) |

### 3. Postgresを起動する

```bash
cd api
docker compose up -d db
```

`api/.env` の `POSTGRES_*` は、そのままローカル接続用に以下の値で揃えると動きます。

```bash
POSTGRES_HOST=localhost
POSTGRES_USER=knot
POSTGRES_PASSWORD=knot
POSTGRES_DB=knot_api
POSTGRES_PORT=5432
```

### 4. APIサーバーを起動する

```bash
# api ディレクトリで実行
go run ./cmd/server
```

起動時にマイグレーションが自動実行される。`http://localhost:8080` で待ち受ける。

### 5. フロントエンドを起動する

```bash
cd ../app
bun install
bun run dev
```

`http://localhost:3000` で待ち受ける。

### 6. 動作確認

ブラウザで [http://localhost:3000](http://localhost:3000) を開く。

## ER図

```mermaid
erDiagram
    ACCOUNTS ||--o{ MEMBERSHIPS : has
    USERS ||--o{ MEMBERSHIPS : has
    MEMBERSHIPS ||--o{ MEMBERSHIP_EVENTS : has
    MEMBERSHIPS ||--o{ MEMBERSHIP_REMOVAL_REQUESTS : has
    USERS ||--o{ MEMBERSHIP_REMOVAL_REQUESTS : requests
    ACCOUNTS ||--o{ ACCOUNT_STATUS_EVENTS : has
    ACCOUNTS ||--o{ INFORMATIONS : owns
    USERS ||--o{ INFORMATIONS : creates
    INFORMATIONS ||--o{ SOURCES : has
    SOURCES ||--o{ OPTIONS : has
    INFORMATIONS ||--o{ RECIPIENTS : "shared with"
    USERS ||--o{ RECIPIENTS : "can view"
    USERS ||--o| PREFERENCES : "has (user_id)"
    INFORMATIONS ||--o{ RESPONSES : "has"
    USERS ||--o{ RESPONSES : responds
    RESPONSES ||--o{ RESPONSE_ITEMS : has
    SOURCES ||--o{ RESPONSE_ITEMS : "answered by"
    OPTIONS ||--o{ RESPONSE_ITEMS : "selected by"

    ACCOUNTS {
        uuid id PK
        text provider_id UK
        account_type account_type
        text name "nullable, organization必須"
        account_role role
        account_status status
        timestamptz created_at
        timestamptz updated_at
    }

    USERS {
        uuid id PK
        text last_name
        text first_name
        text language
        timestamptz created_at
        timestamptz updated_at
    }

    MEMBERSHIPS {
        uuid id PK
        uuid account_id FK
        uuid user_id FK
        membership_role role
        membership_status status
        timestamptz created_at
        timestamptz updated_at
    }

    MEMBERSHIP_EVENTS {
        uuid id PK
        uuid membership_id FK
        membership_event_type event_type
        timestamptz created_at
    }

    MEMBERSHIP_REMOVAL_REQUESTS {
        uuid id PK
        uuid membership_id FK
        uuid requested_by_user_id FK
        membership_removal_request_status status
        timestamptz created_at
        timestamptz updated_at
    }

    ACCOUNT_STATUS_EVENTS {
        uuid id PK
        uuid account_id FK
        account_status_event_type event_type
        timestamptz created_at
    }

    INFORMATIONS {
        uuid id PK
        uuid account_id FK
        uuid created_by_user_id FK
        text title
        information_access_type access_type
        information_response_policy response_policy
        timestamptz created_at
        timestamptz updated_at
    }

    SOURCES {
        uuid id PK
        uuid information_id FK
        source_type type
        text key
        text value
        source_status status
        source_interaction_type interaction_type
        timestamptz created_at
        timestamptz updated_at
    }

    OPTIONS {
        uuid id PK
        uuid source_id FK
        text value
        int sort_order
        timestamptz created_at
        timestamptz updated_at
    }

    RECIPIENTS {
        uuid id PK
        uuid information_id FK
        uuid user_id FK
        timestamptz created_at
    }

    PREFERENCES {
        uuid id PK
        uuid user_id UK
        jsonb items
        timestamptz created_at
        timestamptz updated_at
    }

    RESPONSES {
        uuid id PK
        uuid information_id FK
        uuid user_id FK "nullable, 匿名回答"
        timestamptz created_at
        timestamptz updated_at
    }

    RESPONSE_ITEMS {
        uuid id PK
        uuid response_id FK
        uuid source_id FK
        uuid option_id FK "nullable"
        text value "nullable"
        timestamptz created_at
    }
```

詳細は [`api/README.md`](./api/README.md) を参照。
