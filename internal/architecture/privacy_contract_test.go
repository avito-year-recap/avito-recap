package architecture

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestPublicRecapProtoDoesNotExposeActionableState(t *testing.T) {
	_, current, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("cannot resolve test path")
	}
	root := filepath.Clean(filepath.Join(filepath.Dir(current), "..", ".."))
	data, err := os.ReadFile(
		filepath.Join(root, "proto", "recap", "v1", "recap.proto"),
	)
	if err != nil {
		t.Fatal(err)
	}
	proto := string(data)
	for _, forbidden := range []string{
		"ActionableState actionable_state",
		"message ActionableState",
		"draft_listing_id",
		"open_dialog_id",
		"active_listing_id",
		"last_purchased_listing_id",
	} {
		if strings.Contains(proto, forbidden) {
			t.Fatalf("public proto exposes internal actionable-state detail %q", forbidden)
		}
	}
}
