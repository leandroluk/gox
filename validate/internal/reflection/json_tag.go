// internal/reflection/json_tag.go
package reflection

import "strings"

type JSONTag struct {
	Name      string
	OmitEmpty bool
	Ignored   bool
	HasTag    bool
}

func ParseJSONTag(tag string) JSONTag {
	if tag == "" {
		return JSONTag{}
	}

	head, tail, _ := strings.Cut(tag, ",")
	if head == "-" {
		return JSONTag{Ignored: true, HasTag: true}
	}

	result := JSONTag{
		Name:   head,
		HasTag: true,
	}

	for tail != "" {
		var part string
		part, tail, _ = strings.Cut(tail, ",")
		if part == "omitempty" {
			result.OmitEmpty = true
		}
	}

	return result
}
