# おむすび Markdown — Goldmark

[Goldmark v2](https://github.com/yuin/goldmark) 向けのおむすび Markdown 拡張です。

- [goldmark-omusubi-ruby](packages/goldmark-omusubi-ruby): ルビ構文のパーサー・HTML レンダラー。
- [goldmark-omusubi-indent](packages/goldmark-omusubi-indent): 全角スペースによる段落開始。専用レンダラーは不要です。

各パッケージは独立した Go モジュールです。ルビは単一行構文に対応します。
インデント拡張の互換性と未確定の境界仕様は [設計メモ](packages/goldmark-omusubi-indent/DESIGN.md) を参照してください。
ルビとインデントの併用テストは、公開パッケージと独立した [integration](integration) モジュールにあります。

各モジュールで実行します。

```sh
cd packages/goldmark-omusubi-indent # または packages/goldmark-omusubi-ruby / integration
go test ./...
go test -race ./...
go vet ./...
```
