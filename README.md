# ganttgen

CSV から、ガントチャートの単一 HTML を生成する Go 製 CLI です。

<img width="1057" height="844" alt="image" src="https://github.com/user-attachments/assets/f82307a5-cd29-4b24-ab81-ba3c6c5298aa" />

[ガントチャート作成ツールを作った - 003 備考が書けるようになりました](https://youtube.com/shorts/iGupb_EMSDc)

## Features:

- シンプルな CSV 入力でガントチャートを生成
- 単一 HTML ファイルに完結。追加のサーバやリソース不要
- 予定と実績の両方を表示可能
- 稼働日は月〜金とし、依存関係に応じてタスクを自動リスケジューリング
- 任意で yaml 形式の祝日リストを渡すことも可能
- 変更監視モードで、CSV 更新時に自動再生成
- Livereload サーバを内蔵し、ブラウザで自動更新も可能


## Getting Started

以下コマンドでインストールするか、リリースページよりバイナリをダウンロードしてください。

```sh
go install github.com/mikoto2000/ganttgen@latest
```


## Usage:

```sh
Usage of ./dist/ganttgen:
  -holidays string
        optional YAML file listing YYYY-MM-DD holidays
  -holidays-as-workdays
        treat holidays as workdays even if --holidays is provided
  -gen-template string
        output an empty CSV template and exit
  -livereload
        enable livereload server and inject client script
  -livereload-port int
        port for livereload server (default 35729) (default 35729)
  -o string
        output HTML file (default "gantt.html")
  -output string
        output HTML file (default "gantt.html")
  -version
        print version and exit
  -watch
        watch input CSV and regenerate on changes
```

`ganttgen <input.csv>` で CSV からガントチャート HTML を生成します。

デフォルト出力は入力 CSV と同じディレクトリの `gantt.html` です。`-o`/`--output` で出力先を変更できます。`--holidays` で YYYY-MM-DD の配列を持つ yaml を渡すと、その日付を非稼働日として扱います。
`--holidays-as-workdays` を付けると、`--holidays` を指定していても祝日を稼働日として扱います。
`--gen-template` を付けると、`sample/sample.csv` と同じヘッダを持つ空の CSV テンプレートを出力して終了します。

`--watch` を付けると CSV の更新を1秒間隔で検知し、都度再生成します（Ctrl+C で終了）。

`--livereload` を付けるとローカルに SSE ベースのライブリロードサーバを立ち上げ、生成 HTML にクライアントスクリプトを埋め込みます。CSV を保存するたびに生成とブラウザ更新まで自動で行います。ポートは `--livereload-port`（デフォルト 35729）で変更できます。


### コマンド例

```sh
# 変更監視しながら生成
ganttgen --watch [-o output.html] [--holidays holidays.yaml] <input.csv>

# Livereload 付きで監視生成（HTML を開いたまま自動更新）
ganttgen --livereload [-o output.html] [--holidays holidays.yaml] <input.csv>
```


## 入力フォーマット

### CSV 形式

ヘッダー必須。列は順不同でも可。日付は `YYYY-MM-DD` / `YYYY/MM/DD` のほか、月日が1桁の場合のゼロ省略（例: `2024-6-3`, `2024/6/3`）も受け付けます。

先頭列が `#` で始まる行はセクション区切りとして扱います。セクション名はガントチャート上に表示されます。

文字コードは UTF-8 / Shift_JIS をヘッダ行から自動判定します。

| 列英名(日本語名) | 型 | 必須 | 説明 |
| --- | --- | --- | --- |
| name(タスク名) | string | ✔︎ | タスク名（ユニーク） |
| start(開始) | YYYY-MM-DD |  | 絶対開始日（非稼働日の場合は次稼働日にスライド） |
| end(終了) | YYYY-MM-DD |  | 絶対終了日（duration と併用不可、単独指定不可） |
| duration(期間) | Nd |  | 稼働日ベースの期間（例: `5d`） |
| depends_on(依存) | string list |  | 依存タスク名（`,` または `;` 区切り） |
| actual_start(実績開始) | YYYY-MM-DD |  | 実績開始日（予定と同じ稼働日ルールで補正、予定の計算には影響なし） |
| actual_end(実績終了) | YYYY-MM-DD |  | 実績終了日（actual_duration と併用不可、単独指定不可） |
| actual_duration(実績期間) | Nd |  | 実績期間（稼働日ベース。actual_start とセットで使用） |
| notes(備考) | string |  | タスク備考（ガントチャート上に表示） |

サンプル CSV のように日本語ヘッダも使用できます（英語ヘッダと同義）。

CSV サンプル:

```csv
タスク名,開始,終了,期間,依存,実績開始,実績終了,実績期間,備考
#要件定義,,,,,,,,
タスク1,2025/12/1,,2d,,2025/12/11,2025/12/15,,備考1
#設計,,,,,,,,
タスク2,,,3d,タスク1,,,,備考2
```

`sample/sample.csv` を参照表計算アプリで開くのを推奨。


### 祝日 yaml 形式

```yaml
# 配列だけでも OK
holidays:
  - 2025-01-01
  - 2025-01-08
  - 2025-02-11
  # ...
```


## 主なバリデーション

- end 単独指定不可 / end と duration 併用不可
- actual_end 単独指定不可 / actual_end と actual_duration 併用不可 / actual_duration のみ指定不可
- name 重複不可
- 存在しないタスクへの depends_on 禁止
- 循環依存禁止
- 全フィールド空はエラー


## 実績について

- 実績列は任意。未指定の場合は予定のみ描画されます。
- 実績の開始・終了・期間は予定と同じく稼働日（週末＋祝日を除外）前提で補正されます。
- 実績はスケジューリングには使わず、ガント上で「予定（青）」と「実績（オレンジ）」を上下に並べて比較表示します。
- 全カラム空の行は無視します（エラーにしません）。
- タスク表示順は CSV の行順を維持します（並び替えしません）。


## サンプル

`sample.csv` を同梱しています。生成例:

```bash
ganttgen --holidays sample/sample_holiday.yaml sample/sample.csv
```


## Build:

```bash
make test
make
```

`dist` にバイナリが生成されます。


## License:

MIT License (`LICENSE` を参照)。


## Author:

mikoto2000 <mikoto2000@gmail.com>
