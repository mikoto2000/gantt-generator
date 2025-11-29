# ganttgen

CSV を唯一の入力として、純 CSS のガントチャート HTML を生成する Go 製 CLI です。稼働日は月〜金のみを考慮し、依存関係に応じてタスクを自動リスケします。

## 必要環境
- Go 1.22 以上

## 使い方
```bash
go run ./cmd/ganttgen [-o output.html] <input.csv>
```
デフォルト出力は `gantt.html` です。`-o`/`--output` で出力先を変更できます。

## CSV 形式
ヘッダー必須。列は順不同でも可。

| 列名 | 型 | 必須 | 説明 |
| --- | --- | --- | --- |
| name | string | ✔︎ | タスク名（ユニーク） |
| start | YYYY-MM-DD |  | 絶対開始日（非稼働日の場合は次稼働日にスライド） |
| end | YYYY-MM-DD |  | 絶対終了日（duration と併用不可、単独指定不可） |
| duration | Nd |  | 稼働日ベースの期間（例: `5d`） |
| depends_on | string list |  | 依存タスク名（`,` または `;` 区切り） |

### 主なバリデーション
- end 単独指定不可 / end と duration 併用不可
- name 重複不可
- 存在しないタスクへの depends_on 禁止
- 循環依存禁止
- 全フィールド空はエラー

## サンプル
`sample.csv` を同梱しています。生成例:
```bash
go run ./cmd/ganttgen -o sample.html sample.csv
```

## 開発メモ
- レンダリングは HTML 内に `<style>` を埋め込み、CSS Grid で日付軸・バーを配置
- 1 日 30px 幅。今日の縦線を常に表示
- 稼働日判定は週末のみ考慮（祝日未対応）

## テスト
```bash
go test ./...
```
※ 現在の環境で Go ツールチェインが利用できない場合があります。その際は Go をインストールしてから実行してください。

## ライセンス
MIT License (`LICENSE` を参照)。
