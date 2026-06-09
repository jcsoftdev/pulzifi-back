package mocks

import "context"

// MockMediaStore is a hand-rolled mock for services.MediaStore.
type MockMediaStore struct {
	StoreResult string
	StoreErr    error

	StoreFn func(ctx context.Context, sourceURL, key string) (string, error)

	StoreCalls int
}

func (m *MockMediaStore) Store(ctx context.Context, sourceURL, key string) (string, error) {
	m.StoreCalls++
	if m.StoreFn != nil {
		return m.StoreFn(ctx, sourceURL, key)
	}
	return m.StoreResult, m.StoreErr
}
