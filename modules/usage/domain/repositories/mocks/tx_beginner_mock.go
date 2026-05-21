package mocks

import "context"

// MockTxBeginner is a hand-rolled mock of repositories.TxBeginner.
type MockTxBeginner struct {
	BeginErr error
}

func (m *MockTxBeginner) Begin(_ context.Context) (interface{}, error) {
	if m.BeginErr != nil {
		return nil, m.BeginErr
	}
	return &struct{}{}, nil // opaque non-nil handle
}

func (m *MockTxBeginner) Commit(_ interface{}) error { return nil }
func (m *MockTxBeginner) Rollback(_ interface{})     {}
