package main

import (
	"path/filepath"
	"sync"
	"testing"
)

func TestConcurrentWithdrawalsNeverOverdraw(t *testing.T) {
	store, err := NewStore(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	u, err := store.CreateUser("A", "a@example.com", "hash", "MIE0000000001")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.ApplyTransaction(u.ID, "deposit", 1000); err != nil {
		t.Fatal(err)
	}

	var wg sync.WaitGroup
	var mu sync.Mutex
	ok := 0
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := store.ApplyTransaction(u.ID, "withdraw", -100); err == nil {
				mu.Lock()
				ok++
				mu.Unlock()
			}
		}()
	}
	wg.Wait()

	got, _ := store.GetByID(u.ID)
	if ok != 10 || got.BalanceCents != 0 {
		t.Fatalf("expected 10 successful withdrawals and 0 balance, got %d and %d", ok, got.BalanceCents)
	}
}
