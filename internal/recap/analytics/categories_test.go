package analytics_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"

	"github.com/year-recap/internal/recap/analytics"
)

func TestSeedCategoryCatalogueMatchesCode(t *testing.T) {
	var seeded []analytics.CategoryDefinition
	data, err := os.ReadFile(projectFile(t, "seeds", "categories.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, &seeded); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(seeded, analytics.CategoryCatalogue()) {
		t.Fatalf("seed category catalogue differs from code:\nseed: %+v\ncode: %+v", seeded, analytics.CategoryCatalogue())
	}
}

func TestCategoryCatalogueContainsUniqueSafeCodes(t *testing.T) {
	seen := make(map[string]struct{})
	for _, category := range analytics.CategoryCatalogue() {
		if category.Code == "" || category.Title == "" {
			t.Fatalf("invalid category: %+v", category)
		}
		for _, r := range category.Code {
			if !((r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '_' || r == '-') {
				t.Fatalf("unsafe category code %q", category.Code)
			}
		}
		if _, exists := seen[category.Code]; exists {
			t.Fatalf("duplicate category code %q", category.Code)
		}
		seen[category.Code] = struct{}{}
	}
}

func projectFile(t *testing.T, parts ...string) string {
	t.Helper()
	_, current, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("cannot resolve test file path")
	}
	root := filepath.Clean(filepath.Join(filepath.Dir(current), "..", "..", ".."))
	return filepath.Join(append([]string{root}, parts...)...)
}
