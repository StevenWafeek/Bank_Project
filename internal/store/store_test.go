package store

import (
	"path/filepath"
	"sync"
	"testing"

	"miebank/internal/models"
)

func newTestStore(t *testing.T) *Store {
	s, err := New(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { s.Close() })
	return s
}

func TestConcurrentWithdrawalsNeverOverdraw(t *testing.T) {
	s := newTestStore(t)
	u, err := s.CreateUser("A", "a@example.com", "hash", "MIE0000000001")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.ApplyTransaction(u.ID, models.TxDeposit, 1000); err != nil {
		t.Fatal(err)
	}

	var wg sync.WaitGroup
	var mu sync.Mutex
	ok := 0
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := s.ApplyTransaction(u.ID, models.TxWithdraw, -100); err == nil {
				mu.Lock()
				ok++
				mu.Unlock()
			}
		}()
	}
	wg.Wait()

	got, _ := s.GetByID(u.ID)
	if ok != 10 || got.BalanceCents != 0 {
		t.Fatalf("expected 10 successful withdrawals and 0 balance, got %d and %d", ok, got.BalanceCents)
	}
}

func TestUsersAndHistory(t *testing.T) {
	s := newTestStore(t)
	u, err := s.CreateUser("A", "a@example.com", "hash", "MIE0000000001")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.CreateUser("B", "A@Example.com", "hash", "MIE0000000002"); err != ErrUserExists {
		t.Fatalf("duplicate email (case-insensitive): got %v", err)
	}
	if _, err := s.GetByEmail("nobody@example.com"); err != ErrUserNotFound {
		t.Fatalf("missing user: got %v", err)
	}
	if _, err := s.ApplyTransaction(u.ID, models.TxWithdraw, -1); err != ErrInsufficientFunds {
		t.Fatalf("overdraw: got %v", err)
	}
	s.ApplyTransaction(u.ID, models.TxDeposit, 5000)
	s.ApplyTransaction(u.ID, models.TxWithdraw, -2000)
	txs, err := s.ListTransactions(u.ID, 50)
	if err != nil || len(txs) != 2 || txs[0].Type != models.TxWithdraw || txs[0].BalanceCents != 3000 {
		t.Fatalf("history: %+v %v", txs, err)
	}
}
