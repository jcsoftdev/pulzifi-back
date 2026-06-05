package html

import (
	"fmt"
	"strings"
)

// DiffOp represents the type of change detected between content blocks.
type DiffOp string

const (
	DiffAdded   DiffOp = "added"
	DiffRemoved DiffOp = "removed"
	DiffChanged DiffOp = "changed"
)

// BlockDiff represents a single difference between two content block sequences.
type BlockDiff struct {
	Op       DiffOp        `json:"op"`
	Block    ContentBlock  `json:"block"`
	OldBlock *ContentBlock `json:"old_block,omitempty"`
	// Section grouping for the UI. Assigned by EnrichSections. SectionOrder is the
	// zero-based display order of the section this diff belongs to.
	SectionName  string `json:"section_name,omitempty"`
	SectionOrder int    `json:"section_order,omitempty"`
}

// ContentDiff represents the result of comparing two content block sequences.
type ContentDiff struct {
	HasChanges   bool        `json:"has_changes"`
	TotalChanges int         `json:"total_changes"`
	Diffs        []BlockDiff `json:"diffs"`
}

// SectionNamer names a section's blocks. Implementations may call out to an AI
// backend. Returning "" means "no opinion" (caller applies a fallback).
type SectionNamer interface {
	NameSection(blocks []ContentBlock) string
}

// diffSectionBlock returns the block whose container defines this diff's section:
// the new block for added/changed, the old block for removed.
func diffSectionBlock(d BlockDiff) ContentBlock {
	if d.Op == DiffRemoved && d.OldBlock != nil {
		return *d.OldBlock
	}
	if d.Op == DiffRemoved {
		return d.Block
	}
	return d.Block
}

// EnrichSections groups the diffs by container signature (first-seen order) and
// assigns each diff a SectionName + SectionOrder. Naming precedence:
//  1. heuristicSectionName (deterministic, free)
//  2. namer.NameSection (AI; skipped when namer is nil or returns "")
//  3. "Sección N" fallback
func EnrichSections(diff *ContentDiff, namer SectionNamer) {
	if diff == nil || len(diff.Diffs) == 0 {
		return
	}

	order := make(map[string]int)
	var groupBlocks [][]ContentBlock

	for i := range diff.Diffs {
		b := diffSectionBlock(diff.Diffs[i])
		sig := sectionSignature(b)
		idx, ok := order[sig]
		if !ok {
			idx = len(groupBlocks)
			order[sig] = idx
			groupBlocks = append(groupBlocks, nil)
		}
		diff.Diffs[i].SectionOrder = idx
		groupBlocks[idx] = append(groupBlocks[idx], b)
	}

	names := make([]string, len(groupBlocks))
	for idx, blocks := range groupBlocks {
		name := heuristicSectionName(blocks)
		if name == "" && namer != nil {
			name = strings.TrimSpace(namer.NameSection(blocks))
		}
		if name == "" {
			name = fmt.Sprintf("Sección %d", idx+1)
		}
		names[idx] = name
	}

	for i := range diff.Diffs {
		diff.Diffs[i].SectionName = names[diff.Diffs[i].SectionOrder]
	}
}

// DiffContentBlocks compares two sequences of content blocks and produces a diff.
// It uses a two-pass approach:
//  1. LCS (Longest Common Subsequence) to find exact-match blocks that haven't changed.
//  2. Greedy type-matching on unmatched blocks to identify changes vs additions/removals.
func DiffContentBlocks(prev, curr []ContentBlock) *ContentDiff {
	if len(prev) == 0 && len(curr) == 0 {
		return &ContentDiff{}
	}

	// Pass 1: Compute LCS to find unchanged blocks.
	unmatchedPrev, unmatchedCurr := collectUnmatched(prev, curr)

	// Pass 2: Greedy type-matching on unmatched blocks.
	diffs := matchUnmatched(unmatchedPrev, unmatchedCurr)

	return &ContentDiff{
		HasChanges:   len(diffs) > 0,
		TotalChanges: len(diffs),
		Diffs:        diffs,
	}
}

// collectUnmatched computes the LCS of prev and curr, then returns the blocks
// from each side that are not part of the longest common subsequence.
func collectUnmatched(prev, curr []ContentBlock) (unmatchedPrev, unmatchedCurr []indexedBlock) {
	lcsTable := computeLCS(prev, curr)
	prevMatched := make([]bool, len(prev))
	currMatched := make([]bool, len(curr))

	// Backtrack LCS to mark matched positions.
	i, j := len(prev), len(curr)
	for i > 0 && j > 0 {
		switch {
		case blocksEqual(prev[i-1], curr[j-1]):
			prevMatched[i-1] = true
			currMatched[j-1] = true
			i--
			j--
		case lcsTable[i-1][j] >= lcsTable[i][j-1]:
			i--
		default:
			j--
		}
	}

	for idx, b := range prev {
		if !prevMatched[idx] {
			unmatchedPrev = append(unmatchedPrev, indexedBlock{idx: idx, block: b})
		}
	}
	for idx, b := range curr {
		if !currMatched[idx] {
			unmatchedCurr = append(unmatchedCurr, indexedBlock{idx: idx, block: b})
		}
	}
	return unmatchedPrev, unmatchedCurr
}

// matchUnmatched greedily pairs unmatched prev/curr blocks of the same type and
// level into "changed" diffs, emitting the leftovers as removed/added.
func matchUnmatched(unmatchedPrev, unmatchedCurr []indexedBlock) []BlockDiff {
	var diffs []BlockDiff
	currUsed := make([]bool, len(unmatchedCurr))

	for _, pb := range unmatchedPrev {
		matched := false
		for ci, cb := range unmatchedCurr {
			if currUsed[ci] {
				continue
			}
			if cb.block.Type == pb.block.Type && cb.block.Level == pb.block.Level {
				// Same type — this is a "changed" block.
				old := pb.block
				diffs = append(diffs, BlockDiff{
					Op:       DiffChanged,
					Block:    cb.block,
					OldBlock: &old,
				})
				currUsed[ci] = true
				matched = true
				break
			}
		}
		if !matched {
			diffs = append(diffs, BlockDiff{
				Op:    DiffRemoved,
				Block: pb.block,
			})
		}
	}

	// Remaining unmatched curr blocks are additions.
	for ci, cb := range unmatchedCurr {
		if !currUsed[ci] {
			diffs = append(diffs, BlockDiff{
				Op:    DiffAdded,
				Block: cb.block,
			})
		}
	}

	return diffs
}

type indexedBlock struct {
	idx   int
	block ContentBlock
}

// blocksEqual returns true if two content blocks are identical.
func blocksEqual(a, b ContentBlock) bool {
	return a.Type == b.Type &&
		a.Level == b.Level &&
		a.Text == b.Text &&
		a.Href == b.Href &&
		a.Src == b.Src &&
		a.Alt == b.Alt
}

// computeLCS computes the LCS table for two content block sequences.
func computeLCS(prev, curr []ContentBlock) [][]int {
	m, n := len(prev), len(curr)
	table := make([][]int, m+1)
	for i := range table {
		table[i] = make([]int, n+1)
	}
	for i := 1; i <= m; i++ {
		for j := 1; j <= n; j++ {
			switch {
			case blocksEqual(prev[i-1], curr[j-1]):
				table[i][j] = table[i-1][j-1] + 1
			case table[i-1][j] >= table[i][j-1]:
				table[i][j] = table[i-1][j]
			default:
				table[i][j] = table[i][j-1]
			}
		}
	}
	return table
}

const maxDiffTextLen = 200

// FormatDiffForAI produces a compact text representation of a ContentDiff
// suitable for sending to an LLM. Only changed/added/removed blocks are included.
// Text is truncated at 200 chars per block.
func FormatDiffForAI(diff *ContentDiff) string {
	if diff == nil || !diff.HasChanges {
		return ""
	}

	var sb strings.Builder
	for _, d := range diff.Diffs {
		switch d.Op {
		case DiffAdded:
			writeAddedRemoved(&sb, "ADDED", d.Block)
		case DiffRemoved:
			writeAddedRemoved(&sb, "REMOVED", d.Block)
		case DiffChanged:
			writeChanged(&sb, d)
		}
	}

	return sb.String()
}

// writeAddedRemoved formats an added or removed block (label is "ADDED" or "REMOVED").
func writeAddedRemoved(sb *strings.Builder, label string, b ContentBlock) {
	_, _ = fmt.Fprintf(sb, "[%s] %s: %q\n", label, formatBlockLabel(b), truncateDiff(b.Text, maxDiffTextLen))
	if b.Href != "" {
		_, _ = fmt.Fprintf(sb, "  (href: %s)\n", b.Href)
	}
	if b.Src != "" {
		_, _ = fmt.Fprintf(sb, "  (src: %s)\n", b.Src)
	}
}

// writeChanged formats a changed block, including an href transition line when relevant.
func writeChanged(sb *strings.Builder, d BlockDiff) {
	oldText := ""
	if d.OldBlock != nil {
		oldText = truncateDiff(d.OldBlock.Text, maxDiffTextLen)
	}
	newText := truncateDiff(d.Block.Text, maxDiffTextLen)
	_, _ = fmt.Fprintf(sb, "[CHANGED] %s: %q → %q\n", formatBlockLabel(d.Block), oldText, newText)
	if d.Block.Href != "" && (d.OldBlock == nil || d.OldBlock.Href != d.Block.Href) {
		oldHref := ""
		if d.OldBlock != nil {
			oldHref = d.OldBlock.Href
		}
		_, _ = fmt.Fprintf(sb, "  (href: %s → %s)\n", oldHref, d.Block.Href)
	}
}

// formatBlockLabel returns a human-readable label for a content block type.
func formatBlockLabel(b ContentBlock) string {
	switch b.Type {
	case BlockHeading:
		return fmt.Sprintf("H%d", b.Level)
	case BlockParagraph:
		return "P"
	case BlockLink:
		return "A"
	case BlockImage:
		return "IMG"
	case BlockListItem:
		return "LI"
	case BlockTableCell:
		return "TD"
	case BlockCode:
		return "CODE"
	case BlockBlockquote:
		return "BLOCKQUOTE"
	default:
		return string(b.Type)
	}
}

func truncateDiff(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}
