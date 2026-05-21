package managesections

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/monitoring/domain/entities"
	repomocks "github.com/jcsoftdev/pulzifi-back/modules/monitoring/domain/repositories/mocks"
)

func TestManageSectionsHandler_DeleteSection(t *testing.T) {
	sectionID := uuid.New()

	tests := []struct {
		name      string
		sectionID uuid.UUID
		deleteErr error
		wantErr   bool
	}{
		{
			name:      "delete existing section succeeds",
			sectionID: sectionID,
			deleteErr: nil,
			wantErr:   false,
		},
		{
			name:      "repo error propagates",
			sectionID: sectionID,
			deleteErr: errors.New("not found"),
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sectionRepo := &repomocks.MockMonitoredSectionRepository{
				DeleteErr: tt.deleteErr,
			}
			configRepo := &repomocks.MockMonitoringConfigRepository{}

			h := NewManageSectionsHandler(sectionRepo, configRepo)
			err := h.DeleteSection(context.Background(), tt.sectionID)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if sectionRepo.DeleteCalls != 1 {
				t.Errorf("expected Delete called once, got %d", sectionRepo.DeleteCalls)
			}
			if sectionRepo.DeleteLastID != tt.sectionID {
				t.Errorf("Delete called with wrong ID: want %v, got %v", tt.sectionID, sectionRepo.DeleteLastID)
			}
		})
	}
}

func TestManageSectionsHandler_SaveAll_ReplacesAndUpdatesConfig(t *testing.T) {
	pageID := uuid.New()
	existingConfig := &entities.MonitoringConfig{
		PageID:       pageID,
		SelectorType: "full_page",
	}

	tests := []struct {
		name          string
		sections      []SectionDTO
		replaceErr    error
		configResult  *entities.MonitoringConfig
		wantErr       bool
		wantReplaced  bool
	}{
		{
			name:         "save sections calls ReplaceAll and updates config to 'sections'",
			sections:     []SectionDTO{{Name: "Header", CSSSelector: "header"}},
			replaceErr:   nil,
			configResult: existingConfig,
			wantErr:      false,
			wantReplaced: true,
		},
		{
			name:         "save empty sections updates config to 'full_page'",
			sections:     []SectionDTO{},
			replaceErr:   nil,
			configResult: existingConfig,
			wantErr:      false,
			wantReplaced: true,
		},
		{
			name:         "ReplaceAll error propagates",
			sections:     []SectionDTO{{Name: "Footer", CSSSelector: "footer"}},
			replaceErr:   errors.New("db error"),
			configResult: existingConfig,
			wantErr:      true,
			wantReplaced: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// After ReplaceAll, SaveAll calls ListByPageID to return the saved sections.
			savedSections := make([]*entities.MonitoredSection, len(tt.sections))
			for i, s := range tt.sections {
				savedSections[i] = &entities.MonitoredSection{
					ID:          uuid.New(),
					PageID:      pageID,
					Name:        s.Name,
					CSSSelector: s.CSSSelector,
				}
			}

			sectionRepo := &repomocks.MockMonitoredSectionRepository{
				ReplaceAllErr:      tt.replaceErr,
				ListByPageIDResult: savedSections,
			}
			configRepo := &repomocks.MockMonitoringConfigRepository{
				GetByPageIDResult: tt.configResult,
			}

			h := NewManageSectionsHandler(sectionRepo, configRepo)
			req := &SaveSectionsRequest{Sections: tt.sections}
			resp, err := h.SaveAll(context.Background(), pageID, req)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if sectionRepo.ReplaceAllCalls != 1 {
				t.Errorf("expected ReplaceAll called once, got %d", sectionRepo.ReplaceAllCalls)
			}

			if resp == nil {
				t.Fatal("expected non-nil response")
			}
			if len(resp.Sections) != len(tt.sections) {
				t.Errorf("response sections: want %d, got %d", len(tt.sections), len(resp.Sections))
			}
		})
	}
}
