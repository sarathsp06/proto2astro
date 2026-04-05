package generator

import (
	"flag"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

var update = flag.Bool("update", false, "update golden files")

// testdataDir returns the absolute path to the project's testdata directory.
func testdataDir(t *testing.T) string {
	t.Helper()
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("cannot determine test file location")
	}
	// emit_test.go is in internal/generator/ — testdata is at project root
	return filepath.Join(filepath.Dir(filename), "..", "..", "testdata")
}

func goldenDir(t *testing.T) string {
	t.Helper()
	return filepath.Join(testdataDir(t), "golden")
}

func TestMarshalServiceTS_Golden(t *testing.T) {
	svc := TSService{
		Service:     "ItemService",
		Package:     "testpkg.v1",
		Description: "ItemService manages items.",
		RPCs: []TSRPC{
			{
				Name:        "CreateItem",
				Description: "CreateItem creates a new item.",
				Request: []TSField{
					{
						Name:        "name",
						Type:        "string",
						Required:    true,
						Description: "The name of the item.",
						Example:     "My Item",
					},
					{
						Name:        "count",
						Type:        "int32",
						Description: "A description with HTML chars: values < 100 & > 0.",
						Example:     float64(50),
					},
					{
						Name:        "priority",
						Type:        "Priority",
						Description: "Priority level for the item.",
					},
					{
						Name:        "tags",
						Type:        "string[]",
						Description: "Optional tags for the item.",
						Example:     []any{"alpha", "beta"},
					},
				},
				Response: []TSField{
					{
						Name:        "item",
						Type:        "ItemDetail",
						Description: "The created item.",
					},
				},
				Errors: []TSErrorCode{
					{Code: "ALREADY_EXISTS", Description: "An item with the same name exists."},
					{Code: "INVALID_ARGUMENT", Description: "The name is empty."},
				},
			},
		},
	}

	got, err := marshalServiceTS(svc)
	if err != nil {
		t.Fatalf("marshalServiceTS() error = %v", err)
	}

	goldenFile := filepath.Join(goldenDir(t), "item-service.ts")

	if *update {
		if err := os.MkdirAll(goldenDir(t), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(goldenFile, []byte(got), 0o644); err != nil {
			t.Fatalf("write golden: %v", err)
		}
		t.Log("updated golden file:", goldenFile)
		return
	}

	want, err := os.ReadFile(goldenFile)
	if err != nil {
		t.Fatalf("read golden file %s: %v\n  (run with -update to create)", goldenFile, err)
	}

	if got != string(want) {
		t.Errorf("marshalServiceTS() output differs from golden file %s\n--- got ---\n%s\n--- want ---\n%s",
			goldenFile, got, string(want))
	}
}

func TestMarshalEnumTS_Golden(t *testing.T) {
	enum := TSEnum{
		Name:        "ItemStatus",
		Package:     "testpkg.v1",
		Description: "ItemStatus represents the status of an item.",
		Values: []TSEnumValue{
			{Name: "ITEM_STATUS_UNSPECIFIED", Number: 0, Description: "Unspecified status."},
			{Name: "ITEM_STATUS_ACTIVE", Number: 1, Description: "Item is active."},
			{Name: "ITEM_STATUS_ARCHIVED", Number: 2, Description: "Item is archived."},
		},
		UsedBy: []string{"ItemService"},
	}

	got, err := marshalEnumTS(enum)
	if err != nil {
		t.Fatalf("marshalEnumTS() error = %v", err)
	}

	goldenFile := filepath.Join(goldenDir(t), "enum-item-status.ts")

	if *update {
		if err := os.MkdirAll(goldenDir(t), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(goldenFile, []byte(got), 0o644); err != nil {
			t.Fatalf("write golden: %v", err)
		}
		t.Log("updated golden file:", goldenFile)
		return
	}

	want, err := os.ReadFile(goldenFile)
	if err != nil {
		t.Fatalf("read golden file %s: %v\n  (run with -update to create)", goldenFile, err)
	}

	if got != string(want) {
		t.Errorf("marshalEnumTS() output differs from golden file %s\n--- got ---\n%s\n--- want ---\n%s",
			goldenFile, got, string(want))
	}
}

// TestMarshalServiceTS_HTMLChars verifies that HTML characters are preserved
// without escaping. Go's default json.Marshal escapes < > & but we use
// SetEscapeHTML(false) to prevent this.
func TestMarshalServiceTS_HTMLChars(t *testing.T) {
	svc := TSService{
		Service:     "TestService",
		Package:     "test",
		Description: "Values < 100 & > 0",
		RPCs:        nil,
	}

	got, err := marshalServiceTS(svc)
	if err != nil {
		t.Fatalf("marshalServiceTS() error = %v", err)
	}

	// After fix: HTML chars should NOT be escaped
	if strings.Contains(got, `\u003c`) {
		t.Error("should not contain \\u003c — HTML escaping should be disabled")
	}
	if strings.Contains(got, `\u003e`) {
		t.Error("should not contain \\u003e — HTML escaping should be disabled")
	}
	if strings.Contains(got, `\u0026`) {
		t.Error("should not contain \\u0026 — HTML escaping should be disabled")
	}
	// Should contain the literal characters
	if !strings.Contains(got, `< 100`) {
		t.Error("should contain literal < character")
	}
	if !strings.Contains(got, `> 0`) {
		t.Error("should contain literal > character")
	}
	if !strings.Contains(got, `&`) {
		t.Error("should contain literal & character")
	}
}
