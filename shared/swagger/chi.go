package swagger

import (
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	httpSwagger "github.com/swaggo/http-swagger"
)

// SetupSwaggerForChi configures Swagger UI for Chi router
// Add this to your router setup (usually the v1 router):
// swagger.SetupSwaggerForChi(v1Router)
func SetupSwaggerForChi(router chi.Router) {
	// Register the swagger spec manually
	swaggerHandler := httpSwagger.Handler(
		httpSwagger.URL("/api/v1/docs/doc.json"),
		httpSwagger.DeepLinking(true),
		httpSwagger.DocExpansion("none"),
		httpSwagger.DomID("swagger-ui"),
	)

	// Route for doc.json - serve the swagger spec JSON
	router.Get("/docs/doc.json", func(w http.ResponseWriter, r *http.Request) {
		// Read and serve the swagger.json file
		data, err := ioutil.ReadFile("docs/swagger.json")
		if err != nil {
			// If file doesn't exist, try alternate paths
			data, err = ioutil.ReadFile("/app/docs/swagger.json")
			if err != nil {
				http.Error(w, "Swagger spec not found", http.StatusNotFound)
				return
			}
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(data)
	})

	// Redirect /docs to /api/v1/docs/index.html
	router.Get("/docs", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/api/v1/docs/index.html", http.StatusMovedPermanently)
	})

	// Serve swagger UI and assets
	router.Get("/docs/*", func(w http.ResponseWriter, r *http.Request) {
		// Strip the /docs prefix
		if strings.HasPrefix(r.URL.Path, "/docs") {
			r.URL.Path = strings.TrimPrefix(r.URL.Path, "/docs")
		}
		if r.URL.Path == "" {
			r.URL.Path = "/"
		}
		swaggerHandler.ServeHTTP(w, r)
	})
}
