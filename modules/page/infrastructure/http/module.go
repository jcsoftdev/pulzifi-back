package http

import (
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	bulkdeletepages "github.com/jcsoftdev/pulzifi-back/modules/page/application/bulk_delete_pages"
	createpage "github.com/jcsoftdev/pulzifi-back/modules/page/application/create_page"
	deletepage "github.com/jcsoftdev/pulzifi-back/modules/page/application/delete_page"
	getpage "github.com/jcsoftdev/pulzifi-back/modules/page/application/get_page"
	listpages "github.com/jcsoftdev/pulzifi-back/modules/page/application/list_pages"
	updatepage "github.com/jcsoftdev/pulzifi-back/modules/page/application/update_page"
	pageservices "github.com/jcsoftdev/pulzifi-back/modules/page/domain/services"
	"github.com/jcsoftdev/pulzifi-back/modules/page/infrastructure/persistence"
	"github.com/jcsoftdev/pulzifi-back/shared/contextkeys"
	"github.com/jcsoftdev/pulzifi-back/shared/logger"
	"github.com/jcsoftdev/pulzifi-back/shared/middleware"
	"github.com/jcsoftdev/pulzifi-back/shared/router"
	"go.uber.org/zap"
)

// Module implements the router.ModuleRegisterer interface for the Page module
type Module struct {
	db              *sql.DB
	previewStreamer  pageservices.PagePreviewStreamer
}

// NewModule creates a new instance of the Page module
func NewModule() router.ModuleRegisterer {
	return &Module{}
}

// NewModuleWithDB creates a new instance with database connection
func NewModuleWithDB(db *sql.DB) router.ModuleRegisterer {
	return &Module{
		db: db,
	}
}

// NewModuleWithExtractor creates a new instance with database and preview streamer
func NewModuleWithExtractor(db *sql.DB, previewStreamer pageservices.PagePreviewStreamer) router.ModuleRegisterer {
	return &Module{
		db:             db,
		previewStreamer: previewStreamer,
	}
}

// ModuleName returns the name of the module
func (m *Module) ModuleName() string {
	return "Page"
}

// RegisterHTTPRoutes registers all HTTP routes for the Page module
func (m *Module) RegisterHTTPRoutes(router chi.Router) {
	router.Route("/pages", func(r chi.Router) {
		r.Use(middleware.AuthMiddleware.Authenticate)
		r.Use(middleware.OrgMiddleware.RequireOrganizationMembership)

		// Preview endpoint doesn't need tenant context
		r.Post("/preview", m.handlePreviewPage)

		r.Group(func(r chi.Router) {
			r.Use(middleware.RequireTenant)
			r.Post("/", m.handleCreatePage)
			r.Post("/bulk-delete", m.handleBulkDeletePages)
			r.Get("/", m.handleListPages)
			r.Get("/{id}", m.handleGetPage)
			r.Put("/{id}", m.handleUpdatePage)
			r.Delete("/{id}", m.handleDeletePage)
		})
	})
}

// handlePreviewPage proxies the SSE stream from the extractor service.
// The extractor emits "progress", "result", and "error" SSE events which
// are forwarded to the client in real-time for live progress feedback.
func (m *Module) handlePreviewPage(w http.ResponseWriter, r *http.Request) {
	if m.previewStreamer == nil {
		http.Error(w, "Extractor service not configured", http.StatusServiceUnavailable)
		return
	}

	var req struct {
		URL             string `json:"url"`
		BlockAdsCookies bool   `json:"block_ads_cookies"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if req.URL == "" {
		http.Error(w, "URL is required", http.StatusBadRequest)
		return
	}

	body, err := m.previewStreamer.PreviewStream(r.Context(), req.URL, req.BlockAdsCookies)
	if err != nil {
		logger.Error("Failed to start preview stream", zap.Error(err))
		http.Error(w, "Failed to preview page", http.StatusInternalServerError)
		return
	}
	defer body.Close()

	// Set SSE headers and proxy the stream
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")
	w.WriteHeader(http.StatusOK)

	rc := http.NewResponseController(w)
	// Flush headers immediately so the client sees the 200 OK right away.
	if err := rc.Flush(); err != nil {
		return
	}

	// Stream the extractor response to the client, flushing after each chunk.
	// A keepalive ticker sends SSE comments every 15s to prevent intermediate
	// proxies (Railway, Cloudflare, etc.) from killing idle connections.
	keepalive := time.NewTicker(15 * time.Second)
	defer keepalive.Stop()

	// readCh delivers chunks from the extractor body in a goroutine so we can
	// select between data arriving and the keepalive ticker firing.
	type readResult struct {
		data []byte
		err  error
	}
	readCh := make(chan readResult, 1)

	go func() {
		buf := make([]byte, 4096)
		for {
			n, rErr := body.Read(buf)
			if n > 0 {
				chunk := make([]byte, n)
				copy(chunk, buf[:n])
				readCh <- readResult{data: chunk}
			}
			if rErr != nil {
				readCh <- readResult{err: rErr}
				return
			}
		}
	}()

	for {
		select {
		case rr := <-readCh:
			if rr.data != nil {
				if _, writeErr := w.Write(rr.data); writeErr != nil {
					return
				}
				if flushErr := rc.Flush(); flushErr != nil {
					return
				}
			}
			if rr.err != nil {
				if rr.err != io.EOF {
					logger.Error("Error reading preview stream", zap.Error(rr.err))
				}
				return
			}
		case <-keepalive.C:
			if _, writeErr := w.Write([]byte(": keepalive\n\n")); writeErr != nil {
				return
			}
			if flushErr := rc.Flush(); flushErr != nil {
				return
			}
		case <-r.Context().Done():
			return
		}
	}
}

// handleCreatePage creates a new page
// @Summary Create Page
// @Description Create a new page
// @Tags pages
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body createpage.CreatePageRequest true "Create Page Request"
// @Success 201 {object} createpage.CreatePageResponse
// @Router /pages [post]
func (m *Module) handleCreatePage(w http.ResponseWriter, r *http.Request) {
	// If db is not available, return mock response
	if m.db == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":      "550e8400-e29b-41d4-a716-446655440000",
			"message": "create page (mock - db not initialized)",
		})
		return
	}

	var req createpage.CreatePageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if req.WorkspaceID == uuid.Nil || req.Name == "" || req.URL == "" {
		http.Error(w, "workspace_id, name, and url are required", http.StatusBadRequest)
		return
	}

	userIDStr, ok := r.Context().Value(contextkeys.UserIDKey).(string)
	if !ok {
		logger.Error("User ID not found in context")
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	createdBy, err := uuid.Parse(userIDStr)
	if err != nil {
		logger.Error("Invalid user ID", zap.Error(err))
		http.Error(w, "invalid user ID", http.StatusBadRequest)
		return
	}

	// Get tenant from context
	tenant := middleware.GetTenantFromContext(r.Context())
	repo := persistence.NewPagePostgresRepository(m.db, tenant)

	handler := createpage.NewCreatePageHandler(repo)
	resp, err := handler.Handle(r.Context(), &req, createdBy)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

// handleListPages lists all pages
// @Summary List Pages
// @Description List all pages for a workspace
// @Tags pages
// @Security BearerAuth
// @Produce json
// @Param workspace_id query string true "Workspace ID"
// @Success 200 {object} listpages.ListPagesResponse
// @Router /pages [get]
func (m *Module) handleListPages(w http.ResponseWriter, r *http.Request) {
	// If db is not available, return mock response
	if m.db == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"pages":   []interface{}{},
			"message": "list pages (mock - db not initialized)",
		})
		return
	}

	workspaceIDStr := r.URL.Query().Get("workspace_id")
	if workspaceIDStr == "" {
		logger.ErrorWithContext(r.Context(), "workspace_id query parameter is required")
		http.Error(w, "workspace_id query parameter is required", http.StatusBadRequest)
		return
	}
	workspaceID, err := uuid.Parse(workspaceIDStr)
	if err != nil {
		logger.ErrorWithContext(r.Context(), "Invalid workspace_id format",
			zap.String("workspace_id", workspaceIDStr),
			zap.Error(err),
		)
		http.Error(w, "invalid workspace_id", http.StatusBadRequest)
		return
	}

	tenant := middleware.GetTenantFromContext(r.Context())
	repo := persistence.NewPagePostgresRepository(m.db, tenant)

	handler := listpages.NewListPagesHandler(repo)
	response, err := handler.Handle(r.Context(), workspaceID)
	if err != nil {
		logger.ErrorWithContext(r.Context(), "Failed to list pages",
			zap.String("workspace_id", workspaceID.String()),
			zap.Error(err),
		)
		http.Error(w, "failed to list pages\n", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// handleGetPage gets a page by ID
// @Summary Get Page
// @Description Get a page by ID
// @Tags pages
// @Security BearerAuth
// @Produce json
// @Param id path string true "Page ID"
// @Success 200 {object} getpage.GetPageResponse
// @Router /pages/{id} [get]
func (m *Module) handleGetPage(w http.ResponseWriter, r *http.Request) {
	if m.db == nil {
		http.Error(w, "Database not initialized", http.StatusInternalServerError)
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid page ID", http.StatusBadRequest)
		return
	}

	tenant := middleware.GetTenantFromContext(r.Context())
	repo := persistence.NewPagePostgresRepository(m.db, tenant)
	handler := getpage.NewGetPageHandler(repo)

	resp, err := handler.Handle(r.Context(), id)
	if err != nil {
		logger.Error("Failed to get page", zap.Error(err))
		http.Error(w, "Failed to get page", http.StatusInternalServerError)
		return
	}
	if resp == nil {
		http.Error(w, "Page not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

// handleUpdatePage updates a page
// @Summary Update Page
// @Description Update a page
// @Tags pages
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Page ID"
// @Param request body updatepage.UpdatePageRequest true "Update Page Request"
// @Success 200 {object} updatepage.UpdatePageResponse
// @Router /pages/{id} [put]
func (m *Module) handleUpdatePage(w http.ResponseWriter, r *http.Request) {
	if m.db == nil {
		http.Error(w, "Database not initialized", http.StatusInternalServerError)
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid page ID", http.StatusBadRequest)
		return
	}

	var req updatepage.UpdatePageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	tenant := middleware.GetTenantFromContext(r.Context())
	repo := persistence.NewPagePostgresRepository(m.db, tenant)
	handler := updatepage.NewUpdatePageHandler(repo)

	resp, err := handler.Handle(r.Context(), id, &req)
	if err != nil {
		logger.Error("Failed to update page", zap.Error(err))
		http.Error(w, "Failed to update page", http.StatusInternalServerError)
		return
	}
	if resp == nil {
		http.Error(w, "Page not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

// handleDeletePage deletes a page
// @Summary Delete Page
// @Description Delete a page
// @Tags pages
// @Security BearerAuth
// @Success 204
// @Router /pages/{id} [delete]
func (m *Module) handleDeletePage(w http.ResponseWriter, r *http.Request) {
	if m.db == nil {
		http.Error(w, "Database not initialized", http.StatusInternalServerError)
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid page ID", http.StatusBadRequest)
		return
	}

	tenant := middleware.GetTenantFromContext(r.Context())
	repo := persistence.NewPagePostgresRepository(m.db, tenant)
	handler := deletepage.NewDeletePageHandler(repo)

	if err := handler.Handle(r.Context(), id); err != nil {
		logger.Error("Failed to delete page", zap.Error(err))
		http.Error(w, "Failed to delete page", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// handleBulkDeletePages deletes multiple pages at once
// @Summary Bulk Delete Pages
// @Description Delete multiple pages by ID
// @Tags pages
// @Security BearerAuth
// @Accept json
// @Param request body bulkdeletepages.BulkDeletePagesRequest true "Bulk Delete Request"
// @Success 204
// @Router /pages/bulk-delete [post]
func (m *Module) handleBulkDeletePages(w http.ResponseWriter, r *http.Request) {
	if m.db == nil {
		http.Error(w, "Database not initialized", http.StatusInternalServerError)
		return
	}

	var req bulkdeletepages.BulkDeletePagesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if len(req.IDs) == 0 {
		http.Error(w, "ids must not be empty", http.StatusBadRequest)
		return
	}

	ids := make([]uuid.UUID, 0, len(req.IDs))
	for _, idStr := range req.IDs {
		id, err := uuid.Parse(idStr)
		if err != nil {
			http.Error(w, "Invalid page ID: "+idStr, http.StatusBadRequest)
			return
		}
		ids = append(ids, id)
	}

	tenant := middleware.GetTenantFromContext(r.Context())
	repo := persistence.NewPagePostgresRepository(m.db, tenant)
	handler := bulkdeletepages.NewBulkDeletePagesHandler(repo)

	if err := handler.Handle(r.Context(), ids); err != nil {
		logger.Error("Failed to bulk delete pages", zap.Error(err))
		http.Error(w, "Failed to delete pages", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
