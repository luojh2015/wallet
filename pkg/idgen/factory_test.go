package idgen

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// mockIDGenerator implements IDGenerator for testing
type mockIDGenerator struct {
	counter   atomic.Int64
	errToReturn error
}

func newMockIDGenerator() *mockIDGenerator {
	return &mockIDGenerator{}
}

func (m *mockIDGenerator) Generate() (int64, error) {
	if m.errToReturn != nil {
		return 0, m.errToReturn
	}
	return m.counter.Add(1), nil
}

func (m *mockIDGenerator) GenerateString() (string, error) {
	if m.errToReturn != nil {
		return "", m.errToReturn
	}
	id := m.counter.Add(1)
	return fmt.Sprintf("%d", id), nil
}

func (m *mockIDGenerator) Clone() IDGenerator {
	c := newMockIDGenerator()
	c.errToReturn = m.errToReturn
	return c
}

func (m *mockIDGenerator) Parse(id int64) ParsedID {
	return ParsedID{Timestamp: time.Now(), MachineID: 1, Sequence: id}
}

func TestGenWalletID(t *testing.T) {
	gen := newMockIDGenerator()
	factory := NewIDFactory(gen)

	id, err := factory.GenWalletID()
	if err != nil {
		t.Fatalf("GenWalletID() error = %v", err)
	}
	if !strings.HasPrefix(id, "W") {
		t.Errorf("GenWalletID() = %q, want prefix 'W'", id)
	}
}

func TestGenTransactionID(t *testing.T) {
	gen := newMockIDGenerator()
	factory := NewIDFactory(gen)

	id, err := factory.GenTransactionID()
	if err != nil {
		t.Fatalf("GenTransactionID() error = %v", err)
	}
	if !strings.HasPrefix(id, "T") {
		t.Errorf("GenTransactionID() = %q, want prefix 'T'", id)
	}
}

func TestGenSessionID(t *testing.T) {
	gen := newMockIDGenerator()
	factory := NewIDFactory(gen)

	id, err := factory.GenSessionID()
	if err != nil {
		t.Fatalf("GenSessionID() error = %v", err)
	}
	if !strings.HasPrefix(id, "S") {
		t.Errorf("GenSessionID() = %q, want prefix 'S'", id)
	}
}

func TestGenWalletID_Error(t *testing.T) {
	gen := newMockIDGenerator()
	gen.errToReturn = errors.New("id gen failure")
	factory := NewIDFactory(gen)

	_, err := factory.GenWalletID()
	if err == nil {
		t.Fatal("GenWalletID() expected error, got nil")
	}
}

func TestGenTransactionID_Error(t *testing.T) {
	gen := newMockIDGenerator()
	gen.errToReturn = errors.New("id gen failure")
	factory := NewIDFactory(gen)

	_, err := factory.GenTransactionID()
	if err == nil {
		t.Fatal("GenTransactionID() expected error, got nil")
	}
}

func TestGenSessionID_Error(t *testing.T) {
	gen := newMockIDGenerator()
	gen.errToReturn = errors.New("id gen failure")
	factory := NewIDFactory(gen)

	_, err := factory.GenSessionID()
	if err == nil {
		t.Fatal("GenSessionID() expected error, got nil")
	}
}

func TestIDFactory_UniqueIDs(t *testing.T) {
	gen := newMockIDGenerator()
	factory := NewIDFactory(gen)

	ids := make(map[string]bool)
	for i := 0; i < 100; i++ {
		id, err := factory.GenWalletID()
		if err != nil {
			t.Fatalf("GenWalletID() error = %v", err)
		}
		if ids[id] {
			t.Fatalf("duplicate wallet ID: %s", id)
		}
		ids[id] = true
	}
}

func TestIDFactory_ConcurrentGeneration(t *testing.T) {
	gen := newMockIDGenerator()
	factory := NewIDFactory(gen)

	var mu sync.Mutex
	ids := make(map[string]bool)
	var wg sync.WaitGroup

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				id, err := factory.GenWalletID()
				if err != nil {
					t.Errorf("GenWalletID() error = %v", err)
					return
				}
				mu.Lock()
				if ids[id] {
					t.Errorf("duplicate ID: %s", id)
				}
				ids[id] = true
				mu.Unlock()
			}
		}()
	}
	wg.Wait()
}
