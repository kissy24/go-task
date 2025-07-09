# US-017: ソート機能 完了報告

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

```
=== RUN   TestSortTasks
=== RUN   TestSortTasks/Sort_by_CreatedAt_Ascending
=== RUN   TestSortTasks/Sort_by_CreatedAt_Descending
=== RUN   TestSortTasks/Sort_by_UpdatedAt_Ascending
=== RUN   TestSortTasks/Sort_by_UpdatedAt_Descending
=== RUN   TestSortTasks/Sort_by_Priority_Ascending_(Low_to_High)
=== RUN   TestSortTasks/Sort_by_Priority_Descending_(High_to_Low)
=== RUN   TestSortTasks/Default_sort_(unknown_sortBy,_CreatedAt_Descending)
--- PASS: TestSortTasks (0.00s)
    --- PASS: TestSortTasks/Sort_by_CreatedAt_Ascending (0.00s)
    --- PASS: TestSortTasks/Sort_by_CreatedAt_Descending (0.00s)
    --- PASS: TestSortTasks/Sort_by_UpdatedAt_Ascending (0.00s)
    --- PASS: TestSortTasks/Sort_by_UpdatedAt_Descending (0.00s)
    --- PASS: TestSortTasks/Sort_by_Priority_Ascending_(Low_to_High) (0.00s)
    --- PASS: TestSortTasks/Sort_by_Priority_Descending_(High_to_Low) (0.00s)
    --- PASS: TestSortTasks/Default_sort_(unknown_sortBy,_CreatedAt_Descending) (0.00s)
PASS
ok  	zan/internal/app	0.235s