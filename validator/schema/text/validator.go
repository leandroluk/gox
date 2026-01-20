// schema/text/validator.go
package text

import (
	"github.com/leandroluk/go/validator/internal/ast"
	"github.com/leandroluk/go/validator/internal/defaults"
	"github.com/leandroluk/go/validator/internal/engine"
	"github.com/leandroluk/go/validator/internal/reflection"
	"github.com/leandroluk/go/validator/internal/ruleset"
	"github.com/leandroluk/go/validator/schema"
)

func (schemaValue *Schema) validateWithOptions(input any, options schema.Options) (string, error) {
	context := engine.NewContext(options)

	value, err := engine.InputToASTWithOptions(input, options)
	if err != nil {
		return "", err
	}

	output, _ := schemaValue.validateAST(context, value)
	return output, context.Error()
}

func (schemaValue *Schema) validateAST(context *engine.Context, value ast.Value) (string, bool) {
	if defaultValue, ok := defaults.Apply(value.Presence, context.Options, schemaValue.defaultProvider); ok {
		return defaultValue, false
	}

	if value.IsMissing() || value.IsNull() {
		if schemaValue.required {
			stop := context.AddIssue(CodeRequired, "required")
			return "", stop
		}
		return "", false
	}

	output, stopParse := parseTextValue(context, value)
	if stopParse {
		return "", true
	}

	if schemaValue.isDefault && reflection.IsDefault(output) {
		return output, false
	}

	output, stopRules := schemaValue.rules.ApplyAll(output, context)
	if stopRules {
		return output, true
	}

	if len(schemaValue.customRules) > 0 {
		if ruleset.Apply(output, context, schemaValue.customRules...) {
			return output, true
		}
	}

	return output, false
}
