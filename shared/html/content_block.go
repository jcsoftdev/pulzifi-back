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

	// Section metadata — describes the nearest structural container this block
	// belongs to. Used ONLY for grouping/naming in the diff view; deliberately
	// excluded from HashContentBlocks so it never triggers false content changes.
	SectionIndex   int    `json:"section_index,omitempty"`
	SectionTag     string `json:"section_tag,omitempty"`
	SectionRole    string `json:"section_role,omitempty"`
	SectionAria    string `json:"section_aria,omitempty"`
	SectionClass   string `json:"section_class,omitempty"`
	SectionHeading string `json:"section_heading,omitempty"`
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

// sectionBoundaryTags start a new logical section when entered during the walk.
var sectionBoundaryTags = map[string]bool{
	"section": true, "article": true, "header": true, "footer": true,
	"main": true, "nav": true, "aside": true,
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

	w := &blockWalker{}
	w.walk(doc)
	return w.blocks
}

// sectionMeta holds the container hints for one logical section.
type sectionMeta struct {
	index   int
	tag     string
	role    string
	aria    string
	class   string
	heading string
}

// blockWalker carries the mutable state for a single ExtractContentBlocks DOM walk.
type blockWalker struct {
	blocks       []ContentBlock
	listDepth    int
	sectionStack []sectionMeta
	sectionSeq   int
}

// curSection returns the current top-of-stack section metadata (zero value when
// the block is outside any structural container).
func (w *blockWalker) curSection() sectionMeta {
	if len(w.sectionStack) == 0 {
		return sectionMeta{}
	}
	return w.sectionStack[len(w.sectionStack)-1]
}

// tagBlock stamps the current section metadata onto a block before appending.
func (w *blockWalker) tagBlock(b ContentBlock) ContentBlock {
	s := w.curSection()
	b.SectionIndex = s.index
	b.SectionTag = s.tag
	b.SectionRole = s.role
	b.SectionAria = s.aria
	b.SectionClass = s.class
	b.SectionHeading = s.heading
	return b
}

// walk performs the recursive DOM traversal, dispatching each element to its
// emitter. Returns after emitting a block when the element type owns its subtree.
func (w *blockWalker) walk(n *html.Node) {
	if n.Type == html.ElementNode && contentSkipTags[n.Data] {
		return
	}

	// Enter a structural section container — push its metadata for descendants.
	pushed := false
	if n.Type == html.ElementNode && sectionBoundaryTags[n.Data] {
		w.sectionSeq++
		w.sectionStack = append(w.sectionStack, sectionMeta{
			index:   w.sectionSeq,
			tag:     n.Data,
			role:    getAttr(n, "role"),
			aria:    getAttr(n, "aria-label"),
			class:   getAttr(n, "class"),
			heading: firstHeadingText(n),
		})
		pushed = true
	}
	if pushed {
		defer func() { w.sectionStack = w.sectionStack[:len(w.sectionStack)-1] }()
	}

	if n.Type == html.ElementNode {
		if handled := w.emitElement(n); handled {
			return
		}
	}

	// Transparent containers (div, section, span, etc.) — if they contain only
	// inline content (no block-level children), emit as a paragraph block so
	// text in non-semantic markup isn't lost.
	if n.Type == html.ElementNode && isTransparentContainer(n.Data) && !hasBlockChild(n) {
		w.appendText(BlockParagraph, normalizeText(collectText(n)))
		return
	}

	// For all other nodes, recurse into children.
	w.walkChildren(n)
}

// firstHeadingText returns the text of the first heading descendant of n, or "".
func firstHeadingText(n *html.Node) string {
	var found string
	var rec func(*html.Node)
	rec = func(node *html.Node) {
		if found != "" {
			return
		}
		if node.Type == html.ElementNode && contentSkipTags[node.Data] {
			return
		}
		if node.Type == html.ElementNode && headingLevel(node.Data) > 0 {
			found = normalizeText(collectText(node))
			return
		}
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			rec(c)
		}
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		rec(c)
	}
	return found
}

// emitElement handles a single element node, returning true when the element
// type fully owns its subtree (so the caller must not recurse further).
func (w *blockWalker) emitElement(n *html.Node) bool {
	tag := n.Data

	// Headings
	if lvl := headingLevel(tag); lvl > 0 {
		if text := normalizeText(collectText(n)); text != "" {
			w.blocks = append(w.blocks, w.tagBlock(ContentBlock{Type: BlockHeading, Level: lvl, Text: text}))
		}
		return true // don't recurse into heading children (already collected)
	}

	switch tag {
	case "p":
		w.appendText(BlockParagraph, normalizeText(collectText(n)))
		return true
	case "a":
		w.emitLink(n)
		return true
	case "img":
		w.emitImage(n)
		return true
	case "li":
		w.emitListItem(n)
		return true
	case "ul", "ol":
		w.emitList(n)
		return true
	case "td", "th":
		w.appendText(BlockTableCell, normalizeText(collectText(n)))
		return true
	case "pre", "code":
		w.appendText(BlockCode, normalizeText(collectText(n)))
		return true
	case "blockquote":
		w.appendText(BlockBlockquote, normalizeText(collectText(n)))
		return true
	}

	return false
}

// emitLink emits a standalone link block (does not recurse into children).
func (w *blockWalker) emitLink(n *html.Node) {
	text := normalizeText(collectText(n))
	href := getAttr(n, "href")
	if text != "" || href != "" {
		w.blocks = append(w.blocks, w.tagBlock(ContentBlock{Type: BlockLink, Text: text, Href: href}))
	}
}

// emitImage emits an image block.
func (w *blockWalker) emitImage(n *html.Node) {
	src := getAttr(n, "src")
	alt := getAttr(n, "alt")
	if src != "" || alt != "" {
		w.blocks = append(w.blocks, w.tagBlock(ContentBlock{Type: BlockImage, Src: src, Alt: alt}))
	}
}

// emitListItem collects the item's own text (excluding nested lists) then
// recurses into any nested lists.
func (w *blockWalker) emitListItem(n *html.Node) {
	var ownText strings.Builder
	hasNestedList := false
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.ElementNode && (c.Data == "ul" || c.Data == "ol") {
			hasNestedList = true
		} else {
			ownText.WriteString(collectText(c))
		}
	}
	if text := normalizeText(ownText.String()); text != "" {
		w.blocks = append(w.blocks, w.tagBlock(ContentBlock{Type: BlockListItem, Level: w.listDepth, Text: text}))
	}
	if hasNestedList {
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if c.Type == html.ElementNode && (c.Data == "ul" || c.Data == "ol") {
				w.walk(c)
			}
		}
	}
}

// emitList walks a list's children while tracking nesting depth.
func (w *blockWalker) emitList(n *html.Node) {
	w.listDepth++
	w.walkChildren(n)
	w.listDepth--
}

// walkChildren recurses into all direct children of n.
func (w *blockWalker) walkChildren(n *html.Node) {
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		w.walk(c)
	}
}

// appendText appends a block of the given type when text is non-empty.
func (w *blockWalker) appendText(t BlockType, text string) {
	if text != "" {
		w.blocks = append(w.blocks, w.tagBlock(ContentBlock{Type: t, Text: text}))
	}
}

// HashContentBlocks computes a deterministic SHA-256 hash of a slice of content blocks.
func HashContentBlocks(blocks []ContentBlock) string {
	h := sha256.New()
	for _, b := range blocks {
		_, _ = fmt.Fprintf(h, "%s|%d|%s|%s|%s|%s\n", b.Type, b.Level, b.Text, b.Href, b.Src, b.Alt)
	}
	return hex.EncodeToString(h.Sum(nil))
}
