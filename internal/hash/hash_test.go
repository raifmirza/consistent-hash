package hash

import "testing"

func TestHash_IsDeterministic(t *testing.T) {
	key := "user-42"

	first := Hash(key)
	second := Hash(key)

	if first != second {
		t.Fatalf("expected deterministic hash, got %d and %d", first, second)
	}
}

func TestHash_DiffersForDifferentKeys(t *testing.T) {
	if Hash("user-42") == Hash("user-99") {
		t.Fatal("expected different keys to produce different hashes")
	}
}
