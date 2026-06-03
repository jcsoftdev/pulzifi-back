package listusers_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	listusers "github.com/jcsoftdev/pulzifi-back/modules/auth/application/list_users"
)

// fakeAdminUserReader is an in-memory implementation of AdminUserReader for tests.
type fakeAdminUserReader struct {
	rows  []listusers.AdminUserRow
	total int
	err   error

	// Call tracking
	lastSearch string
	lastLimit  int
	lastOffset int
}

func (f *fakeAdminUserReader) ListUsers(_ context.Context, search string, limit, offset int) ([]listusers.AdminUserRow, int, error) {
	f.lastSearch = search
	f.lastLimit = limit
	f.lastOffset = offset
	if f.err != nil {
		return nil, 0, f.err
	}
	return f.rows, f.total, nil
}

func TestListUsersHandler_Handle(t *testing.T) {
	id1 := uuid.New()
	id2 := uuid.New()
	rows := []listusers.AdminUserRow{
		{ID: id1, Email: "alice@example.com", FirstName: "Alice", LastName: "Smith", IsSuperAdmin: false},
		{ID: id2, Email: "bob@example.com", FirstName: "Bob", LastName: "Jones", IsSuperAdmin: true},
	}

	tests := []struct {
		name        string
		req         listusers.Request
		readerRows  []listusers.AdminUserRow
		readerTotal int
		readerErr   error
		wantErr     bool
		checkResult func(t *testing.T, resp *listusers.Response)
		checkReader func(t *testing.T, r *fakeAdminUserReader)
	}{
		{
			name:        "returns paginated users",
			req:         listusers.Request{Search: "", Page: 1, PageSize: 5},
			readerRows:  rows,
			readerTotal: 2,
			checkResult: func(t *testing.T, resp *listusers.Response) {
				t.Helper()
				if len(resp.Users) != 2 {
					t.Fatalf("expected 2 users, got %d", len(resp.Users))
				}
				if resp.Total != 2 {
					t.Errorf("total: want 2, got %d", resp.Total)
				}
				if resp.Page != 1 {
					t.Errorf("page: want 1, got %d", resp.Page)
				}
				if resp.PageSize != 5 {
					t.Errorf("pageSize: want 5, got %d", resp.PageSize)
				}
				if resp.Users[0].ID != id1.String() {
					t.Errorf("first user ID mismatch")
				}
				if resp.Users[1].IsSuperAdmin != true {
					t.Errorf("expected second user to be super admin")
				}
			},
			checkReader: func(t *testing.T, r *fakeAdminUserReader) {
				t.Helper()
				if r.lastLimit != 5 {
					t.Errorf("limit: want 5, got %d", r.lastLimit)
				}
				if r.lastOffset != 0 {
					t.Errorf("offset: want 0, got %d", r.lastOffset)
				}
				if r.lastSearch != "" {
					t.Errorf("search: want empty, got %q", r.lastSearch)
				}
			},
		},
		{
			name:        "computes correct offset for page 2",
			req:         listusers.Request{Search: "alice", Page: 2, PageSize: 10},
			readerRows:  rows[:1],
			readerTotal: 11,
			checkResult: func(t *testing.T, resp *listusers.Response) {
				t.Helper()
				if resp.Page != 2 {
					t.Errorf("page: want 2, got %d", resp.Page)
				}
			},
			checkReader: func(t *testing.T, r *fakeAdminUserReader) {
				t.Helper()
				if r.lastOffset != 10 {
					t.Errorf("offset: want 10 (page 2 * size 10), got %d", r.lastOffset)
				}
				if r.lastSearch != "alice" {
					t.Errorf("search: want %q, got %q", "alice", r.lastSearch)
				}
			},
		},
		{
			name: "clamps pageSize above 50 to 50",
			req:  listusers.Request{Search: "", Page: 1, PageSize: 200},
			checkResult: func(t *testing.T, resp *listusers.Response) {
				t.Helper()
				if resp.PageSize != 50 {
					t.Errorf("pageSize: want 50 (clamped), got %d", resp.PageSize)
				}
			},
			checkReader: func(t *testing.T, r *fakeAdminUserReader) {
				t.Helper()
				if r.lastLimit != 50 {
					t.Errorf("limit passed to reader: want 50, got %d", r.lastLimit)
				}
			},
		},
		{
			name:    "reader error propagates",
			req:     listusers.Request{Search: "", Page: 1, PageSize: 5},
			readerErr: errors.New("db error"),
			wantErr: true,
		},
		{
			name:        "page defaults to minimum 1 when zero",
			req:         listusers.Request{Search: "", Page: 0, PageSize: 5},
			readerRows:  nil,
			readerTotal: 0,
			checkResult: func(t *testing.T, resp *listusers.Response) {
				t.Helper()
				if resp.Page != 1 {
					t.Errorf("page: want 1 (normalized), got %d", resp.Page)
				}
			},
			checkReader: func(t *testing.T, r *fakeAdminUserReader) {
				t.Helper()
				if r.lastOffset != 0 {
					t.Errorf("offset: want 0, got %d", r.lastOffset)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := &fakeAdminUserReader{
				rows:  tt.readerRows,
				total: tt.readerTotal,
				err:   tt.readerErr,
			}
			h := listusers.NewHandler(reader)
			resp, err := h.Handle(context.Background(), tt.req)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.checkResult != nil {
				tt.checkResult(t, resp)
			}
			if tt.checkReader != nil {
				tt.checkReader(t, reader)
			}
		})
	}
}
