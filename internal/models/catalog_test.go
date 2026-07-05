package models

import "testing"

func TestResolveByAlias(t *testing.T) {
	m, err := Resolve("haiku")
	if err != nil {
		t.Fatalf("Resolve(%q) error: %v", "haiku", err)
	}
	if m.ID != "anthropic.claude-haiku-4-5-20251001-v1:0" {
		t.Errorf("ID = %q, want %q", m.ID, "anthropic.claude-haiku-4-5-20251001-v1:0")
	}
	if m.Provider != "Anthropic" {
		t.Errorf("Provider = %q, want %q", m.Provider, "Anthropic")
	}
}

func TestResolveByID(t *testing.T) {
	id := "meta.llama3-70b-instruct-v1:0"
	m, err := Resolve(id)
	if err != nil {
		t.Fatalf("Resolve(%q) error: %v", id, err)
	}
	if m.Alias != "llama-3-70b" {
		t.Errorf("Alias = %q, want %q", m.Alias, "llama-3-70b")
	}
}

func TestResolveUnknown(t *testing.T) {
	if _, err := Resolve("does-not-exist"); err == nil {
		t.Error("Resolve() should error on an unknown model")
	}
}

func TestAllReturnsDefensiveCopy(t *testing.T) {
	got := All()
	if len(got) == 0 {
		t.Fatal("All() returned no models")
	}

	// Mutating the returned slice must not corrupt the underlying catalog.
	original := got[0].Alias
	got[0].Alias = "tampered"

	again := All()
	if again[0].Alias != original {
		t.Errorf("All() leaked internal state: got %q after mutation, want %q", again[0].Alias, original)
	}
}

func TestByProviderPartitionsCatalog(t *testing.T) {
	groups := ByProvider()
	if len(groups) == 0 {
		t.Fatal("ByProvider() returned no groups")
	}

	// Every model must appear in exactly one provider bucket, so the grouped
	// counts have to sum back to the full catalog size.
	total := 0
	for provider, list := range groups {
		if len(list) == 0 {
			t.Errorf("provider %q has an empty group", provider)
		}
		for _, m := range list {
			if m.Provider != provider {
				t.Errorf("model %q grouped under %q, but its Provider is %q", m.ID, provider, m.Provider)
			}
		}
		total += len(list)
	}

	if total != len(All()) {
		t.Errorf("grouped model count = %d, want %d", total, len(All()))
	}
}
