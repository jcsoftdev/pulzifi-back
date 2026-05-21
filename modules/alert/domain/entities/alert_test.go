package entities

import (
	"testing"

	"github.com/google/uuid"
)

func TestNewAlert(t *testing.T) {
	workspaceID := uuid.New()
	pageID := uuid.New()
	checkID := uuid.New()

	tests := []struct {
		name        string
		alertType   string
		title       string
		description string
	}{
		{
			name:        "change detected alert",
			alertType:   "change_detected",
			title:       "Content changed",
			description: "Major changes detected on the page",
		},
		{
			name:        "empty description allowed",
			alertType:   "change_detected",
			title:       "Content changed",
			description: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange/Act
			alert := NewAlert(workspaceID, pageID, checkID, tt.alertType, tt.title, tt.description)

			// Assert
			if alert == nil {
				t.Fatal("expected non-nil alert")
			}
			if alert.ID == uuid.Nil {
				t.Error("alert ID should not be nil UUID")
			}
			if alert.WorkspaceID != workspaceID {
				t.Errorf("WorkspaceID: want %v, got %v", workspaceID, alert.WorkspaceID)
			}
			if alert.PageID != pageID {
				t.Errorf("PageID: want %v, got %v", pageID, alert.PageID)
			}
			if alert.CheckID != checkID {
				t.Errorf("CheckID: want %v, got %v", checkID, alert.CheckID)
			}
			if alert.Type != tt.alertType {
				t.Errorf("Type: want %q, got %q", tt.alertType, alert.Type)
			}
			if alert.Title != tt.title {
				t.Errorf("Title: want %q, got %q", tt.title, alert.Title)
			}
			if alert.Description != tt.description {
				t.Errorf("Description: want %q, got %q", tt.description, alert.Description)
			}
			if alert.CreatedAt.IsZero() {
				t.Error("CreatedAt should not be zero")
			}
			if alert.ReadAt != nil {
				t.Error("ReadAt should be nil for a new alert")
			}
		})
	}
}

func TestAlertWithPage_Fields(t *testing.T) {
	// Verify AlertWithPage embeds Alert correctly
	base := NewAlert(uuid.New(), uuid.New(), uuid.New(), "change_detected", "Title", "Desc")
	awp := &AlertWithPage{
		Alert:    *base,
		PageName: "My Page",
		PageURL:  "https://example.com",
	}

	if awp.ID != base.ID {
		t.Errorf("embedded ID mismatch: want %v, got %v", base.ID, awp.ID)
	}
	if awp.PageName != "My Page" {
		t.Errorf("PageName: want %q, got %q", "My Page", awp.PageName)
	}
	if awp.PageURL != "https://example.com" {
		t.Errorf("PageURL: want %q, got %q", "https://example.com", awp.PageURL)
	}
}
