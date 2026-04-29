package html

import (
	"strings"
	"testing"
)

func TestDiffContentBlocks_NoChanges(t *testing.T) {
	blocks := []ContentBlock{
		{Type: BlockHeading, Level: 1, Text: "Title"},
		{Type: BlockParagraph, Text: "Content"},
	}

	diff := DiffContentBlocks(blocks, blocks)
	if diff.HasChanges {
		t.Error("expected no changes for identical blocks")
	}
	if diff.TotalChanges != 0 {
		t.Errorf("expected 0 total changes, got %d", diff.TotalChanges)
	}
}

func TestDiffContentBlocks_AddedBlocks(t *testing.T) {
	prev := []ContentBlock{
		{Type: BlockHeading, Level: 1, Text: "Title"},
	}
	curr := []ContentBlock{
		{Type: BlockHeading, Level: 1, Text: "Title"},
		{Type: BlockParagraph, Text: "New paragraph"},
	}

	diff := DiffContentBlocks(prev, curr)
	if !diff.HasChanges {
		t.Error("expected changes")
	}
	if diff.TotalChanges != 1 {
		t.Errorf("expected 1 change, got %d", diff.TotalChanges)
	}
	if diff.Diffs[0].Op != DiffAdded {
		t.Errorf("expected added, got %s", diff.Diffs[0].Op)
	}
	if diff.Diffs[0].Block.Text != "New paragraph" {
		t.Errorf("expected 'New paragraph', got %q", diff.Diffs[0].Block.Text)
	}
}

func TestDiffContentBlocks_RemovedBlocks(t *testing.T) {
	prev := []ContentBlock{
		{Type: BlockHeading, Level: 1, Text: "Title"},
		{Type: BlockParagraph, Text: "Old paragraph"},
	}
	curr := []ContentBlock{
		{Type: BlockHeading, Level: 1, Text: "Title"},
	}

	diff := DiffContentBlocks(prev, curr)
	if !diff.HasChanges {
		t.Error("expected changes")
	}
	if diff.TotalChanges != 1 {
		t.Errorf("expected 1 change, got %d", diff.TotalChanges)
	}
	if diff.Diffs[0].Op != DiffRemoved {
		t.Errorf("expected removed, got %s", diff.Diffs[0].Op)
	}
}

func TestDiffContentBlocks_ChangedBlocks(t *testing.T) {
	prev := []ContentBlock{
		{Type: BlockHeading, Level: 1, Text: "Old Title"},
		{Type: BlockParagraph, Text: "Same paragraph"},
	}
	curr := []ContentBlock{
		{Type: BlockHeading, Level: 1, Text: "New Title"},
		{Type: BlockParagraph, Text: "Same paragraph"},
	}

	diff := DiffContentBlocks(prev, curr)
	if !diff.HasChanges {
		t.Error("expected changes")
	}
	if diff.TotalChanges != 1 {
		t.Errorf("expected 1 change, got %d", diff.TotalChanges)
	}
	if diff.Diffs[0].Op != DiffChanged {
		t.Errorf("expected changed, got %s", diff.Diffs[0].Op)
	}
	if diff.Diffs[0].OldBlock == nil {
		t.Fatal("expected old block to be set")
	}
	if diff.Diffs[0].OldBlock.Text != "Old Title" {
		t.Errorf("expected old text 'Old Title', got %q", diff.Diffs[0].OldBlock.Text)
	}
	if diff.Diffs[0].Block.Text != "New Title" {
		t.Errorf("expected new text 'New Title', got %q", diff.Diffs[0].Block.Text)
	}
}

func TestDiffContentBlocks_MixedChanges(t *testing.T) {
	prev := []ContentBlock{
		{Type: BlockHeading, Level: 1, Text: "Title"},
		{Type: BlockParagraph, Text: "Para 1"},
		{Type: BlockParagraph, Text: "Para 2"},
		{Type: BlockLink, Text: "Old link", Href: "/old"},
	}
	curr := []ContentBlock{
		{Type: BlockHeading, Level: 1, Text: "Title"},
		{Type: BlockParagraph, Text: "Para 1 modified"},
		{Type: BlockParagraph, Text: "Para 2"},
		{Type: BlockHeading, Level: 2, Text: "New Section"},
	}

	diff := DiffContentBlocks(prev, curr)
	if !diff.HasChanges {
		t.Error("expected changes")
	}
	// Para 1 changed, link removed, new H2 added
	if diff.TotalChanges != 3 {
		t.Errorf("expected 3 changes, got %d: %+v", diff.TotalChanges, diff.Diffs)
	}
}

func TestDiffContentBlocks_EmptyPrev(t *testing.T) {
	curr := []ContentBlock{
		{Type: BlockHeading, Level: 1, Text: "Title"},
	}

	diff := DiffContentBlocks(nil, curr)
	if !diff.HasChanges {
		t.Error("expected changes")
	}
	if diff.TotalChanges != 1 {
		t.Errorf("expected 1 change, got %d", diff.TotalChanges)
	}
	if diff.Diffs[0].Op != DiffAdded {
		t.Errorf("expected added, got %s", diff.Diffs[0].Op)
	}
}

func TestDiffContentBlocks_EmptyCurr(t *testing.T) {
	prev := []ContentBlock{
		{Type: BlockHeading, Level: 1, Text: "Title"},
	}

	diff := DiffContentBlocks(prev, nil)
	if !diff.HasChanges {
		t.Error("expected changes")
	}
	if diff.Diffs[0].Op != DiffRemoved {
		t.Errorf("expected removed, got %s", diff.Diffs[0].Op)
	}
}

func TestDiffContentBlocks_BothEmpty(t *testing.T) {
	diff := DiffContentBlocks(nil, nil)
	if diff.HasChanges {
		t.Error("expected no changes for two empty slices")
	}
}

func TestFormatDiffForAI_Added(t *testing.T) {
	diff := &ContentDiff{
		HasChanges:   true,
		TotalChanges: 1,
		Diffs: []BlockDiff{
			{Op: DiffAdded, Block: ContentBlock{Type: BlockHeading, Level: 2, Text: "New Section"}},
		},
	}

	result := FormatDiffForAI(diff)
	if !strings.Contains(result, "[ADDED]") {
		t.Error("expected [ADDED] prefix")
	}
	if !strings.Contains(result, "H2") {
		t.Error("expected H2 label")
	}
	if !strings.Contains(result, "New Section") {
		t.Error("expected text content")
	}
}

func TestFormatDiffForAI_Changed(t *testing.T) {
	old := ContentBlock{Type: BlockParagraph, Text: "old text"}
	diff := &ContentDiff{
		HasChanges:   true,
		TotalChanges: 1,
		Diffs: []BlockDiff{
			{Op: DiffChanged, Block: ContentBlock{Type: BlockParagraph, Text: "new text"}, OldBlock: &old},
		},
	}

	result := FormatDiffForAI(diff)
	if !strings.Contains(result, "[CHANGED]") {
		t.Error("expected [CHANGED] prefix")
	}
	if !strings.Contains(result, "old text") {
		t.Error("expected old text")
	}
	if !strings.Contains(result, "new text") {
		t.Error("expected new text")
	}
	if !strings.Contains(result, "→") {
		t.Error("expected arrow between old and new")
	}
}

func TestFormatDiffForAI_Removed(t *testing.T) {
	diff := &ContentDiff{
		HasChanges:   true,
		TotalChanges: 1,
		Diffs: []BlockDiff{
			{Op: DiffRemoved, Block: ContentBlock{Type: BlockLink, Text: "old link", Href: "/old"}},
		},
	}

	result := FormatDiffForAI(diff)
	if !strings.Contains(result, "[REMOVED]") {
		t.Error("expected [REMOVED] prefix")
	}
	if !strings.Contains(result, "(href: /old)") {
		t.Error("expected href info")
	}
}

func TestFormatDiffForAI_NilDiff(t *testing.T) {
	result := FormatDiffForAI(nil)
	if result != "" {
		t.Errorf("expected empty string for nil diff, got %q", result)
	}
}

func TestFormatDiffForAI_NoChanges(t *testing.T) {
	diff := &ContentDiff{HasChanges: false}
	result := FormatDiffForAI(diff)
	if result != "" {
		t.Errorf("expected empty string for no changes, got %q", result)
	}
}

func TestFormatDiffForAI_Truncation(t *testing.T) {
	longText := strings.Repeat("a", 300)
	diff := &ContentDiff{
		HasChanges:   true,
		TotalChanges: 1,
		Diffs: []BlockDiff{
			{Op: DiffAdded, Block: ContentBlock{Type: BlockParagraph, Text: longText}},
		},
	}

	result := FormatDiffForAI(diff)
	// The text should be truncated
	if strings.Contains(result, longText) {
		t.Error("expected text to be truncated")
	}
	if !strings.Contains(result, "...") {
		t.Error("expected truncation marker")
	}
}
