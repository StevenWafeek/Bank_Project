package store

import (
	"path/filepath"
	"strings"
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

func TestTransfer(t *testing.T) {
	s := newTestStore(t)
	alice, err := s.CreateUser("Alice", "alice@example.com", "hash1", "MIE1111111111")
	if err != nil {
		t.Fatal(err)
	}
	bob, err := s.CreateUser("Bob", "bob@example.com", "hash2", "MIE2222222222")
	if err != nil {
		t.Fatal(err)
	}

	// Fund Alice with $100 (10000 cents)
	s.ApplyTransaction(alice.ID, models.TxDeposit, 10000)

	// Transfer $40 from Alice to Bob by account number
	txA, recBob, err := s.Transfer(alice.ID, "MIE2222222222", 4000, "Dinner split")
	if err != nil {
		t.Fatalf("transfer: %v", err)
	}
	if recBob.Name != "Bob" || txA.BalanceCents != 6000 || txA.Type != models.TxTransferOut {
		t.Fatalf("unexpected sender transfer tx: %+v", txA)
	}
	if !strings.Contains(txA.Description, "Dinner split") {
		t.Fatalf("expected note in description: %s", txA.Description)
	}

	// Check Bob's balance and history
	bobUser, _ := s.GetByID(bob.ID)
	if bobUser.BalanceCents != 4000 {
		t.Fatalf("expected Bob balance 4000, got %d", bobUser.BalanceCents)
	}
	bobTxs, _ := s.ListTransactions(bob.ID, 10)
	if len(bobTxs) != 1 || bobTxs[0].Type != models.TxTransferIn || bobTxs[0].AmountCents != 4000 {
		t.Fatalf("unexpected Bob txs: %+v", bobTxs)
	}

	// Transfer using email
	_, _, err = s.Transfer(alice.ID, "bob@example.com", 2000, "Second transfer")
	if err != nil {
		t.Fatalf("transfer by email: %v", err)
	}
	aliceUser, _ := s.GetByID(alice.ID)
	if aliceUser.BalanceCents != 4000 {
		t.Fatalf("expected Alice balance 4000, got %d", aliceUser.BalanceCents)
	}

	// Self-transfer should fail
	if _, _, err := s.Transfer(alice.ID, "MIE1111111111", 1000, ""); err != ErrSelfTransfer {
		t.Fatalf("expected ErrSelfTransfer, got %v", err)
	}

	// Non-existent recipient
	if _, _, err := s.Transfer(alice.ID, "MIE9999999999", 1000, ""); err != ErrRecipientNotFound {
		t.Fatalf("expected ErrRecipientNotFound, got %v", err)
	}

	// Insufficient funds
	if _, _, err := s.Transfer(alice.ID, "MIE2222222222", 999999, ""); err != ErrInsufficientFunds {
		t.Fatalf("expected ErrInsufficientFunds, got %v", err)
	}
}

func TestPayBillAndDonate(t *testing.T) {
	s := newTestStore(t)
	u, err := s.CreateUser("Charlie", "charlie@example.com", "hash", "MIE3333333333")
	if err != nil {
		t.Fatal(err)
	}
	s.ApplyTransaction(u.ID, models.TxDeposit, 15000)

	// Pay bill
	txBill, err := s.PayBill(u.ID, "Electric Grid", "INV-101", 5000)
	if err != nil {
		t.Fatalf("pay bill: %v", err)
	}
	if txBill.Type != models.TxBillPayment || txBill.BalanceCents != 10000 {
		t.Fatalf("bill tx: %+v", txBill)
	}
	if !strings.Contains(txBill.Description, "Electric Grid") || !strings.Contains(txBill.Description, "INV-101") {
		t.Fatalf("bill description: %s", txBill.Description)
	}

	// Donate
	txDonation, err := s.Donate(u.ID, "Red Cross", "Emergency Relief", 3000)
	if err != nil {
		t.Fatalf("donate: %v", err)
	}
	if txDonation.Type != models.TxDonation || txDonation.BalanceCents != 7000 {
		t.Fatalf("donation tx: %+v", txDonation)
	}
	if !strings.Contains(txDonation.Description, "Red Cross") || !strings.Contains(txDonation.Description, "Emergency Relief") {
		t.Fatalf("donation description: %s", txDonation.Description)
	}

	// Verify transactions history
	txs, err := s.ListTransactions(u.ID, 10)
	if err != nil || len(txs) != 3 {
		t.Fatalf("expected 3 txs, got %d: %v", len(txs), err)
	}
}

func TestAccountManagement(t *testing.T) {
	s := newTestStore(t)
	u, err := s.CreateUser("David", "david@example.com", "hash_old", "MIE4444444444")
	if err != nil {
		t.Fatal(err)
	}

	// Update name
	if err := s.UpdateUserName(u.ID, "David Updated"); err != nil {
		t.Fatalf("update name: %v", err)
	}
	got, _ := s.GetByID(u.ID)
	if got.Name != "David Updated" {
		t.Fatalf("expected David Updated, got %s", got.Name)
	}

	// Update password
	if err := s.UpdateUserPassword(u.ID, "hash_new"); err != nil {
		t.Fatalf("update password: %v", err)
	}
	got, _ = s.GetByID(u.ID)
	if got.PasswordHash != "hash_new" {
		t.Fatalf("expected hash_new, got %s", got.PasswordHash)
	}
}

func TestCardAndIBANStore(t *testing.T) {
	s := newTestStore(t)
	cardNumber := "5342000000005555"
	cardExpiry := "12/30"
	iban := "EG96MIEB0000000055555555"
	u, err := s.CreateUser("Frank", "frank@example.com", "hash", "MIE5555555555", cardNumber, cardExpiry, iban)
	if err != nil {
		t.Fatal(err)
	}

	if u.CardNumber != cardNumber {
		t.Fatalf("expected card number %q, got %q", cardNumber, u.CardNumber)
	}
	if u.CardExpiry != cardExpiry {
		t.Fatalf("expected card expiry %q, got %q", cardExpiry, u.CardExpiry)
	}
	if u.IBAN != iban {
		t.Fatalf("expected IBAN %q, got %q", iban, u.IBAN)
	}

	// Lookup by card number
	byCard, err := s.GetByCardNumber(u.CardNumber)
	if err != nil || byCard.ID != u.ID {
		t.Fatalf("GetByCardNumber: %v %v", err, byCard)
	}

	// Lookup by card with spaces
	spaced := u.CardNumber[0:4] + " " + u.CardNumber[4:8] + " " + u.CardNumber[8:12] + " " + u.CardNumber[12:16]
	byCardSpaced, err := s.GetByCardNumber(spaced)
	if err != nil || byCardSpaced.ID != u.ID {
		t.Fatalf("GetByCardNumber spaced: %v %v", err, byCardSpaced)
	}

	// Lookup by IBAN
	byIBAN, err := s.GetByIBAN(u.IBAN)
	if err != nil || byIBAN.ID != u.ID {
		t.Fatalf("GetByIBAN: %v %v", err, byIBAN)
	}

	// Lookup by login identifier with IBAN
	byLoginIBAN, err := s.GetByLoginIdentifier(u.IBAN)
	if err != nil || byLoginIBAN.ID != u.ID {
		t.Fatalf("GetByLoginIdentifier IBAN: %v %v", err, byLoginIBAN)
	}

	// Lookup by login identifier with card number
	byLoginCard, err := s.GetByLoginIdentifier(spaced)
	if err != nil || byLoginCard.ID != u.ID {
		t.Fatalf("GetByLoginIdentifier Card: %v %v", err, byLoginCard)
	}
}
