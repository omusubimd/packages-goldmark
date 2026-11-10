# goldmark-omusubi-ruby

おむすび Markdown のルビ構文を Goldmark **v2** に追加します。Go 1.25 以降が必要です。

## 利用方法

このモジュールはリポジトリの `packages/goldmark-omusubi-ruby` にあります。
未リリースのため、動作確認にはこの変更を含むコミットまたはブランチを指定してください。

```sh
go get github.com/omusubimd/packages-goldmark/packages/goldmark-omusubi-ruby@<commit-or-branch>
```

```go
package main

import (
    "os"

    ruby "github.com/omusubimd/packages-goldmark/packages/goldmark-omusubi-ruby"
    "github.com/yuin/goldmark/v2/parser"
    "github.com/yuin/goldmark/v2/renderer/html"
)

func main() {
    p := parser.New(parser.WithExtensions(ruby.Parser))
    r := html.New(html.WithExtensions(ruby.HTMLRenderer))
    source := []byte("[青天|の|霹靂]<<せいてん||へきれき>>")
    if err := r.Render(os.Stdout, source, p.Parse(source)); err != nil {
        panic(err)
    }
}
```

出力:

```html
<p><ruby>青天<rt>せいてん</rt></ruby>の<ruby>霹靂<rt>へきれき</rt></ruby></p>
```

`Parser` / `HTMLRenderer` は既定の拡張です。`NewParser()` / `NewHTMLRenderer()` で
拡張を生成することもできます。Goldmark v2 のパーサー拡張と HTML 拡張は別々に登録します。
Goldmark v1 の `goldmark.WithExtensions` API には対応していません。

## 構文と AST

- `[青空]<<あおぞら>>` は単一セグメントのルビです。
- `[青|空]<<あお|ぞら>>` は一つの ruby 要素内に複数の読みを出力します。
- 読みが空の区間は ruby 要素の外へ出します。`[青空]<<>>` は通常のテキストになります。
- 空のラベル区間、区間数の不一致、閉じマーカーの欠落は通常の Markdown として処理します。
- ラベルの強調・インラインコード・自動リンク・エスケープ・文字参照を扱います。
  読みは Markdown として解析せず、エスケープと文字参照をデコードして HTML 用にエスケープします。
- `|` と最初の `]` はラベルの構文区切りです。コードやエスケープで区切りを隠せません。
  最初の `>` は読みの終端の開始として扱い、続く文字も `>` である必要があります。

`RubyNode` と `RubySegmentNode`、および `KindRuby` / `KindRubySegment` を公開しています。
`RubyNode` の子は `RubySegmentNode`、その子は通常のインラインノードです。
`RubySegmentNode.RubyText.Value(source)` でデコード済みの読みを取得できます。
読みが空の区間は AST 変換時に外へ移すため、最終 AST に空読みのセグメントは残りません。
ラベルの値は別のソースを必要としない owned 値です。ラベル内のテキストの元バイト範囲は保持しません。

## 互換性と未対応範囲

[unified 版](https://github.com/omusubimd/packages-unified) の単一行のルビ構文と HTML 出力例を基準にしています。
Goldmark は HTML ブロック・インライン HTML・通常の Markdown の出力方針を決めます。
既定では生 HTML を `<!-- raw HTML omitted -->` にし、`html.WithUnsafe()` では出力します。
読みはこの設定にかかわらず HTML 用にエスケープします。段落の後ろには Goldmark 既定の改行が付きます。

現時点では改行をまたぐルビは認識しません。unified 版のトークナイザーは改行も受け取れるため、完全な構文互換ではありません。
ラベルは区間ごとに標準 CommonMark インラインパーサーで解析します。
ホストに設定した `parser.WithParseDelimiterFunc(...)` と `parser.WithEscapedSpace()` はラベル内にも引き継ぎます。
例えば `parser.ParseDelimiterSimple` を指定した CJK 向け強調規則は、通常の本文とルビのベースで一致します。
エスケープ空白はベース・読みで同じ設定を使いますが、インラインコード内の文字はそのままです。
外側で登録した任意の独自インライン拡張や、上記以外のパーサー設定はラベル内に引き継ぎません。
この引き継ぎ方と複数行の読みの AST 表現は、今後の設計項目です。

## 検証

```sh
go test ./...
go test -race -cover ./...
go vet ./...
go test -run '^$' -fuzz FuzzConversion -fuzztime=15s -parallel=2
```

unified 版由来の26件の比較用データと出典は [testdata](testdata/README.md) を参照してください。
