import DOMPurify from "isomorphic-dompurify";

// AIが生成する表示本文(TailwindCSSクラスで装飾されたHTML断片)を描画するための
// 共通コンポーネント。bodyの「事実」自体はSource of Truthに厳密に基づいて
// バックエンド側で担保されるが、レンダリング時はscript/イベントハンドラ属性・
// 外部リソース読み込みなどを許可しない安全なタグ/属性のみに絞ってサニタイズし、
// classによるTailwindスタイリングのみを許可する。
const ALLOWED_TAGS = [
  "div",
  "span",
  "p",
  "br",
  "hr",
  "h1",
  "h2",
  "h3",
  "h4",
  "h5",
  "h6",
  "ul",
  "ol",
  "li",
  "strong",
  "em",
  "b",
  "i",
  "u",
  "a",
  "table",
  "thead",
  "tbody",
  "tfoot",
  "tr",
  "th",
  "td",
  "small",
  "sub",
  "sup",
];

const ALLOWED_ATTR = ["class", "href", "target", "rel"];

// リンクは常に新規タブ+安全なrelで開かせる(reverse tabnabbing対策)。
DOMPurify.addHook("afterSanitizeAttributes", (node) => {
  if (node.tagName === "A") {
    node.setAttribute("target", "_blank");
    node.setAttribute("rel", "noopener noreferrer");
  }
});

export function HtmlBody({
  children,
  className = "",
}: {
  children: string;
  className?: string;
}) {
  const safeHtml = DOMPurify.sanitize(children, {
    ALLOWED_TAGS,
    ALLOWED_ATTR,
    ALLOW_DATA_ATTR: false,
    ALLOWED_URI_REGEXP: /^(?:https?|mailto):/i,
  });

  return (
    <div
      className={`text-sm leading-relaxed text-zinc-700 ${className}`}
      dangerouslySetInnerHTML={{ __html: safeHtml }}
    />
  );
}
