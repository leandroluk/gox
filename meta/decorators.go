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

// Title sets a title for an object.
type titleDecorator struct{ Text string }

func Title(text string) titleDecorator                          { return titleDecorator{Text: text} }
func (d titleDecorator) applyToObject(_ any, m *ObjectMetadata) { m.Title = d.Text }

// Required marks a field as required in the schema.
type requiredDecorator struct{}

func Required() requiredDecorator { return requiredDecorator{} }
func (d requiredDecorator) applyToField(m *FieldMetadata) {
	m.Required = true
}

// Format sets an OpenAPI format hint for a field (e.g. "date-time", "uuid", "email").
type formatDecorator struct{ Value string }

func Format(format string) formatDecorator              { return formatDecorator{Value: format} }
func (d formatDecorator) applyToField(m *FieldMetadata) { m.Format = d.Value }

// Constraints

type minDecorator struct{ Value float64 }

func Min(val float64) minDecorator                   { return minDecorator{val} }
func (d minDecorator) applyToField(m *FieldMetadata) { m.Min = &d.Value }

type maxDecorator struct{ Value float64 }

func Max(val float64) maxDecorator                   { return maxDecorator{val} }
func (d maxDecorator) applyToField(m *FieldMetadata) { m.Max = &d.Value }

type multipleOfDecorator struct{ Value float64 }

func MultipleOf(val float64) multipleOfDecorator            { return multipleOfDecorator{val} }
func (d multipleOfDecorator) applyToField(m *FieldMetadata) { m.MultipleOf = &d.Value }

type minLengthDecorator struct{ Value int }

func MinLength(val int) minLengthDecorator                 { return minLengthDecorator{val} }
func (d minLengthDecorator) applyToField(m *FieldMetadata) { m.MinLength = &d.Value }

type maxLengthDecorator struct{ Value int }

func MaxLength(val int) maxLengthDecorator                 { return maxLengthDecorator{val} }
func (d maxLengthDecorator) applyToField(m *FieldMetadata) { m.MaxLength = &d.Value }

type patternDecorator struct{ Value string }

func Pattern(val string) patternDecorator                { return patternDecorator{val} }
func (d patternDecorator) applyToField(m *FieldMetadata) { m.Pattern = d.Value }

type minItemsDecorator struct{ Value int }

func MinItems(val int) minItemsDecorator                  { return minItemsDecorator{val} }
func (d minItemsDecorator) applyToField(m *FieldMetadata) { m.MinItems = &d.Value }

type maxItemsDecorator struct{ Value int }

func MaxItems(val int) maxItemsDecorator                  { return maxItemsDecorator{val} }
func (d maxItemsDecorator) applyToField(m *FieldMetadata) { m.MaxItems = &d.Value }

// Visibility

type readOnlyDecorator struct{}

func ReadOnly() readOnlyDecorator                         { return readOnlyDecorator{} }
func (d readOnlyDecorator) applyToField(m *FieldMetadata) { m.ReadOnly = true }

type writeOnlyDecorator struct{}

func WriteOnly() writeOnlyDecorator                        { return writeOnlyDecorator{} }
func (d writeOnlyDecorator) applyToField(m *FieldMetadata) { m.WriteOnly = true }

type deprecatedDecorator struct{}

func Deprecated() deprecatedDecorator { return deprecatedDecorator{} }
func (d deprecatedDecorator) applyToObject(_ any, m *ObjectMetadata) {
	m.Deprecated = true
}
func (d deprecatedDecorator) applyToField(m *FieldMetadata) {
	m.Deprecated = true
}

// ExternalDocs

type externalDocsDecorator struct{ Docs ExternalDocs }

func ExtDocs(description, url string) externalDocsDecorator {
	return externalDocsDecorator{ExternalDocs{Description: description, URL: url}}
}
func (d externalDocsDecorator) applyToObject(_ any, m *ObjectMetadata) {
	m.ExternalDocs = &d.Docs
}
func (d externalDocsDecorator) applyToField(m *FieldMetadata) {
	m.ExternalDocs = &d.Docs
}

// Throws documents potential errors.
type throwsDecorator ThrowsMetadata

func Throws[T error](optionalDescription ...string) throwsDecorator {
	errType := reflect.TypeFor[T]()
	description := errType.Name()
	if len(optionalDescription) > 0 {
		description = optionalDescription[0]
	}
	return throwsDecorator{
		ErrorType:   errType,
		Description: description,
	}
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
	name, jsonName, fieldType := ResolveFieldName(sPtr, d.FieldPointer)
	if name == "" {
		panic("meta: could not resolve field name - ensure you are passing a pointer to the struct's field")
	}

	fMeta, exists := m.Fields[name]
	if !exists {
		fMeta = &FieldMetadata{}
		m.Fields[name] = fMeta
	}

	fMeta.JSONName = jsonName

	if fieldType != nil {
		resolvedType := fieldType
		fMeta.Nullable = false

		if resolvedType.Kind() == reflect.Pointer ||
			resolvedType.Kind() == reflect.Slice ||
			resolvedType.Kind() == reflect.Map ||
			resolvedType.Kind() == reflect.Interface {
			fMeta.Nullable = true
			if resolvedType.Kind() == reflect.Pointer {
				resolvedType = resolvedType.Elem()
			}
		}

		fMeta.Type = resolvedType

		// Auto-detect Enumerable interface to populate Enum values.
		instance := reflect.New(resolvedType).Elem().Interface()
		if e, ok := instance.(Enumerable); ok {
			fMeta.Enum = e.Values()
		}
	}

	for _, opt := range d.Options {
		opt.applyToField(fMeta)
	}

	// Update the object's Required list if necessary
	if fMeta.Required {
		alreadyPresent := false
		for _, r := range m.Required {
			if r == fMeta.JSONName {
				alreadyPresent = true
				break
			}
		}
		if !alreadyPresent {
			m.Required = append(m.Required, fMeta.JSONName)
		}
	}
}
