// schema/text/schema.go
package text

import (
	"reflect"
	"regexp"

	"github.com/leandroluk/go/validate/internal/defaults"
	"github.com/leandroluk/go/validate/internal/ruleset"
	"github.com/leandroluk/go/validate/schema"
	"github.com/leandroluk/go/validate/schema/text/rule"
)

type Schema struct {
	required  bool
	isDefault bool

	defaultProvider defaults.Provider[string]

	rules       *ruleset.Set[string]
	customRules []ruleset.RuleFn[string]
}

func New() *Schema {
	return &Schema{
		defaultProvider: defaults.None[string](),
		rules:           ruleset.NewSet[string](),
		customRules:     make([]ruleset.RuleFn[string], 0),
	}
}

func (schemaValue *Schema) Required() *Schema {
	schemaValue.required = true
	return schemaValue
}

func (schemaValue *Schema) IsDefault() *Schema {
	schemaValue.isDefault = true
	return schemaValue
}

func (schemaValue *Schema) Len(length int) *Schema {
	if length < 0 {
		schemaValue.rules.Remove(CodeLen)
		return schemaValue
	}
	schemaValue.rules.Put(rule.Len(CodeLen, length))
	return schemaValue
}

func (schemaValue *Schema) Min(length int) *Schema {
	if length < 0 {
		schemaValue.rules.Remove(CodeMin)
		return schemaValue
	}
	schemaValue.rules.Put(rule.Min(CodeMin, length))
	return schemaValue
}

func (schemaValue *Schema) Max(length int) *Schema {
	if length < 0 {
		schemaValue.rules.Remove(CodeMax)
		return schemaValue
	}
	schemaValue.rules.Put(rule.Max(CodeMax, length))
	return schemaValue
}

func (schemaValue *Schema) Eq(value string) *Schema {
	schemaValue.rules.Put(rule.Eq(CodeEq, value))
	return schemaValue
}

func (schemaValue *Schema) Ne(value string) *Schema {
	schemaValue.rules.Put(rule.Ne(CodeNe, value))
	return schemaValue
}

func (schemaValue *Schema) EqIgnoreCase(value string) *Schema {
	schemaValue.rules.Put(rule.EqIgnoreCase(CodeEqI, value))
	return schemaValue
}

func (schemaValue *Schema) NeIgnoreCase(value string) *Schema {
	schemaValue.rules.Put(rule.NeIgnoreCase(CodeNeI, value))
	return schemaValue
}

func (schemaValue *Schema) OneOf(values ...string) *Schema {
	if len(values) == 0 {
		schemaValue.rules.Remove(CodeOneOf)
		return schemaValue
	}
	schemaValue.rules.Put(rule.OneOf(CodeOneOf, values...))
	return schemaValue
}

func (schemaValue *Schema) Contains(value string) *Schema {
	if value == "" {
		schemaValue.rules.Remove(CodeContains)
		return schemaValue
	}
	schemaValue.rules.Put(rule.Contains(CodeContains, value))
	return schemaValue
}

func (schemaValue *Schema) Excludes(value string) *Schema {
	if value == "" {
		schemaValue.rules.Remove(CodeExcludes)
		return schemaValue
	}
	schemaValue.rules.Put(rule.Excludes(CodeExcludes, value))
	return schemaValue
}

func (schemaValue *Schema) StartsWith(value string) *Schema {
	if value == "" {
		schemaValue.rules.Remove(CodeStartsWith)
		return schemaValue
	}
	schemaValue.rules.Put(rule.StartsWith(CodeStartsWith, value))
	return schemaValue
}

func (schemaValue *Schema) NotStartsWith(value string) *Schema {
	if value == "" {
		schemaValue.rules.Remove(CodeNotStartsWith)
		return schemaValue
	}
	schemaValue.rules.Put(rule.NotStartsWith(CodeNotStartsWith, value))
	return schemaValue
}

func (schemaValue *Schema) EndsWith(value string) *Schema {
	if value == "" {
		schemaValue.rules.Remove(CodeEndsWith)
		return schemaValue
	}
	schemaValue.rules.Put(rule.EndsWith(CodeEndsWith, value))
	return schemaValue
}

func (schemaValue *Schema) NotEndsWith(value string) *Schema {
	if value == "" {
		schemaValue.rules.Remove(CodeNotEndsWith)
		return schemaValue
	}
	schemaValue.rules.Put(rule.NotEndsWith(CodeNotEndsWith, value))
	return schemaValue
}

func (schemaValue *Schema) Lowercase() *Schema {
	schemaValue.rules.Put(rule.Lowercase(CodeLowercase))
	return schemaValue
}

func (schemaValue *Schema) Uppercase() *Schema {
	schemaValue.rules.Put(rule.Uppercase(CodeUppercase))
	return schemaValue
}

func (schemaValue *Schema) Pattern(value string) *Schema {
	if value == "" {
		schemaValue.rules.Remove(CodePattern)
		return schemaValue
	}
	compiled, err := regexp.Compile(value)
	if err != nil {
		return schemaValue
	}
	schemaValue.rules.Put(rule.Pattern(CodePattern, compiled))
	return schemaValue
}

func (schemaValue *Schema) PatternRegexp(value *regexp.Regexp) *Schema {
	if value == nil {
		schemaValue.rules.Remove(CodePattern)
		return schemaValue
	}
	schemaValue.rules.Put(rule.Pattern(CodePattern, value))
	return schemaValue
}

func (schemaValue *Schema) Email() *Schema {
	schemaValue.rules.Put(rule.Email(CodeEmail))
	return schemaValue
}

func (schemaValue *Schema) URL() *Schema {
	schemaValue.rules.Put(rule.URL(CodeURL))
	return schemaValue
}

func (schemaValue *Schema) HTTPURL() *Schema {
	schemaValue.rules.Put(rule.HTTPURL(CodeHTTPURL))
	return schemaValue
}

func (schemaValue *Schema) HttpURL() *Schema {
	return schemaValue.HTTPURL()
}

func (schemaValue *Schema) URI() *Schema {
	schemaValue.rules.Put(rule.URI(CodeURI))
	return schemaValue
}

func (schemaValue *Schema) URNRFC2141() *Schema {
	schemaValue.rules.Put(rule.URNRFC2141(CodeURNRFC2141))
	return schemaValue
}

func (schemaValue *Schema) UrnRFC2141() *Schema {
	return schemaValue.URNRFC2141()
}

func (schemaValue *Schema) File() *Schema {
	schemaValue.rules.Put(rule.File(CodeFile))
	return schemaValue
}

func (schemaValue *Schema) FilePath() *Schema {
	schemaValue.rules.Put(rule.FilePath(CodeFilePath))
	return schemaValue
}

func (schemaValue *Schema) Filepath() *Schema {
	return schemaValue.FilePath()
}

func (schemaValue *Schema) Dir() *Schema {
	schemaValue.rules.Put(rule.Dir(CodeDir))
	return schemaValue
}

func (schemaValue *Schema) DirPath() *Schema {
	schemaValue.rules.Put(rule.DirPath(CodeDirPath))
	return schemaValue
}

func (schemaValue *Schema) Dirpath() *Schema {
	return schemaValue.DirPath()
}

func (schemaValue *Schema) Image() *Schema {
	schemaValue.rules.Put(rule.Image(CodeImage))
	return schemaValue
}

func (schemaValue *Schema) UUID() *Schema {
	schemaValue.rules.Put(rule.UUID(CodeUUID))
	return schemaValue
}

func (schemaValue *Schema) UUID3() *Schema {
	schemaValue.rules.Put(rule.UUID3(CodeUUID3))
	return schemaValue
}

func (schemaValue *Schema) UUID4() *Schema {
	schemaValue.rules.Put(rule.UUID4(CodeUUID4))
	return schemaValue
}

func (schemaValue *Schema) UUID5() *Schema {
	schemaValue.rules.Put(rule.UUID5(CodeUUID5))
	return schemaValue
}

func (schemaValue *Schema) IP() *Schema {
	schemaValue.rules.Put(rule.IP(CodeIP))
	return schemaValue
}

func (schemaValue *Schema) IPv4() *Schema {
	schemaValue.rules.Put(rule.IPv4(CodeIPv4))
	return schemaValue
}

func (schemaValue *Schema) IPv6() *Schema {
	schemaValue.rules.Put(rule.IPv6(CodeIPv6))
	return schemaValue
}

func (schemaValue *Schema) CIDR() *Schema {
	schemaValue.rules.Put(rule.CIDR(CodeCIDR))
	return schemaValue
}

func (schemaValue *Schema) MAC() *Schema {
	schemaValue.rules.Put(rule.MAC(CodeMAC))
	return schemaValue
}

func (schemaValue *Schema) Hostname() *Schema {
	schemaValue.rules.Put(rule.Hostname(CodeHostname))
	return schemaValue
}

func (schemaValue *Schema) FQDN() *Schema {
	schemaValue.rules.Put(rule.FQDN(CodeFQDN))
	return schemaValue
}

func (schemaValue *Schema) Port() *Schema {
	schemaValue.rules.Put(rule.Port(CodePort))
	return schemaValue
}

func (schemaValue *Schema) Numeric() *Schema {
	schemaValue.rules.Put(rule.Numeric(CodeNumeric))
	return schemaValue
}

func (schemaValue *Schema) Number() *Schema {
	schemaValue.rules.Put(rule.Number(CodeNumber))
	return schemaValue
}

func (schemaValue *Schema) Hexadecimal() *Schema {
	schemaValue.rules.Put(rule.Hexadecimal(CodeHexadecimal))
	return schemaValue
}

func (schemaValue *Schema) HexColor() *Schema {
	schemaValue.rules.Put(rule.HexColor(CodeHexColor))
	return schemaValue
}

func (schemaValue *Schema) RGB() *Schema {
	schemaValue.rules.Put(rule.RGB(CodeRGB))
	return schemaValue
}

func (schemaValue *Schema) RGBA() *Schema {
	schemaValue.rules.Put(rule.RGBA(CodeRGBA))
	return schemaValue
}

func (schemaValue *Schema) HSL() *Schema {
	schemaValue.rules.Put(rule.HSL(CodeHSL))
	return schemaValue
}

func (schemaValue *Schema) HSLA() *Schema {
	schemaValue.rules.Put(rule.HSLA(CodeHSLA))
	return schemaValue
}

func (schemaValue *Schema) Base64() *Schema {
	schemaValue.rules.Put(rule.Base64(CodeBase64))
	return schemaValue
}

func (schemaValue *Schema) Base64URL() *Schema {
	schemaValue.rules.Put(rule.Base64URL(CodeBase64URL))
	return schemaValue
}

func (schemaValue *Schema) Base64Url() *Schema {
	return schemaValue.Base64URL()
}

func (schemaValue *Schema) Base64RawURL() *Schema {
	schemaValue.rules.Put(rule.Base64RawURL(CodeBase64RawURL))
	return schemaValue
}

func (schemaValue *Schema) Base64RawUrl() *Schema {
	return schemaValue.Base64RawURL()
}

func (schemaValue *Schema) DataURI() *Schema {
	schemaValue.rules.Put(rule.DataURI(CodeDataURI))
	return schemaValue
}

func (schemaValue *Schema) ASCII() *Schema {
	schemaValue.rules.Put(rule.ASCII(CodeASCII))
	return schemaValue
}

func (schemaValue *Schema) PrintASCII() *Schema {
	schemaValue.rules.Put(rule.PrintASCII(CodePrintASCII))
	return schemaValue
}

func (schemaValue *Schema) Multibyte() *Schema {
	schemaValue.rules.Put(rule.Multibyte(CodeMultibyte))
	return schemaValue
}

func (schemaValue *Schema) CreditCard() *Schema {
	schemaValue.rules.Put(rule.CreditCard(CodeCreditCard))
	return schemaValue
}

func (schemaValue *Schema) LuhnChecksum() *Schema {
	schemaValue.rules.Put(rule.LuhnChecksum(CodeLuhnChecksum))
	return schemaValue
}

func (schemaValue *Schema) ISBN() *Schema {
	schemaValue.rules.Put(rule.ISBN(CodeISBN))
	return schemaValue
}

func (schemaValue *Schema) ISBN10() *Schema {
	schemaValue.rules.Put(rule.ISBN10(CodeISBN10))
	return schemaValue
}

func (schemaValue *Schema) ISBN13() *Schema {
	schemaValue.rules.Put(rule.ISBN13(CodeISBN13))
	return schemaValue
}

func (schemaValue *Schema) ISSN() *Schema {
	schemaValue.rules.Put(rule.ISSN(CodeISSN))
	return schemaValue
}

func (schemaValue *Schema) E164() *Schema {
	schemaValue.rules.Put(rule.E164(CodeE164))
	return schemaValue
}

func (schemaValue *Schema) SemVer() *Schema {
	schemaValue.rules.Put(rule.SemVer(CodeSemVer))
	return schemaValue
}

func (schemaValue *Schema) Semver() *Schema {
	return schemaValue.SemVer()
}

func (schemaValue *Schema) CVE() *Schema {
	schemaValue.rules.Put(rule.CVE(CodeCVE))
	return schemaValue
}

func (schemaValue *Schema) MD4() *Schema {
	schemaValue.rules.Put(rule.MD4(CodeMD4))
	return schemaValue
}

func (schemaValue *Schema) MD5() *Schema {
	schemaValue.rules.Put(rule.MD5(CodeMD5))
	return schemaValue
}

func (schemaValue *Schema) SHA1() *Schema {
	schemaValue.rules.Put(rule.SHA1(CodeSHA1))
	return schemaValue
}

func (schemaValue *Schema) SHA224() *Schema {
	schemaValue.rules.Put(rule.SHA224(CodeSHA224))
	return schemaValue
}

func (schemaValue *Schema) SHA256() *Schema {
	schemaValue.rules.Put(rule.SHA256(CodeSHA256))
	return schemaValue
}

func (schemaValue *Schema) SHA384() *Schema {
	schemaValue.rules.Put(rule.SHA384(CodeSHA384))
	return schemaValue
}

func (schemaValue *Schema) SHA512() *Schema {
	schemaValue.rules.Put(rule.SHA512(CodeSHA512))
	return schemaValue
}

func (schemaValue *Schema) SHA512_224() *Schema {
	schemaValue.rules.Put(rule.SHA512_224(CodeSHA512224))
	return schemaValue
}

func (schemaValue *Schema) SHA512224() *Schema {
	return schemaValue.SHA512_224()
}

func (schemaValue *Schema) SHA512_256() *Schema {
	schemaValue.rules.Put(rule.SHA512_256(CodeSHA512256))
	return schemaValue
}

func (schemaValue *Schema) SHA512256() *Schema {
	return schemaValue.SHA512_256()
}

func (schemaValue *Schema) SHA3_224() *Schema {
	schemaValue.rules.Put(rule.SHA3_224(CodeSHA3224))
	return schemaValue
}

func (schemaValue *Schema) SHA3224() *Schema {
	return schemaValue.SHA3_224()
}

func (schemaValue *Schema) SHA3_256() *Schema {
	schemaValue.rules.Put(rule.SHA3_256(CodeSHA3256))
	return schemaValue
}

func (schemaValue *Schema) SHA3256() *Schema {
	return schemaValue.SHA3_256()
}

func (schemaValue *Schema) SHA3_384() *Schema {
	schemaValue.rules.Put(rule.SHA3_384(CodeSHA3384))
	return schemaValue
}

func (schemaValue *Schema) SHA3384() *Schema {
	return schemaValue.SHA3_384()
}

func (schemaValue *Schema) SHA3_512() *Schema {
	schemaValue.rules.Put(rule.SHA3_512(CodeSHA3512))
	return schemaValue
}

func (schemaValue *Schema) SHA3512() *Schema {
	return schemaValue.SHA3_512()
}

func (schemaValue *Schema) RIPEMD160() *Schema {
	schemaValue.rules.Put(rule.RIPEMD160(CodeRIPEMD160))
	return schemaValue
}

func (schemaValue *Schema) BLAKE2B_256() *Schema {
	schemaValue.rules.Put(rule.BLAKE2B_256(CodeBLAKE2B256))
	return schemaValue
}

func (schemaValue *Schema) Blake2b256() *Schema {
	return schemaValue.BLAKE2B_256()
}

func (schemaValue *Schema) BLAKE2B_384() *Schema {
	schemaValue.rules.Put(rule.BLAKE2B_384(CodeBLAKE2B384))
	return schemaValue
}

func (schemaValue *Schema) Blake2b384() *Schema {
	return schemaValue.BLAKE2B_384()
}

func (schemaValue *Schema) BLAKE2B_512() *Schema {
	schemaValue.rules.Put(rule.BLAKE2B_512(CodeBLAKE2B512))
	return schemaValue
}

func (schemaValue *Schema) Blake2b512() *Schema {
	return schemaValue.BLAKE2B_512()
}

func (schemaValue *Schema) BLAKE2S_256() *Schema {
	schemaValue.rules.Put(rule.BLAKE2S_256(CodeBLAKE2S256))
	return schemaValue
}

func (schemaValue *Schema) Blake2s256() *Schema {
	return schemaValue.BLAKE2S_256()
}

func (schemaValue *Schema) Default(value string) *Schema {
	schemaValue.defaultProvider = defaults.Value(value)
	return schemaValue
}

func (schemaValue *Schema) DefaultFunc(fn func() string) *Schema {
	schemaValue.defaultProvider = defaults.Func(fn)
	return schemaValue
}

func (schemaValue *Schema) Custom(ruleFn ruleset.RuleFn[string]) *Schema {
	if ruleFn != nil {
		schemaValue.customRules = append(schemaValue.customRules, ruleFn)
	}
	return schemaValue
}

func (schemaValue *Schema) Validate(input any, optionList ...schema.Option) (string, error) {
	options := schema.ApplyOptions(optionList...)
	return schemaValue.validateWithOptions(input, options)
}

func (schemaValue *Schema) ValidateAny(input any, options schema.Options) (any, error) {
	return schemaValue.validateWithOptions(input, options)
}

func (schemaValue *Schema) OutputType() reflect.Type {
	return reflect.TypeOf((*string)(nil)).Elem()
}
