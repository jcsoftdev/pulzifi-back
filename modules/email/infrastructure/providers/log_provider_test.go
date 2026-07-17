package providers

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	emailservices "github.com/jcsoftdev/pulzifi-back/modules/email/domain/services"
	"github.com/jcsoftdev/pulzifi-back/shared/logger"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

// Compile-time assertion that LogProvider satisfies the shared EmailProvider port.
var _ emailservices.EmailProvider = (*LogProvider)(nil)

// TestLogProvider_Send_LogsAndSucceeds verifies that a send with no capture
// file configured still succeeds and logs the recipient/subject at info level.
func TestLogProvider_Send_LogsAndSucceeds(t *testing.T) {
	t.Setenv("EMAIL_CAPTURE_FILE", "")

	core, recorded := observer.New(zapcore.InfoLevel)
	restore := swapLogger(zap.New(core))
	defer restore()

	p := NewLogProvider()
	err := p.Send(context.Background(), "user@example.com", "Test Subject", "<p>hi</p>")
	if err != nil {
		t.Fatalf("Send() error = %v, want nil", err)
	}

	found := false
	for _, entry := range recorded.All() {
		fields := entry.ContextMap()
		if fields["to"] == "user@example.com" && fields["subject"] == "Test Subject" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected an info log entry with to=%q subject=%q, got entries=%+v",
			"user@example.com", "Test Subject", recorded.All())
	}
}

// TestLogProvider_Send_WritesCaptureFile verifies that when EMAIL_CAPTURE_FILE
// is set, Send() appends a single JSON line with to/subject/sent_at fields.
func TestLogProvider_Send_WritesCaptureFile(t *testing.T) {
	capturePath := filepath.Join(t.TempDir(), "capture.jsonl")
	t.Setenv("EMAIL_CAPTURE_FILE", capturePath)

	p := NewLogProvider()
	if err := p.Send(context.Background(), "to@example.com", "Hello", "body"); err != nil {
		t.Fatalf("Send() error = %v, want nil", err)
	}

	data, err := os.ReadFile(capturePath)
	if err != nil {
		t.Fatalf("failed to read capture file: %v", err)
	}

	line := strings.TrimSpace(string(data))
	var entry struct {
		To      string `json:"to"`
		Subject string `json:"subject"`
		SentAt  string `json:"sent_at"`
	}
	if err := json.Unmarshal([]byte(line), &entry); err != nil {
		t.Fatalf("failed to parse capture line %q: %v", line, err)
	}
	if entry.To != "to@example.com" {
		t.Errorf("captured to = %q, want %q", entry.To, "to@example.com")
	}
	if entry.Subject != "Hello" {
		t.Errorf("captured subject = %q, want %q", entry.Subject, "Hello")
	}
	if entry.SentAt == "" {
		t.Error("captured sent_at is empty, want a timestamp")
	}
}

// TestLogProvider_Send_NoCaptureFileWhenEnvUnset verifies Send() succeeds and
// never attempts to write a capture file when EMAIL_CAPTURE_FILE is unset.
func TestLogProvider_Send_NoCaptureFileWhenEnvUnset(t *testing.T) {
	t.Setenv("EMAIL_CAPTURE_FILE", "")

	p := NewLogProvider()
	if p.captureFile != "" {
		t.Fatalf("captureFile = %q, want empty when EMAIL_CAPTURE_FILE is unset", p.captureFile)
	}
	if err := p.Send(context.Background(), "a@b.com", "s", "b"); err != nil {
		t.Fatalf("Send() error = %v, want nil", err)
	}
}

// TestLogProvider_Send_CaptureFileWriteFailureIsSwallowed verifies that a
// capture file write failure (e.g. parent directory does not exist) never
// bubbles up as an error from Send() — file capture is best-effort.
func TestLogProvider_Send_CaptureFileWriteFailureIsSwallowed(t *testing.T) {
	badPath := filepath.Join(t.TempDir(), "does-not-exist", "capture.jsonl")
	t.Setenv("EMAIL_CAPTURE_FILE", badPath)

	p := NewLogProvider()
	if err := p.Send(context.Background(), "a@b.com", "s", "b"); err != nil {
		t.Fatalf("Send() error = %v, want nil even when capture file write fails", err)
	}
}

// swapLogger temporarily replaces the package-level logger.Logger and returns
// a restore func. Needed because logger.Info/Error write through the global.
func swapLogger(l *zap.Logger) (restore func()) {
	original := logger.Logger
	logger.Logger = l
	return func() { logger.Logger = original }
}
