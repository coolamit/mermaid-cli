package icons

import (
	"strings"
	"testing"
)

// --- ParseIconPacks ---

func TestParseIconPacks(t *testing.T) {
	packs := ParseIconPacks([]string{"@iconify-json/logos"})
	if len(packs) != 1 {
		t.Fatalf("expected 1 pack, got %d", len(packs))
	}
	if packs[0].Name != "logos" {
		t.Errorf("expected name %q, got %q", "logos", packs[0].Name)
	}
	want := "https://unpkg.com/@iconify-json/logos/icons.json"
	if packs[0].URL != want {
		t.Errorf("expected URL %q, got %q", want, packs[0].URL)
	}
}

func TestParseIconPacks_Multiple(t *testing.T) {
	packs := ParseIconPacks([]string{"@iconify-json/logos", "@iconify-json/mdi"})
	if len(packs) != 2 {
		t.Fatalf("expected 2 packs, got %d", len(packs))
	}
	if packs[0].Name != "logos" {
		t.Errorf("expected first name %q, got %q", "logos", packs[0].Name)
	}
	if packs[1].Name != "mdi" {
		t.Errorf("expected second name %q, got %q", "mdi", packs[1].Name)
	}
}

func TestParseIconPacks_Empty(t *testing.T) {
	packs := ParseIconPacks([]string{})
	if len(packs) != 0 {
		t.Errorf("expected 0 packs, got %d", len(packs))
	}
}

// --- ParseIconPacksNamesAndUrls ---

func TestParseIconPacksNamesAndUrls(t *testing.T) {
	packs := ParseIconPacksNamesAndUrls([]string{"myicons#https://example.com/icons.json"})
	if len(packs) != 1 {
		t.Fatalf("expected 1 pack, got %d", len(packs))
	}
	if packs[0].Name != "myicons" {
		t.Errorf("expected name %q, got %q", "myicons", packs[0].Name)
	}
	if packs[0].URL != "https://example.com/icons.json" {
		t.Errorf("expected URL %q, got %q", "https://example.com/icons.json", packs[0].URL)
	}
}

func TestParseIconPacksNamesAndUrls_NoHash(t *testing.T) {
	packs := ParseIconPacksNamesAndUrls([]string{"no-hash-here"})
	if len(packs) != 0 {
		t.Errorf("expected 0 packs for entries without #, got %d", len(packs))
	}
}

func TestParseIconPacksNamesAndUrls_Empty(t *testing.T) {
	packs := ParseIconPacksNamesAndUrls([]string{})
	if len(packs) != 0 {
		t.Errorf("expected 0 packs, got %d", len(packs))
	}
}

// --- GenerateIconPackJS ---

func TestGenerateIconPackJS_Empty(t *testing.T) {
	js := GenerateIconPackJS([]IconPack{})
	if js != "" {
		t.Errorf("expected empty string, got %q", js)
	}
}

func TestGenerateIconPackJS_Single(t *testing.T) {
	packs := []IconPack{{Name: "logos", URL: "https://example.com/logos.json"}}
	js := GenerateIconPackJS(packs)

	if !strings.Contains(js, "mermaid.registerIconPacks") {
		t.Error("expected output to contain mermaid.registerIconPacks")
	}
	if !strings.Contains(js, `"logos"`) {
		t.Error("expected output to contain pack name")
	}
	if !strings.Contains(js, `"https://example.com/logos.json"`) {
		t.Error("expected output to contain pack URL")
	}
}

func TestGenerateIconPackJS_Multiple(t *testing.T) {
	packs := []IconPack{
		{Name: "logos", URL: "https://example.com/logos.json"},
		{Name: "mdi", URL: "https://example.com/mdi.json"},
	}
	js := GenerateIconPackJS(packs)

	if !strings.Contains(js, `"logos"`) {
		t.Error("expected output to contain first pack name")
	}
	if !strings.Contains(js, `"mdi"`) {
		t.Error("expected output to contain second pack name")
	}
}
