package http

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	sendemail "github.com/jcsoftdev/pulzifi-back/modules/email/application/send_email"
	"github.com/jcsoftdev/pulzifi-back/modules/email/domain/services"
	"github.com/jcsoftdev/pulzifi-back/modules/email/infrastructure/persistence"
	"github.com/jcsoftdev/pulzifi-back/shared/middleware"
	"github.com/jcsoftdev/pulzifi-back/shared/router"
)

// Module implements router.ModuleRegisterer for the Email module.
type Module struct {
	provider services.EmailProvider
}

// NewModule creates a new Email module.
func NewModule(provider services.EmailProvider) router.ModuleRegisterer {
	return &Module{provider: provider}
}

func (m *Module) ModuleName() string {
	return "Email"
}

func (m *Module) RegisterHTTPRoutes(r chi.Router) {
	r.Route("/emails", func(r chi.Router) {
		r.Use(middleware.AuthMiddleware.Authenticate)
		r.Use(middleware.OrgMiddleware.RequireOrganizationMembership)
		r.Post("/send", m.handleSendEmail)
	})
}

func (m *Module) handleSendEmail(w http.ResponseWriter, r *http.Request) {
	var req sendemail.SendEmailRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid request body"})
		return
	}

	emailRepo := persistence.NewMemoryEmailRepository()
	emailService := services.NewEmailService()
	handler := sendemail.NewSendEmailHandler(emailRepo, emailService, m.provider)

	resp, err := handler.Handle(r.Context(), &req)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

// Provider returns the email provider for use by other modules.
func (m *Module) Provider() services.EmailProvider {
	return m.provider
}
