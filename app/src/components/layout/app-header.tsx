"use client";

import Link from "next/link";
import { usePathname, useRouter } from "next/navigation";

import { useMyAccount } from "@/features/account/hooks";
import { useSignOut } from "@/features/auth/hooks";

const NAV_LINKS = [
  { href: "/", label: "ホーム" },
  { href: "/informations/new", label: "新しい情報" },
  { href: "/account/members", label: "メンバー" },
  { href: "/settings", label: "設定" },
];

const ADMIN_LINK = { href: "/admin/accounts", label: "管理" };

export function AppHeader() {
  const pathname = usePathname();
  const router = useRouter();
  const signOut = useSignOut();
  const { data: myAccount } = useMyAccount();

  const navLinks = myAccount?.role === "admin" ? [...NAV_LINKS, ADMIN_LINK] : NAV_LINKS;

  function handleSignOut() {
    signOut.mutate(undefined, {
      onSuccess: () => router.push("/signin"),
    });
  }

  return (
    <header className="border-b border-zinc-200 bg-white">
      <div className="mx-auto flex h-14 w-full max-w-5xl items-center gap-6 px-4">
        <Link href="/" className="text-lg font-semibold text-primary">
          knot
        </Link>

        <nav className="flex flex-1 items-center gap-1">
          {navLinks.map((link) => {
            const active = pathname === link.href;
            return (
              <Link
                key={link.href}
                href={link.href}
                className={`rounded-full px-3 py-1.5 text-sm font-medium transition-colors ${
                  active
                    ? "bg-primary-soft text-primary-hover"
                    : "text-zinc-600 hover:bg-zinc-100 hover:text-zinc-900"
                }`}
              >
                {link.label}
              </Link>
            );
          })}
        </nav>

        <button
          type="button"
          onClick={handleSignOut}
          disabled={signOut.isPending}
          className="text-sm font-medium text-zinc-500 transition-colors hover:text-zinc-900 disabled:opacity-50"
        >
          サインアウト
        </button>
      </div>
    </header>
  );
}
