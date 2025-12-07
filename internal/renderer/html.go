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
  --line: #e0e5ef;
  --today: #ff5a5f;
  --bg: #f5f7fb;
  --timeline-header-height: 72px;
  --row-height: 32px;
  --row-gap: 8px;
  --name-col-width: 200px;
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
  align-items: center;
  height: var(--row-height);
  z-index: 1;
}

.bar {
  height: 24px;
  border-radius: 8px;
  background: linear-gradient(135deg, var(--accent), var(--accent-2));
  color: #fff;
  display: flex;
  align-items: center;
  padding: 0 10px;
  font-size: 13px;
  box-shadow: 0 6px 14px rgba(76, 111, 255, 0.25);
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
    <div class="gantt" style="--day-count:{{.DayCount}};--today-index:{{.TodayIndex}};">
      <div class="name-list">
        <div class="name header">Task</div>
        {{range .Tasks}}
          <div class="name">{{.Name}}</div>
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
            {{range .Tasks}}
              <div class="bar-row grid">
                <div class="bar" style="grid-column:{{add1 .StartIndex}} / span {{.Span}};" title="{{formatDate .Start}} - {{formatDate .End}}">{{.Name}}</div>
              </div>
            {{end}}
          </div>
        </div>
      </div>
    </div>
  </div>
</body>
</html>`
