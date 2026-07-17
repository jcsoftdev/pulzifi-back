package providers

import (
	"context"
	"encoding/json"
	"os"
	"time"

	"github.com/jcsoftdev/pulzifi-back/shared/logger"
	"go.uber.org/zap"
)

// LogProvider implements EmailProvider by logging sends instead of calling a
// real email API. It never sends real emails — intended for local development
// and E2E tests so the notification pipeline can run without RESEND_API_KEY.
type LogProvider struct {
	captureFile string
}

// NewLogProvider creates a log-only email provider. If EMAIL_CAPTURE_FILE is
// set, each Send() additionally appends a JSON line to that file so tests can
// assert on delivered emails.
func NewLogProvider() *LogProvider {
	return &LogProvider{captureFile: os.Getenv("EMAIL_CAPTURE_FILE")}
}

// captureEntry is the JSON line format appended to the capture file.
type captureEntry struct {
	To      string `json:"to"`
	Subject string `json:"subject"`
	SentAt  string `json:"sent_at"`
}

// Send logs the email instead of delivering it, and optionally records it to
// EMAIL_CAPTURE_FILE. Capture-file errors are logged but never returned —
// this provider must never fail a caller relying on real email delivery
// semantics.
func (p *LogProvider) Send(_ context.Context, to, subject, _ string) error {
	logger.Info("Email send (log provider — no real email sent)",
		zap.String("to", to),
		zap.String("subject", subject),
	)

	if p.captureFile == "" {
		return nil
	}

	entry := captureEntry{To: to, Subject: subject, SentAt: time.Now().UTC().Format(time.RFC3339)}
	line, err := json.Marshal(entry)
	if err != nil {
		logger.Error("log email provider: failed to marshal capture entry", zap.Error(err))
		return nil
	}

	f, err := os.OpenFile(p.captureFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		logger.Error("log email provider: failed to open capture file",
			zap.String("path", p.captureFile), zap.Error(err))
		return nil
	}
	defer func() { _ = f.Close() }()

	if _, err := f.Write(append(line, '\n')); err != nil {
		logger.Error("log email provider: failed to write capture entry",
			zap.String("path", p.captureFile), zap.Error(err))
	}

	return nil
}
