package parser

// resolveFieldTypes marks IsMessage and IsEnum on all fields in a package
// by looking up each field's RawType against the package's merged message
// and enum maps.
func resolveFieldTypes(pkg *ProtoPackage) {
	for _, msg := range pkg.Messages {
		for i := range msg.Fields {
			f := &msg.Fields[i]
			rawType := f.RawType

			// Check map value type for map fields
			if f.IsMap {
				rawType = f.MapValue
			}

			if _, ok := pkg.Messages[rawType]; ok {
				f.IsMessage = true
			}
			if _, ok := pkg.Enums[rawType]; ok {
				f.IsEnum = true
			}
		}
	}
}
