package store

import (
	"testing")

func TestKVStoreBasicOperations(t *testing.T) {
	store := NewKVStore()

	// Test Set and Get
	store.Set("name", "alice")
	val, exists := store.Get("name")

	if !exists {
		t.Errorf("Get() key doesn't exist")
	}

	if val != "alice" {
		t.Errorf("Get() = %v, want %v", val, "alice")
	}

	// Test Delete
	store.Delete("name")
	_, exists = store.Get("name")

	if exists {
		t.Errorf("Delete() failed: key still exists")
	}

	// Test non-existent key
	_, exists = store.Get("nonexistent")
	if exists {
		t.Errorf("Expected non-existent key to return false")
	}
}
