// バックエンド domain.PredefinedPreferenceKeys の代表的なキー候補。
// キー・値ともに自由記述の文字列であり、これらはあくまで選びやすくするための候補。
export const PREFERENCE_KEY_OPTIONS: { value: string; label: string }[] = [
  { value: "reading_level", label: "読みやすさ" },
  { value: "verbosity", label: "詳しさ" },
  { value: "tone", label: "トーン" },
  { value: "font_size", label: "文字サイズ" },
  { value: "information_priority", label: "情報の優先度" },
  { value: "visual_style", label: "見た目のスタイル" },
];

export const CUSTOM_PREFERENCE_KEY = "__custom__";
