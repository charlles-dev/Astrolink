package jobs

import (
	"context"
	"testing"
	"time"
)

func TestExpireSessions_ChamaStoreQuandoImplementado(t *testing.T) {
	now := time.Date(2026, 5, 21, 12, 0, 0, 0, time.UTC)
	store := &sessionExpirer{expired: 3}

	expired, err := ExpireSessions(context.Background(), store, now)

	if err != nil {
		t.Fatal(err)
	}
	if expired != 3 {
		t.Fatalf("expired = %d, want 3", expired)
	}
	if !store.called || !store.now.Equal(now) {
		t.Fatalf("store nao chamado corretamente: %+v", store)
	}
}

func TestExpireSessions_RetornaZeroQuandoStoreNaoImplementa(t *testing.T) {
	expired, err := ExpireSessions(context.Background(), struct{}{}, time.Now().UTC())

	if err != nil {
		t.Fatal(err)
	}
	if expired != 0 {
		t.Fatalf("expired = %d, want 0", expired)
	}
}

type sessionExpirer struct {
	called  bool
	now     time.Time
	expired int
}

func (s *sessionExpirer) ExpireSessions(_ context.Context, now time.Time) (int, error) {
	s.called = true
	s.now = now
	return s.expired, nil
}
