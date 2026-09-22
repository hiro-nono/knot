input=$(cat)
tool=$(echo "$input" | jq -r '.tool_name')
command=$(echo "$input" | jq -r '.tool_input.command // empty')
filepath=$(echo "$input" | jq -r '.tool_input.file_path // empty')

# 1. 危険コマンドのブロック
if echo "$command" | grep -qE "rm -rf|git add -A|git push --force"; then
  echo "❌ 危険なコマンドです: $command" >&2
  exit 2
fi

# 2. 触ってはいけないファイルの保護
if echo "$filepath" | grep -qE "\.env$|go\.sum$|/migrations/"; then
  echo "❌ このファイルはAIによる直接編集を禁止しています: $filepath" >&2
  exit 2
fi

exit 0