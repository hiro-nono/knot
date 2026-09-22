export default function AuthLayout({ children }: { children: React.ReactNode }) {
  return (
    <div className="flex flex-1 items-center justify-center px-4 py-16">
      <div className="w-full max-w-sm rounded-2xl border border-zinc-200 bg-white p-6 shadow-sm">
        {children}
      </div>
    </div>
  );
}
