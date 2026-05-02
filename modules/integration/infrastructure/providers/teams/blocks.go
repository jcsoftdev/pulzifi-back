package teamsprovider

import (
	"fmt"
	"html"

	"github.com/jcsoftdev/pulzifi-back/modules/integration/domain/entities"
)

func buildHTMLBody(p *entities.NotificationPayload) string {
	title := html.EscapeString(p.Title)
	body := html.EscapeString(p.Body)
	pageURL := html.EscapeString(p.PageURL)

	out := fmt.Sprintf("<h2>%s</h2><p>%s</p>", title, body)
	if pageURL != "" {
		out += fmt.Sprintf(`<p><a href="%s">View page</a></p>`, pageURL)
	}
	return out
}
