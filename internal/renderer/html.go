package renderer

import (
	"fmt"
	"html"
	"strings"
	"time"

	"ganttgen/internal/calendar"
)

func renderHTML(ctx renderContext) string {
	var b strings.Builder
	b.WriteString("<!DOCTYPE html>\n<html lang=\"ja\">\n<head>\n<meta charset=\"UTF-8\">\n")
	b.WriteString("<meta name=\"viewport\" content=\"width=device-width, initial-scale=1\">\n")
	b.WriteString("<title>Gantt Chart</title>\n<style>\n")
	b.WriteString(baseCSS())
	b.WriteString("\n</style>\n</head>\n<body>\n")
	b.WriteString("<div class=\"page\">\n<h1>Gantt Chart</h1>\n")
	fmt.Fprintf(&b, "<div class=\"gantt\" style=\"--day-count:%d;--today-index:%d;\">\n", ctx.DayCount, ctx.TodayIndex)

	b.WriteString("<div class=\"name-list\">")
	b.WriteString("<div class=\"name header\">Task</div>")
	for _, task := range ctx.Tasks {
		b.WriteString("<div class=\"name\">")
		b.WriteString(html.EscapeString(task.Name))
		b.WriteString("</div>")
	}
	b.WriteString("</div>") // name-list

	b.WriteString("<div class=\"timeline-wrapper\">")
	b.WriteString("<div class=\"grid-surface\">")
	b.WriteString("<div class=\"today-line\"></div>")

	b.WriteString("<div class=\"timeline-grid grid\">")
	for _, day := range ctx.Days {
		class := "day"
		if !calendar.IsWorkday(day) {
			class += " weekend"
		}
		fmt.Fprintf(&b, "<div class=\"%s\">%s</div>", class, day.Format("2006-01-02"))
	}
	b.WriteString("</div>") // timeline-grid

	b.WriteString("<div class=\"bars\">")
	for _, task := range ctx.Tasks {
		b.WriteString("<div class=\"bar-row grid\">")
		fmt.Fprintf(&b, "<div class=\"bar\" style=\"grid-column:%d / span %d;\" title=\"%s - %s\">%s</div>",
			task.StartIndex+1, task.Span, formatDate(task.Start), formatDate(task.End), html.EscapeString(task.Name))
		b.WriteString("</div>")
	}
	b.WriteString("</div>") // bars

	b.WriteString("</div>") // grid-surface
	b.WriteString("</div>") // timeline-wrapper

	b.WriteString("</div>") // gantt
	b.WriteString("</div>") // page
	b.WriteString("\n</body>\n</html>")
	return b.String()
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
  gap: 8px;
}

.name {
  min-width: 180px;
  background: #fff;
  border: 1px solid var(--line);
  border-radius: 8px;
  padding: 10px 12px;
  font-weight: 500;
  box-shadow: 0 6px 18px rgba(15, 23, 42, 0.05);
}

.name.header {
  background: linear-gradient(120deg, #fff, #f0f4ff);
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
}

.timeline-grid {
  gap: 2px;
  padding-bottom: 12px;
  border-bottom: 1px solid var(--line);
  position: relative;
  z-index: 1;
}

.day {
  font-size: 11px;
  text-align: left;
  line-height: 1.2;
  transform: rotate(-60deg);
  transform-origin: left bottom;
  height: 56px;
  color: #1f2937;
}

.day.weekend {
  color: #9ca3af;
}

.bars {
  display: flex;
  flex-direction: column;
  gap: 8px;
  padding-top: 8px;
}

.bar-row {
  position: relative;
  align-items: center;
  min-height: 32px;
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
