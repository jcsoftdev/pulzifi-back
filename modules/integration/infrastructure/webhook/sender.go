package webhook

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/jcsoftdev/pulzifi-back/modules/integration/domain/entities"
)

// Sender dispatches webhook payloads to integration endpoints (Slack, Discord, Teams, etc.).
type Sender struct {
	client *http.Client
}

func NewSender() *Sender {
	return &Sender{
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

// Dispatch sends a JSON payload to the integration's configured URL.
func (s *Sender) Dispatch(ctx context.Context, integration *entities.Integration, pageURL, changeType string) error {
	urlVal, ok := integration.Config["url"].(string)
	if !ok || urlVal == "" {
		return fmt.Errorf("integration %s has no configured URL", integration.ID)
	}

	payload := map[string]string{
		"text": fmt.Sprintf("🔔 Pulzifi Alert: A *%s* change was detected on %s", changeType, pageURL),
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, urlVal, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("webhook returned status %d for integration %s", resp.StatusCode, integration.ID)
	}
	return nil
}
