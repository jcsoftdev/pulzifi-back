package http

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	createreport "github.com/jcsoftdev/pulzifi-back/modules/report/application/create_report"
	getreport "github.com/jcsoftdev/pulzifi-back/modules/report/application/get_report"
	listreports "github.com/jcsoftdev/pulzifi-back/modules/report/application/list_reports"
	"github.com/jcsoftdev/pulzifi-back/modules/report/domain/entities"
	"github.com/jcsoftdev/pulzifi-back/modules/report/infrastructure/persistence"
	"github.com/jcsoftdev/pulzifi-back/shared/contextkeys"
	"github.com/jcsoftdev/pulzifi-back/shared/middleware"
	"github.com/jcsoftdev/pulzifi-back/shared/router"
)

// ModuleDeps carries all injected dependencies for the Report module.
type ModuleDeps struct {
	DB                  *sql.DB
	CreateReportHandler *createreport.Handler
	ListReportsHandler  *listreports.Handler
	GetReportHandler    *getreport.Handler
}

// Module implements the router.ModuleRegisterer interface for the Report module.
type Module struct {
	deps ModuleDeps
}

// NewModule creates a new Report module with injected dependencies.
func NewModule(deps ModuleDeps) router.ModuleRegisterer {
	return &Module{deps: deps}
}

// NewModuleWithDB creates a new instance with database connection (backward-compat shim).
// Builds handlers directly from the DB; prefer NewModule for full DI in production.
func NewModuleWithDB(db *sql.DB) router.ModuleRegisterer {
	return &Module{deps: ModuleDeps{DB: db}}
}

// ModuleName returns the name of the module.
func (m *Module) ModuleName() string {
	return "Report"
}

// RegisterHTTPRoutes registers all HTTP routes for the Report module.
func (m *Module) RegisterHTTPRoutes(r chi.Router) {
	r.Route("/reports", func(r chi.Router) {
		r.Use(middleware.AuthMiddleware.Authenticate)
		r.Use(middleware.OrgMiddleware.RequireOrganizationMembership)
		r.Use(middleware.RequireTenant)
		r.Post("/", m.handleCreateReport)
		r.Get("/", m.handleListReports)
		r.Get("/{id}", m.handleGetReport)
	})
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v) //nolint:errcheck
}

func (m *Module) handleCreateReport(w http.ResponseWriter, r *http.Request) {
	if m.deps.DB == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "database not initialized"})
		return
	}

	var body struct {
		PageID     string                 `json:"page_id"`
		Title      string                 `json:"title"`
		ReportDate string                 `json:"report_date"`
		Content    map[string]interface{} `json:"content"`
		PDFURL     string                 `json:"pdf_url"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}
	if body.Title == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "title is required"})
		return
	}
	pageID, err := uuid.Parse(body.PageID)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid page_id"})
		return
	}
	reportDate := time.Now()
	if body.ReportDate != "" {
		parsed, err := time.Parse("2006-01-02", body.ReportDate)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid report_date format, use YYYY-MM-DD"})
			return
		}
		reportDate = parsed
	}
	userIDStr, _ := r.Context().Value(contextkeys.UserIDKey).(string)
	createdBy, err := uuid.Parse(userIDStr)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}
	if body.Content == nil {
		body.Content = map[string]interface{}{}
	}

	tenant := middleware.GetTenantFromContext(r.Context())
	req := &createreport.Request{
		PageID:     pageID,
		Title:      body.Title,
		ReportDate: reportDate,
		Content:    entities.Content(body.Content),
		PDFURL:     body.PDFURL,
		CreatedBy:  createdBy,
	}

	var handler *createreport.Handler
	if m.deps.CreateReportHandler != nil {
		handler = m.deps.CreateReportHandler
	} else {
		handler = createreport.NewHandler(persistence.NewReportPostgresRepository(m.deps.DB, tenant))
	}

	resp, err := handler.Handle(r.Context(), req)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to create report"})
		return
	}
	writeJSON(w, http.StatusCreated, resp.Report)
}

func (m *Module) handleListReports(w http.ResponseWriter, r *http.Request) {
	if m.deps.DB == nil {
		writeJSON(w, http.StatusOK, map[string]interface{}{"data": []interface{}{}, "count": 0})
		return
	}

	tenant := middleware.GetTenantFromContext(r.Context())
	var req listreports.Request
	if pageIDStr := r.URL.Query().Get("page_id"); pageIDStr != "" {
		pageID, err := uuid.Parse(pageIDStr)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid page_id"})
			return
		}
		req.PageID = &pageID
	}

	var handler *listreports.Handler
	if m.deps.ListReportsHandler != nil {
		handler = m.deps.ListReportsHandler
	} else {
		handler = listreports.NewHandler(persistence.NewReportPostgresRepository(m.deps.DB, tenant))
	}

	resp, err := handler.Handle(r.Context(), &req)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to list reports"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"data": resp.Reports, "count": resp.Count})
}

func (m *Module) handleGetReport(w http.ResponseWriter, r *http.Request) {
	if m.deps.DB == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "database not initialized"})
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid report id"})
		return
	}

	tenant := middleware.GetTenantFromContext(r.Context())
	var handler *getreport.Handler
	if m.deps.GetReportHandler != nil {
		handler = m.deps.GetReportHandler
	} else {
		handler = getreport.NewHandler(persistence.NewReportPostgresRepository(m.deps.DB, tenant))
	}

	resp, err := handler.Handle(r.Context(), &getreport.Request{ReportID: id})
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to get report"})
		return
	}
	if resp.Report == nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "report not found"})
		return
	}
	writeJSON(w, http.StatusOK, resp.Report)
}
