package snowflake

import (
	"testing"
	"time"
)

func TestSnowflakeGenerator_Generate(t *testing.T) {
	gen, err := NewSnowflakeGenerator(1)
	if err != nil {
		t.Fatalf("Failed to create generator: %v", err)
	}

	// 生成多个 ID，确保它们都是唯一的
	ids := make(map[int64]bool)
	for i := 0; i < 10000; i++ {
		id, err := gen.Generate()
		if err != nil {
			t.Fatalf("Failed to generate ID: %v", err)
		}

		if ids[id] {
			t.Fatalf("Duplicate ID generated: %d", id)
		}
		ids[id] = true
	}
}

func TestSnowflakeGenerator_GenerateString(t *testing.T) {
	gen, err := NewSnowflakeGenerator(1)
	if err != nil {
		t.Fatalf("Failed to create generator: %v", err)
	}

	id, err := gen.GenerateString()
	if err != nil {
		t.Fatalf("Failed to generate string ID: %v", err)
	}

	if id == "" {
		t.Error("Generated ID is empty")
	}

	t.Logf("Generated ID: %s", id)
}

func TestSnowflakeGenerator_Parse(t *testing.T) {
	gen, err := NewSnowflakeGenerator(1)
	if err != nil {
		t.Fatalf("Failed to create generator: %v", err)
	}

	id, err := gen.Generate()
	if err != nil {
		t.Fatalf("Failed to generate ID: %v", err)
	}

	parsed := gen.Parse(id)

	if parsed.MachineID != 1 {
		t.Errorf("Expected machine ID 1, got %d", parsed.MachineID)
	}

	if parsed.Timestamp.After(time.Now()) {
		t.Error("Parsed timestamp is in the future")
	}

	if parsed.Sequence < 0 {
		t.Error("Sequence is negative")
	}
}

func TestSnowflakeGenerator_InvalidMachineID(t *testing.T) {
	_, err := NewSnowflakeGenerator(-1)
	if err != ErrInvalidMachineID {
		t.Error("Expected error for negative machine ID")
	}

	_, err = NewSnowflakeGenerator(64)
	if err != ErrInvalidMachineID {
		t.Error("Expected error for machine ID > 63")
	}
}

func TestSnowflakeGenerator_Concurrent(t *testing.T) {
	gen, err := NewSnowflakeGenerator(1)
	if err != nil {
		t.Fatalf("Failed to create generator: %v", err)
	}

	const numGoroutines = 100
	const numIDsPerGoroutine = 1000

	ids := make(chan int64, numGoroutines*numIDsPerGoroutine)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			for j := 0; j < numIDsPerGoroutine; j++ {
				id, err := gen.Generate()
				if err != nil {
					t.Errorf("Failed to generate ID: %v", err)
					return
				}
				ids <- id
			}
		}()
	}

	uniqueIDs := make(map[int64]bool)
	for i := 0; i < numGoroutines*numIDsPerGoroutine; i++ {
		id := <-ids
		if uniqueIDs[id] {
			t.Errorf("Duplicate ID: %d", id)
		}
		uniqueIDs[id] = true
	}

	t.Logf("Generated %d unique IDs concurrently", len(uniqueIDs))
}

func TestSnowflakeGenerator_MultipleMachines(t *testing.T) {
	gen1, _ := NewSnowflakeGenerator(1)
	gen2, _ := NewSnowflakeGenerator(2)

	id1, _ := gen1.Generate()
	id2, _ := gen2.Generate()

	parsed1 := gen1.Parse(id1)
	parsed2 := gen2.Parse(id2)

	if parsed1.MachineID == parsed2.MachineID {
		t.Error("Different machines should have different machine IDs")
	}

	t.Logf("Machine 1 ID: %d, Machine 2 ID: %d", parsed1.MachineID, parsed2.MachineID)
}
