package generator

import (
	"testing"

	"github.com/sarathsp06/proto2astro/internal/config"
	"github.com/sarathsp06/proto2astro/internal/parser"
)

func TestCleanDescription(t *testing.T) {
	tests := []struct {
		name string
		desc string
		want string
	}{
		{"strips Required. prefix", "Required. The URL.", "The URL."},
		{"strips Required prefix", "Required The URL.", "The URL."},
		{"no Required", "The webhook URL.", "The webhook URL."},
		{"empty", "", ""},
		{"only whitespace", "   ", ""},
		// Annotations ARE now stripped after Phase 2 fix.
		{
			"strips @example from description",
			`The item name. @example "test"`,
			"The item name.",
		},
		{
			"strips Default: from description",
			"Max retries. Default: 5.",
			"Max retries.",
		},
		{
			"strips Range: from description",
			"Page size. Range: 1-100.",
			"Page size.",
		},
		// @-prefix annotations
		{
			"strips @required from description",
			"@required The item ID to update.",
			"The item ID to update.",
		},
		{
			"strips @default from description",
			"Name for the item. @default Untitled",
			"Name for the item.",
		},
		{
			"strips @range from description",
			"Count value. @range 1-200",
			"Count value.",
		},
		{
			"strips @deprecated from description",
			"@deprecated Use new_field instead.",
			"",
		},
		{
			"strips multiple @-prefix annotations",
			"@required Count. @range 1-100 @default 50 @example 25",
			"Count.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := cleanDescription(tt.desc)
			if got != tt.want {
				t.Errorf("cleanDescription(%q) = %q, want %q", tt.desc, got, tt.want)
			}
		})
	}
}

func TestShouldFlatten(t *testing.T) {
	entityTypes := []string{"ItemDetail", "UserProfile"}

	tests := []struct {
		name        string
		typeName    string
		entityTypes []string
		want        bool
	}{
		{"non-entity type", "CreateItemRequest", entityTypes, true},
		{"entity type", "ItemDetail", entityTypes, false},
		{"another entity", "UserProfile", entityTypes, false},
		{"empty entity list", "ItemDetail", nil, true},
		{"empty type name", "", entityTypes, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := shouldFlatten(tt.typeName, tt.entityTypes)
			if got != tt.want {
				t.Errorf("shouldFlatten(%q, %v) = %v, want %v",
					tt.typeName, tt.entityTypes, got, tt.want)
			}
		})
	}
}

func TestToKebab(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"PascalCase", "ItemService", "item-service"},
		{"single word", "Item", "item"},
		{"already lowercase", "item", "item"},
		{"multiple caps", "HTTPService", "h-t-t-p-service"},
		{"empty", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := toKebab(tt.in)
			if got != tt.want {
				t.Errorf("toKebab(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestMessageUsesEnum(t *testing.T) {
	// Build a package with messages for recursive lookup
	pkg := &parser.ProtoPackage{
		Name: "testpkg.v1",
		Messages: map[string]*parser.ProtoMessage{
			"CreateItemRequest": {
				Name: "CreateItemRequest",
				Fields: []parser.ProtoField{
					{Name: "name", Type: "string", RawType: "string"},
					{Name: "status", Type: "ItemStatus", RawType: "ItemStatus", IsEnum: true},
				},
			},
			"ItemDetail": {
				Name: "ItemDetail",
				Fields: []parser.ProtoField{
					{Name: "id", Type: "string", RawType: "string"},
					{Name: "status", Type: "ItemStatus", RawType: "ItemStatus", IsEnum: true},
				},
			},
			"GetItemResponse": {
				Name: "GetItemResponse",
				Fields: []parser.ProtoField{
					{Name: "item", Type: "ItemDetail", RawType: "ItemDetail", IsMessage: true},
				},
			},
		},
	}

	// Direct enum usage — should be found
	if !messageUsesEnum(pkg.Messages["CreateItemRequest"], "ItemStatus", pkg, nil) {
		t.Error("messageUsesEnum should find direct enum usage")
	}
	if messageUsesEnum(pkg.Messages["CreateItemRequest"], "Priority", pkg, nil) {
		t.Error("messageUsesEnum should not find unused enum")
	}

	// Nested enum usage — NOW found after Phase 2 recursive fix
	if !messageUsesEnum(pkg.Messages["GetItemResponse"], "ItemStatus", pkg, nil) {
		t.Error("messageUsesEnum should find nested enum usage (recursive)")
	}
}

func TestFindEnumUsage(t *testing.T) {
	pkg := &parser.ProtoPackage{
		Name: "testpkg.v1",
		Services: map[string]*parser.ProtoService{
			"ItemService": {
				Name: "ItemService",
				RPCs: []parser.ProtoRPC{
					{
						Name:         "CreateItem",
						RequestType:  "CreateItemRequest",
						ResponseType: "CreateItemResponse",
					},
				},
			},
		},
		Messages: map[string]*parser.ProtoMessage{
			"CreateItemRequest": {
				Name: "CreateItemRequest",
				Fields: []parser.ProtoField{
					{Name: "priority", Type: "Priority", RawType: "Priority", IsEnum: true},
				},
			},
			"CreateItemResponse": {
				Name: "CreateItemResponse",
				Fields: []parser.ProtoField{
					{Name: "item", Type: "ItemDetail", RawType: "ItemDetail", IsMessage: true},
				},
			},
			"ItemDetail": {
				Name: "ItemDetail",
				Fields: []parser.ProtoField{
					{Name: "status", Type: "ItemStatus", RawType: "ItemStatus", IsEnum: true},
				},
			},
		},
		Enums: map[string]*parser.ProtoEnum{
			"Priority":   {Name: "Priority"},
			"ItemStatus": {Name: "ItemStatus"},
		},
	}

	// Priority is used directly in CreateItemRequest — should be found
	priorityUsage := findEnumUsage("Priority", pkg)
	if len(priorityUsage) != 1 || priorityUsage[0] != "ItemService" {
		t.Errorf("findEnumUsage(Priority) = %v, want [ItemService]", priorityUsage)
	}

	// ItemStatus is used in ItemDetail (nested in CreateItemResponse) —
	// NOW found after Phase 2 recursive messageUsesEnum fix.
	statusUsage := findEnumUsage("ItemStatus", pkg)
	if len(statusUsage) != 1 || statusUsage[0] != "ItemService" {
		t.Errorf("findEnumUsage(ItemStatus) = %v, want [ItemService]", statusUsage)
	}
}

func TestSortedServiceNames(t *testing.T) {
	pkg := &parser.ProtoPackage{
		Services: map[string]*parser.ProtoService{
			"Zebra": {Name: "Zebra"},
			"Alpha": {Name: "Alpha"},
			"Bravo": {Name: "Bravo"},
		},
	}

	// No ordering specified — alphabetical
	got := sortedServiceNames(pkg, nil)
	want := []string{"Alpha", "Bravo", "Zebra"}
	if len(got) != len(want) {
		t.Fatalf("sortedServiceNames() len = %d, want %d", len(got), len(want))
	}
	for i := range got {
		if got[i] != want[i] {
			t.Errorf("sortedServiceNames()[%d] = %q, want %q", i, got[i], want[i])
		}
	}

	// With explicit ordering
	got = sortedServiceNames(pkg, []string{"Bravo", "Zebra"})
	want = []string{"Bravo", "Zebra", "Alpha"}
	if len(got) != len(want) {
		t.Fatalf("sortedServiceNames(ordered) len = %d, want %d", len(got), len(want))
	}
	for i := range got {
		if got[i] != want[i] {
			t.Errorf("sortedServiceNames(ordered)[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestFormatPackageLabel(t *testing.T) {
	tests := []struct {
		name string
		pkg  string
		want string
	}{
		{"default package", "_default", "API Reference"},
		{"dotted package", "webhook.v1", "Webhook V1"},
		{"single segment", "mypackage", "Mypackage"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatPackageLabel(tt.pkg)
			if got != tt.want {
				t.Errorf("formatPackageLabel(%q) = %q, want %q", tt.pkg, got, tt.want)
			}
		})
	}
}

func TestBuildFlatField_EntityExamples(t *testing.T) {
	entityExamples := map[string]any{
		"ItemDetail": map[string]any{
			"id":     "item-abc",
			"status": "active",
		},
	}

	tests := []struct {
		name           string
		field          parser.ProtoField
		entityExamples map[string]any
		wantExample    any
	}{
		{
			name: "entity example used as fallback",
			field: parser.ProtoField{
				Name:      "item",
				Type:      "ItemDetail",
				RawType:   "ItemDetail",
				IsMessage: true,
			},
			entityExamples: entityExamples,
			wantExample:    entityExamples["ItemDetail"],
		},
		{
			name: "proto example takes precedence over entity example",
			field: parser.ProtoField{
				Name:      "item",
				Type:      "ItemDetail",
				RawType:   "ItemDetail",
				IsMessage: true,
				Example:   "custom-example",
			},
			entityExamples: entityExamples,
			wantExample:    "custom-example",
		},
		{
			name: "no entity example for unknown type",
			field: parser.ProtoField{
				Name:      "user",
				Type:      "UserProfile",
				RawType:   "UserProfile",
				IsMessage: true,
			},
			entityExamples: entityExamples,
			wantExample:    nil,
		},
		{
			name: "non-message field ignores entity examples",
			field: parser.ProtoField{
				Name:    "name",
				Type:    "string",
				RawType: "string",
			},
			entityExamples: entityExamples,
			wantExample:    nil,
		},
		{
			name: "nil entity examples map",
			field: parser.ProtoField{
				Name:      "item",
				Type:      "ItemDetail",
				RawType:   "ItemDetail",
				IsMessage: true,
			},
			entityExamples: nil,
			wantExample:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ov := config.OverlayField{}
			got := buildFlatField(tt.field, tt.field.Name, ov, tt.entityExamples)
			if tt.wantExample == nil && got.Example != nil {
				t.Errorf("expected nil example, got %v", got.Example)
			} else if tt.wantExample != nil && got.Example == nil {
				t.Errorf("expected example %v, got nil", tt.wantExample)
			}
			// For non-nil cases, verify the example is the expected value
			if tt.wantExample != nil && got.Example != nil {
				// Simple string comparison for the proto-example-takes-precedence case
				if s, ok := tt.wantExample.(string); ok {
					if gs, ok := got.Example.(string); ok && gs != s {
						t.Errorf("expected example %q, got %q", s, gs)
					}
				}
			}
		})
	}
}

func TestBuildFlatField_OverlayExampleTakesPrecedence(t *testing.T) {
	entityExamples := map[string]any{
		"ItemDetail": map[string]any{"id": "fallback"},
	}
	trueBool := true
	ov := config.OverlayField{
		Example:  "overlay-value",
		Required: &trueBool,
	}
	field := parser.ProtoField{
		Name:      "item",
		Type:      "ItemDetail",
		RawType:   "ItemDetail",
		IsMessage: true,
	}

	got := buildFlatField(field, "item", ov, entityExamples)
	if got.Example != "overlay-value" {
		t.Errorf("overlay example should take precedence, got %v", got.Example)
	}
	if !got.Required {
		t.Error("overlay required should be applied")
	}
}
