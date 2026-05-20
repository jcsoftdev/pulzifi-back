package snapshotwiring

import (
	monEntities "github.com/jcsoftdev/pulzifi-back/modules/monitoring/domain/entities"
	snapEntities "github.com/jcsoftdev/pulzifi-back/modules/snapshot/domain/entities"
)

// toMonCheck maps a snapshot Check to a monitoring Check for persistence.
func toMonCheck(c *snapEntities.Check) *monEntities.Check {
	return &monEntities.Check{
		ID:                  c.ID,
		PageID:              c.PageID,
		SectionID:           c.SectionID,
		ParentCheckID:       c.ParentCheckID,
		Status:              c.Status,
		ScreenshotURL:       c.ScreenshotURL,
		HTMLSnapshotURL:     c.HTMLSnapshotURL,
		ContentHash:         c.ContentHash,
		ChangeDetected:      c.ChangeDetected,
		ChangeType:          c.ChangeType,
		ErrorMessage:        c.ErrorMessage,
		DurationMs:          c.DurationMs,
		ScreenshotHash:      c.ScreenshotHash,
		VisionChangeSummary: c.VisionChangeSummary,
		ContentBlockHash:    c.ContentBlockHash,
		ContentDiffJSON:     c.ContentDiffJSON,
		DiffImageURL:        c.DiffImageURL,
		CheckedAt:           c.CheckedAt,
	}
}

// fromMonCheck maps a monitoring Check to a snapshot Check.
func fromMonCheck(c *monEntities.Check) *snapEntities.Check {
	return &snapEntities.Check{
		ID:                  c.ID,
		PageID:              c.PageID,
		SectionID:           c.SectionID,
		ParentCheckID:       c.ParentCheckID,
		Status:              c.Status,
		ScreenshotURL:       c.ScreenshotURL,
		HTMLSnapshotURL:     c.HTMLSnapshotURL,
		ContentHash:         c.ContentHash,
		ChangeDetected:      c.ChangeDetected,
		ChangeType:          c.ChangeType,
		ErrorMessage:        c.ErrorMessage,
		DurationMs:          c.DurationMs,
		ScreenshotHash:      c.ScreenshotHash,
		VisionChangeSummary: c.VisionChangeSummary,
		ContentBlockHash:    c.ContentBlockHash,
		ContentDiffJSON:     c.ContentDiffJSON,
		DiffImageURL:        c.DiffImageURL,
		CheckedAt:           c.CheckedAt,
	}
}
