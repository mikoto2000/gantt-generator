# ganttgen

CSV を唯一の入力として、純 CSS のガントチャート HTML を生成する Go 製 CLI です。稼働日は月〜金のみを考慮し、依存関係に応じてタスクを自動リスケします。任意で YAML 形式の祝日リストを渡すこともできます。

## 必要環境
- Go 1.22 以上

## 使い方
```bash
go run ./cmd/ganttgen [-o output.html] [--holidays holidays.yaml] <input.csv>
# 変更監視しながら生成
go run ./cmd/ganttgen --watch [-o output.html] [--holidays holidays.yaml] <input.csv>
# livereload 付きで監視生成（HTML を開いたまま自動更新）
go run ./cmd/ganttgen --livereload [-o output.html] [--holidays holidays.yaml] <input.csv>
```
デフォルト出力は `gantt.html` です。`-o`/`--output` で出力先を変更できます。`--holidays` で YYYY-MM-DD の配列を持つ YAML を渡すと、その日付を非稼働日として扱います。
`--watch` を付けると CSV の更新を1秒間隔で検知し、都度再生成します（Ctrl+C で終了）。
`--livereload` を付けるとローカルに SSE ベースのライブリロードサーバを立ち上げ、生成 HTML にクライアントスクリプトを埋め込みます。CSV を保存するたびに生成とブラウザ更新まで自動で行います。ポートは `--livereload-port`（デフォルト 35729）で変更できます。

## CSV 形式
ヘッダー必須。列は順不同でも可。日付は `YYYY-MM-DD` / `YYYY/MM/DD` のほか、月日が1桁の場合のゼロ省略（例: `2024-6-3`, `2024/6/3`）も受け付けます。文字コードは UTF-8 / Shift_JIS をヘッダ行から自動判定します。

| 列名 | 型 | 必須 | 説明 |
| --- | --- | --- | --- |
| name | string | ✔︎ | タスク名（ユニーク） |
| start | YYYY-MM-DD |  | 絶対開始日（非稼働日の場合は次稼働日にスライド） |
| end | YYYY-MM-DD |  | 絶対終了日（duration と併用不可、単独指定不可） |
| duration | Nd |  | 稼働日ベースの期間（例: `5d`） |
| depends_on | string list |  | 依存タスク名（`,` または `;` 区切り） |
| actual_start | YYYY-MM-DD |  | 実績開始日（予定と同じ稼働日ルールで補正、予定の計算には影響なし） |
| actual_end | YYYY-MM-DD |  | 実績終了日（actual_duration と併用不可、単独指定不可） |
| actual_duration | Nd |  | 実績期間（稼働日ベース。actual_start とセットで使用） |
| （日本語ヘッダ例）タスク名/開始/終了/期間/依存/実績開始/実績終了/実績期間 |  |  | サンプル CSV のように日本語ヘッダも使用できます（英語ヘッダと同義） |

### 主なバリデーション
- end 単独指定不可 / end と duration 併用不可
- actual_end 単独指定不可 / actual_end と actual_duration 併用不可 / actual_duration のみ指定不可
- name 重複不可
- 存在しないタスクへの depends_on 禁止
- 循環依存禁止
- 全フィールド空はエラー

### 実績について
- 実績列は任意。未指定の場合は予定のみ描画されます。
- 実績の開始・終了・期間は予定と同じく稼働日（週末＋祝日を除外）前提で補正されます。
- 実績はスケジューリングには使わず、ガント上で「予定（青）」と「実績（オレンジ）」を上下に並べて比較表示します。

## サンプル
`sample.csv` を同梱しています。生成例:
```bash
go run ./cmd/ganttgen -o sample.html sample.csv
```

祝日を考慮させる例:
```bash
go run ./cmd/ganttgen --holidays sample_holidays.yaml sample.csv
```
`sample_holidays.yaml` 例:
```yaml
# 配列だけでも OK
holidays:
  - 2024-09-16
  - 2024-09-23
```

## 開発メモ
- レンダリングは HTML 内に `<style>` を埋め込み、CSS Grid で日付軸・バーを配置
- 1 日 30px 幅。今日の縦線を常に表示
- 稼働日判定は週末に加えて、任意の祝日 YAML を考慮可能

## テスト
```bash
go test ./...
```
※ 現在の環境で Go ツールチェインが利用できない場合があります。その際は Go をインストールしてから実行してください。

## ライセンス
MIT License (`LICENSE` を参照)。
