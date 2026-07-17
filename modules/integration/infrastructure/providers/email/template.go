package emailprovider

import (
	"bytes"
	"html/template"
	"strings"
	"time"
	"unicode"

	"github.com/jcsoftdev/pulzifi-back/modules/integration/domain/entities"
)

// sanitizeText strips control characters (including CR/LF and tabs) from a
// string and trims surrounding whitespace. It leaves printable Unicode intact,
// so legitimate accented content survives. Used for both the email subject —
// where a stray CR/LF would enable header injection — and the rendered body
// fields, keeping the output free of stray/invisible special characters.
func sanitizeText(s string) string {
	cleaned := strings.Map(func(r rune) rune {
		if unicode.IsControl(r) {
			return -1
		}
		return r
	}, s)
	return strings.TrimSpace(cleaned)
}

// changeEmailTmpl is a self-contained, inline-styled HTML email. Email clients
// strip <style> blocks and external CSS, so every style is inline. Optional
// blocks (badge, page name, diff image) render only when their data is present.
var changeEmailTmpl = template.Must(template.New("change-email").Parse(`<div style="margin:0;padding:0;background:#f4f4f7;">
  <div style="max-width:480px;margin:0 auto;padding:24px 0;font-family:-apple-system,Segoe UI,Roboto,Helvetica,Arial,sans-serif;color:#1a1a2e;">
    <div style="padding:0 8px 16px;font-size:18px;font-weight:700;letter-spacing:-0.02em;color:#6d28d9;">Pulzifi</div>
    <div style="background:#ffffff;border-radius:12px;overflow:hidden;border:1px solid #ececf1;">
      <div style="padding:24px;">
        <h1 style="margin:0 0 8px;font-size:20px;line-height:1.3;font-weight:700;">{{.Title}}</h1>
        <div style="margin:0 0 16px;font-size:13px;color:#6b7280;">
          {{if .PageTitle}}<span>{{.PageTitle}}</span>{{end}}{{if .Badge}} <span style="color:#c4b5fd;">|</span> <span style="display:inline-block;padding:2px 8px;border-radius:999px;background:#f3e8ff;color:#6d28d9;font-weight:600;font-size:11px;text-transform:capitalize;">{{.Badge}}</span>{{end}}{{if .ChangedAt}} <span style="color:#c4b5fd;">|</span> <span>{{.ChangedAt}}</span>{{end}}
        </div>
        {{if .Summary}}<div style="margin:0 0 20px;padding:14px 16px;background:#faf9ff;border-left:3px solid #6d28d9;border-radius:6px;font-size:14px;line-height:1.5;color:#374151;"><div style="font-size:11px;text-transform:uppercase;letter-spacing:0.04em;color:#9ca3af;margin-bottom:4px;font-weight:600;">What changed</div>{{.Summary}}</div>{{end}}
        {{if .ImageURL}}{{if .CTAURL}}<a href="{{.CTAURL}}" style="display:block;margin:0 0 20px;"><img src="{{.ImageURL}}" alt="Visual diff preview" style="display:block;width:100%;max-width:100%;border-radius:8px;border:1px solid #ececf1;" /></a>{{else}}<img src="{{.ImageURL}}" alt="Visual diff preview" style="display:block;margin:0 0 20px;width:100%;max-width:100%;border-radius:8px;border:1px solid #ececf1;" />{{end}}{{end}}
        {{if .CTAURL}}<a href="{{.CTAURL}}" style="display:inline-block;background:#6d28d9;color:#ffffff;text-decoration:none;font-weight:600;font-size:14px;padding:11px 20px;border-radius:8px;">{{.CTALabel}}</a>{{end}}
      </div>
    </div>
    <div style="padding:16px 8px 0;font-size:11px;color:#9ca3af;line-height:1.5;">
      You are receiving this because you monitor this page on Pulzifi.
    </div>
  </div>
</div>`))

// emailView is the flattened, render-ready projection of a NotificationPayload.
type emailView struct {
	Title     string
	PageTitle string
	Badge     string
	ChangedAt string
	Summary   string
	ImageURL  string
	CTALabel  string
	CTAURL    string
}

// humanTime renders an RFC3339 timestamp as a compact, ASCII-only UTC label
// (e.g. "Jul 17, 2026 14:18 UTC"). Non-RFC3339 input is returned unchanged.
func humanTime(rfc string) string {
	if rfc == "" {
		return ""
	}
	t, err := time.Parse(time.RFC3339, rfc)
	if err != nil {
		return rfc
	}
	return t.UTC().Format("Jan 2, 2006 15:04 UTC")
}

// renderEmailHTML builds the branded HTML body for a notification. It works for
// both enriched change.detected payloads and plainer events (alert.created),
// showing only the fields that are present.
func renderEmailHTML(p *entities.NotificationPayload) string {
	v := emailView{
		Title:     sanitizeText(p.Title),
		PageTitle: sanitizeText(p.PageTitle),
		Badge:     sanitizeText(p.ChangeType),
		ChangedAt: humanTime(p.ChangedAt),
		Summary:   sanitizeText(p.Body),
		ImageURL:  p.DiffImageURL,
	}

	// Prefer the dashboard deep link; fall back to the raw monitored page URL.
	if p.DashboardURL != "" {
		v.CTAURL = p.DashboardURL
		v.CTALabel = "View changes"
	} else {
		v.CTAURL = p.PageURL
		v.CTALabel = "View page"
	}

	var buf bytes.Buffer
	if err := changeEmailTmpl.Execute(&buf, v); err != nil {
		// Template execution is deterministic; a failure means a programming
		// error, not runtime data. Degrade to a minimal body rather than fail
		// the whole delivery.
		return "<h2>" + template.HTMLEscapeString(p.Title) + "</h2><p>" +
			template.HTMLEscapeString(p.Body) + "</p>"
	}
	return buf.String()
}
