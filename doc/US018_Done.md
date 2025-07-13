# US-018: 設定機能 完了報告

## 完了定義

- [x] デフォルト優先度の設定が可能
- [x] 自動保存の有効/無効が設定可能
- [x] テーマの変更が可能

## 完了のエビデンス

### 実装内容

- `internal/config/config.go` に `Settings` 構造体を定義し、設定の読み込み・保存機能 (`LoadConfig`, `SaveConfig`) を実装しました。
- `cmd/main.go` の `model` 構造体に `cfg` フィールドを追加し、アプリケーション起動時に設定をロードするようにしました。
- メインビューのフッターに設定オプションの表示を追加し、`g` キーによる設定ビューへの遷移、および設定ロジックを実装しました。
- 設定ビューでデフォルト優先度、自動保存、テーマの変更ができるようにしました。

### テスト結果

`go test -v ./internal/config` を実行し、全てのテストがパスすることを確認しました。

```
=== RUN   TestGetConfigFilePath
--- PASS: TestGetConfigFilePath (0.00s)
=== RUN   TestNewDefaultConfig
--- PASS: TestNewDefaultConfig (0.00s)
=== RUN   TestSaveAndLoadConfig
--- PASS: TestSaveAndLoadConfig (0.00s)
PASS
ok  	zan/internal/config	0.275s