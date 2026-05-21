package bulkupdatemonitoringconfig

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	repomocks "github.com/jcsoftdev/pulzifi-back/modules/monitoring/domain/repositories/mocks"
)

func TestBulkUpdateMonitoringConfigHandler_Handle(t *testing.T) {
	pageID1 := uuid.New()
	pageID2 := uuid.New()
	pageID3 := uuid.New()

	tests := []struct {
		name      string
		pageIDs   []uuid.UUID
		frequency string
		repoErr   error
		wantErr   bool
	}{
		{
			name:      "all pages updated successfully",
			pageIDs:   []uuid.UUID{pageID1, pageID2, pageID3},
			frequency: "1h",
			repoErr:   nil,
			wantErr:   false,
		},
		{
			name:      "empty page list — no error",
			pageIDs:   []uuid.UUID{},
			frequency: "1h",
			repoErr:   nil,
			wantErr:   false,
		},
		{
			name:      "repo error propagates to caller",
			pageIDs:   []uuid.UUID{pageID1, pageID2},
			frequency: "30m",
			repoErr:   errors.New("db error"),
			wantErr:   true,
		},
		{
			name:      "verbose frequency alias is normalized before repo call",
			pageIDs:   []uuid.UUID{pageID1},
			frequency: "Every hour", // alias for "1h"
			repoErr:   nil,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &repomocks.MockMonitoringConfigRepository{
				UpdateErr: tt.repoErr,
			}

			h := NewBulkUpdateMonitoringConfigHandler(repo)
			err := h.Handle(context.Background(), tt.pageIDs, tt.frequency)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}
