package recap

import (
	"reflect"
	"testing"
)

func TestSeedCategoryCatalogueMatchesCode(t *testing.T) {
	var seeded []CategoryDefinition
	readJSONFile(t, projectFile(t, "seeds", "categories.json"), &seeded)
	if !reflect.DeepEqual(seeded, CategoryCatalogue()) {
		t.Fatalf("seed category catalogue differs from code:\nseed: %+v\ncode: %+v", seeded, CategoryCatalogue())
	}
}

func TestCategoryCatalogueContainsUniqueSafeCodes(t *testing.T) {
	seen := make(map[string]struct{})
	for _, category := range CategoryCatalogue() {
		if !isSafeCategoryCode(category.Code) || category.Title == "" {
			t.Fatalf("invalid category: %+v", category)
		}
		if _, exists := seen[category.Code]; exists {
			t.Fatalf("duplicate category code %q", category.Code)
		}
		seen[category.Code] = struct{}{}
	}
}
