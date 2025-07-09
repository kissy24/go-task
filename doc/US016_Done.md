# US-016: キーワード検索機能 完了報告

## 完了定義

- [x] タイトルでの部分一致検索が可能
- [x] 詳細説明での部分一致検索が可能
- [x] 大文字小文字を区別しない検索が可能

## 完了のエビデンス

### 実装内容

- `internal/task/task.go` に `SearchTasks` 関数を実装し、タイトルと詳細説明に対して大文字小文字を区別しない部分一致検索を可能にしました。
- `internal/app/app.go` の `App` 構造体に `Search` メソッドを追加し、`task.SearchTasks` を呼び出すようにしました。
- `cmd/main.go` に検索用のUI (`searchInput`) と、`s` キーによる検索ビューへの遷移、および検索ロジックを実装しました。
- 検索結果のタスクリストにおいて、検索キーワードにマッチする部分をハイライト表示するようにしました。

### テスト結果

`go test -v ./internal/app` を実行し、`TestSearchTasks` を含む全てのテストがパスすることを確認しました。

```
=== RUN   TestSearchTasks
    app_test.go:523: Task 0: ID=..., Title='Buy groceries', Description='Milk, eggs, bread'
    app_test.go:523: Task 1: ID=..., Title='Finish report', Description='Complete Q3 financial report'
    app_test.go:523: Task 2: ID=..., Title='Call John', Description='Discuss project updates'
    app_test.go:523: Task 3: ID=..., Title='Prepare presentation', Description='Review slides for meeting'
    app_test.go:523: Task 4: ID=..., Title='Grocery shopping list', Description='Fruits and vegetables'
=== RUN   TestSearchTasks/Search_by_title_keyword_'report'
    app_test.go:566: Search() for keyword 'report' returned 1 tasks:
    app_test.go:568:   - Found Task: Title='Finish report'
=== RUN   TestSearchTasks/Search_by_description_keyword_'milk'
    app_test.go:566: Search() for keyword 'milk' returned 1 tasks:
    app_test.go:568:   - Found Task: Title='Buy groceries'
=== RUN   TestSearchTasks/Case-insensitive_search_'grocery'
    app_test.go:566: Search() for keyword 'grocery' returned 1 tasks:
    app_test.go:568:   - Found Task: Title='Grocery shopping list'
=== RUN   TestSearchTasks/Search_by_partial_keyword_'proj'
    app_test.go:566: Search() for keyword 'proj' returned 1 tasks:
    app_test.go:568:   - Found Task: Title='Call John'
=== RUN   TestSearchTasks/No_matching_keyword
    app_test.go:566: Search() for keyword 'nonexistent' returned 0 tasks:
=== RUN   TestSearchTasks/Empty_keyword_returns_all_tasks
    app_test.go:566: Search() for keyword '' returned 5 tasks:
    app_test.go:568:   - Found Task: Title='Buy groceries'
    app_test.go:568:   - Found Task: Title='Finish report'
    app_test.go:568:   - Found Task: Title='Call John'
    app_test.go:568:   - Found Task: Title='Prepare presentation'
    app_test.go:568:   - Found Task: Title='Grocery shopping list'
--- PASS: TestSearchTasks (0.00s)
    --- PASS: TestSearchTasks/Search_by_title_keyword_'report' (0.00s)
    --- PASS: TestSearchTasks/Search_by_description_keyword_'milk' (0.00s)
    --- PASS: TestSearchTasks/Case-insensitive_search_'grocery' (0.00s)
    --- PASS: TestSearchTasks/Search_by_partial_keyword_'proj' (0.00s)
    --- PASS: TestSearchTasks/No_matching_keyword (0.00s)
    --- PASS: TestSearchTasks/Empty_keyword_returns_all_tasks (0.00s)
PASS
ok  	zan/internal/app	0.235s
```

### 補足

当初、`Case-insensitive search 'grocery'`のテストケースで期待値の誤りがありましたが、修正し、正規表現ではなく元の`strings.Contains`による部分一致検索で要件を満たせることを確認しました。