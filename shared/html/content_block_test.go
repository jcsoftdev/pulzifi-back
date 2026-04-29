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

func TestHashContentBlocks_Empty(t *testing.T) {
	h := HashContentBlocks(nil)
	if h == "" {
		t.Error("expected non-empty hash for nil blocks")
	}
}
