# 規約

## まずこれらの資料を一読してください。

- doc配下に下記の資料があります。
  - requirement_definition.md : 要件定義書です。必ずタスクを始める前に一読してください。
  - product_backlog.md : バックログ(US-XXX、以下US)です。該当のスプリントを完了させてください。

## 開発は下記のように進めてください。

1. 該当スプリントの残USを確認します。
2. スプリント内すべてのUSを確認して、対応内容で要件定義書の機能を実装できるかを確認してください。もし、要件が足りない場合、追加でUSを足す提案をしてください。
3. USを上から順に実装します。t_wadaさんのTDDで開発します。
  - USは、**受け入れ条件**と**task**のチェックボックスにすべてチェックが付いている、かつ完了理由を記載したmarkdown資料を作成する必要があります。
  - 完了理由を記載した資料はdoc配下に`USXXX_Done.md`として**完了定義**と**完了のエビデンス**を記載します。
  - **受け入れ条件**と**task**については、`product_backlog.md`のチェックボックスにチェックしてください。
4. USの対応が完了した場合、次のUSに移る前にGitHubにCommit,Pushするプロセスを踏んでください。
5. スプリント内のUSが残っている場合は、手順3に戻る。全て完了した場合は手順6に移ります。
5. このスプリント成果物の動作披露をしてください。動作に不備がある場合は、手順2に戻り、フィードバックの対応を行います。

## リポジトリ構造が変わった場合の対応です。

- 要件定義書を満たすような環境が準備できていない場合、適宜環境のセットアップをしてください。
- 環境構築や、ディレクトリ構成に変更が入った場合、バックログの対応とは別に`README.md`を修正してください。

## USXXX_Done.md の記載例です。

完了の定義と完了のエビデンスを提出します。

```md
# US-XXX: ソート機能 完了報告

## 完了定義

- [x] 作成日時、更新日時でソート可能
- [x] 優先度でソート可能
- [x] 昇順・降順の選択が可能

## 完了のエビデンス

### 実装内容

- `internal/app/app.go` に `SortTasks` 関数を実装し、作成日時、更新日時、優先度で昇順・降順にソートできるようにしました。
- `cmd/main.go` の `model` 構造体にソート条件 (`sortBy`) とソート順 (`sortAsc`) を保持するフィールドを追加しました。
- `cmd/main.go` にソート用のUI (`sortInput`) と、`o` キーによるソートビューへの遷移、およびソートロジックを実装しました。
- メインビューのフッターにソートオプションの表示を追加しました。

### テスト結果

`go test -v ./internal/app` を実行し、`TestSortTasks` を含む全てのテストがパスすることを確認しました。

{テスト結果...}

```