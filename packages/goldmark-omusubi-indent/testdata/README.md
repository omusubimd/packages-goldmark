# unified 版との比較データ

`unified.json` は以下のリビジョンの `omusubiIndentSyntax()` と micromark を実行して生成した81例です。

- Repository: https://github.com/omusubimd/packages-unified
- Commit: `5be268bd93b87c7e5a9293a6ac1a88ed7e5007b8`
- [構文実装](https://github.com/omusubimd/packages-unified/blob/5be268bd93b87c7e5a9293a6ac1a88ed7e5007b8/packages/micromark-extension-omusubi-indent/src/syntax.ts)
- [基本4例の元テスト](https://github.com/omusubimd/packages-unified/blob/5be268bd93b87c7e5a9293a6ac1a88ed7e5007b8/packages/micromark-extension-omusubi-indent/test/compile.test.ts)
- License: MIT ([LICENSE](LICENSE))

基本4例は元テストの入力を使用しました。他の77例は互換性を確認するために追加した境界例です。
`unifiedOutput` は実際の unified 出力です。文書前後の CR / LF のみを比較時に正規化します。
それ以外に差がある10例は `goldmarkOutput` と `difference` を付け、レビューした期待値を明示しています。

差の分類と判断の理由は [DESIGN.md](../DESIGN.md) を参照してください。
