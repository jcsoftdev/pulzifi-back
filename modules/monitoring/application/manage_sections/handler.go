package managesections

import (
	"context"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/monitoring/domain/entities"
	"github.com/jcsoftdev/pulzifi-back/modules/monitoring/domain/repositories"
)

// ManageSectionsHandler handles CRUD operations for monitored sections.
type ManageSectionsHandler struct {
	sectionRepo repositories.MonitoredSectionRepository
	configRepo  repositories.MonitoringConfigRepository
}

// NewManageSectionsHandler creates a new handler.
func NewManageSectionsHandler(sectionRepo repositories.MonitoredSectionRepository, configRepo repositories.MonitoringConfigRepository) *ManageSectionsHandler {
	return &ManageSectionsHandler{
		sectionRepo: sectionRepo,
		configRepo:  configRepo,
	}
}

// List returns all sections for a page.
func (h *ManageSectionsHandler) List(ctx context.Context, pageID uuid.UUID) (*ListSectionsResponse, error) {
	sections, err := h.sectionRepo.ListByPageID(ctx, pageID)
	if err != nil {
		return nil, err
	}

	resp := &ListSectionsResponse{
		Sections: make([]*SectionResponse, len(sections)),
	}
	for i, s := range sections {
		resp.Sections[i] = toSectionResponse(s)
	}
	return resp, nil
}

// SaveAll replaces all sections for a page atomically.
// It also updates the monitoring config selector_type to "sections" when sections are provided,
// or back to "full_page" when the list is empty.
func (h *ManageSectionsHandler) SaveAll(ctx context.Context, pageID uuid.UUID, req *SaveSectionsRequest) (*ListSectionsResponse, error) {
	// Build domain entities
	domainSections := make([]*entities.MonitoredSection, len(req.Sections))
	for i, dto := range req.Sections {
		var offsets *entities.SelectorOffsets
		if dto.SelectorOffsets != nil {
			offsets = &entities.SelectorOffsets{
				Top:    dto.SelectorOffsets.Top,
				Right:  dto.SelectorOffsets.Right,
				Bottom: dto.SelectorOffsets.Bottom,
				Left:   dto.SelectorOffsets.Left,
			}
		}
		var rect *entities.SectionRect
		if dto.Rect != nil {
			rect = &entities.SectionRect{
				X: dto.Rect.X,
				Y: dto.Rect.Y,
				W: dto.Rect.W,
				H: dto.Rect.H,
			}
		}
		domainSections[i] = entities.NewMonitoredSection(
			pageID, dto.Name, dto.CSSSelector, dto.XPathSelector, offsets, rect, dto.ViewportWidth, dto.SortOrder,
		)
	}

	// Replace all sections
	if err := h.sectionRepo.ReplaceAll(ctx, pageID, domainSections); err != nil {
		return nil, err
	}

	// Update monitoring config selector_type
	config, err := h.configRepo.GetByPageID(ctx, pageID)
	if err != nil {
		return nil, err
	}
	if config != nil {
		if len(domainSections) > 0 {
			config.SelectorType = "sections"
		} else {
			config.SelectorType = "full_page"
		}
		// Clear legacy single-selector fields when switching to multi-section
		config.CSSSelector = ""
		config.XPathSelector = ""
		config.SelectorOffsets = nil
		if err := h.configRepo.Update(ctx, config); err != nil {
			return nil, err
		}
	}

	// Return the saved sections
	saved, err := h.sectionRepo.ListByPageID(ctx, pageID)
	if err != nil {
		return nil, err
	}

	resp := &ListSectionsResponse{
		Sections: make([]*SectionResponse, len(saved)),
	}
	for i, s := range saved {
		resp.Sections[i] = toSectionResponse(s)
	}
	return resp, nil
}

// DeleteSection removes a single section by ID.
func (h *ManageSectionsHandler) DeleteSection(ctx context.Context, sectionID uuid.UUID) error {
	return h.sectionRepo.Delete(ctx, sectionID)
}

func toSectionResponse(s *entities.MonitoredSection) *SectionResponse {
	resp := &SectionResponse{
		ID:            s.ID,
		PageID:        s.PageID,
		Name:          s.Name,
		CSSSelector:   s.CSSSelector,
		XPathSelector: s.XPathSelector,
		SortOrder:     s.SortOrder,
		ViewportWidth: s.ViewportWidth,
		CreatedAt:     s.CreatedAt,
		UpdatedAt:     s.UpdatedAt,
	}
	if s.SelectorOffsets != nil {
		resp.SelectorOffsets = &SectionOffsetsDTO{
			Top:    s.SelectorOffsets.Top,
			Right:  s.SelectorOffsets.Right,
			Bottom: s.SelectorOffsets.Bottom,
			Left:   s.SelectorOffsets.Left,
		}
	}
	if s.Rect != nil {
		resp.Rect = &SectionRectDTO{
			X: s.Rect.X,
			Y: s.Rect.Y,
			W: s.Rect.W,
			H: s.Rect.H,
		}
	}
	return resp
}
