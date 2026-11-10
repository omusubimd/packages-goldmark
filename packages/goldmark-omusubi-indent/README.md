# goldmark-omusubi-indent

おむすび Markdown の「全角スペースによる段落開始」を Goldmark **v2** に追加します。
Go 1.25 以降が必要です。

## 利用方法

```sh
go get github.com/omusubimd/packages-goldmark/packages/goldmark-omusubi-indent@<commit-or-branch>
```

```go
package main

import (
    "os"

    indent "github.com/omusubimd/packages-goldmark/packages/goldmark-omusubi-indent"
    "github.com/yuin/goldmark/v2/parser"
    "github.com/yuin/goldmark/v2/renderer/html"
)

func main() {
    p := parser.New(parser.WithExtensions(indent.Parser))
    source := []byte("こんにちは\n　おむすび\n　**大作戦**")
    if err := html.New().Render(os.Stdout, source, p.Parse(source)); err != nil {
        panic(err)
    }
}
```

出力:

```html
<p>こんにちは</p>
<p>おむすび</p>
<p><strong>大作戦</strong></p>
```

`Parser` は既定の拡張、`NewParser()` は新しい拡張を返します。オプションはありません。
通常の段落・インラインノードを出すため専用レンダラーは不要です。
Goldmark v1 の API には対応していません。

## 構文

- ブロック開始位置の U+3000 を1文字取り除き、新しい段落を始めます。
- 直前に空行がなくても段落を区切ります。印のない後続行は同じ段落に続きます。
- 印の前の半角スペース0～3個は通常の Markdown のインデントとして扱います。
  4桁以上のインデントやタブはコード・段落継続など Goldmark の通常の規則に従います。
- 印の直後の半角スペースや2個目以降の全角スペースは本文に残します。
- 印だけ、または印の後ろが半角空白だけの行はブロックを区切り、空の段落を作りません。
  リスト項目内・項目間を区切る場合も通常の空行と同じ tight/loose 判定にします。
- `　# heading` や `　- item` は、その行では見出し・リストではなく段落の本文です。
  次の行の見出し・リスト・Setext 見出しなどは通常どおり解析します。
- 引用・リストの内部では、そのコンテナの規則に従って適用します。
- コード・インラインコード・HTML ブロック内、および行の途中の全角スペースは変更しません。

ソース全体を書き換える前処理は行いません。段落の `Source()` は印を除いた元のバイト範囲を参照します。
内部の一時的な区切りノードは最終 AST から取り除きます。独自ノードや印を使った書き戻し機能は提供しません。
本文はホストのインラインパーサーで解析するため、独自拡張やルビ拡張を併用できます。

```go
p := parser.New(parser.WithExtensions(indent.Parser, ruby.Parser))
r := html.New(html.WithExtensions(ruby.HTMLRenderer))
```

上の併用例の `ruby` は [goldmark-omusubi-ruby](../goldmark-omusubi-ruby) です。
登録順序を逆にした場合も [integration](../../integration) で検証しています。

## 互換性と設計上の選択

unified 版の仕様・基本例に従いますが、完全な出力互換ではありません。
81例の比較と意図的な差、採用しなかった選択肢を [設計メモ](DESIGN.md) に記載しています。
印だけの行と下線見出しの境界は CommonMark の空行と同じ解釈です。
`one\n　\n---` は `one\n\n---` と同じ段落＋水平線になり、Setext 見出しにはなりません。
参照定義、HTML、リスト内段落の整形、改行コードは Goldmark の既定処理に従います。

## 開発

```sh
go test ./...
go test -race -cover ./...
go vet ./...
go test -run '^$' -fuzz FuzzParsing -fuzztime=15s -parallel=2
```

元になる unified 版テストには MIT ライセンスを添付しています。[testdata](testdata/README.md) を参照してください。
