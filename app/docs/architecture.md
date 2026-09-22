## Next.jsによるアーキテクチャ

App Routerを採用します。

## 基本アーキテクチャ

機能単位でコードを整理し、各責務を明確にします。

src/
├── app/
│   ├── (auth)/
│   ├── (dashboard)/
│   └── ...
├── components/
│   ├── ui/
│   └── ...
├── features/
│   ├── information/
│   ├── response/
│   ├── membership/
│   ├── preference/
│   └── account/
├── lib/
│   ├── api/
│   ├── auth/
│   └── ...
├── hooks/
├── types/
└── ...

基本的に機能単位でコードを整理します。

汎用的なUIプリミティブは components/ui に配置します。

特定の機能にのみ使用するコンポーネントは、対応する features 配下に配置します。