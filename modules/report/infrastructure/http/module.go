package http

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	authmw "github.com/jcsoftdev/pulzifi-back/modules/auth/infrastructure/middleware"
	"github.com/jcsoftdev/pulzifi-back/modules/report/domain/entities"
	"github.com/jcsoftdev/pulzifi-back/modules/report/infrastructure/persistence"
	"github.com/jcsoftdev/pulzifi-back/shared/logger"
	"github.com/jcsoftdev/pulzifi-back/shared/middleware"
	"github.com/jcsoftdev/pulzifi-back/shared/router"
	"go.uber.org/zap"
)

// Module implements the router.ModuleRegisterer interface for the Report module
type Module struct {
	db *sql.DB
}

// NewModule creates a new instance of the Report module
func NewModule() router.ModuleRegisterer {
	return &Module{}
}

// NewModuleWithDB creates a new instance with database connection
func NewModuleWithDB(db *sql.DB) router.ModuleRegisterer {
	return &Module{db: db}
}

// ModuleName returns the name of the module
func (m *Module) ModuleName() string {
	return "Report"
}

// RegisterHTTPRoutes registers all HTTP routes for the Report module
func (m *Module) RegisterHTTPRoutes(router chi.Router) {
	router.Route("/reports", func(r chi.Router) {
		r.Use(middleware.AuthMiddleware.Authenticate)
		r.Use(middleware.OrgMiddleware.RequireOrganizationMembership)
		r.Use(middleware.RequireTenant)
		r.Post("/", m.handleCreateReport)
		r.Get("/", m.handleListReports)
		r.Get("/{id}", m.handleGetReport)
	})
}

// handleCreateReport creates a new report
// @Summary Create Report
// @Description Create a new report
// @Tags reports
// @Security BearerAuth
// @Accept json
// @Produce json
// @Success 201 {object} map[string]interface{}
// @Router /reports [post]
func (m *Module) handleCreateReport(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if m.db == nil {
		http.Error(w, "database not initialized", http.StatusServiceUnavailable)
		return
	}

	var req struct {
		PageID     string                 `json:"page_id"`
		Title      string                 `json:"title"`
		ReportDate string                 `json:"report_date"`
		Content    map[string]interface{} `json:"content"`
		PDFURL     string                 `json:"pdf_url"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid request body"})
		return
	}

	if req.Title == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "title is required"})
		return
	}

	pageID, err := uuid.Parse(req.PageID)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid page_id"})
		return
	}

	reportDate := time.Now()
	if req.ReportDate != "" {
		parsed, err := time.Parse("2006-01-02", req.ReportDate)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "invalid report_date format, use YYYY-MM-DD"})
			return
		}
		reportDate = parsed
	}

	userIDStr, _ := r.Context().Value(authmw.UserIDKey).(string)
	createdBy, err := uuid.Parse(userIDStr)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "unauthorized"})
		return
	}

	if req.Content == nil {
		req.Content = map[string]interface{}{}
	}

	report := &entities.Report{
		ID:         uuid.New(),
		PageID:     pageID,
		Title:      req.Title,
		ReportDate: reportDate,
		Content:    entities.Content(req.Content),
		PDFURL:     req.PDFURL,
		CreatedBy:  createdBy,
		CreatedAt:  time.Now(),
	}

	tenant := middleware.GetTenantFromContext(r.Context())
	repo := persistence.NewReportPostgresRepository(m.db, tenant)

	if err := repo.Create(r.Context(), report); err != nil {
		logger.Error("Failed to create report", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "failed to create report"})
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(report)
}

// handleListReports lists all reports
// @Summary List Reports
// @Description List all reports, optionally filtered by page_id
// @Tags reports
// @Security BearerAuth
// @Produce json
// @Param page_id query string false "Page ID filter"
// @Success 200 {object} map[string]interface{}
// @Router /reports [get]
func (m *Module) handleListReports(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if m.db == nil {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{"data": []interface{}{}, "count": 0})
		return
	}

	tenant := middleware.GetTenantFromContext(r.Context())
	repo := persistence.NewReportPostgresRepository(m.db, tenant)

	pageIDStr := r.URL.Query().Get("page_id")

	var reports []*entities.Report
	var err error

	if pageIDStr != "" {
		pageID, parseErr := uuid.Parse(pageIDStr)
		if parseErr != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "invalid page_id"})
			return
		}
		reports, err = repo.ListByPage(r.Context(), pageID)
	} else {
		reports, err = repo.List(r.Context())
	}

	if err != nil {
		logger.Error("Failed to list reports", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "failed to list reports"})
		return
	}

	if reports == nil {
		reports = []*entities.Report{}
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"data":  reports,
		"count": len(reports),
	})
}

// handleGetReport gets a report by ID
// @Summary Get Report
// @Description Get a report by ID
// @Tags reports
// @Security BearerAuth
// @Produce json
// @Param id path string true "Report ID"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]string
// @Router /reports/{id} [get]
func (m *Module) handleGetReport(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if m.db == nil {
		http.Error(w, "database not initialized", http.StatusServiceUnavailable)
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid report id"})
		return
	}

	tenant := middleware.GetTenantFromContext(r.Context())
	repo := persistence.NewReportPostgresRepository(m.db, tenant)

	report, err := repo.GetByID(r.Context(), id)
	if err != nil {
		logger.Error("Failed to get report", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "failed to get report"})
		return
	}
	if report == nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "report not found"})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(report)
}
