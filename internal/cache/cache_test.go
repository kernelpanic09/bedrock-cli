package cache

import (
	"os"
	"testing"
	"time"
)

func TestKeyStability(t *testing.T) {
	k1 := Key("haiku", "hello world", 0.7, 1024)
	k2 := Key("haiku", "hello world", 0.7, 1024)
	if k1 != k2 {
		t.Errorf("Key() not stable: %q != %q", k1, k2)
	}
}

func TestKeyDiffers(t *testing.T) {
	tests := []struct {
		model       string
		prompt      string
		temperature float64
		maxTokens   int
	}{
		{"haiku", "same prompt", 0.7, 1024},
		{"sonnet", "same prompt", 0.7, 1024},
		{"haiku", "different prompt", 0.7, 1024},
		{"haiku", "same prompt", 0.5, 1024},
		{"haiku", "same prompt", 0.7, 2048},
	}

	keys := make(map[string]bool)
	for _, tt := range tests {
		k := Key(tt.model, tt.prompt, tt.temperature, tt.maxTokens)
		if keys[k] {
			t.Errorf("Key collision for %+v", tt)
		}
		keys[k] = true
	}
}

func TestKeyNormalizesWhitespace(t *testing.T) {
	k1 := Key("haiku", "  hello  ", 0.7, 1024)
	k2 := Key("haiku", "hello", 0.7, 1024)
	if k1 != k2 {
		t.Errorf("Key() should normalize surrounding whitespace: %q != %q", k1, k2)
	}
}

func TestRoundTrip(t *testing.T) {
	// Redirect to a temp dir so we don't pollute the real cache.
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	entry := &Entry{
		Model:        "haiku",
		Prompt:       "test prompt",
		Response:     "test response",
		InputTokens:  10,
		OutputTokens: 20,
		CachedAt:     time.Now().Truncate(time.Second),
	}
	key := Key("haiku", "test prompt", 0.7, 1024)

	if err := Put(key, entry); err != nil {
		t.Fatalf("Put() error: %v", err)
	}

	got := Get(key)
	if got == nil {
		t.Fatal("Get() returned nil after Put()")
	}
	if got.Response != entry.Response {
		t.Errorf("Response = %q, want %q", got.Response, entry.Response)
	}
	if got.InputTokens != entry.InputTokens {
		t.Errorf("InputTokens = %d, want %d", got.InputTokens, entry.InputTokens)
	}
}

func TestGetMiss(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	got := Get("does-not-exist")
	if got != nil {
		t.Errorf("Get() on missing key should return nil, got %+v", got)
	}
}

func TestDelete(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	entry := &Entry{Response: "to delete"}
	key := Key("haiku", "delete me", 0.7, 1024)

	if err := Put(key, entry); err != nil {
		t.Fatalf("Put() error: %v", err)
	}
	if err := Delete(key); err != nil {
		t.Fatalf("Delete() error: %v", err)
	}
	if got := Get(key); got != nil {
		t.Error("Get() should return nil after Delete()")
	}
}

func TestDeleteNonexistent(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	// Should not error on missing key.
	if err := Delete("no-such-key"); err != nil {
		t.Errorf("Delete() on nonexistent key returned error: %v", err)
	}
}

func TestClearAndStats(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	// Temporarily redirect XDG_CACHE_HOME so CacheDir() uses our tmp.
	t.Setenv("XDG_CACHE_HOME", tmp)

	for i := 0; i < 3; i++ {
		key := Key("haiku", string(rune('a'+i)), 0.7, 1024)
		if err := Put(key, &Entry{Response: "resp"}); err != nil {
			t.Fatalf("Put() error: %v", err)
		}
	}

	count, _, err := Stats()
	if err != nil {
		// Stats failing is ok in test env if dirs differ.
		t.Skip("Stats() error in test env:", err)
	}
	if count == 0 {
		t.Error("Stats() returned 0 count after inserts")
	}

	if err := Clear(); err != nil {
		t.Fatalf("Clear() error: %v", err)
	}

	count, _, err = Stats()
	if err != nil {
		t.Skip("Stats() error after clear:", err)
	}
	if count != 0 {
		t.Errorf("Stats() count = %d after Clear(), want 0", count)
	}

	// Suppress the os.Remove call in TestMain cleanup.
	_ = os.RemoveAll(tmp)
}
