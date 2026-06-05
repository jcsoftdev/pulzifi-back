package ai

import (
	"context"
	"testing"

	sharedhtml "github.com/jcsoftdev/pulzifi-back/shared/html"
)

type stubCompleter struct {
	out   string
	calls int
}

func (s *stubCompleter) Complete(_ context.Context, _ string) (string, error) {
	s.calls++
	return s.out, nil
}

func TestOllamaSectionNamer_ParsesAndCaches(t *testing.T) {
	stub := &stubCompleter{out: `{"name":"Pricing"}`}
	namer := NewOllamaSectionNamer(stub)

	blocks := []sharedhtml.ContentBlock{{Type: sharedhtml.BlockParagraph, Text: "Plan A", SectionClass: "pricing"}}
	if got := namer.NameSection(blocks); got != "Pricing" {
		t.Fatalf("NameSection = %q, want Pricing", got)
	}
	_ = namer.NameSection(blocks) // same signature → cached
	if stub.calls != 1 {
		t.Errorf("completer called %d times, want 1 (cached)", stub.calls)
	}
}
