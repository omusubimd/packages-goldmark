# ルビの比較用データ

`ruby.json` は以下の unified 版の26件の HTML コンパイル例を JSON に移したものです。

- Repository: https://github.com/omusubimd/packages-unified
- Commit: `4778af5aac9c67b9d36a4c89fa76864351e843b2`
- File: `packages/micromark-extension-omusubi-ruby/test/compile.test.ts`
- Source: https://github.com/omusubimd/packages-unified/blob/4778af5aac9c67b9d36a4c89fa76864351e843b2/packages/micromark-extension-omusubi-ruby/test/compile.test.ts
- License: MIT ([LICENSE](LICENSE))

`output` は unified 版の期待値に、Goldmark が段落末尾に出す改行を加えたものです。
生 HTML を含む1件には `goldmarkOutput` を追加し、Goldmark 既定の HTML 非出力方針を明示しています。
その他の25件は `output` を直接比較します。
