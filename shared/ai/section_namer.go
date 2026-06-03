package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	sharedhtml "github.com/jcsoftdev/pulzifi-back/shared/html"
)

// completer is the subset of OllamaClient used for naming (allows test stubs).
type completer interface {
	Complete(ctx context.Context, prompt string) (string, error)
}

// OllamaSectionNamer implements sharedhtml.SectionNamer using an Ollama backend,
// with an in-memory cache keyed by section signature. On any error it returns ""
// so the diff layer applies its "Sección N" fallback.
type OllamaSectionNamer struct {
	client completer
	mu     sync.Mutex
	cache  map[string]string
}

// NewOllamaSectionNamer builds a namer over the given completer.
func NewOllamaSectionNamer(client completer) *OllamaSectionNamer {
	return &OllamaSectionNamer{client: client, cache: make(map[string]string)}
}

func sigKey(b sharedhtml.ContentBlock) string {
	return strings.Join([]string{b.SectionTag, b.SectionRole, b.SectionAria, b.SectionClass}, "|")
}

// NameSection returns a short semantic name for the section, or "" on failure.
func (n *OllamaSectionNamer) NameSection(blocks []sharedhtml.ContentBlock) string {
	if len(blocks) == 0 {
		return ""
	}
	key := sigKey(blocks[0])

	n.mu.Lock()
	if v, ok := n.cache[key]; ok {
		n.mu.Unlock()
		return v
	}
	n.mu.Unlock()

	var sample strings.Builder
	for i, b := range blocks {
		if i >= 6 {
			break
		}
		t := strings.TrimSpace(b.Text)
		if t == "" {
			continue
		}
		if len(t) > 80 {
			t = t[:80]
		}
		fmt.Fprintf(&sample, "- %s\n", t)
	}

	prompt := fmt.Sprintf(
		"You label web page sections. Given these text snippets from one section, "+
			"return a JSON object {\"name\": \"<1-2 word Spanish label>\"} naming the section "+
			"(e.g. Hero, Estadísticas, Precios, Testimonios, CTA, Footer). Snippets:\n%s",
		sample.String())

	out, err := n.client.Complete(context.Background(), prompt)
	if err != nil {
		return ""
	}

	var parsed struct {
		Name string `json:"name"`
	}
	if err := json.Unmarshal([]byte(out), &parsed); err != nil {
		return ""
	}
	name := strings.TrimSpace(parsed.Name)

	n.mu.Lock()
	n.cache[key] = name
	n.mu.Unlock()
	return name
}

// compile-time assertion: *OllamaSectionNamer must satisfy html.SectionNamer.
var _ sharedhtml.SectionNamer = (*OllamaSectionNamer)(nil)
