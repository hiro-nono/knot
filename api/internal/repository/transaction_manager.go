package repository

import "context"

// TransactionManager は複数のRepository呼び出しを1つのトランザクションとして
// 実行することを抽象化する。
type TransactionManager interface {
	// WithinTransaction はfnを1つのトランザクション内で実行する。
	// fnがエラーを返した場合はロールバックし、そうでなければコミットする。
	// fnに渡されるctxを使ってRepositoryを呼び出すことで、
	// そのRepositoryの操作が同じトランザクションに参加する。
	WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}
