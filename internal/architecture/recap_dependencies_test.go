package architecture_test

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

const modulePath = "github.com/year-recap"

func TestRecapDependencyRules(t *testing.T) {
	root := projectRoot(t)
	recapRoot := filepath.Join(root, "internal", "recap")

	entries, err := os.ReadDir(recapRoot)
	if err != nil {
		t.Fatalf("read recap root: %v", err)
	}
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".go") {
			t.Errorf("Go file must not live directly in internal/recap: %s", entry.Name())
		}
	}

	domainPackages := map[string]bool{
		"achievement":  true,
		"analytics":    true,
		"behavior":     true,
		"model":        true,
		"nextaction":   true,
		"presentation": true,
		"ruleset":      true,
		"validation":   true,
	}
	presentationForbidden := map[string]bool{
		"achievement": true,
		"analytics":   true,
		"behavior":    true,
	}

	err = filepath.WalkDir(recapRoot, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") || strings.HasSuffix(entry.Name(), "_test.go") {
			return nil
		}
		rel, err := filepath.Rel(recapRoot, path)
		if err != nil {
			return err
		}
		segments := strings.Split(filepath.ToSlash(rel), "/")
		owner := segments[0]

		parsed, err := parser.ParseFile(token.NewFileSet(), path, nil, parser.ImportsOnly)
		if err != nil {
			t.Errorf("parse %s: %v", rel, err)
			return nil
		}
		for _, spec := range parsed.Imports {
			importPath, err := strconv.Unquote(spec.Path.Value)
			if err != nil {
				t.Errorf("unquote import in %s: %v", rel, err)
				continue
			}
			if domainPackages[owner] && importPath == modulePath+"/internal/recap/application" {
				t.Errorf("domain package %s must not import application: %s", owner, rel)
			}
			if presentationForbidden[owner] && strings.HasPrefix(importPath, modulePath+"/internal/recap/presentation") {
				t.Errorf("%s must not import presentation: %s", owner, rel)
			}
			if strings.HasPrefix(importPath, modulePath+"/internal/storage") {
				t.Errorf("recap package must not import storage: %s imports %s", rel, importPath)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk recap packages: %v", err)
	}
}

func projectRoot(t *testing.T) string {
	t.Helper()
	current, err := os.Getwd()
	if err != nil {
		t.Fatalf("get working directory: %v", err)
	}
	for {
		if _, err := os.Stat(filepath.Join(current, "go.mod")); err == nil {
			return current
		}
		parent := filepath.Dir(current)
		if parent == current {
			t.Fatal("go.mod not found")
		}
		current = parent
	}
}
