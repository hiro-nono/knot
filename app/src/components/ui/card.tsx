export function Card({ className = "", children }: { className?: string; children: React.ReactNode }) {
  return (
    <div
      className={`rounded-2xl border border-zinc-200 bg-white p-6 shadow-sm ${className}`}
    >
      {children}
    </div>
  );
}
