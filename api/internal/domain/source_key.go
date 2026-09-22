package domain

// PredefinedSourceKeys はtype別によく使われるkeyの候補をあらかじめ定義したもの。
//
// Source.Keyは自由記述のstringのままだが、似たような表現(例: "date"と"departure_date"と
// "travel_date")が乱立することを防ぐため、該当する概念があればここに定義したkeyを
// 優先的に再利用することが望ましい。該当するものが無い場合は、新しいkeyを自由に作ってよい。
//
// DBの実データを見て動的に増やすのではなく、ここに列挙したものだけを「既知のkey」として扱う。
// 新しい概念を頻繁に扱うようになった場合は、このリストに追記して育てていく。
var PredefinedSourceKeys = map[SourceType][]string{
	SourceTypeIdentity: {
		"name",
		"organizer",
	},
	SourceTypeFact: {
		"description",
	},
	SourceTypeSchedule: {
		"date",
		"start_time",
		"end_time",
		"location",
	},
	SourceTypeCondition: {
		"requirement",
	},
	SourceTypeInteraction: {
		"required_action",
	},
}
