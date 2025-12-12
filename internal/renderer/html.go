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
  --today: #ff5a5f;
  --bg: #f5f7fb;
  --timeline-header-height: 72px;
  --bar-height: 20px;
  --row-height: 56px;
  --row-gap: 10px;
  --name-col-width: 200px;
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
  margin: 8px 0 4px;
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

.gantt {
  display: flex;
  gap: 12px;
  align-items: flex-start;
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

.heading {
  background: linear-gradient(120deg, #fff, #f7f7ff);
  border: 1px solid var(--line);
  border-radius: 8px;
  padding: 0 12px;
  font-weight: 700;
  color: #0f172a;
  height: var(--row-height);
  display: flex;
  align-items: center;
}

.heading-spacer {
  height: var(--row-height);
}

.name.header {
  background: linear-gradient(120deg, #fff, #f0f4ff);
  height: var(--timeline-header-height);
  align-items: flex-end;
}

.timeline-wrapper {
  overflow: auto;
  max-width: 100%;
}

.grid-surface {
  position: relative;
  background: #fff;
  border: 1px solid var(--line);
  border-radius: 12px;
  padding: 12px 12px 16px 12px;
  min-width: calc(var(--day-count) * var(--cell-width) + 24px);
  box-shadow: 0 10px 24px rgba(15, 23, 42, 0.08);
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
  padding-top: var(--row-gap);
}

.bar-row {
  position: relative;
  align-items: start;
  min-height: var(--row-height);
  grid-auto-rows: max-content;
  row-gap: 6px;
  padding: 2px 0;
  z-index: 1;
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
  box-shadow: 0 6px 14px rgba(76, 111, 255, 0.25);
}

.bar.actual {
  background: linear-gradient(135deg, var(--actual), var(--actual-2));
  color: #0f172a;
  box-shadow: 0 5px 12px rgba(249, 115, 22, 0.28);
}

.heading-spacer {
  height: var(--heading-row-height);
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
<body>
  <div class="page">
    <h1>Gantt Chart</h1>
    <div class="legend">
      <div class="legend-item"><span class="legend-swatch plan"></span><span>予定</span></div>
      {{if .HasActual}}<div class="legend-item"><span class="legend-swatch actual"></span><span>実績</span></div>{{end}}
    </div>
    <div class="gantt" style="--day-count:{{.DayCount}};--today-index:{{.TodayIndex}};">
      <div class="name-list">
        <div class="name header">Task</div>
        {{range .Rows}}
          {{if .Heading}}
            <div class="heading">{{.Heading}}</div>
          {{else if .Task}}
            <div class="name">{{.Task.Name}}</div>
          {{end}}
        {{end}}
      </div>
      <div class="timeline-wrapper">
        <div class="grid-surface">
          <div class="today-line"></div>
          <div class="timeline-grid grid">
            {{range .Days}}
              <div class="day{{if isWeekend .}} weekend{{end}}"><span class="day-label">{{formatDate .}}</span></div>
            {{end}}
          </div>
          <div class="bars">
            {{range .Rows}}
              {{if .Heading}}
                <div class="heading-spacer"></div>
              {{else if .Task}}
                <div class="bar-row grid">
                  <div class="bar plan" style="grid-column:{{add1 .Task.StartIndex}} / span {{.Task.Span}};" title="予定: {{formatDate .Task.Start}} - {{formatDate .Task.End}}">予定</div>
                  {{if .Task.Actual}}
                    <div class="bar actual" style="grid-column:{{add1 .Task.Actual.StartIndex}} / span {{.Task.Actual.Span}};" title="実績: {{formatDate .Task.Actual.Start}} - {{formatDate .Task.Actual.End}}">実績</div>
                  {{end}}
                </div>
              {{end}}
            {{end}}
          </div>
        </div>
      </div>
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
</body>
</html>`
