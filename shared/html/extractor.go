package html

import (
	"strings"

	"golang.org/x/net/html"
)

var blockTags = map[string]bool{
	"p": true, "h1": true, "h2": true, "h3": true, "h4": true, "h5": true, "h6": true,
	"li": true, "td": true, "th": true, "blockquote": true, "pre": true,
	"div": true, "article": true, "section": true, "header": true,
	"footer": true, "nav": true, "main": true, "aside": true, "figcaption": true, "tr": true,
}

var skipTags = map[string]bool{
	"script": true, "style": true, "noscript": true, "head": true,
}

// ExtractText extracts plain text from an HTML string.
// Block-level elements produce paragraph breaks; script/style nodes are skipped.
func ExtractText(htmlContent string) string {
	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		return htmlContent
	}

	var sb strings.Builder
	walkNode(&sb, doc)

	lines := strings.Split(sb.String(), "\n")
	result := make([]string, 0, len(lines))
	for _, line := range lines {
		normalized := strings.Join(strings.Fields(line), " ")
		if normalized != "" {
			result = append(result, normalized)
		}
	}
	return strings.Join(result, "\n")
}

func walkNode(sb *strings.Builder, n *html.Node) {
	if n.Type == html.TextNode {
		sb.WriteString(n.Data)
		return
	}

	if n.Type == html.ElementNode {
		if skipTags[n.Data] {
			return
		}
		if blockTags[n.Data] {
			sb.WriteString("\n")
		}
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		walkNode(sb, c)
	}

	if n.Type == html.ElementNode && blockTags[n.Data] {
		sb.WriteString("\n")
	}
}
