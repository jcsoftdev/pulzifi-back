# HTML Package (`shared/html/`)

HTML text extraction via DOM tree walking.

## Files

- `extractor.go` — HTML-to-text converter

## Exported API

### Functions
- `ExtractText(htmlContent string) string` — Parses HTML and extracts readable plain text

## Behavior

- Skips `script`, `style`, `noscript`, `head` elements entirely
- Inserts newlines at block-level elements (`p`, `h1`-`h6`, `li`, `td`, `th`, `blockquote`, `pre`, `div`, `article`, `section`, `header`, `footer`, `nav`, `main`, `aside`, `figcaption`, `tr`)
- Normalizes whitespace within lines
- On parse error, returns raw HTML content as fallback

## Usage

Used by the `insight` module to extract readable text from page HTML snapshots before sending to the LLM for analysis.

## Dependencies

- `golang.org/x/net/html`
