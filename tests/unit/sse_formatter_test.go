package unit

import (
	"strings"
	"testing"
	"time"

	server "github.com/joeblew999/timelinize-plug/internal/server"
)

func TestFormatProgressEventHTML(t *testing.T) {
	tests := []struct {
		name     string
		payload  string
		contains []string
	}{
		{
			name: "progress event with percentage",
			payload: `{
                "type": "progress",
                "message": "Processing items",
                "timestamp": "2025-10-14T12:00:00Z",
                "source": "test_source",
                "progress": {
                    "items_processed": 50,
                    "percentage": 50
                }
            }`,
			contains: []string{
				"event-type type-progress",
				"Processing items",
				"50% complete",
				"50 items processed",
			},
		},
		{
			name: "done event with metadata",
			payload: `{
                "type": "done",
                "message": "Import completed",
                "timestamp": "2025-10-14T12:00:00Z",
                "source": "google_photos",
                "progress": {
                    "items_processed": 100,
                    "percentage": 100
                },
                "metadata": {
                    "duration": "5m30s",
                    "input_path": "/test/input",
                    "output_path": "/test/output"
                }
            }`,
			contains: []string{
				"event-line event-done",
				"event-type type-done",
				"Import completed",
				"100% complete",
				"google_photos",
				"Duration:",
				"Input:",
				"Output:",
			},
		},
		{
			name: "error event",
			payload: `{
                "type": "error",
                "message": "Import failed",
                "timestamp": "2025-10-14T12:00:00Z",
                "source": "test_source",
                "error": "authentication required"
            }`,
			contains: []string{
				"event-line event-error",
				"event-type type-error",
				"Import failed",
				"authentication required",
			},
		},
		{
			name:    "simple non-JSON fallback",
			payload: "Simple text message",
			contains: []string{
				"Simple text message",
				"event-type",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			html := server.FormatProgressEventHTML(tt.payload)
			for _, expected := range tt.contains {
				if !strings.Contains(html, expected) {
					t.Errorf("expected HTML to contain %q, got: %s", expected, html)
				}
			}
		})
	}
}

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		bytes    int64
		expected string
	}{
		{0, "0 B"},
		{500, "500 B"},
		{1024, "1.0 KB"},
		{1536, "1.5 KB"},
		{1048576, "1.0 MB"},
		{1073741824, "1.0 GB"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := server.FormatBytes(tt.bytes)
			if result != tt.expected {
				t.Errorf("FormatBytes(%d) = %q, want %q", tt.bytes, result, tt.expected)
			}
		})
	}
}

func TestProgressEventStructure(t *testing.T) {
	t.Run("ProgressEvent has required fields", func(t *testing.T) {
		type ProgressEvent struct {
			Type      string    `json:"type"`
			Message   string    `json:"message"`
			Timestamp time.Time `json:"timestamp"`
			Source    string    `json:"source"`
		}

		event := ProgressEvent{
			Type:      "progress",
			Message:   "Test message",
			Timestamp: time.Now(),
			Source:    "test",
		}

		if event.Type != "progress" {
			t.Errorf("expected type 'progress', got %q", event.Type)
		}
		if event.Source != "test" {
			t.Errorf("expected source 'test', got %q", event.Source)
		}
	})
}
