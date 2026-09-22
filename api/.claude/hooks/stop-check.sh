# 1. 全テスト実行(重い処理はここでまとめて)
test_result=$(go test ./... -v 2>&1)
if [ $? -ne 0 ]; then
  echo "❌ テストが失敗しています: $test_result" >&2
  exit 2
fi

# 2. カバレッジの最低ライン
coverage=$(go test ./... -cover 2>&1 | grep -oP 'coverage: \K[0-9.]+')
# (必要なら閾値チェックのロジックを追加)

# 3. lintの最終チェック
lint_result=$(golangci-lint run ./... 2>&1)
if [ $? -ne 0 ]; then
  echo "❌ lintエラー: $lint_result" >&2
  exit 2
fi

exit 0