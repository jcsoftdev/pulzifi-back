package html

import "strings"

// sectionSignature returns a stable key identifying a block's container, used to
// group diffs. Two blocks under the same kind of container (same tag/role/aria/
// class) share a signature — so sibling cards with the same wrapper group together
// and the same card across snapshots merges. Block text is intentionally excluded.
func sectionSignature(b ContentBlock) string {
	return strings.Join([]string{
		b.SectionTag, b.SectionRole, b.SectionAria, b.SectionClass,
	}, "|")
}

// heuristicSectionName returns a deterministic semantic name for a section's
// blocks, or "" when no rule matches (caller may then fall back to AI/heading).
func heuristicSectionName(blocks []ContentBlock) string {
	if len(blocks) == 0 {
		return ""
	}
	b := blocks[0]

	// An explicit aria-label is the strongest signal.
	if a := strings.TrimSpace(b.SectionAria); a != "" {
		return a
	}

	role := strings.ToLower(b.SectionRole)
	tag := strings.ToLower(b.SectionTag)

	switch {
	case tag == "header" || role == "banner":
		return "Hero"
	case tag == "footer" || role == "contentinfo":
		return "Footer"
	case tag == "nav" || role == "navigation":
		return "Navegación"
	}

	// Stats pattern: several short blocks that are mostly numeric/currency.
	if looksLikeStats(blocks) {
		return "Estadísticas"
	}

	return ""
}

// looksLikeStats reports whether the section contains 3+ short numeric/currency
// blocks (e.g. "+$900K", "137.000", "40").
func looksLikeStats(blocks []ContentBlock) bool {
	numeric := 0
	for _, b := range blocks {
		t := strings.TrimSpace(b.Text)
		if t == "" || len(t) > 12 {
			continue
		}
		if strings.IndexAny(t, "0123456789") >= 0 {
			numeric++
		}
	}
	return numeric >= 3
}
