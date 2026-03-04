package adapter

import (
	"regexp"

	"github.com/gofiber/fiber/v2"
	"github.com/leandroluk/gox/oas"
)

// GroupBuilder constructor for group configuration
type GroupBuilder struct {
	group *Group
}

// Tag adds an OAS tag to the group. All routes in the group will inherit this tag.
func (b *GroupBuilder) Tag(tag string) *GroupBuilder {
	b.group.tags = append(b.group.tags, tag)
	return b
}

// Description sets the group description, registered in the OAS document
// alongside the group tags.
func (b *GroupBuilder) Description(desc string) *GroupBuilder {
	b.group.description = desc
	return b
}

// Group represents a route group with OpenAPI context
type Group struct {
	fiber.Router
	document    *oas.Document
	tags        []string
	pathPrefix  string
	description string
}

// OAS configures the group with tags and description.
// The description is registered in the OAS document linked to the group tags.
func (g *Group) OAS(fn func(*GroupBuilder)) *Group {
	builder := &GroupBuilder{group: g}
	fn(builder)

	// Registers the description in the OAS document for each tag in the group
	if g.description != "" && len(g.tags) > 0 {
		for _, tag := range g.tags {
			tagName := tag
			desc := g.description
			g.document.Tag(tagName, func(t *oas.Tag) {
				t.Description(desc)
			})
		}
	}

	return g
}

// Group creates a subgroup inheriting the parent group's tags and pathPrefix.
func (g *Group) Group(prefix string, handlers ...fiber.Handler) *Group {
	return &Group{
		Router:     g.Router.Group(prefix, handlers...),
		document:   g.document,
		tags:       append([]string{}, g.tags...), // inherit a copy of the parent tags
		pathPrefix: g.pathPrefix + prefix,
	}
}

// ExtractPathParams extracts path parameter names in Fiber format.
// Supports normal parameters (:param) and optional ones (:param?).
//
// Examples:
//
//	/users/:id           → ["id"]
//	/users/:id/:field    → ["id", "field"]
//	/files/:name?        → ["name"]
func (g *Group) ExtractPathParams(routePath string) []string {
	re := regexp.MustCompile(`:([a-zA-Z0-9_]+)\??`)
	matches := re.FindAllStringSubmatch(routePath, -1)

	params := make([]string, 0, len(matches))
	for _, match := range matches {
		if len(match) > 1 {
			params = append(params, match[1])
		}
	}
	return params
}
