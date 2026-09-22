// Package supabase はSupabaseが発行するJWTの検証を行う。
package supabase

// Claims は検証済みJWTから抽出した、アプリケーション側で扱う認証情報を表す。
type Claims struct {
	// UserID はSupabase側のユーザーID(JWTのsubクレーム)。
	// domain.Account.ProviderIDに対応する。
	UserID string
}
