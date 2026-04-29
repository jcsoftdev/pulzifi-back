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
}

// ContentDiff represents the result of comparing two content block sequences.
type ContentDiff struct {
	HasChanges   bool        `json:"has_changes"`
	TotalChanges int         `json:"total_changes"`
	Diffs        []BlockDiff `json:"diffs"`
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
	lcsTable := computeLCS(prev, curr)
	prevMatched := make([]bool, len(prev))
	currMatched := make([]bool, len(curr))

	// Backtrack LCS to mark matched positions.
	i, j := len(prev), len(curr)
	for i > 0 && j > 0 {
		if blocksEqual(prev[i-1], curr[j-1]) {
			prevMatched[i-1] = true
			currMatched[j-1] = true
			i--
			j--
		} else if lcsTable[i-1][j] >= lcsTable[i][j-1] {
			i--
		} else {
			j--
		}
	}

	// Collect unmatched blocks from prev and curr.
	var unmatchedPrev []indexedBlock
	var unmatchedCurr []indexedBlock
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

	// Pass 2: Greedy type-matching on unmatched blocks.
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

	return &ContentDiff{
		HasChanges:   len(diffs) > 0,
		TotalChanges: len(diffs),
		Diffs:        diffs,
	}
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
			if blocksEqual(prev[i-1], curr[j-1]) {
				table[i][j] = table[i-1][j-1] + 1
			} else if table[i-1][j] >= table[i][j-1] {
				table[i][j] = table[i-1][j]
			} else {
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
			sb.WriteString(fmt.Sprintf("[ADDED] %s: %q\n", formatBlockLabel(d.Block), truncateDiff(d.Block.Text, maxDiffTextLen)))
			if d.Block.Href != "" {
				sb.WriteString(fmt.Sprintf("  (href: %s)\n", d.Block.Href))
			}
			if d.Block.Src != "" {
				sb.WriteString(fmt.Sprintf("  (src: %s)\n", d.Block.Src))
			}

		case DiffRemoved:
			sb.WriteString(fmt.Sprintf("[REMOVED] %s: %q\n", formatBlockLabel(d.Block), truncateDiff(d.Block.Text, maxDiffTextLen)))
			if d.Block.Href != "" {
				sb.WriteString(fmt.Sprintf("  (href: %s)\n", d.Block.Href))
			}
			if d.Block.Src != "" {
				sb.WriteString(fmt.Sprintf("  (src: %s)\n", d.Block.Src))
			}

		case DiffChanged:
			oldText := ""
			if d.OldBlock != nil {
				oldText = truncateDiff(d.OldBlock.Text, maxDiffTextLen)
			}
			newText := truncateDiff(d.Block.Text, maxDiffTextLen)
			sb.WriteString(fmt.Sprintf("[CHANGED] %s: %q → %q\n", formatBlockLabel(d.Block), oldText, newText))
			if d.Block.Href != "" && (d.OldBlock == nil || d.OldBlock.Href != d.Block.Href) {
				oldHref := ""
				if d.OldBlock != nil {
					oldHref = d.OldBlock.Href
				}
				sb.WriteString(fmt.Sprintf("  (href: %s → %s)\n", oldHref, d.Block.Href))
			}
		}
	}

	return sb.String()
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
