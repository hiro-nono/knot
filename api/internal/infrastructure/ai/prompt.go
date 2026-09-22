package ai

import (
	"fmt"
	"sort"
	"strings"

	"knot-api/internal/domain"
)

// BuildMessages はシステムプロンプトとユーザー発話から、
// リクエスト送信用のメッセージ列を組み立てる。
func BuildMessages(systemPrompt string, userInput string) []Message {
	return []Message{
		{Role: RoleSystem, Content: systemPrompt},
		{Role: RoleUser, Content: userInput},
	}
}

// BuildSystemPrompt は発信者との対話からInformation/Sourceを構造化するための
// デフォルトのシステムプロンプトを組み立てる。
//
// 曖昧な項目についてはSourceのstatusをconfirmed以外にしたうえで、
// 発信者に確認するための質問をquestionに含めるよう指示する。
// この質問への回答をどう扱い、いつ対話を終えてDBへ保存するかはusecase側の責務であり、
// このプロンプトはAIの応答内容(1回分)の生成方針のみを定める。
func BuildSystemPrompt() string {
	return fmt.Sprintf(`あなたは発信者との対話から、発信したい情報を構造化するアシスタントです。
これまでの会話から、Informationのタイトルと、それに属するSourceの一覧をJSONで出力してください。

各Sourceは次のいずれかのtypeに分類してください。
- %s: 何についての情報か
- %s: 何が事実か
- %s: いつ・期間
- %s: どんな条件で何が起こるか
- %s: 受信者に何をしてほしいか

各Sourceのstatusは次のいずれかにしてください。
- %s: 発信者の発言から明確に確定している
- %s: 発言はあるが曖昧、または未確定
- %s: まだ情報が得られていない

statusがconfirmed以外の場合は、発信者に確認するための具体的な質問をquestionに含めてください。
confirmedの場合、questionはnullにしてください。

これまでの会話で既に尋ねた質問に対し、発信者の直前の回答が曖昧・的外れ・無回答などで
期待する情報が得られなかった場合、全く同じ表現の質問をそのまま繰り返してはいけません。
次のいずれかの方法で聞き方を変えてください。
- 具体例や選択肢を挙げて、より答えやすい聞き方にする
- より粒度の小さい、具体的な質問に分解する
- 表現を平易にする、言い換える

optionsは、interaction_typeが%sまたは%sの場合にのみ選択肢を設定してください。
それ以外(%sまたはnull)の場合、optionsは空配列にしてください。`,
		domain.SourceTypeIdentity,
		domain.SourceTypeFact,
		domain.SourceTypeSchedule,
		domain.SourceTypeCondition,
		domain.SourceTypeInteraction,
		domain.SourceStatusConfirmed,
		domain.SourceStatusUndecided,
		domain.SourceStatusUnknown,
		domain.SourceInteractionTypeRadio,
		domain.SourceInteractionTypeCheck,
		domain.SourceInteractionTypeText,
	)
}

// BuildKnownKeysPrompt は既存のSourceで使われているkeyをtype別にまとめ、
// AIが新しいkeyを作る前に再利用を検討できるようにするための追加プロンプトを組み立てる。
//
// keyは自由記述の文字列だが、アプリ側で一貫した語彙にするため、
// 該当する概念であれば新しいkeyを作らずここに挙げたkeyを再利用させる。
// 既存のどれにも当てはまらない全く新しい概念の場合のみ、新しいkeyの作成を許可する。
//
// knownKeysが空の場合(まだ何も保存されていない場合)は空文字列を返す。
func BuildKnownKeysPrompt(knownKeys map[domain.SourceType][]string) string {
	if len(knownKeys) == 0 {
		return ""
	}

	types := make([]string, 0, len(knownKeys))
	for t := range knownKeys {
		types = append(types, string(t))
	}
	sort.Strings(types)

	var b strings.Builder
	b.WriteString("次は、typeごとに既に使われているkeyの一覧です。\n")
	b.WriteString("該当する概念であれば、新しいkeyを作らずこれらを再利用してください。\n")
	b.WriteString("既存のどれにも当てはまらない全く新しい概念の場合のみ、新しいkeyを作成してください。\n\n")

	for _, t := range types {
		keys := knownKeys[domain.SourceType(t)]
		sortedKeys := append([]string(nil), keys...)
		sort.Strings(sortedKeys)
		fmt.Fprintf(&b, "- %s: %s\n", t, strings.Join(sortedKeys, ", "))
	}

	return b.String()
}
