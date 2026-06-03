package html

import (
	"testing"
)

func TestExtractContentBlocks_BasicElements(t *testing.T) {
	htmlContent := `<html><body>
		<h1>Main Title</h1>
		<p>First paragraph with some text.</p>
		<h2>Subtitle</h2>
		<p>Second paragraph.</p>
		<a href="/page">Click here</a>
		<img src="/img.png" alt="An image">
		<ul>
			<li>Item 1</li>
			<li>Item 2</li>
		</ul>
	</body></html>`

	blocks := ExtractContentBlocks(htmlContent)

	expected := []struct {
		typ   BlockType
		text  string
		level int
		href  string
		src   string
		alt   string
	}{
		{BlockHeading, "Main Title", 1, "", "", ""},
		{BlockParagraph, "First paragraph with some text.", 0, "", "", ""},
		{BlockHeading, "Subtitle", 2, "", "", ""},
		{BlockParagraph, "Second paragraph.", 0, "", "", ""},
		{BlockLink, "Click here", 0, "/page", "", ""},
		{BlockImage, "", 0, "", "/img.png", "An image"},
		{BlockListItem, "Item 1", 1, "", "", ""},
		{BlockListItem, "Item 2", 1, "", "", ""},
	}

	if len(blocks) != len(expected) {
		t.Fatalf("expected %d blocks, got %d: %+v", len(expected), len(blocks), blocks)
	}

	for i, exp := range expected {
		b := blocks[i]
		if b.Type != exp.typ {
			t.Errorf("block %d: type = %q, want %q", i, b.Type, exp.typ)
		}
		if b.Text != exp.text {
			t.Errorf("block %d: text = %q, want %q", i, b.Text, exp.text)
		}
		if b.Level != exp.level {
			t.Errorf("block %d: level = %d, want %d", i, b.Level, exp.level)
		}
		if b.Href != exp.href {
			t.Errorf("block %d: href = %q, want %q", i, b.Href, exp.href)
		}
		if b.Src != exp.src {
			t.Errorf("block %d: src = %q, want %q", i, b.Src, exp.src)
		}
		if b.Alt != exp.alt {
			t.Errorf("block %d: alt = %q, want %q", i, b.Alt, exp.alt)
		}
	}
}

func TestExtractContentBlocks_SkipsNoiseElements(t *testing.T) {
	htmlContent := `<html>
		<head><title>Skip me</title></head>
		<body>
			<script>var x = 1;</script>
			<style>.foo { color: red; }</style>
			<noscript>Enable JS</noscript>
			<svg><text>SVG text</text></svg>
			<iframe src="/frame"></iframe>
			<p>Visible content</p>
		</body>
	</html>`

	blocks := ExtractContentBlocks(htmlContent)
	if len(blocks) != 1 {
		t.Fatalf("expected 1 block, got %d: %+v", len(blocks), blocks)
	}
	if blocks[0].Text != "Visible content" {
		t.Errorf("expected 'Visible content', got %q", blocks[0].Text)
	}
}

func TestExtractContentBlocks_TransparentContainers(t *testing.T) {
	htmlContent := `<html><body>
		<div>
			<section>
				<article>
					<h3>Nested heading</h3>
					<div><span><p>Deeply nested paragraph</p></span></div>
				</article>
			</section>
		</div>
	</body></html>`

	blocks := ExtractContentBlocks(htmlContent)
	if len(blocks) != 2 {
		t.Fatalf("expected 2 blocks, got %d: %+v", len(blocks), blocks)
	}
	if blocks[0].Type != BlockHeading || blocks[0].Level != 3 {
		t.Errorf("expected H3, got %s level %d", blocks[0].Type, blocks[0].Level)
	}
	if blocks[1].Type != BlockParagraph {
		t.Errorf("expected paragraph, got %s", blocks[1].Type)
	}
}

func TestExtractContentBlocks_NestedLists(t *testing.T) {
	htmlContent := `<html><body>
		<ul>
			<li>Top level</li>
			<li>
				<ul>
					<li>Nested item</li>
				</ul>
			</li>
		</ul>
	</body></html>`

	blocks := ExtractContentBlocks(htmlContent)
	if len(blocks) < 2 {
		t.Fatalf("expected at least 2 blocks, got %d: %+v", len(blocks), blocks)
	}
	if blocks[0].Level != 1 {
		t.Errorf("top-level list item: level = %d, want 1", blocks[0].Level)
	}
	// The second li contains a nested ul with an li, so its text includes "Nested item"
	// The nested li should have depth 2
	found := false
	for _, b := range blocks {
		if b.Text == "Nested item" && b.Level == 2 {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected nested list item at level 2, blocks: %+v", blocks)
	}
}

func TestExtractContentBlocks_TableCells(t *testing.T) {
	htmlContent := `<html><body>
		<table>
			<tr><th>Header</th><td>Cell 1</td></tr>
			<tr><td>Cell 2</td><td>Cell 3</td></tr>
		</table>
	</body></html>`

	blocks := ExtractContentBlocks(htmlContent)
	if len(blocks) != 4 {
		t.Fatalf("expected 4 blocks, got %d: %+v", len(blocks), blocks)
	}
	for _, b := range blocks {
		if b.Type != BlockTableCell {
			t.Errorf("expected table_cell, got %s", b.Type)
		}
	}
}

func TestExtractContentBlocks_WhitespaceNormalization(t *testing.T) {
	htmlContent := `<html><body>
		<p>  Multiple   spaces   and
			tabs	and
			newlines  </p>
	</body></html>`

	blocks := ExtractContentBlocks(htmlContent)
	if len(blocks) != 1 {
		t.Fatalf("expected 1 block, got %d", len(blocks))
	}
	if blocks[0].Text != "Multiple spaces and tabs and newlines" {
		t.Errorf("unexpected normalized text: %q", blocks[0].Text)
	}
}

func TestExtractContentBlocks_EmptyHTML(t *testing.T) {
	blocks := ExtractContentBlocks("")
	if len(blocks) != 0 {
		t.Errorf("expected 0 blocks for empty HTML, got %d", len(blocks))
	}
}

func TestExtractContentBlocks_CodeAndBlockquote(t *testing.T) {
	htmlContent := `<html><body>
		<pre>func main() { fmt.Println("hello") }</pre>
		<blockquote>A wise person once said something.</blockquote>
	</body></html>`

	blocks := ExtractContentBlocks(htmlContent)
	if len(blocks) != 2 {
		t.Fatalf("expected 2 blocks, got %d: %+v", len(blocks), blocks)
	}
	if blocks[0].Type != BlockCode {
		t.Errorf("expected code block, got %s", blocks[0].Type)
	}
	if blocks[1].Type != BlockBlockquote {
		t.Errorf("expected blockquote, got %s", blocks[1].Type)
	}
}

func TestHashContentBlocks_Deterministic(t *testing.T) {
	blocks := []ContentBlock{
		{Type: BlockHeading, Level: 1, Text: "Title"},
		{Type: BlockParagraph, Text: "Content"},
	}

	h1 := HashContentBlocks(blocks)
	h2 := HashContentBlocks(blocks)

	if h1 != h2 {
		t.Errorf("hash not deterministic: %s vs %s", h1, h2)
	}
	if len(h1) != 64 {
		t.Errorf("expected SHA-256 hex (64 chars), got %d chars", len(h1))
	}
}

func TestHashContentBlocks_DiffersOnChange(t *testing.T) {
	blocks1 := []ContentBlock{
		{Type: BlockHeading, Level: 1, Text: "Title"},
	}
	blocks2 := []ContentBlock{
		{Type: BlockHeading, Level: 1, Text: "Different Title"},
	}

	h1 := HashContentBlocks(blocks1)
	h2 := HashContentBlocks(blocks2)

	if h1 == h2 {
		t.Error("hashes should differ for different content")
	}
}

func TestExtractContentBlocks_NestedAndOutsideSections(t *testing.T) {
	htmlContent := `<html><body>
		<p>Outside paragraph</p>
		<section>
			<article>
				<p>Nested paragraph</p>
			</article>
		</section>
		<section>
			<p>Sibling paragraph</p>
		</section>
	</body></html>`

	blocks := ExtractContentBlocks(htmlContent)

	find := func(text string) *ContentBlock {
		for i := range blocks {
			if blocks[i].Text == text {
				return &blocks[i]
			}
		}
		return nil
	}

	outside := find("Outside paragraph")
	if outside == nil {
		t.Fatalf("block 'Outside paragraph' not found; all blocks: %+v", blocks)
	}
	if outside.SectionIndex != 0 {
		t.Errorf("outside block: SectionIndex = %d, want 0", outside.SectionIndex)
	}
	if outside.SectionTag != "" {
		t.Errorf("outside block: SectionTag = %q, want empty", outside.SectionTag)
	}

	nested := find("Nested paragraph")
	if nested == nil {
		t.Fatalf("block 'Nested paragraph' not found; all blocks: %+v", blocks)
	}
	if nested.SectionTag != "article" {
		t.Errorf("nested block: SectionTag = %q, want \"article\" (innermost section)", nested.SectionTag)
	}

	sibling := find("Sibling paragraph")
	if sibling == nil {
		t.Fatalf("block 'Sibling paragraph' not found; all blocks: %+v", blocks)
	}

	// The two top-level <section> elements must get distinct SectionIndex values.
	// nested.SectionIndex reflects the inner <article>; sibling.SectionIndex reflects
	// the second <section>. All three must be non-zero and mutually distinct.
	if nested.SectionIndex == 0 {
		t.Errorf("nested block: SectionIndex = 0, want > 0")
	}
	if sibling.SectionIndex == 0 {
		t.Errorf("sibling block: SectionIndex = 0, want > 0")
	}
	if nested.SectionIndex == sibling.SectionIndex {
		t.Errorf("nested and sibling blocks share SectionIndex %d; want distinct values",
			nested.SectionIndex)
	}
}

func TestHashContentBlocks_Empty(t *testing.T) {
	h := HashContentBlocks(nil)
	if h == "" {
		t.Error("expected non-empty hash for nil blocks")
	}
}

func TestExtractContentBlocks_SectionMetadata(t *testing.T) {
	htmlContent := `<html><body>
		<header><h1>Title</h1></header>
		<section class="stats" aria-label="Numbers"><p>+$900K</p><p>137.000</p></section>
		<footer><p>Contact us</p></footer>
	</body></html>`

	blocks := ExtractContentBlocks(htmlContent)

	var header, stats, footer *ContentBlock
	for i := range blocks {
		switch blocks[i].Text {
		case "Title":
			header = &blocks[i]
		case "+$900K":
			stats = &blocks[i]
		case "Contact us":
			footer = &blocks[i]
		}
	}
	if header == nil || stats == nil || footer == nil {
		t.Fatalf("missing blocks: header=%v stats=%v footer=%v", header, stats, footer)
	}
	if header.SectionTag != "header" {
		t.Errorf("header.SectionTag = %q, want header", header.SectionTag)
	}
	if stats.SectionTag != "section" || stats.SectionClass != "stats" || stats.SectionAria != "Numbers" {
		t.Errorf("stats section meta wrong: tag=%q class=%q aria=%q", stats.SectionTag, stats.SectionClass, stats.SectionAria)
	}
	if footer.SectionTag != "footer" {
		t.Errorf("footer.SectionTag = %q, want footer", footer.SectionTag)
	}
}

func TestHashContentBlocks_IgnoresSectionMetadata(t *testing.T) {
	a := []ContentBlock{{Type: BlockParagraph, Text: "x"}}
	b := []ContentBlock{{Type: BlockParagraph, Text: "x", SectionTag: "section", SectionClass: "card", SectionIndex: 3}}
	if HashContentBlocks(a) != HashContentBlocks(b) {
		t.Error("hash changed due to section metadata; must hash content only")
	}
}
