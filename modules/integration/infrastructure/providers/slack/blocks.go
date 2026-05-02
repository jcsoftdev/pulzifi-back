package slackprovider

import "github.com/jcsoftdev/pulzifi-back/modules/integration/domain/entities"

// buildBlocks converts a NotificationPayload into Slack Block Kit blocks.
func buildBlocks(p *entities.NotificationPayload) []map[string]any {
	out := []map[string]any{
		{
			"type": "header",
			"text": map[string]any{
				"type": "plain_text",
				"text": p.Title,
			},
		},
		{
			"type": "section",
			"text": map[string]any{
				"type": "mrkdwn",
				"text": p.Body,
			},
		},
	}

	if p.PageURL != "" {
		out = append(out, map[string]any{
			"type": "actions",
			"elements": []map[string]any{
				{
					"type": "button",
					"text": map[string]any{
						"type": "plain_text",
						"text": "View page",
					},
					"url": p.PageURL,
				},
			},
		})
	}

	return out
}
