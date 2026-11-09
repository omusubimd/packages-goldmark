# おむすび Markdown — Goldmark

[Goldmark v2](https://github.com/yuin/goldmark) 向けのおむすび Markdown 拡張です。

- [goldmark-omusubi-ruby](packages/goldmark-omusubi-ruby): ルビ構文のパーサー・HTML レンダラー。

各パッケージは独立した Go モジュールです。現在はルビの単一行構文を実装しています。
インデント拡張はまだ実装していません。

```sh
cd packages/goldmark-omusubi-ruby
go test ./...
go test -race ./...
go vet ./...
```
