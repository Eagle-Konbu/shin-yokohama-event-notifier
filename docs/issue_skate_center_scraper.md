# KOSÉ新横浜スケートセンターのスクレイパー実装

## Title
KOSÉ新横浜スケートセンターのスクレイパーを実装する

## Description

### 概要
KOSÉ新横浜スケートセンターの公式サイトからイベント情報を取得するスクレイパーを実装します。

### 対象URL
https://ticketjam.jp/venues/3442

### 現状
- `internal/infrastructure/scraper/skate_center.go` にスケルトン実装が存在
- `FetchEvents` メソッドは "not implemented" エラーを返す状態

### 実装要件

1. **イベント情報の取得**
   - KOSÉ新横浜スケートセンターの公式サイトからイベント情報をスクレイピング
   - 当日のイベント情報を取得

2. **データ構造**
   - `event.Event` 型のスライスを返す
   - 必須フィールド: イベント名、開催日、会場ID
   - オプション: 開始時間、終了時間、説明など

3. **エラーハンドリング**
   - HTTP リクエストのエラーを適切に処理
   - HTML パース失敗時のエラーハンドリング
   - イベントが見つからない場合の処理

4. **テスト**
   - ユニットテストの実装（`skate_center_test.go`）
   - モックを使用したテストケースの作成

### 受け入れ基準

- [ ] `FetchEvents` メソッドが正常にイベント情報を取得できる
- [ ] エラーケースが適切に処理される
- [ ] ユニットテストが実装され、すべてパスする
- [ ] コードが既存のコーディング規約に準拠している
- [ ] `task ci-check` が成功する

### 参考情報
- スクレイピング対象: https://ticketjam.jp/venues/3442
- 既存のスケルトン実装: `internal/infrastructure/scraper/skate_center.go`
- インターフェース: `internal/domain/ports/event_fetcher.go`
- ドメインモデル: `internal/domain/event/event.go`

---

## 共通の注意事項

### スクレイピングポリシー
- 各サイトの利用規約を遵守してください
- 過度なアクセスを避け、適切な間隔でリクエストを行ってください
- User-Agent を適切に設定してください

### アーキテクチャ
- ヘキサゴナルアーキテクチャに従ってください
- `domain/ports` のインターフェースを実装してください
- 外部依存（HTTP クライアント等）は注入可能にしてください

### コーディング規約
- INSTRUCTIONS.md に記載されているガイドラインに従ってください
- コミット前に必ず `task tidy` と `task ci-check` を実行してください
- golangci-lint のエラーは `task lint-fix` で自動修正してください
- import の整理は `task goreg-fix` または `goreg -w <file>` で行ってください

### 依存関係
- 可能な限り Go 標準ライブラリを使用してください
- 新しい外部ライブラリの追加が必要な場合は、理由を明確にしてください
