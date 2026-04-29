package html

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"

	"golang.org/x/net/html"
)

// BlockType identifies the semantic type of a content block.
type BlockType string

const (
	BlockHeading    BlockType = "heading"
	BlockParagraph  BlockType = "paragraph"
	BlockLink       BlockType = "link"
	BlockImage      BlockType = "image"
	BlockListItem   BlockType = "list_item"
	BlockTableCell  BlockType = "table_cell"
	BlockCode       BlockType = "code"
	BlockBlockquote BlockType = "blockquote"
)

// ContentBlock represents a single piece of semantic content extracted from HTML.
type ContentBlock struct {
	Type  BlockType `json:"type"`
	Level int       `json:"level,omitempty"`
	Text  string    `json:"text"`
	Href  string    `json:"href,omitempty"`
	Src   string    `json:"src,omitempty"`
	Alt   string    `json:"alt,omitempty"`
}

// contentSkipTags are elements whose subtrees should be skipped entirely.
var contentSkipTags = map[string]bool{
	"script": true, "style": true, "noscript": true, "head": true,
	"svg": true, "iframe": true,
}

// headingLevel returns the heading level (1-6) or 0 if not a heading.
func headingLevel(tag string) int {
	switch tag {
	case "h1":
		return 1
	case "h2":
		return 2
	case "h3":
		return 3
	case "h4":
		return 4
	case "h5":
		return 5
	case "h6":
		return 6
	}
	return 0
}

// getAttr returns the value of the named attribute, or "".
func getAttr(n *html.Node, name string) string {
	for _, a := range n.Attr {
		if a.Key == name {
			return a.Val
		}
	}
	return ""
}

// collectText recursively collects visible text from a node subtree.
func collectText(n *html.Node) string {
	if n.Type == html.TextNode {
		return n.Data
	}
	if n.Type == html.ElementNode && contentSkipTags[n.Data] {
		return ""
	}
	var sb strings.Builder
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		sb.WriteString(collectText(c))
	}
	return sb.String()
}

// normalizeText collapses whitespace and trims.
func normalizeText(s string) string {
	return strings.Join(strings.Fields(s), " ")
}

// transparentContainers are elements that don't carry semantic meaning but may
// wrap text content directly (e.g. <div class="text">Hello</div>).
var transparentContainers = map[string]bool{
	"div": true, "section": true, "article": true, "main": true,
	"aside": true, "header": true, "footer": true, "nav": true,
	"span": true, "figure": true, "figcaption": true,
}

// isTransparentContainer reports whether the tag is a transparent container.
func isTransparentContainer(tag string) bool {
	return transparentContainers[tag]
}

// blockElements are elements that form their own content blocks or contain block-level content.
var blockElements = map[string]bool{
	"p": true, "h1": true, "h2": true, "h3": true, "h4": true, "h5": true, "h6": true,
	"ul": true, "ol": true, "li": true, "table": true, "tr": true, "td": true, "th": true,
	"pre": true, "blockquote": true, "hr": true, "form": true,
	"div": true, "section": true, "article": true, "main": true,
	"aside": true, "header": true, "footer": true, "nav": true,
	"figure": true, "figcaption": true,
}

// hasBlockChild reports whether node n has any direct child that is a block element.
func hasBlockChild(n *html.Node) bool {
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.ElementNode && blockElements[c.Data] {
			return true
		}
	}
	return false
}

// ExtractContentBlocks parses HTML and extracts a flat list of semantic content blocks.
// It performs a single-pass DOM walk, emitting blocks for content-bearing elements
// (headings, paragraphs, links, images, list items, table cells, code, blockquotes)
// while skipping noise (script, style, svg, iframe, etc.). Transparent container
// elements (div, section, article, span, etc.) are walked through without emitting blocks.
func ExtractContentBlocks(htmlContent string) []ContentBlock {
	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		return nil
	}

	var blocks []ContentBlock
	var listDepth int

	var walk func(n *html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode && contentSkipTags[n.Data] {
			return
		}

		if n.Type == html.ElementNode {
			tag := n.Data

			// Headings
			if lvl := headingLevel(tag); lvl > 0 {
				text := normalizeText(collectText(n))
				if text != "" {
					blocks = append(blocks, ContentBlock{Type: BlockHeading, Level: lvl, Text: text})
				}
				return // don't recurse into heading children (already collected)
			}

			// Paragraphs
			if tag == "p" {
				text := normalizeText(collectText(n))
				if text != "" {
					blocks = append(blocks, ContentBlock{Type: BlockParagraph, Text: text})
				}
				return
			}

			// Links (standalone — emit block, don't recurse)
			if tag == "a" {
				text := normalizeText(collectText(n))
				href := getAttr(n, "href")
				if text != "" || href != "" {
					blocks = append(blocks, ContentBlock{Type: BlockLink, Text: text, Href: href})
				}
				return
			}

			// Images
			if tag == "img" {
				src := getAttr(n, "src")
				alt := getAttr(n, "alt")
				if src != "" || alt != "" {
					blocks = append(blocks, ContentBlock{Type: BlockImage, Src: src, Alt: alt})
				}
				return
			}

			// List items — collect own text (excluding nested lists), then recurse for nested lists.
			if tag == "li" {
				var ownText strings.Builder
				hasNestedList := false
				for c := n.FirstChild; c != nil; c = c.NextSibling {
					if c.Type == html.ElementNode && (c.Data == "ul" || c.Data == "ol") {
						hasNestedList = true
					} else {
						ownText.WriteString(collectText(c))
					}
				}
				text := normalizeText(ownText.String())
				if text != "" {
					blocks = append(blocks, ContentBlock{Type: BlockListItem, Level: listDepth, Text: text})
				}
				if hasNestedList {
					for c := n.FirstChild; c != nil; c = c.NextSibling {
						if c.Type == html.ElementNode && (c.Data == "ul" || c.Data == "ol") {
							walk(c)
						}
					}
				}
				return
			}

			// Lists — track depth
			if tag == "ul" || tag == "ol" {
				listDepth++
				for c := n.FirstChild; c != nil; c = c.NextSibling {
					walk(c)
				}
				listDepth--
				return
			}

			// Table cells
			if tag == "td" || tag == "th" {
				text := normalizeText(collectText(n))
				if text != "" {
					blocks = append(blocks, ContentBlock{Type: BlockTableCell, Text: text})
				}
				return
			}

			// Code blocks
			if tag == "pre" || tag == "code" {
				text := normalizeText(collectText(n))
				if text != "" {
					blocks = append(blocks, ContentBlock{Type: BlockCode, Text: text})
				}
				return
			}

			// Blockquotes
			if tag == "blockquote" {
				text := normalizeText(collectText(n))
				if text != "" {
					blocks = append(blocks, ContentBlock{Type: BlockBlockquote, Text: text})
				}
				return
			}
		}

		// Transparent containers (div, section, span, etc.) — if they contain only
		// inline content (no block-level children), emit as a paragraph block so
		// text in non-semantic markup isn't lost.
		if n.Type == html.ElementNode && isTransparentContainer(n.Data) && !hasBlockChild(n) {
			text := normalizeText(collectText(n))
			if text != "" {
				blocks = append(blocks, ContentBlock{Type: BlockParagraph, Text: text})
			}
			return
		}

		// For all other nodes, recurse into children.
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}

	walk(doc)
	return blocks
}

// HashContentBlocks computes a deterministic SHA-256 hash of a slice of content blocks.
func HashContentBlocks(blocks []ContentBlock) string {
	h := sha256.New()
	for _, b := range blocks {
		fmt.Fprintf(h, "%s|%d|%s|%s|%s|%s\n", b.Type, b.Level, b.Text, b.Href, b.Src, b.Alt)
	}
	return hex.EncodeToString(h.Sum(nil))
}
