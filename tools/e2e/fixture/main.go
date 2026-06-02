// Command fixture serves mutable HTML for the E2E monitoring pipeline test.
// GET /page   → current version's HTML
// POST /switch → flip v1 → v2 (forces a deterministic content change)
// GET /health → 200 readiness probe
package main

import (
	"net/http"
	"os"
	"sync"
)

const v1 = `<!doctype html><html><head><title>Pulzifi E2E Fixture</title></head>
<body><h1>Pulzifi E2E Fixture</h1>
<p>Version ONE — baseline content for monitoring. Nothing has changed yet.</p>
</body></html>`

const v2 = `<!doctype html><html><head><title>Pulzifi E2E Fixture</title></head>
<body><h1>Pulzifi E2E Fixture</h1>
<p>Version TWO — content CHANGED. Price updated, new section added, layout shifted.</p>
</body></html>`

type server struct {
	mu  sync.RWMutex
	v2  bool
	mux *http.ServeMux
}

func newServer() *server {
	s := &server{mux: http.NewServeMux()}
	s.mux.HandleFunc("/page", s.handlePage)
	s.mux.HandleFunc("/switch", s.handleSwitch)
	s.mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	return s
}

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) { s.mux.ServeHTTP(w, r) }

func (s *server) handlePage(w http.ResponseWriter, _ *http.Request) {
	s.mu.RLock()
	body := v1
	if s.v2 {
		body = v2
	}
	s.mu.RUnlock()
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	_, _ = w.Write([]byte(body))
}

func (s *server) handleSwitch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	s.mu.Lock()
	s.v2 = true
	s.mu.Unlock()
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("switched"))
}

func main() {
	addr := ":8080"
	if p := os.Getenv("PORT"); p != "" {
		addr = ":" + p
	}
	_ = http.ListenAndServe(addr, newServer())
}
