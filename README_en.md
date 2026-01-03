[日本語](README.md) / **English**

# ganttgen

Go CLI that generates a single HTML Gantt chart from CSV.

<img width="1391" height="1130" alt="image" src="https://github.com/user-attachments/assets/760d1ddc-d09b-4759-8d76-51da8340d390" />

[I made a Gantt chart tool - 003 You can now write notes](https://youtube.com/shorts/iGupb_EMSDc)

## Features:

- Generate a Gantt chart from a simple CSV input
- Output is a single HTML file. No extra server or resources needed
- Show both planned and actual schedules
- Workdays are Mon-Fri and tasks are auto-rescheduled based on dependencies
- Optional YAML holiday list is supported
- Watch mode for auto-regeneration on CSV updates
- Built-in Livereload server for automatic browser refresh


## Getting Started

Install with the command below, or download a binary from the releases page.

```sh
go install github.com/mikoto2000/ganttgen@latest
```


## Usage:

```sh
Usage of ./dist/ganttgen:
  -holidays string
        optional YAML file listing YYYY-MM-DD holidays
  -all-workdays
        treat weekends and holidays as workdays
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

Run `ganttgen <input.csv>` to generate an HTML Gantt chart from a CSV file.

By default, the output is `gantt.html` in the same directory as the input CSV. You can change the output with `-o`/`--output`. With `--holidays`, pass a YAML file that contains a list of YYYY-MM-DD holidays; those dates are treated as non-working days.
Add `--all-workdays` to treat weekends and holidays as working days.
Add `--gen-template` to output an empty CSV template with the same header as `sample/sample.csv`, then exit.

With `--watch`, the tool checks for CSV updates every second and regenerates on changes (exit with Ctrl+C).

With `--livereload`, a local SSE-based livereload server is started and a client script is embedded in the generated HTML. Each CSV save triggers regeneration and browser refresh. The port can be changed with `--livereload-port` (default 35729).


### Command Examples

```sh
# Generate while watching for changes
ganttgen --watch [-o output.html] [--holidays holidays.yaml] <input.csv>

# Generate with livereload (auto refresh while HTML is open)
ganttgen --livereload [-o output.html] [--holidays holidays.yaml] <input.csv>
```


## Input Format

### CSV Format

Header is required. Column order does not matter. Dates accept `YYYY-MM-DD` / `YYYY/MM/DD`, and also allow single-digit month/day without zero padding (e.g. `2024-6-3`, `2024/6/3`).

Rows starting with `#` in the first column are treated as section headings, and the section name is displayed in the chart.

Encoding is auto-detected from the header line: UTF-8 or Shift_JIS.

| Column (JP label) | Type | Required | Description |
| --- | --- | --- | --- |
| name(タスク名) | string | ✔︎ | Task name (unique) |
| status(状態) | string |  | `cancelled` / `中止` marks the task as cancelled |
| progress(進捗) | 0-100(%) |  | Progress percentage (0-100, trailing `%` is allowed) |
| start(開始) | YYYY-MM-DD |  | Absolute start date (moved to next workday if needed) |
| end(終了) | YYYY-MM-DD |  | Absolute end date (cannot be combined with duration, cannot be alone) |
| duration(期間) | Nd |  | Duration in workdays (e.g. `5d`) |
| depends_on(依存) | string list |  | Dependency task names (`,` or `;` separated) |
| actual_start(実績開始) | YYYY-MM-DD |  | Actual start date (same workday rules; does not affect planned schedule) |
| actual_end(実績終了) | YYYY-MM-DD |  | Actual end date (cannot be combined with actual_duration, cannot be alone) |
| actual_duration(実績期間) | Nd |  | Actual duration in workdays (used with actual_start) |
| notes(備考) | string |  | Task notes (shown on the chart) |

If the `progress(進捗)` column exists, the planned bar color changes according to progress.

Columns not listed above are treated as custom columns and shown on the left side of the HTML.

Japanese headers are accepted, as in the sample CSV, and are equivalent to the English headers.

CSV sample:

```csv
タスク名,状態,進捗,開始,終了,期間,依存,実績開始,実績終了,実績期間,備考
#要件定義,,,,,,,,,,
タスク1,,45%,2025/12/1,,2d,,2025/12/11,2025/12/15,,,備考1
#設計,,,,,,,,,,
タスク2,,,,3d,タスク1,,,,,備考2
```

See `sample/sample.csv`. Opening it in a spreadsheet app is recommended.


### Holidays YAML Format

```yaml
# A plain list is also OK
holidays:
  - 2025-01-01
  - 2025-01-08
  - 2025-02-11
  # ...
```


## Main Validations

- `end` cannot be specified alone / cannot be combined with `duration`
- `actual_end` cannot be specified alone / cannot be combined with `actual_duration` / `actual_duration` cannot be used alone
- `name` must be unique
- `depends_on` cannot reference unknown tasks
- Circular dependencies are not allowed
- A row with all empty fields is an error


## About Actuals

- Actual columns are optional. If missing, only the planned bars are rendered.
- Actual start/end/duration are adjusted using the same workday rules (exclude weekends and holidays).
- Actuals are not used for scheduling; the chart shows planned (blue) and actual (orange) bars stacked for comparison.
- Rows with all fields empty are ignored (not treated as errors).
- Task order follows the CSV row order (no sorting).


## Samples

`sample.csv` is included. Example:

```bash
ganttgen --holidays sample/sample_holiday.yaml sample/sample.csv
```


## Build:

```bash
make test
make
```

The binary is generated in `dist`.


## License:

MIT License (see `LICENSE`).


## Author:

mikoto2000 <mikoto2000@gmail.com>
