package server

import (
	"encoding/json"
	"fmt"
	"html/template"
	"time"
)

// ProgressEvent matches the structure from runner package
type ProgressEvent struct {
	Type      string    `json:"type"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
	Source    string    `json:"source"`
	Progress  *Progress `json:"progress,omitempty"`
	Error     string    `json:"error,omitempty"`
	Metadata  Metadata  `json:"metadata,omitempty"`
}

type Progress struct {
	ItemsProcessed int    `json:"items_processed"`
	ItemsTotal     int    `json:"items_total,omitempty"`
	Percentage     int    `json:"percentage,omitempty"`
	CurrentItem    string `json:"current_item,omitempty"`
}

type Metadata struct {
	StartTime    time.Time `json:"start_time,omitempty"`
	EndTime      time.Time `json:"end_time,omitempty"`
	Duration     string    `json:"duration,omitempty"`
	InputPath    string    `json:"input_path,omitempty"`
	OutputPath   string    `json:"output_path,omitempty"`
	BytesRead    int64     `json:"bytes_read,omitempty"`
	FilesCreated int       `json:"files_created,omitempty"`
}

// formatProgressEventHTML takes a JSON payload and returns formatted HTML for the Datastar UI
func formatProgressEventHTML(payload string) string {
	var event ProgressEvent
	if err := json.Unmarshal([]byte(payload), &event); err != nil {
		// Fallback to simple formatting if JSON parsing fails
		return formatSimpleEvent(payload)
	}

	// Determine event class and type class
	eventClass := ""
	typeClass := "event-type"
	switch event.Type {
	case "error":
		eventClass = " event-error"
		typeClass += " type-error"
	case "done":
		eventClass = " event-done"
		typeClass += " type-done"
	case "progress":
		typeClass += " type-progress"
	}

	ts := event.Timestamp.Format("15:04:05")
	if event.Timestamp.IsZero() {
		ts = time.Now().Format("15:04:05")
	}

	html := fmt.Sprintf(`<li class="event-line%s">`, eventClass)

	// Event header with type and timestamp
	html += `<div class="event-header">`
	html += fmt.Sprintf(`<span class="%s">%s</span>`, typeClass, template.HTMLEscapeString(event.Type))
	html += fmt.Sprintf(`<span class="event-ts">%s</span>`, ts)
	html += `</div>`

	// Event message
	if event.Message != "" {
		html += fmt.Sprintf(`<div class="event-message">%s</div>`, template.HTMLEscapeString(event.Message))
	}

	// Progress information
	if event.Progress != nil {
		html += `<div class="event-progress">`

		if event.Progress.Percentage > 0 {
			html += fmt.Sprintf(`<div class="progress-bar"><div class="progress-bar-fill" style="width: %d%%"></div></div>`, event.Progress.Percentage)
			html += fmt.Sprintf(`<span>%d%% complete • %d items processed</span>`,
				event.Progress.Percentage, event.Progress.ItemsProcessed)
		} else {
			html += fmt.Sprintf(`<span>%d items processed`, event.Progress.ItemsProcessed)
			if event.Progress.ItemsTotal > 0 {
				html += fmt.Sprintf(` of %d`, event.Progress.ItemsTotal)
			}
			html += `</span>`
		}

		if event.Progress.CurrentItem != "" {
			html += fmt.Sprintf(`<span>Current: %s</span>`, template.HTMLEscapeString(event.Progress.CurrentItem))
		}

		html += `</div>`
	}

	// Error details
	if event.Error != "" {
		html += fmt.Sprintf(`<div class="event-error-detail">%s</div>`, template.HTMLEscapeString(event.Error))
	}

	// Metadata
	if hasMetadata(event.Metadata) {
		html += `<dl class="event-metadata">`

		if event.Source != "" {
			html += `<dt>Source:</dt>`
			html += fmt.Sprintf(`<dd>%s</dd>`, template.HTMLEscapeString(event.Source))
		}

		if event.Metadata.Duration != "" {
			html += `<dt>Duration:</dt>`
			html += fmt.Sprintf(`<dd>%s</dd>`, template.HTMLEscapeString(event.Metadata.Duration))
		}

		if event.Metadata.InputPath != "" {
			html += `<dt>Input:</dt>`
			html += fmt.Sprintf(`<dd>%s</dd>`, template.HTMLEscapeString(event.Metadata.InputPath))
		}

		if event.Metadata.OutputPath != "" {
			html += `<dt>Output:</dt>`
			html += fmt.Sprintf(`<dd>%s</dd>`, template.HTMLEscapeString(event.Metadata.OutputPath))
		}

		if event.Metadata.FilesCreated > 0 {
			html += `<dt>Files:</dt>`
			html += fmt.Sprintf(`<dd>%d created</dd>`, event.Metadata.FilesCreated)
		}

		if event.Metadata.BytesRead > 0 {
			html += `<dt>Data:</dt>`
			html += fmt.Sprintf(`<dd>%s</dd>`, formatBytes(event.Metadata.BytesRead))
		}

		html += `</dl>`
	}

	html += `</li>`
	return html
}

// formatSimpleEvent provides a fallback for non-JSON payloads
func formatSimpleEvent(payload string) string {
	ts := time.Now().Format("15:04:05")
	return fmt.Sprintf(
		`<li class="event-line"><div class="event-header"><span class="event-type">event</span><span class="event-ts">%s</span></div><div class="event-message"><code>%s</code></div></li>`,
		ts,
		template.HTMLEscapeString(payload),
	)
}

// hasMetadata checks if metadata contains any non-zero values
func hasMetadata(m Metadata) bool {
	return m.Duration != "" || m.InputPath != "" || m.OutputPath != "" ||
		m.FilesCreated > 0 || m.BytesRead > 0
}

// formatBytes formats byte counts in human-readable format
func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// FormatProgressEventHTML is an exported helper for rendering progress events.
// Primarily used by tests and Datastar templates.
func FormatProgressEventHTML(payload string) string {
    return formatProgressEventHTML(payload)
}

// FormatBytes exposes the byte formatter for consumers and tests.
func FormatBytes(bytes int64) string {
    return formatBytes(bytes)
}
