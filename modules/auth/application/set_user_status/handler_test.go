package setuserstatus_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	setuserstatus "github.com/jcsoftdev/pulzifi-back/modules/auth/application/set_user_status"
)

type fakeWriter struct {
	gotID     uuid.UUID
	gotStatus string
	err       error
}

func (f *fakeWriter) SetStatus(_ context.Context, id uuid.UUID, status string) error {
	f.gotID, f.gotStatus = id, status
	return f.err
}

func TestSetUserStatus_Suspend(t *testing.T) {
	w := &fakeWriter{}
	h := setuserstatus.NewHandler(w)
	target := uuid.New()
	actor := uuid.New()

	err := h.Handle(context.Background(), setuserstatus.Request{TargetID: target, ActorID: actor, Status: "suspended"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if w.gotStatus != "suspended" || w.gotID != target {
		t.Errorf("writer called wrong: %s %s", w.gotID, w.gotStatus)
	}
}

func TestSetUserStatus_InvalidStatusRejected(t *testing.T) {
	h := setuserstatus.NewHandler(&fakeWriter{})
	err := h.Handle(context.Background(), setuserstatus.Request{TargetID: uuid.New(), ActorID: uuid.New(), Status: "banned"})
	if err != setuserstatus.ErrInvalidStatus {
		t.Fatalf("want ErrInvalidStatus, got %v", err)
	}
}

func TestSetUserStatus_SelfSuspendBlocked(t *testing.T) {
	id := uuid.New()
	h := setuserstatus.NewHandler(&fakeWriter{})
	err := h.Handle(context.Background(), setuserstatus.Request{TargetID: id, ActorID: id, Status: "suspended"})
	if err != setuserstatus.ErrCannotSuspendSelf {
		t.Fatalf("want ErrCannotSuspendSelf, got %v", err)
	}
}
