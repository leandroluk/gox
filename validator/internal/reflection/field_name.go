// internal/reflection/field_name.go
package reflection

import "reflect"

func FieldJSONName(field reflect.StructField) string {
	tag := ParseJSONTag(field.Tag.Get("json"))
	if tag.Ignored {
		return ""
	}
	if tag.HasTag && tag.Name != "" {
		return tag.Name
	}
	return field.Name
}

func IsExported(field reflect.StructField) bool {
	return field.PkgPath == ""
}
