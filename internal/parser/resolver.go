package parser

import "strings"

// resolveFieldTypes marks IsMessage and IsEnum on all fields in each package
// by looking up each field's RawType against the package's merged message
// and enum maps. Supports both short names ("Foo") and fully-qualified names
// ("pkg.Foo"). Also resolves cross-package references when the full result
// is available.
func resolveFieldTypes(pkg *ProtoPackage) {
	for _, msg := range pkg.Messages {
		for i := range msg.Fields {
			f := &msg.Fields[i]
			rawType := f.RawType

			// Check map value type for map fields
			if f.IsMap {
				rawType = f.MapValue
			}

			if resolveType(rawType, pkg) {
				if _, ok := pkg.Messages[rawType]; ok {
					f.IsMessage = true
				} else if _, ok := pkg.Messages[stripPackagePrefix(rawType)]; ok {
					f.IsMessage = true
				}
				if _, ok := pkg.Enums[rawType]; ok {
					f.IsEnum = true
				} else if _, ok := pkg.Enums[stripPackagePrefix(rawType)]; ok {
					f.IsEnum = true
				}
			}
		}
	}
}

// resolveAllFieldTypes resolves field types across all packages in a parse result.
// This enables cross-package references (e.g., package A referencing a message
// defined in package B).
func resolveAllFieldTypes(result *ParseResult) {
	// First pass: resolve within each package
	for _, pkg := range result.Packages {
		resolveFieldTypes(pkg)
	}

	// Second pass: resolve unresolved cross-package references
	allMessages := make(map[string]bool)
	allEnums := make(map[string]bool)
	for pkgName, pkg := range result.Packages {
		for msgName := range pkg.Messages {
			allMessages[pkgName+"."+msgName] = true
			allMessages[msgName] = true
		}
		for enumName := range pkg.Enums {
			allEnums[pkgName+"."+enumName] = true
			allEnums[enumName] = true
		}
	}

	for _, pkg := range result.Packages {
		for _, msg := range pkg.Messages {
			for i := range msg.Fields {
				f := &msg.Fields[i]
				if f.IsMessage || f.IsEnum {
					continue // already resolved
				}

				rawType := f.RawType
				if f.IsMap {
					rawType = f.MapValue
				}

				if allMessages[rawType] {
					f.IsMessage = true
				}
				if allEnums[rawType] {
					f.IsEnum = true
				}
			}
		}
	}
}

// resolveType checks if a type name matches any known message or enum in the package.
func resolveType(typeName string, pkg *ProtoPackage) bool {
	if _, ok := pkg.Messages[typeName]; ok {
		return true
	}
	if _, ok := pkg.Enums[typeName]; ok {
		return true
	}
	// Try stripping package prefix (e.g., "pkg.Foo" -> "Foo")
	short := stripPackagePrefix(typeName)
	if short != typeName {
		if _, ok := pkg.Messages[short]; ok {
			return true
		}
		if _, ok := pkg.Enums[short]; ok {
			return true
		}
	}
	return false
}

// stripPackagePrefix removes the leading package qualifier from a type name.
// e.g., "webhook.v1.WebhookStatus" -> "WebhookStatus"
func stripPackagePrefix(typeName string) string {
	if idx := strings.LastIndex(typeName, "."); idx != -1 {
		return typeName[idx+1:]
	}
	return typeName
}
