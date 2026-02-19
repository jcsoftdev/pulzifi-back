package http

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	generateinsights "github.com/jcsoftdev/pulzifi-back/modules/insight/application/generate_insights"
	listinsights "github.com/jcsoftdev/pulzifi-back/modules/insight/application/list_insights"
	insightAI "github.com/jcsoftdev/pulzifi-back/modules/insight/infrastructure/ai"
	"github.com/jcsoftdev/pulzifi-back/modules/insight/infrastructure/persistence"
	monPersistence "github.com/jcsoftdev/pulzifi-back/modules/monitoring/infrastructure/persistence"
	sharedAI "github.com/jcsoftdev/pulzifi-back/shared/ai"
	"github.com/jcsoftdev/pulzifi-back/shared/config"
	sharedHTML "github.com/jcsoftdev/pulzifi-back/shared/html"
	"github.com/jcsoftdev/pulzifi-back/shared/logger"
	"github.com/jcsoftdev/pulzifi-back/shared/middleware"
	"github.com/jcsoftdev/pulzifi-back/shared/pubsub"
	"github.com/jcsoftdev/pulzifi-back/shared/router"
	"go.uber.org/zap"
)

// Module implements the router.ModuleRegisterer interface for the Insight module
type Module struct {
	db             *sql.DB
	insightHandler *generateinsights.GenerateInsightsHandler
	broker         *pubsub.InsightBroker
}

// NewModule creates a new instance of the Insight module
func NewModule() router.ModuleRegisterer {
	return &Module{}
}

// NewModuleWithDB creates a new instance with database connection and pub/sub broker.
func NewModuleWithDB(db *sql.DB, broker *pubsub.InsightBroker) router.ModuleRegisterer {
	m := &Module{db: db, broker: broker}

	cfg := config.Load()
	if cfg.OpenRouterAPIKey != "" {
		openRouterClient := sharedAI.NewOpenRouterClient(cfg.OpenRouterAPIKey, cfg.OpenRouterModel)
		generator := insightAI.NewOpenRouterGenerator(openRouterClient)
		m.insightHandler = generateinsights.NewGenerateInsightsHandler(generator, db)
	}

	return m
}

// ModuleName returns the name of the module
func (m *Module) ModuleName() string {
	return "Insight"
}

// RegisterHTTPRoutes registers all HTTP routes for the Insight module
func (m *Module) RegisterHTTPRoutes(router chi.Router) {
	router.Route("/insights", func(r chi.Router) {
		// SSE endpoint — auth only (broker is keyed by globally-unique check UUID,
		// so no tenant resolution is required here).
		r.With(middleware.AuthMiddleware.Authenticate).Get("/sse", m.handleInsightSSE)

		// All other endpoints require full auth + tenant.
		r.Group(func(r chi.Router) {
			r.Use(middleware.AuthMiddleware.Authenticate)
			r.Use(middleware.OrgMiddleware.RequireOrganizationMembership)
			r.Use(middleware.RequireTenant)
			r.Post("/generate", m.handleGenerateInsight)
			r.Get("/", m.handleListInsights)
			r.Get("/{id}", m.handleGetInsight)
		})
	})
}

// handleInsightSSE streams an insight-ready event to the client using SSE.
// The client connects with ?check_id=<uuid> and waits; when the background
// generation finishes the broker publishes the insights and this handler
// writes them as a single SSE "data:" frame, then closes the connection.
func (m *Module) handleInsightSSE(w http.ResponseWriter, r *http.Request) {
	checkIDStr := r.URL.Query().Get("check_id")
	if _, err := uuid.Parse(checkIDStr); err != nil {
		http.Error(w, "invalid check_id", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.WriteHeader(http.StatusOK)

	rc := http.NewResponseController(w)
	// Flush headers immediately so the client sees the 200 OK right away
	// and doesn't close the connection while waiting for the response header.
	if err := rc.Flush(); err != nil {
		return
	}

	ch, unsubscribe := m.broker.Subscribe(checkIDStr)
	defer unsubscribe()

	// Give the LLM up to 120 s; after that we tell the client to fall back.
	ctx, cancel := context.WithTimeout(r.Context(), 120*time.Second)
	defer cancel()

	select {
	case payload, ok := <-ch:
		if ok {
			fmt.Fprintf(w, "data: %s\n\n", payload)
			rc.Flush() //nolint:errcheck
		}
	case <-ctx.Done():
		fmt.Fprintf(w, "data: {\"timeout\":true}\n\n")
		rc.Flush() //nolint:errcheck
	}
}

// handleGenerateInsight generates insights for a specific check using the page's HTML content.
// @Summary Generate Insight
// @Description Generate insights for a page check using its HTML content
// @Tags insights
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body object true "Generate Insight Request" SchemaExample({"page_id":"uuid","check_id":"uuid"})
// @Success 201 {object} listinsights.ListInsightsResponse
// @Router /insights/generate [post]
func (m *Module) handleGenerateInsight(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if m.db == nil || m.insightHandler == nil {
		http.Error(w, "insight generation not available", http.StatusServiceUnavailable)
		return
	}

	var req struct {
		PageID  string `json:"page_id"`
		CheckID string `json:"check_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	pageID, err := uuid.Parse(req.PageID)
	if err != nil {
		http.Error(w, "invalid page_id", http.StatusBadRequest)
		return
	}
	checkID, err := uuid.Parse(req.CheckID)
	if err != nil {
		http.Error(w, "invalid check_id", http.StatusBadRequest)
		return
	}

	tenant := middleware.GetTenantFromContext(r.Context())

	// Validate that the check exists before starting generation
	checkRepo := monPersistence.NewCheckPostgresRepository(m.db, tenant)
	check, err := checkRepo.GetByID(r.Context(), checkID)
	if err != nil || check == nil {
		http.Error(w, "check not found", http.StatusNotFound)
		return
	}

	// Return 202 immediately — generation runs in the background.
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]string{"status": "generating"})

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
		defer cancel()

		allChecks, err := checkRepo.ListByPage(ctx, pageID)
		if err != nil {
			logger.Error("Failed to list checks for insight generation", zap.Error(err))
			return
		}
		var prevHTMLURL string
		for _, c := range allChecks {
			if c.ID != checkID && c.Status == "success" {
				prevHTMLURL = c.HTMLSnapshotURL
				break
			}
		}

		newText := fetchHTMLText(check.HTMLSnapshotURL)
		prevText := fetchHTMLText(prevHTMLURL)

		configRepo := monPersistence.NewMonitoringConfigPostgresRepository(m.db, tenant)
		pageConfig, _ := configRepo.GetByPageID(ctx, pageID)
		enabledTypes := []string{"marketing", "market_analysis"}
		if pageConfig != nil && len(pageConfig.EnabledInsightTypes) > 0 {
			enabledTypes = pageConfig.EnabledInsightTypes
		}

		pageURL, _ := configRepo.GetPageURL(ctx, pageID)

		if err := m.insightHandler.Handle(ctx, &generateinsights.Request{
			PageID:              pageID,
			CheckID:             checkID,
			PageURL:             pageURL,
			PrevText:            prevText,
			NewText:             newText,
			SchemaName:          tenant,
			EnabledInsightTypes: enabledTypes,
		}); err != nil {
			logger.Error("Failed to generate insights on demand", zap.Error(err))
			return
		}

		// Publish generated insights to any SSE subscriber waiting on this check.
		if m.broker != nil {
			repo := persistence.NewInsightPostgresRepository(m.db, tenant)
			handler := listinsights.NewListInsightsHandler(repo)
			resp, err := handler.HandleByCheckID(ctx, checkID)
			if err == nil {
				payload, _ := json.Marshal(resp)
				m.broker.Publish(checkID.String(), payload)
			}
		}

		logger.Info("On-demand insight generation completed", zap.String("check_id", checkID.String()))
	}()
}

// fetchHTMLText downloads HTML from url and extracts plain text.
func fetchHTMLText(url string) string {
	if url == "" {
		return ""
	}
	resp, err := http.Get(url) //nolint:noctx
	if err != nil {
		logger.Error("Failed to fetch HTML for text extraction", zap.String("url", url), zap.Error(err))
		return ""
	}
	defer resp.Body.Close()
	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return ""
	}
	return sharedHTML.ExtractText(string(content))
}

// handleListInsights lists all insights
// @Summary List Insights
// @Description List all insights
// @Tags insights
// @Security BearerAuth
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /insights [get]
func (m *Module) handleListInsights(w http.ResponseWriter, r *http.Request) {
	if m.db == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"insights": []interface{}{},
			"message":  "list insights (mock - db not initialized)",
		})
		return
	}

	tenant := middleware.GetTenantFromContext(r.Context())
	repo := persistence.NewInsightPostgresRepository(m.db, tenant)
	handler := listinsights.NewListInsightsHandler(repo)
	handler.HandleHTTP(w, r)
}

// handleGetInsight gets an insight by ID
// @Summary Get Insight
// @Description Get an insight by ID
// @Tags insights
// @Security BearerAuth
// @Produce json
// @Param id path string true "Insight ID"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]string
// @Router /insights/{id} [get]
func (m *Module) handleGetInsight(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid insight id"})
		return
	}

	tenant := middleware.GetTenantFromContext(r.Context())
	repo := persistence.NewInsightPostgresRepository(m.db, tenant)

	insight, err := repo.GetByID(r.Context(), id)
	if err != nil {
		logger.Error("Failed to get insight", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "failed to get insight"})
		return
	}
	if insight == nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "insight not found"})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(insight)
}
