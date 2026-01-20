package types

import (
	"encoding/json"

	"github.com/leandroluk/gox/oas/enums"
)

// Schema represents an OAS 3.1 schema.
type Schema struct {
	type_                 []enums.SchemaType
	format                *string
	description           *string
	properties            map[string]*Schema
	required              []string
	items                 *Schema
	title                 *string
	multipleOf            *float64
	maximum               *float64
	exclusiveMaximum      *bool
	minimum               *float64
	exclusiveMinimum      *bool
	maxLength             *int64
	minLength             *int64
	pattern               *string
	maxItems              *int64
	minItems              *int64
	uniqueItems           *bool
	maxProperties         *int64
	minProperties         *int64
	enum                  []any
	allOf                 []*Schema
	oneOf                 []*Schema
	anyOf                 []*Schema
	not                   *Schema
	additionalProperties  any
	default_              any
	discriminator         *Discriminator
	readOnly              *bool
	writeOnly             *bool
	xml                   *Xml
	externalDocs          *ExternalDocs
	example               any
	deprecated            *bool
	dependentSchemas      map[string]*Schema
	unevaluatedItems      any
	unevaluatedProperties any
	if_                   *Schema
	then                  *Schema
	else_                 *Schema
	contentMediaType      *string
	contentEncoding       *string
	ref                   *string
	const_                any
}

func (s *Schema) Type(values ...enums.SchemaType) *Schema {
	s.type_ = values
	return s
}

// Type helpers
func (s *Schema) String() *Schema {
	return s.Type(enums.SchemaString)
}

func (s *Schema) Object() *Schema {
	return s.Type(enums.SchemaObject)
}

func (s *Schema) Array() *Schema {
	return s.Type(enums.SchemaArray)
}

func (s *Schema) Integer() *Schema {
	return s.Type(enums.SchemaInteger)
}

func (s *Schema) Number() *Schema {
	return s.Type(enums.SchemaNumber)
}

func (s *Schema) Boolean() *Schema {
	return s.Type(enums.SchemaBoolean)
}

func (s *Schema) Null() *Schema {
	return s.Type(enums.SchemaNull)
}

func (s *Schema) Nullable() *Schema {
	for _, t := range s.type_ {
		if t == enums.SchemaNull {
			return s
		}
	}
	s.type_ = append(s.type_, enums.SchemaNull)
	return s
}

func (s *Schema) Format(value string) *Schema {
	s.format = &value
	return s
}

func (s *Schema) Description(value string) *Schema {
	s.description = &value
	return s
}

func (s *Schema) Property(name string, build func(doc *Schema)) *Schema {
	if s.properties == nil {
		s.properties = make(map[string]*Schema)
	}
	schema, ok := s.properties[name]
	if !ok {
		schema = &Schema{}
		s.properties[name] = schema
	}
	if build != nil {
		build(schema)
	}
	return s
}

func (s *Schema) Required(values ...string) *Schema {
	s.required = values
	return s
}

func (s *Schema) Items(build func(doc *Schema)) *Schema {
	if s.items == nil {
		s.items = &Schema{}
	}
	if build != nil {
		build(s.items)
	}
	return s
}

func (s *Schema) Title(value string) *Schema {
	s.title = &value
	return s
}

func (s *Schema) MultipleOf(value float64) *Schema {
	s.multipleOf = &value
	return s
}

func (s *Schema) Maximum(value float64) *Schema {
	s.maximum = &value
	return s
}

func (s *Schema) ExclusiveMaximum(value bool) *Schema {
	s.exclusiveMaximum = &value
	return s
}

func (s *Schema) Minimum(value float64) *Schema {
	s.minimum = &value
	return s
}

func (s *Schema) ExclusiveMinimum(value bool) *Schema {
	s.exclusiveMinimum = &value
	return s
}

func (s *Schema) MaxLength(value int64) *Schema {
	s.maxLength = &value
	return s
}

func (s *Schema) MinLength(value int64) *Schema {
	s.minLength = &value
	return s
}

func (s *Schema) Pattern(value string) *Schema {
	s.pattern = &value
	return s
}

func (s *Schema) MaxItems(value int64) *Schema {
	s.maxItems = &value
	return s
}

func (s *Schema) MinItems(value int64) *Schema {
	s.minItems = &value
	return s
}

func (s *Schema) UniqueItems(value bool) *Schema {
	s.uniqueItems = &value
	return s
}

func (s *Schema) MaxProperties(value int64) *Schema {
	s.maxProperties = &value
	return s
}

func (s *Schema) MinProperties(value int64) *Schema {
	s.minProperties = &value
	return s
}

func (s *Schema) Enum(values ...any) *Schema {
	s.enum = values
	return s
}

func (s *Schema) AllOf(build func(doc *Schema)) *Schema {
	schema := &Schema{}
	if build != nil {
		build(schema)
	}
	s.allOf = append(s.allOf, schema)
	return s
}

func (s *Schema) OneOf(build func(doc *Schema)) *Schema {
	schema := &Schema{}
	if build != nil {
		build(schema)
	}
	s.oneOf = append(s.oneOf, schema)
	return s
}

func (s *Schema) AnyOf(build func(doc *Schema)) *Schema {
	schema := &Schema{}
	if build != nil {
		build(schema)
	}
	s.anyOf = append(s.anyOf, schema)
	return s
}

func (s *Schema) Not(build func(doc *Schema)) *Schema {
	if s.not == nil {
		s.not = &Schema{}
	}
	if build != nil {
		build(s.not)
	}
	return s
}

func (s *Schema) AdditionalProperties(value any) *Schema {
	s.additionalProperties = value
	return s
}

func (s *Schema) Default(value any) *Schema {
	s.default_ = value
	return s
}

func (s *Schema) Discriminator(build func(d *Discriminator)) *Schema {
	if s.discriminator == nil {
		s.discriminator = &Discriminator{}
	}
	if build != nil {
		build(s.discriminator)
	}
	return s
}

func (s *Schema) ReadOnly(value bool) *Schema {
	s.readOnly = &value
	return s
}

func (s *Schema) WriteOnly(value bool) *Schema {
	s.writeOnly = &value
	return s
}

func (s *Schema) Xml(build func(x *Xml)) *Schema {
	if s.xml == nil {
		s.xml = &Xml{}
	}
	if build != nil {
		build(s.xml)
	}
	return s
}

func (s *Schema) ExternalDocs(build func(e *ExternalDocs)) *Schema {
	if s.externalDocs == nil {
		s.externalDocs = &ExternalDocs{}
	}
	if build != nil {
		build(s.externalDocs)
	}
	return s
}

// ExternalDoc sets external documentation with URL as required parameter.
func (s *Schema) ExternalDoc(url string, optionalBuild ...func(e *ExternalDocs)) *Schema {
	if s.externalDocs == nil {
		s.externalDocs = &ExternalDocs{}
	}
	s.externalDocs.URL(url)
	if len(optionalBuild) > 0 && optionalBuild[0] != nil {
		optionalBuild[0](s.externalDocs)
	}
	return s
}

func (s *Schema) Example(value any) *Schema {
	s.example = value
	return s
}

func (s *Schema) Deprecated(value bool) *Schema {
	s.deprecated = &value
	return s
}

func (s *Schema) DependentSchema(name string, build func(doc *Schema)) *Schema {
	if s.dependentSchemas == nil {
		s.dependentSchemas = make(map[string]*Schema)
	}
	schema := &Schema{}
	if build != nil {
		build(schema)
	}
	s.dependentSchemas[name] = schema
	return s
}

func (s *Schema) UnevaluatedItems(value any) *Schema {
	s.unevaluatedItems = value
	return s
}

func (s *Schema) UnevaluatedProperties(value any) *Schema {
	s.unevaluatedProperties = value
	return s
}

func (s *Schema) If(build func(doc *Schema)) *Schema {
	if s.if_ == nil {
		s.if_ = &Schema{}
	}
	if build != nil {
		build(s.if_)
	}
	return s
}

func (s *Schema) Then(build func(doc *Schema)) *Schema {
	if s.then == nil {
		s.then = &Schema{}
	}
	if build != nil {
		build(s.then)
	}
	return s
}

func (s *Schema) Else(build func(doc *Schema)) *Schema {
	if s.else_ == nil {
		s.else_ = &Schema{}
	}
	if build != nil {
		build(s.else_)
	}
	return s
}

func (s *Schema) ContentMediaType(value string) *Schema {
	s.contentMediaType = &value
	return s
}

func (s *Schema) ContentEncoding(value string) *Schema {
	s.contentEncoding = &value
	return s
}

func (s *Schema) Ref(value string) *Schema {
	s.ref = &value
	return s
}

func (s *Schema) Const(value any) *Schema {
	s.const_ = value
	return s
}

func (s Schema) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Type                  []enums.SchemaType `json:"type,omitempty"`
		Format                *string            `json:"format,omitempty"`
		Description           *string            `json:"description,omitempty"`
		Properties            map[string]*Schema `json:"properties,omitempty"`
		Required              []string           `json:"required,omitempty"`
		Items                 *Schema            `json:"items,omitempty"`
		Title                 *string            `json:"title,omitempty"`
		MultipleOf            *float64           `json:"multipleOf,omitempty"`
		Maximum               *float64           `json:"maximum,omitempty"`
		ExclusiveMaximum      *bool              `json:"exclusiveMaximum,omitempty"`
		Minimum               *float64           `json:"minimum,omitempty"`
		ExclusiveMinimum      *bool              `json:"exclusiveMinimum,omitempty"`
		MaxLength             *int64             `json:"maxLength,omitempty"`
		MinLength             *int64             `json:"minLength,omitempty"`
		Pattern               *string            `json:"pattern,omitempty"`
		MaxItems              *int64             `json:"maxItems,omitempty"`
		MinItems              *int64             `json:"minItems,omitempty"`
		UniqueItems           *bool              `json:"uniqueItems,omitempty"`
		MaxProperties         *int64             `json:"maxProperties,omitempty"`
		MinProperties         *int64             `json:"minProperties,omitempty"`
		Enum                  []any              `json:"enum,omitempty"`
		AllOf                 []*Schema          `json:"allOf,omitempty"`
		OneOf                 []*Schema          `json:"oneOf,omitempty"`
		AnyOf                 []*Schema          `json:"anyOf,omitempty"`
		Not                   *Schema            `json:"not,omitempty"`
		AdditionalProperties  any                `json:"additionalProperties,omitempty"`
		Default               any                `json:"default,omitempty"`
		Discriminator         *Discriminator     `json:"discriminator,omitempty"`
		ReadOnly              *bool              `json:"readOnly,omitempty"`
		WriteOnly             *bool              `json:"writeOnly,omitempty"`
		XML                   *Xml               `json:"xml,omitempty"`
		ExternalDocs          *ExternalDocs      `json:"externalDocs,omitempty"`
		Example               any                `json:"example,omitempty"`
		Deprecated            *bool              `json:"deprecated,omitempty"`
		DependentSchemas      map[string]*Schema `json:"dependentSchemas,omitempty"`
		UnevaluatedItems      any                `json:"unevaluatedItems,omitempty"`
		UnevaluatedProperties any                `json:"unevaluatedProperties,omitempty"`
		If                    *Schema            `json:"if,omitempty"`
		Then                  *Schema            `json:"then,omitempty"`
		Else                  *Schema            `json:"else,omitempty"`
		ContentMediaType      *string            `json:"contentMediaType,omitempty"`
		ContentEncoding       *string            `json:"contentEncoding,omitempty"`
		Ref                   *string            `json:"$ref,omitempty"`
		Const                 any                `json:"const,omitempty"`
	}{
		Type:                  s.type_,
		Format:                s.format,
		Description:           s.description,
		Properties:            s.properties,
		Required:              s.required,
		Items:                 s.items,
		Title:                 s.title,
		MultipleOf:            s.multipleOf,
		Maximum:               s.maximum,
		ExclusiveMaximum:      s.exclusiveMaximum,
		Minimum:               s.minimum,
		ExclusiveMinimum:      s.exclusiveMinimum,
		MaxLength:             s.maxLength,
		MinLength:             s.minLength,
		Pattern:               s.pattern,
		MaxItems:              s.maxItems,
		MinItems:              s.minItems,
		UniqueItems:           s.uniqueItems,
		MaxProperties:         s.maxProperties,
		MinProperties:         s.minProperties,
		Enum:                  s.enum,
		AllOf:                 s.allOf,
		OneOf:                 s.oneOf,
		AnyOf:                 s.anyOf,
		Not:                   s.not,
		AdditionalProperties:  s.additionalProperties,
		Default:               s.default_,
		Discriminator:         s.discriminator,
		ReadOnly:              s.readOnly,
		WriteOnly:             s.writeOnly,
		XML:                   s.xml,
		ExternalDocs:          s.externalDocs,
		Example:               s.example,
		Deprecated:            s.deprecated,
		DependentSchemas:      s.dependentSchemas,
		UnevaluatedItems:      s.unevaluatedItems,
		UnevaluatedProperties: s.unevaluatedProperties,
		If:                    s.if_,
		Then:                  s.then,
		Else:                  s.else_,
		ContentMediaType:      s.contentMediaType,
		ContentEncoding:       s.contentEncoding,
		Ref:                   s.ref,
		Const:                 s.const_,
	})
}

// UnmarshalJSON unmarshals the Schema from JSON.
func (s *Schema) UnmarshalJSON(data []byte) error {
	type Alias Schema
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(s),
	}
	return json.Unmarshal(data, &aux)
}
