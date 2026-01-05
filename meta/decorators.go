package meta

import "reflect"

// Description adds a text documentation to an object or field.
type descriptionDecorator struct{ Text string }

func Description(text string) descriptionDecorator                    { return descriptionDecorator{Text: text} }
func (d descriptionDecorator) applyToObject(_ any, m *ObjectMetadata) { m.Description = d.Text }
func (d descriptionDecorator) applyToField(m *FieldMetadata)          { m.Description = d.Text }

// Example adds a sample value to an object or field.
type exampleDecorator struct{ Value any }

func Example[T any](val T) exampleDecorator                       { return exampleDecorator{Value: val} }
func (d exampleDecorator) applyToObject(_ any, m *ObjectMetadata) { m.Example = d.Value }
func (d exampleDecorator) applyToField(m *FieldMetadata)          { m.Example = d.Value }

// Throws documents potential errors.
type throwsDecorator ThrowsMetadata

func Throws[T any](desc string) throwsDecorator {
	return throwsDecorator{ErrorType: reflect.TypeFor[T](), Description: desc}
}
func (d throwsDecorator) applyToObject(_ any, m *ObjectMetadata) {
	m.Throws = append(m.Throws, ThrowsMetadata(d))
}

// Field targets a specific field in the struct for documentation using its memory address.
type fieldDecorator struct {
	FieldPointer any
	Options      []FieldOption
}

func Field(ptr any, opts ...FieldOption) fieldDecorator { return fieldDecorator{ptr, opts} }
func (d fieldDecorator) applyToObject(sPtr any, m *ObjectMetadata) {
	name := resolveFieldName(sPtr, d.FieldPointer)
	if name == "" {
		panic("meta: could not resolve field name - ensure you are passing a pointer to the struct's field")
	}

	fMeta, exists := m.Fields[name]
	if !exists {
		fMeta = &FieldMetadata{}
		m.Fields[name] = fMeta
	}

	val := reflect.ValueOf(d.FieldPointer)
	// Determine type and nullability based on the field pointer
	if val.Kind() == reflect.Pointer {
		fMeta.Type = val.Elem().Type()
		// If the pointer points to a pointer, it's nullable
		if val.Elem().Kind() == reflect.Pointer ||
			val.Elem().Kind() == reflect.Slice ||
			val.Elem().Kind() == reflect.Map ||
			val.Elem().Kind() == reflect.Interface {
			fMeta.Nullable = true
		} else {
			fMeta.Nullable = false
		}
	}

	for _, opt := range d.Options {
		opt.applyToField(fMeta)
	}
}
