# 1. アーキテクチャ依存ルール
arch_result=$(go-arch-lint check 2>&1)
if [ $? -ne 0 ]; then
  echo "❌ アーキテクチャ違反: $arch_result" >&2
  exit 2
fi

# 2. 構文・ビルドエラーの即時検出(重いテストの前段階として)
build_result=$(go build ./... 2>&1)
if [ $? -ne 0 ]; then
  echo "❌ ビルドエラー: $build_result" >&2
  exit 2
fi

# 3. フォーマット崩れ
gofmt_result=$(gofmt -l .)
if [ -n "$gofmt_result" ]; then
  echo "❌ フォーマット未適用のファイル: $gofmt_result" >&2
  exit 2
fi

exit 0