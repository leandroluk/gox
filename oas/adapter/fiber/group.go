// github.com/leandroluk/gox/oas/adapter/fiber/group.go
package adapter

import (
	"github.com/gofiber/fiber/v2"
	"github.com/leandroluk/gox/oas"
)

// Group represents a route group with OpenAPI context
type Group struct {
	fiber.Router
	document    *oas.Document
	tags        []string
	pathPrefix  string
	description string
}

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
