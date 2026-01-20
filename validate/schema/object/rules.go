// schema/object/rules.go
package object

import "fmt"

const (
	CodeRequired = "object.required"
	CodeType     = "object.type"
	CodeInvalid  = "object.invalid"

	CodeFieldDecode = "object.field.decode"

	CodeFieldRequiredIf      = "object.field.required_if"
	CodeFieldRequiredWith    = "object.field.required_with"
	CodeFieldRequiredWithout = "object.field.required_without"
	CodeFieldExcludedIf      = "object.field.excluded_if"

	CodeFieldEqField  = "object.field.eqfield"
	CodeFieldNeField  = "object.field.nefield"
	CodeFieldGtField  = "object.field.gtfield"
	CodeFieldGteField = "object.field.gtefield"
	CodeFieldLtField  = "object.field.ltfield"
	CodeFieldLteField = "object.field.ltefield"

	CodeFieldEqCSField  = "object.field.eqcsfield"
	CodeFieldNeCSField  = "object.field.necsfield"
	CodeFieldGtCSField  = "object.field.gtcsfield"
	CodeFieldGteCSField = "object.field.gtecsfield"
	CodeFieldLtCSField  = "object.field.ltcsfield"
	CodeFieldLteCSField = "object.field.ltecsfield"

	CodeFieldContains = "object.field.fieldcontains"
	CodeFieldExcludes = "object.field.fieldexcludes"
)

var ErrInvalidBuilderUsage = fmt.Errorf("invalid builder usage")
