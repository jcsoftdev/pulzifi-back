package http

import (
	"encoding/json"
	"encoding/xml"
	"net/http"
	"text/template"

	"go.uber.org/zap"
)

// ResponseHelper provides common response patterns for HTTP handlers
type ResponseHelper struct {
	logger *zap.Logger
}

// NewResponseHelper creates a new response helper
func NewResponseHelper(logger *zap.Logger) *ResponseHelper {
	return &ResponseHelper{
		logger: logger,
	}
}

// RespondJSON writes a JSON response
func (rh *ResponseHelper) RespondJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		rh.logger.Error("failed to encode JSON response", zap.Error(err))
	}
}

// RespondError writes an error JSON response
func (rh *ResponseHelper) RespondError(w http.ResponseWriter, statusCode int, message string) {
	rh.RespondJSON(w, statusCode, map[string]string{
		"error": message,
	})
}

// RespondXML writes an XML response
func (rh *ResponseHelper) RespondXML(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/xml")
	w.WriteHeader(statusCode)
	if err := xml.NewEncoder(w).Encode(data); err != nil {
		rh.logger.Error("failed to encode XML response", zap.Error(err))
	}
}

// RespondHTML writes an HTML response
func (rh *ResponseHelper) RespondHTML(w http.ResponseWriter, statusCode int, templateStr string, data interface{}) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(statusCode)

	tmpl, err := template.New("response").Parse(templateStr)
	if err != nil {
		rh.logger.Error("failed to parse template", zap.Error(err))
		return
	}

	if err := tmpl.Execute(w, data); err != nil {
		rh.logger.Error("failed to execute template", zap.Error(err))
	}
}

// RespondText writes a plain text response
func (rh *ResponseHelper) RespondText(w http.ResponseWriter, statusCode int, text string) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(statusCode)
	if _, err := w.Write([]byte(text)); err != nil {
		rh.logger.Error("failed to write text response", zap.Error(err))
	}
}

// RespondNoContent writes a 204 No Content response
func (rh *ResponseHelper) RespondNoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

// RespondCreated writes a 201 Created response with JSON
func (rh *ResponseHelper) RespondCreated(w http.ResponseWriter, data interface{}) {
	rh.RespondJSON(w, http.StatusCreated, data)
}

// RespondOK writes a 200 OK response with JSON
func (rh *ResponseHelper) RespondOK(w http.ResponseWriter, data interface{}) {
	rh.RespondJSON(w, http.StatusOK, data)
}

// RespondBadRequest writes a 400 Bad Request response
func (rh *ResponseHelper) RespondBadRequest(w http.ResponseWriter, message string) {
	rh.RespondError(w, http.StatusBadRequest, message)
}

// RespondUnauthorized writes a 401 Unauthorized response
func (rh *ResponseHelper) RespondUnauthorized(w http.ResponseWriter, message string) {
	rh.RespondError(w, http.StatusUnauthorized, message)
}

// RespondForbidden writes a 403 Forbidden response
func (rh *ResponseHelper) RespondForbidden(w http.ResponseWriter, message string) {
	rh.RespondError(w, http.StatusForbidden, message)
}

// RespondNotFound writes a 404 Not Found response
func (rh *ResponseHelper) RespondNotFound(w http.ResponseWriter, message string) {
	rh.RespondError(w, http.StatusNotFound, message)
}

// RespondConflict writes a 409 Conflict response
func (rh *ResponseHelper) RespondConflict(w http.ResponseWriter, message string) {
	rh.RespondError(w, http.StatusConflict, message)
}

// RespondInternalServerError writes a 500 Internal Server Error response
func (rh *ResponseHelper) RespondInternalServerError(w http.ResponseWriter, message string) {
	rh.RespondError(w, http.StatusInternalServerError, message)
}

// RespondNotImplemented writes a 501 Not Implemented response
func (rh *ResponseHelper) RespondNotImplemented(w http.ResponseWriter, message string) {
	rh.RespondError(w, http.StatusNotImplemented, message)
}

// Package-level convenience functions (for backwards compatibility)

var defaultHelper *ResponseHelper

// SetDefaultLogger sets the logger for package-level functions
func SetDefaultLogger(logger *zap.Logger) {
	defaultHelper = NewResponseHelper(logger)
}

// RespondJSON is a package-level convenience function
func RespondJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

// RespondError is a package-level convenience function
func RespondError(w http.ResponseWriter, statusCode int, message string) {
	RespondJSON(w, statusCode, map[string]string{
		"error": message,
	})
}
