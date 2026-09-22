package domain

// PredefinedPreferenceKeys は代表的なPreferenceのキー候補をあらかじめ定義したもの。
// PredefinedSourceKeysと同様、似たような表現の乱立を防ぐためのガイドであり、
// 強制するものではない。該当するものが無ければ、新しいキーを自由に使ってよい。
//
// languageはUser.Languageで管理するため、ここには含めない。
var PredefinedPreferenceKeys = []string{
	"reading_level",
	"verbosity",
	"tone",
	"font_size",
	"information_priority",
	"visual_style",
}
