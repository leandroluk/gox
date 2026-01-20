// internal/path/builder.go
package path

import "strings"

type Builder struct {
	parts []Part
}

func NewBuilder() Builder {
	return Builder{parts: make([]Part, 0, 8)}
}

func (builder *Builder) Reset() {
	builder.parts = builder.parts[:0]
}

func (builder *Builder) Len() int {
	return len(builder.parts)
}

func (builder *Builder) PushField(name string) {
	builder.parts = append(builder.parts, Part{
		Kind:  PartField,
		Field: name,
	})
}

func (builder *Builder) PushIndex(index int) {
	builder.parts = append(builder.parts, Part{
		Kind:  PartIndex,
		Index: index,
	})
}

func (builder *Builder) PushKey(key string) {
	builder.parts = append(builder.parts, Part{
		Kind: PartKey,
		Key:  key,
	})
}

func (builder *Builder) Pop() {
	if len(builder.parts) == 0 {
		return
	}
	builder.parts = builder.parts[:len(builder.parts)-1]
}

func (builder *Builder) String() string {
	if len(builder.parts) == 0 {
		return ""
	}
	var stringBuilder strings.Builder
	AppendTo(&stringBuilder, builder.parts)
	return stringBuilder.String()
}

func (builder *Builder) Snapshot() []Part {
	return append([]Part(nil), builder.parts...)
}

func (builder *Builder) Restore(parts []Part) {
	builder.parts = append(builder.parts[:0], parts...)
}
