package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/leandroluk/gox/oas/types"
)

func main() {
	// Exemplo usando RequiredProperty e OptionalProperty
	doc := types.New().
		OpenAPI("3.1.0").
		Info(func(i *types.Info) {
			i.Title("Example API").Version("1.0.0")
		}).
		Components(func(c *types.Components) {
			// Usando a nova abordagem
			c.Schema("User", func(s *types.Schema) {
				s.Object().
					Required("id", func(p *types.Schema) {
						p.String().Format("uuid").Example("550e8400-e29b-41d4-a716-446655440000")
					}).
					Required("name", func(p *types.Schema) {
						p.String().MinLength(3).MaxLength(100).Example("John Doe")
					}).
					Required("email", func(p *types.Schema) {
						p.String().Format("email").Example("john@example.com")
					}).
					Optional("age", func(p *types.Schema) {
						p.Integer().Minimum(0).Maximum(150).Example(30)
					}).
					Optional("bio", func(p *types.Schema) {
						p.String().MaxLength(500).Example("Software developer")
					})
			})
		})

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(doc); err != nil {
		fmt.Fprintf(os.Stderr, "Error encoding JSON: %v\n", err)
		os.Exit(1)
	}
}
