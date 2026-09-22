// Package config は環境変数に基づくアプリケーション設定を一元管理する。
package config

import (
	"net"
	"net/url"
	"os"
	"strings"
)

// defaultCORSAllowedOrigins はCORS_ALLOWED_ORIGINSが未設定の場合に使う、
// 開発環境向けのデフォルト許可オリジン。
var defaultCORSAllowedOrigins = []string{"http://localhost:3000"}

// Config は環境変数から読み込んだ実行時設定を保持する。
type Config struct {
	Port    string
	GinMode string

	OrcaRouterBaseURL string
	OrcaRouterAPIKey  string
	AIModel           string

	PostgresHost     string
	PostgresUser     string
	PostgresPassword string
	PostgresDB       string
	PostgresPort     string

	// SupabaseURL はSupabaseプロジェクトのベースURL(例: https://xxxx.supabase.co)。
	// JWT検証用の公開鍵セット(JWKS)を取得するために使う。
	SupabaseURL string

	// CORSAllowedOrigins はCORSで許可するオリジンの一覧。
	CORSAllowedOrigins []string
}

// Load は環境変数から設定を読み込む。
func Load() Config {
	return Config{
		Port:    os.Getenv("PORT"),
		GinMode: os.Getenv("GIN_MODE"),

		OrcaRouterBaseURL: os.Getenv("ORCAROUTER_BASE_URL"),
		OrcaRouterAPIKey:  os.Getenv("ORCAROUTER_API_KEY"),
		AIModel:           os.Getenv("AI_MODEL"),

		PostgresHost:     os.Getenv("POSTGRES_HOST"),
		PostgresUser:     os.Getenv("POSTGRES_USER"),
		PostgresPassword: os.Getenv("POSTGRES_PASSWORD"),
		PostgresDB:       os.Getenv("POSTGRES_DB"),
		PostgresPort:     os.Getenv("POSTGRES_PORT"),

		SupabaseURL: os.Getenv("SUPABASE_URL"),

		CORSAllowedOrigins: parseCORSAllowedOrigins(os.Getenv("CORS_ALLOWED_ORIGINS")),
	}
}

// parseCORSAllowedOrigins はCORS_ALLOWED_ORIGINS(カンマ区切り)をパースする。
// 未設定の場合は開発用のデフォルト(localhost:3000)を返す。
func parseCORSAllowedOrigins(raw string) []string {
	if raw == "" {
		return defaultCORSAllowedOrigins
	}

	var origins []string
	for _, o := range strings.Split(raw, ",") {
		o = strings.TrimSpace(o)
		if o != "" {
			origins = append(origins, o)
		}
	}
	if len(origins) == 0 {
		return defaultCORSAllowedOrigins
	}

	return origins
}

// PostgresDSN は読み込んだ環境変数の値からPostgreSQLの接続文字列を組み立てる。
func (c Config) PostgresDSN() string {
	u := url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(c.PostgresUser, c.PostgresPassword),
		Host:     net.JoinHostPort(c.PostgresHost, c.PostgresPort),
		Path:     "/" + c.PostgresDB,
		RawQuery: "sslmode=disable",
	}
	return u.String()
}
