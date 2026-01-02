package renderer

import (
	"bytes"
	"html/template"
	"time"

	"ganttgen/internal/calendar"
)

func renderHTML(ctx renderContext) (string, error) {
	tmpl := template.Must(template.New("page").Funcs(template.FuncMap{
		"formatDate": formatDate,
		"isWeekend":  func(t time.Time) bool { return !calendar.IsWorkday(t) },
		"add1":       func(v int) int { return v + 1 },
		"isOneDay":   func(span int) bool { return span == 1 },
	}).Parse(pageTemplate))

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, ctx); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func baseCSS() string {
	return `
:root {
  --cell-width: 30px;
  --accent: #4c6fff;
  --accent-2: #67b4ff;
  --actual: #f97316;
  --actual-2: #fdba74;
  --line: #e0e5ef;
  --task-border-line: #888888;
  --today: #ff5a5f;
  --bg: #f5f7fb;
  --timeline-header-height: 72px;
  --bar-height: 20px;
  --row-height: 56px;
  --row-gap: 10px;
  --name-col-width: 200px;
  --note-col-width: 240px;
  --status-col-width: 90px;
  --custom-col-width: 140px;
  --heading-row-height: var(--row-height);
}

* { box-sizing: border-box; }
body {
  margin: 16px;
  font-family: "Segoe UI", "Helvetica Neue", system-ui, sans-serif;
  color: #111827;
  background: var(--bg);
}

.page h1 {
  margin-top: 0;
  font-weight: 700;
  letter-spacing: 0.5px;
}

.legend {
  display: flex;
  gap: 12px;
  align-items: center;
}

.legend-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  margin: 8px 0 4px;
}

.filters {
  display: flex;
  flex-wrap: wrap;
  gap: 10px 16px;
  margin: 10px 0 16px;
}

.filter {
  background: #fff;
  border: 1px solid var(--line);
  border-radius: 8px;
  padding: 8px 10px;
  box-shadow: 0 6px 18px rgba(15, 23, 42, 0.05);
}

.filter summary {
  cursor: pointer;
  font-weight: 600;
  color: #111827;
}

.filter-body {
  margin-top: 8px;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.filter-text {
  border: 1px solid var(--line);
  border-radius: 6px;
  padding: 6px 8px;
  font-size: 12px;
}

.filter-values {
  display: flex;
  flex-direction: column;
  gap: 4px;
  max-height: 140px;
  overflow: auto;
}

.filter-value {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font-size: 12px;
  color: #374151;
}

.legend-item {
  display: flex;
  align-items: center;
  gap: 6px;
  color: #4b5563;
  font-weight: 500;
}

.legend-swatch {
  width: 14px;
  height: 14px;
  border-radius: 4px;
  box-shadow: 0 2px 6px rgba(0, 0, 0, 0.12);
}

.legend-swatch.plan { background: linear-gradient(135deg, var(--accent), var(--accent-2)); }
.legend-swatch.actual { background: linear-gradient(135deg, var(--actual), var(--actual-2)); }

.toggle-notes {
  padding: 8px 12px;
  border-radius: 8px;
  border: 1px solid var(--line);
  background: #fff;
  color: #111827;
  font-weight: 600;
  box-shadow: 0 4px 10px rgba(15, 23, 42, 0.08);
  cursor: pointer;
}

.gantt {
  display: grid;
  grid-template-columns: var(--name-col-width) var(--status-col-width) 1fr var(--note-col-width);
  gap: 12px;
  align-items: flex-start;
}

.has-custom .gantt {
  grid-template-columns: var(--name-col-width) var(--status-col-width) repeat(var(--custom-col-count-visible, var(--custom-col-count)), var(--custom-col-width)) 1fr var(--note-col-width);
}

.custom-hidden.has-custom .gantt {
  grid-template-columns: var(--name-col-width) var(--status-col-width) 1fr var(--note-col-width);
}

.name-list {
  display: flex;
  flex-direction: column;
  gap: var(--row-gap);
  padding: 12px 0 16px 0;
  min-width: var(--name-col-width);
}

.name {
  background: #fff;
  border: 1px solid var(--line);
  border-radius: 8px;
  padding: 0 12px;
  font-weight: 500;
  box-shadow: 0 6px 18px rgba(15, 23, 42, 0.05);
  height: var(--row-height);
  display: flex;
  align-items: center;
}

.row-cancelled {
  background: #eef0f3;
  color: #6b7280;
  border-color: #d1d5db;
}

.bar-row.row-cancelled {
  background: #f1f3f6;
  border-radius: 8px;
  box-shadow: 0 calc(-1 * var(--row-gap)) 0 0 #f1f3f6;
}

.heading {
  background: linear-gradient(120deg, #fff, #f7f7ff);
  border: 1px solid var(--line);
  border-radius: 8px;
  padding: 0 12px;
  font-weight: 700;
  color: #0f172a;
  height: var(--heading-row-height);
  display: flex;
  align-items: center;
}

.name.header {
  background: linear-gradient(120deg, #fff, #f0f4ff);
  height: var(--timeline-header-height);
  align-items: flex-end;
  position: sticky;
  top: 0;
  z-index: 4;
}

.status-list {
  display: flex;
  flex-direction: column;
  gap: var(--row-gap);
  padding: 12px 0 16px 0;
  min-width: var(--status-col-width);
}

.status {
  background: #fff;
  border: 1px solid var(--line);
  border-radius: 8px;
  padding: 0 10px;
  font-weight: 600;
  box-shadow: 0 6px 18px rgba(15, 23, 42, 0.05);
  height: var(--row-height);
  display: flex;
  align-items: center;
  justify-content: center;
  color: #374151;
  font-size: 12px;
}

.status.header {
  background: linear-gradient(120deg, #fff, #f0f4ff);
  height: var(--timeline-header-height);
  align-items: flex-end;
  position: sticky;
  top: 0;
  z-index: 4;
}

.status.heading-row {
  background: linear-gradient(120deg, #fff, #f7f7ff);
  font-weight: 700;
  color: #0f172a;
  height: var(--heading-row-height);
}

.status.empty {
  color: transparent;
}

.status.row-cancelled {
  background: #eef0f3;
  color: #6b7280;
  border-color: #d1d5db;
}

.custom-columns {
  display: flex;
  gap: 12px;
  grid-column: span var(--custom-col-count-visible, var(--custom-col-count));
}

.custom-columns.hidden {
  display: none;
}

.custom-list {
  display: flex;
  flex-direction: column;
  gap: var(--row-gap);
  padding: 12px 0 16px 0;
  min-width: var(--custom-col-width);
}

.custom {
  background: #fff;
  border: 1px solid var(--line);
  border-radius: 8px;
  padding: 0 10px;
  font-weight: 500;
  box-shadow: 0 6px 18px rgba(15, 23, 42, 0.05);
  height: var(--row-height);
  display: flex;
  align-items: center;
  color: #374151;
  font-size: 12px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.custom.header {
  background: linear-gradient(120deg, #fff, #f0f4ff);
  height: var(--timeline-header-height);
  align-items: flex-end;
  position: sticky;
  top: 0;
  z-index: 4;
}

.custom.heading-row {
  background: linear-gradient(120deg, #fff, #f7f7ff);
  font-weight: 700;
  color: #0f172a;
  height: var(--heading-row-height);
}

.custom.empty {
  color: transparent;
}

.custom.row-cancelled {
  background: #eef0f3;
  color: #6b7280;
  border-color: #d1d5db;
}

.custom-list.hidden {
  display: none;
}

.column-toggles {
  display: flex;
  flex-wrap: wrap;
  gap: 8px 12px;
  align-items: center;
  font-size: 12px;
  color: #4b5563;
}

.column-toggles-label {
  font-weight: 600;
  color: #374151;
}

.column-toggle {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  cursor: pointer;
}

.column-toggle input {
  cursor: pointer;
}

.row-hidden {
  display: none !important;
}

.timeline-wrapper {
  display: flex;
  flex-direction: column;
  gap: var(--row-gap);
  max-width: 100%;
}

.grid-surface {
  position: relative;
  background: #fff;
  border: 1px solid var(--line);
  border-radius: 12px;
  margin-top: -10px;
  padding: 12px 12px 16px 12px;
  min-width: calc(var(--day-count) * var(--cell-width) + 24px);
  box-shadow: 0 10px 24px rgba(15, 23, 42, 0.08);
}

.timeline-header-wrap {
  position: sticky;
  top: 0;
  z-index: 3;
}

.timeline-header-scroll,
.timeline-body-scroll {
  overflow-x: auto;
  overflow-y: visible;
}

.timeline-header-surface {
  position: relative;
  background: #fff;
  border: 1px solid var(--line);
  border-radius: 12px;
  padding: 12px 12px 0 12px;
  min-width: calc(var(--day-count) * var(--cell-width) + 24px);
  box-shadow: 0 6px 18px rgba(15, 23, 42, 0.08);
}

.grid-surface::before {
  content: "";
  position: absolute;
  top: 0;
  bottom: 0;
  left: 12px;
  width: calc(var(--day-count) * var(--cell-width));
  background-image: repeating-linear-gradient(90deg, #eef1f7 0, #eef1f7 1px, transparent 1px, transparent var(--cell-width));
  pointer-events: none;
  z-index: 0;
}

.today-line {
  position: absolute;
  top: 0;
  bottom: 0;
  left: calc(12px + var(--today-index) * var(--cell-width));
  width: 2px;
  background: var(--today);
  z-index: 2;
}

.grid {
  display: grid;
  grid-template-columns: repeat(var(--day-count), var(--cell-width));
  align-items: start;
}

.timeline-grid {
  padding-bottom: 0;
  min-height: var(--timeline-header-height);
  height: var(--timeline-header-height);
  border-bottom: 0;
  position: relative;
  z-index: 1;
}

.day {
  font-size: 11px;
  line-height: 1.2;
  height: var(--timeline-header-height);
  color: #1f2937;
  display: grid;
  place-items: start center;
}

.day-label {
  writing-mode: vertical-rl;
  white-space: nowrap;
  padding-top: 8px;
}

.day.weekend {
  color: #9ca3af;
}

.bars {
  display: flex;
  flex-direction: column;
  gap: var(--row-gap);
  padding-top: 0;
}

.bar-row {
  position: relative;
  align-items: start;
  min-height: var(--row-height);
  grid-auto-rows: max-content;
  row-gap: 6px;
  padding: 2px 0;
  z-index: 1;
  border-bottom: 1px dashed var(--task-border-line);
}

.bar {
  height: var(--bar-height);
  border-radius: 8px;
  background: linear-gradient(135deg, var(--accent), var(--accent-2));
  color: #fff;
  display: flex;
  align-items: center;
  padding: 0 10px;
  font-size: 13px;
  white-space: nowrap;
  box-shadow: 0 6px 14px rgba(76, 111, 255, 0.25);
}

.bar.one-day {
  justify-content: center;
  padding: 0;
}

.bar.actual {
  background: linear-gradient(135deg, var(--actual), var(--actual-2));
  color: #0f172a;
  box-shadow: 0 5px 12px rgba(249, 115, 22, 0.28);
}

.heading-spacer {
  height: var(--heading-row-height);
  border-bottom: 1px dashed var(--task-border-line);
}

.notes-list {
  display: flex;
  flex-direction: column;
  gap: var(--row-gap);
  padding: 12px 0 16px 0;
  min-width: var(--note-col-width);
}

.note {
  background: #fff;
  border: 1px solid var(--line);
  border-radius: 8px;
  padding: 10px 12px;
  font-size: 13px;
  color: #111827;
  box-shadow: 0 6px 18px rgba(15, 23, 42, 0.05);
  min-height: var(--row-height);
  display: flex;
  align-items: flex-start;
  justify-content: flex-start;
  white-space: pre-wrap;
  overflow: auto;
  cursor: pointer;
  transition: max-height 0.2s ease;
}

.note.row-cancelled {
  background: #eef0f3;
  color: #6b7280;
  border-color: #d1d5db;
}

.note.header {
  background: linear-gradient(120deg, #fff, #f0f4ff);
  font-weight: 700;
  height: var(--timeline-header-height);
  align-items: flex-end;
  cursor: pointer;
  position: sticky;
  top: 0;
  z-index: 4;
}

.note.empty {
  background: transparent;
  border: none;
  box-shadow: none;
  cursor: default;
}

.note.empty.row-cancelled {
  background: #eef0f3;
  border: 1px solid #d1d5db;
  box-shadow: none;
}

.notes-hidden .notes-list {
  display: none;
}
.notes-hidden .gantt {
  grid-template-columns: var(--name-col-width) var(--status-col-width) 1fr;
}

.notes-hidden.has-custom .gantt {
  grid-template-columns: var(--name-col-width) var(--status-col-width) repeat(var(--custom-col-count-visible, var(--custom-col-count)), var(--custom-col-width)) 1fr;
}

.custom-hidden.notes-hidden.has-custom .gantt {
  grid-template-columns: var(--name-col-width) var(--status-col-width) 1fr;
}

.row-name, .row-bar {
  min-height: var(--row-height);
  transition: min-height 0.2s ease;
}
`
}

func formatDate(t time.Time) string {
	return t.Format("2006-01-02")
}

const pageTemplate = `<!DOCTYPE html>
<html lang="ja">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>Gantt Chart</title>
  <style>{{.CSS}}</style>
</head>
<body{{if .HasCustomColumns}} class="has-custom"{{end}}>
  <div class="page">
    <h1>Gantt Chart</h1>
    <div class="legend-row">
      <div class="legend">
        <div class="legend-item"><span class="legend-swatch plan"></span><span>予定</span></div>
        {{if .HasActual}}<div class="legend-item"><span class="legend-swatch actual"></span><span>実績</span></div>{{end}}
      </div>
      {{if .HasCustomColumns}}
      <div class="column-toggles" id="custom-column-toggles">
        <span class="column-toggles-label">列</span>
        {{range $i, $name := .CustomColumns}}
          <label class="column-toggle"><input type="checkbox" data-custom-col="{{$i}}" checked> {{$name}}</label>
        {{end}}
      </div>
      {{end}}
      {{if .HasNotes}}<button id="toggle-notes" class="toggle-notes" type="button">備考を隠す</button>{{end}}
    </div>
    {{if .FilterColumns}}
    <div class="filters" id="filters">
      {{range .FilterColumns}}
      <details class="filter" data-filter="{{.Key}}">
        <summary>{{.Label}}</summary>
        <div class="filter-body">
          <input class="filter-text" type="text" data-filter-text="{{.Key}}" placeholder="キーワード">
          {{ $key := .Key }}
          <div class="filter-values">
            {{range .Values}}
              {{if eq . ""}}
                <label class="filter-value"><input type="checkbox" data-filter-value="{{$key}}" value="__empty__"> (空)</label>
              {{else}}
                <label class="filter-value"><input type="checkbox" data-filter-value="{{$key}}" value="{{.}}"> {{.}}</label>
              {{end}}
            {{end}}
          </div>
        </div>
      </details>
      {{end}}
    </div>
    {{end}}
    <div class="gantt" style="--day-count:{{.DayCount}};--today-index:{{.TodayIndex}};--custom-col-count:{{.CustomColumnCount}};--custom-col-count-visible:{{.CustomColumnCount}};">
      <div class="name-list">
        <div class="name header">Task</div>
        {{range $i, $row := .Rows}}
          {{if $row.Heading}}
            <div class="heading row-name" data-row="{{$i}}" data-heading="true" data-name="{{$row.FilterName}}" data-status="{{$row.FilterStatus}}" data-notes="{{$row.FilterNotes}}"{{range $ci, $cname := $.CustomColumns}} data-custom-{{$ci}}="{{index $row.CustomValues $ci}}"{{end}}>{{$row.Heading}}</div>
          {{else if $row.DisplayOnly}}
            <div class="name row-name" data-row="{{$i}}" data-name="{{$row.FilterName}}" data-status="{{$row.FilterStatus}}" data-notes="{{$row.FilterNotes}}"{{range $ci, $cname := $.CustomColumns}} data-custom-{{$ci}}="{{index $row.CustomValues $ci}}"{{end}}>{{$row.DisplayOnly}}</div>
          {{else if $row.Task}}
            <div class="name row-name{{if $row.Task.Cancelled}} row-cancelled{{end}}" data-row="{{$i}}" data-name="{{$row.FilterName}}" data-status="{{$row.FilterStatus}}" data-notes="{{$row.FilterNotes}}"{{range $ci, $cname := $.CustomColumns}} data-custom-{{$ci}}="{{index $row.CustomValues $ci}}"{{end}}>{{$row.Task.Name}}</div>
          {{end}}
        {{end}}
      </div>
      <div class="status-list">
        <div class="status header">状態</div>
        {{range $i, $row := .Rows}}
          {{if $row.Heading}}
            {{if $row.HeadingStatus}}
              <div class="status heading-row" data-row="{{$i}}">{{$row.HeadingStatus}}</div>
            {{else}}
              <div class="status heading-row" data-row="{{$i}}">&nbsp;</div>
            {{end}}
          {{else if $row.DisplayOnly}}
            <div class="status empty" data-row="{{$i}}"></div>
          {{else if $row.Task}}
            {{if $row.Task.Status}}
              <div class="status{{if $row.Task.Cancelled}} row-cancelled{{end}}" data-row="{{$i}}">{{$row.Task.Status}}</div>
            {{else}}
              <div class="status empty" data-row="{{$i}}"></div>
            {{end}}
          {{end}}
        {{end}}
      </div>
      {{if .HasCustomColumns}}
      <div class="custom-columns">
        {{range $colIndex, $colName := .CustomColumns}}
        <div class="custom-list" data-col="{{$colIndex}}">
          <div class="custom header">{{$colName}}</div>
          {{range $i, $row := $.Rows}}
            {{$value := index $row.CustomValues $colIndex}}
            {{if $value}}
              <div class="custom custom-cell{{if $row.Heading}} heading-row{{end}}{{if $row.Task}}{{if $row.Task.Cancelled}} row-cancelled{{end}}{{end}}" data-row="{{$i}}">{{$value}}</div>
            {{else}}
              <div class="custom empty custom-cell{{if $row.Heading}} heading-row{{end}}{{if $row.Task}}{{if $row.Task.Cancelled}} row-cancelled{{end}}{{end}}" data-row="{{$i}}"></div>
            {{end}}
          {{end}}
        </div>
        {{end}}
      </div>
      {{end}}
      <div class="timeline-wrapper">
        <div class="timeline-header-wrap">
          <div class="timeline-header-scroll">
            <div class="timeline-header-surface">
              <div class="timeline-grid grid">
                {{range .Days}}
                  <div class="day{{if isWeekend .}} weekend{{end}}"><span class="day-label">{{formatDate .}}</span></div>
                {{end}}
              </div>
            </div>
          </div>
        </div>
        <div class="timeline-body-scroll">
          <div class="grid-surface">
            <div class="today-line"></div>
            <div class="bars">
              {{range $i, $row := .Rows}}
                {{if $row.Heading}}
                  <div class="heading-spacer row-bar" data-row="{{$i}}"></div>
                {{else if $row.DisplayOnly}}
                  <div class="heading-spacer row-bar" data-row="{{$i}}"></div>
              {{else if $row.Task}}
                <div class="bar-row grid row-bar{{if $row.Task.Cancelled}} row-cancelled{{end}}" data-row="{{$i}}">
                  <div class="bar plan{{if isOneDay $row.Task.Span}} one-day{{end}}" style="grid-column:{{add1 $row.Task.StartIndex}} / span {{$row.Task.Span}};" title="予定: {{formatDate $row.Task.Start}} - {{formatDate $row.Task.End}}">予定</div>
                  {{if $row.Task.Actual}}
                    <div class="bar actual{{if isOneDay $row.Task.Actual.Span}} one-day{{end}}" style="grid-column:{{add1 $row.Task.Actual.StartIndex}} / span {{$row.Task.Actual.Span}};" title="実績: {{formatDate $row.Task.Actual.Start}} - {{formatDate $row.Task.Actual.End}}">実績</div>
                  {{end}}
                </div>
              {{end}}
              {{end}}
            </div>
          </div>
        </div>
      </div>
      {{if .HasNotes}}
      <div class="notes-list">
        <div class="note header">備考</div>
        {{range $i, $row := .Rows}}
          {{if $row.Heading}}
            {{if $row.HeadingNotes}}
              <div class="note row-note" data-row="{{$i}}">{{$row.HeadingNotes}}</div>
            {{else}}
              <div class="note empty row-note" data-row="{{$i}}"></div>
            {{end}}
          {{else if $row.DisplayOnly}}
            {{if $row.DisplayOnlyNotes}}
              <div class="note row-note" data-row="{{$i}}">{{$row.DisplayOnlyNotes}}</div>
            {{else}}
              <div class="note empty row-note" data-row="{{$i}}"></div>
            {{end}}
          {{else if $row.Task}}
            {{if $row.Task.Notes}}
              <div class="note row-note{{if $row.Task.Cancelled}} row-cancelled{{end}}" data-row="{{$i}}">{{$row.Task.Notes}}</div>
            {{else}}
              <div class="note empty row-note{{if $row.Task.Cancelled}} row-cancelled{{end}}" data-row="{{$i}}"></div>
            {{end}}
          {{end}}
        {{end}}
      </div>
      {{end}}
    </div>
  </div>
  {{if .LiveReloadURL}}
  <script>
  (function() {
    try {
      var es = new EventSource('{{.LiveReloadURL}}');
      es.onmessage = function(ev) {
        if (ev.data === 'reload') { location.reload(); }
      };
    } catch (e) {
      console.warn('LiveReload unavailable', e);
    }
  })();
  </script>
  {{end}}
  {{if .HasNotes}}
  <script>
    (function() {
      var btn = document.getElementById('toggle-notes');
      var header = document.querySelector('.note.header');
      var toggle = function() {
        document.body.classList.toggle('notes-hidden');
        if (btn) {
          btn.textContent = document.body.classList.contains('notes-hidden') ? '備考を表示' : '備考を隠す';
        }
      };
      if (btn) {
        btn.addEventListener('click', toggle);
      }
      if (header) {
        header.addEventListener('click', toggle);
        header.title = 'クリックで備考の表示/非表示を切替';
      }

      var headerScroll = document.querySelector('.timeline-header-scroll');
      var bodyScroll = document.querySelector('.timeline-body-scroll');
      if (headerScroll && bodyScroll) {
        var syncing = false;
        var syncTo = function(source, target) {
          if (syncing) return;
          syncing = true;
          target.scrollLeft = source.scrollLeft;
          syncing = false;
        };
        headerScroll.addEventListener('scroll', function() { syncTo(headerScroll, bodyScroll); });
        bodyScroll.addEventListener('scroll', function() { syncTo(bodyScroll, headerScroll); });
      }

      var rowHeight = getComputedStyle(document.documentElement).getPropertyValue('--row-height').trim();
      var rowHeightValue = parseFloat(rowHeight);
      var notes = document.querySelectorAll('.row-note[data-row]');
      notes.forEach(function(note) {
        if (note.classList.contains('empty')) return;
        var rowId = note.getAttribute('data-row');
        var name = document.querySelector('.row-name[data-row="' + rowId + '"]');
        var status = document.querySelector('.status[data-row="' + rowId + '"]');
        var bar = document.querySelector('.row-bar[data-row="' + rowId + '"]');
        var customCells = document.querySelectorAll('.custom-cell[data-row="' + rowId + '"]');
        var expanded = false;

        var applyHeight = function(h) {
          var hp = h + 'px';
          note.style.maxHeight = hp;
          if (name) name.style.minHeight = hp;
          if (status) status.style.minHeight = hp;
          if (bar) bar.style.minHeight = hp;
          if (customCells.length) {
            customCells.forEach(function(cell) {
              cell.style.minHeight = hp;
            });
          }
        };

        var collapse = function() {
          expanded = false;
          note.classList.remove('expanded');
          applyHeight(rowHeightValue);
        };
        var expand = function() {
          expanded = true;
          note.classList.add('expanded');
          var h = Math.max(note.scrollHeight, rowHeightValue);
          applyHeight(h);
        };

        collapse();
        note.addEventListener('click', function() {
          if (expanded) {
            collapse();
          } else {
            expand();
          }
        });
      });
    })();
  </script>
  {{end}}
  {{if .HasCustomColumns}}
  <script>
    (function() {
      var container = document.getElementById('custom-column-toggles');
      if (!container) return;
      var checkboxes = container.querySelectorAll('input[data-custom-col]');
      var columnWrap = document.querySelector('.custom-columns');
      var gantt = document.querySelector('.gantt');
      var update = function() {
        var visible = 0;
        checkboxes.forEach(function(cb) {
          var idx = cb.getAttribute('data-custom-col');
          var lists = document.querySelectorAll('.custom-list[data-col="' + idx + '"]');
          if (cb.checked) {
            visible++;
            lists.forEach(function(list) { list.classList.remove('hidden'); });
          } else {
            lists.forEach(function(list) { list.classList.add('hidden'); });
          }
        });
        if (gantt) {
          gantt.style.setProperty('--custom-col-count-visible', visible);
        }
        if (columnWrap) {
          if (visible === 0) {
            columnWrap.classList.add('hidden');
            document.body.classList.add('custom-hidden');
          } else {
            columnWrap.classList.remove('hidden');
            document.body.classList.remove('custom-hidden');
          }
        }
      };
      checkboxes.forEach(function(cb) { cb.addEventListener('change', update); });
      update();
    })();
  </script>
  {{end}}
  {{if .FilterColumns}}
  <script>
    (function() {
      var filters = document.getElementById('filters');
      if (!filters) return;
      var rowNames = document.querySelectorAll('.row-name[data-row]');
      if (!rowNames.length) return;

      var readFilters = function() {
        var states = [];
        var filterBlocks = filters.querySelectorAll('[data-filter]');
        filterBlocks.forEach(function(block) {
          var key = block.getAttribute('data-filter');
          var textInput = block.querySelector('[data-filter-text="' + key + '"]');
          var text = textInput ? textInput.value.trim().toLowerCase() : '';
          var selected = {};
          var anySelected = false;
          var checks = block.querySelectorAll('[data-filter-value="' + key + '"]');
          checks.forEach(function(cb) {
            if (cb.checked) {
              anySelected = true;
              selected[cb.value] = true;
            }
          });
          states.push({ key: key, text: text, selected: selected, anySelected: anySelected });
        });
        return states;
      };

      var getRowValue = function(rowEl, key) {
        if (key === 'name') return rowEl.getAttribute('data-name') || '';
        if (key === 'status') return rowEl.getAttribute('data-status') || '';
        if (key === 'notes') return rowEl.getAttribute('data-notes') || '';
        if (key.indexOf('custom-') === 0) {
          var idx = key.slice('custom-'.length);
          return rowEl.getAttribute('data-custom-' + idx) || '';
        }
        return '';
      };

      var rowMatches = function(rowEl, states) {
        for (var i = 0; i < states.length; i++) {
          var st = states[i];
          var value = getRowValue(rowEl, st.key);
          var trimmed = value.trim();
          if (st.anySelected) {
            var token = trimmed === '' ? '__empty__' : trimmed;
            if (!st.selected[token]) return false;
          }
          if (st.text) {
            if (trimmed.toLowerCase().indexOf(st.text) === -1) return false;
          }
        }
        return true;
      };

      var applyFilters = function() {
        var states = readFilters();
        var matches = [];
        rowNames.forEach(function(rowEl, idx) {
          matches[idx] = rowMatches(rowEl, states);
        });

        // Keep section headers when any row in the section matches.
        for (var i = 0; i < rowNames.length; i++) {
          var rowEl = rowNames[i];
          if (rowEl.getAttribute('data-heading') !== 'true') continue;
          if (matches[i]) continue;
          var hasVisible = false;
          for (var j = i + 1; j < rowNames.length; j++) {
            if (rowNames[j].getAttribute('data-heading') === 'true') break;
            if (matches[j]) {
              hasVisible = true;
              break;
            }
          }
          matches[i] = hasVisible;
        }

        rowNames.forEach(function(rowEl, idx) {
          var rowId = rowEl.getAttribute('data-row');
          var rowEls = document.querySelectorAll('[data-row="' + rowId + '"]');
          rowEls.forEach(function(el) {
            if (matches[idx]) {
              el.classList.remove('row-hidden');
            } else {
              el.classList.add('row-hidden');
            }
          });
        });
      };

      filters.addEventListener('input', applyFilters);
      filters.addEventListener('change', applyFilters);
      applyFilters();
    })();
  </script>
  {{end}}
</body>
</html>`
