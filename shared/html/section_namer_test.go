package html

import "testing"

func TestSectionSignature(t *testing.T) {
	a := ContentBlock{SectionTag: "section", SectionClass: "stats", SectionAria: "x"}
	b := ContentBlock{SectionTag: "section", SectionClass: "stats", SectionAria: "x", Text: "different"}
	if sectionSignature(a) != sectionSignature(b) {
		t.Error("signature must ignore block text, only container hints")
	}
	c := ContentBlock{SectionTag: "footer"}
	if sectionSignature(a) == sectionSignature(c) {
		t.Error("different containers must differ")
	}
}

func TestHeuristicSectionName(t *testing.T) {
	tests := []struct {
		name   string
		blocks []ContentBlock
		want   string
	}{
		{"header→Hero", []ContentBlock{{SectionTag: "header"}}, "Hero"},
		{"role banner→Hero", []ContentBlock{{SectionTag: "div", SectionRole: "banner"}}, "Hero"},
		{"footer→Footer", []ContentBlock{{SectionTag: "footer"}}, "Footer"},
		{"nav→Navegación", []ContentBlock{{SectionTag: "nav"}}, "Navegación"},
		{"aria wins", []ContentBlock{{SectionTag: "section", SectionAria: "Pricing"}}, "Pricing"},
		{"unknown→empty", []ContentBlock{{SectionTag: "section", SectionClass: "xyz"}}, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := heuristicSectionName(tt.blocks); got != tt.want {
				t.Errorf("heuristicSectionName = %q, want %q", got, tt.want)
			}
		})
	}
}
