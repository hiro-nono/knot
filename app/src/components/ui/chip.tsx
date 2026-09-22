export type ChipTone = "neutral" | "primary" | "success" | "warning";

const TONE_STYLES: Record<ChipTone, string> = {
  neutral: "bg-zinc-100 text-zinc-700",
  primary: "bg-primary-soft text-primary-hover",
  success: "bg-green-50 text-green-700",
  warning: "bg-amber-50 text-amber-700",
};

export function Chip({ tone = "neutral", children }: { tone?: ChipTone; children: React.ReactNode }) {
  return (
    <span className={`inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium ${TONE_STYLES[tone]}`}>
      {children}
    </span>
  );
}
