// schema/text/text.go
package text

import (
	"encoding/base64"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"math"
	"net"
	"net/mail"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"runtime"
	"strconv"
	"strings"

	"github.com/leandroluk/gox/validate/internal/ast"
	"github.com/leandroluk/gox/validate/internal/defaults"
	"github.com/leandroluk/gox/validate/internal/engine"
	"github.com/leandroluk/gox/validate/internal/reflection"
	"github.com/leandroluk/gox/validate/internal/ruleset"
	"github.com/leandroluk/gox/validate/internal/types"
	"github.com/leandroluk/gox/validate/schema"
)

const (
	CodeRequired = "string.required"
	CodeType     = "string.type"

	CodeLen = "string.len"
	CodeMin = "string.min"
	CodeMax = "string.max"

	CodeEq  = "string.eq"
	CodeNe  = "string.ne"
	CodeEqI = "string.eq_ignore_case"
	CodeNeI = "string.ne_ignore_case"

	CodeOneOf = "text.oneof"

	CodeContains      = "string.contains"
	CodeExcludes      = "string.excludes"
	CodeStartsWith    = "string.startswith"
	CodeNotStartsWith = "string.not_startswith"
	CodeEndsWith      = "string.endswith"
	CodeNotEndsWith   = "string.not_endswith"

	CodeLowercase = "string.lowercase"
	CodeUppercase = "string.uppercase"

	CodePattern = "string.pattern"

	CodeEmail      = "string.email"
	CodeURL        = "string.url"
	CodeHTTPURL    = "string.http_url"
	CodeURI        = "string.uri"
	CodeURNRFC2141 = "string.urn_rfc2141"

	CodeFile     = "string.file"
	CodeFilePath = "string.filepath"
	CodeDir      = "string.dir"
	CodeDirPath  = "string.dirpath"
	CodeImage    = "string.image"

	CodeUUID  = "string.uuid"
	CodeUUID3 = "string.uuid3"
	CodeUUID4 = "string.uuid4"
	CodeUUID5 = "string.uuid5"

	CodeIP       = "string.ip"
	CodeIPv4     = "string.ipv4"
	CodeIPv6     = "string.ipv6"
	CodeCIDR     = "string.cidr"
	CodeMAC      = "string.mac"
	CodeHostname = "string.hostname"
	CodeFQDN     = "string.fqdn"
	CodePort     = "string.port"

	CodeNumeric     = "string.numeric"
	CodeNumber      = "string.number"
	CodeHexadecimal = "string.hexadecimal"
	CodeHexColor    = "string.hexcolor"
	CodeRGB         = "string.rgb"
	CodeRGBA        = "string.rgba"
	CodeHSL         = "string.hsl"
	CodeHSLA        = "string.hsla"

	CodeBase64       = "string.base64"
	CodeBase64URL    = "string.base64url"
	CodeBase64RawURL = "string.base64rawurl"
	CodeDataURI      = "string.datauri"
	CodeASCII        = "string.ascii"
	CodePrintASCII   = "string.printascii"
	CodeMultibyte    = "string.multibyte"

	CodeCreditCard   = "string.credit_card"
	CodeLuhnChecksum = "string.luhn_checksum"

	CodeISBN   = "string.isbn"
	CodeISBN10 = "string.isbn10"
	CodeISBN13 = "string.isbn13"
	CodeISSN   = "string.issn"

	CodeE164   = "string.e164"
	CodeSemVer = "string.semver"
	CodeCVE    = "string.cve"

	CodeMD4        = "string.md4"
	CodeMD5        = "string.md5"
	CodeSHA1       = "string.sha1"
	CodeSHA224     = "string.sha224"
	CodeSHA256     = "string.sha256"
	CodeSHA384     = "string.sha384"
	CodeSHA512     = "string.sha512"
	CodeSHA512224  = "string.sha512_224"
	CodeSHA512256  = "string.sha512_256"
	CodeSHA3224    = "string.sha3_224"
	CodeSHA3256    = "string.sha3_256"
	CodeSHA3384    = "string.sha3_384"
	CodeSHA3512    = "string.sha3_512"
	CodeRIPEMD160  = "string.ripemd160"
	CodeBLAKE2B256 = "string.blake2b_256"
	CodeBLAKE2B384 = "string.blake2b_384"
	CodeBLAKE2B512 = "string.blake2b_512"
	CodeBLAKE2S256 = "string.blake2s_256"
)

var (
	uuidRegex          = regexp.MustCompile(`(?i)^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)
	uuid3Regex         = regexp.MustCompile(`(?i)^[0-9a-f]{8}-[0-9a-f]{4}-3[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)
	uuid4Regex         = regexp.MustCompile(`(?i)^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)
	uuid5Regex         = regexp.MustCompile(`(?i)^[0-9a-f]{8}-[0-9a-f]{4}-5[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)
	urnRFC2141Regex    = regexp.MustCompile(`(?i)^urn:[a-z0-9][a-z0-9-]{1,31}:[^\s]+$`)
	hexColorRegex      = regexp.MustCompile(`(?i)^#[0-9a-f]{3}([0-9a-f]{3})?$`)
	rgbRegex           = regexp.MustCompile(`(?i)^rgb\(\s*([0-9]{1,3})\s*,\s*([0-9]{1,3})\s*,\s*([0-9]{1,3})\s*\)$`)
	rgbaRegex          = regexp.MustCompile(`(?i)^rgba\(\s*([0-9]{1,3})\s*,\s*([0-9]{1,3})\s*,\s*([0-9]{1,3})\s*,\s*([0-9]+(?:\.[0-9]+)?|\.[0-9]+)\s*\)$`)
	hslRegex           = regexp.MustCompile(`(?i)^hsl\(\s*([0-9]{1,3})\s*,\s*([0-9]{1,3})%\s*,\s*([0-9]{1,3})%\s*\)$`)
	hslaRegex          = regexp.MustCompile(`(?i)^hsla\(\s*([0-9]{1,3})\s*,\s*([0-9]{1,3})%\s*,\s*([0-9]{1,3})%\s*,\s*([0-9]+(?:\.[0-9]+)?|\.[0-9]+)\s*\)$`)
	e164Regex          = regexp.MustCompile(`^\+[1-9]\d{1,14}$`)
	semVerRegex        = regexp.MustCompile(`^(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)(?:-([0-9A-Za-z-]+(?:\.[0-9A-Za-z-]+)*))?(?:\+([0-9A-Za-z-]+(?:\.[0-9A-Za-z-]+)*))?$`)
	cveRegex           = regexp.MustCompile(`(?i)^CVE-\d{4}-\d{4,}$`)
	hostnameLabelRegex = regexp.MustCompile(`^[a-zA-Z0-9]([a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?$`)
	numericRegex       = regexp.MustCompile(`^\d+$`)
	hexadecimalRegex   = regexp.MustCompile(`(?i)^(0x)?[0-9a-f]+$`)
	asciiRegex         = regexp.MustCompile(`^[\x00-\x7F]*$`)
	printASCIIRegex    = regexp.MustCompile(`^[\x20-\x7E]*$`)
	multibyteRegex     = regexp.MustCompile(`[^\x00-\x7F]`)
)

var Msg = struct {
	ASCII         string
	Base64        string
	Base64RawURL  string
	Base64URL     string
	BLAKE2B256    string
	BLAKE2B384    string
	BLAKE2B512    string
	BLAKE2S256    string
	CIDR          string
	Contains      string
	CreditCard    string
	CVE           string
	DataURI       string
	Dir           string
	DirPath       string
	E164          string
	Email         string
	EndsWith      string
	Eq            string
	EqIgnoreCase  string
	Excludes      string
	File          string
	FilePath      string
	FQDN          string
	HexColor      string
	Hexadecimal   string
	Hostname      string
	HSL           string
	HSLA          string
	HTTPURL       string
	Image         string
	IP            string
	IPv4          string
	IPv6          string
	ISBN          string
	ISBN10        string
	ISBN13        string
	ISSN          string
	Len           string
	Lowercase     string
	LuhnChecksum  string
	MAC           string
	Max           string
	MD4           string
	MD5           string
	Min           string
	Multibyte     string
	Ne            string
	NeIgnoreCase  string
	NotEndsWith   string
	NotStartsWith string
	Number        string
	Numeric       string
	OneOf         string
	Pattern       string
	Port          string
	PrintASCII    string
	RGB           string
	RGBA          string
	RIPEMD160     string
	SemVer        string
	SHA1          string
	SHA224        string
	SHA256        string
	SHA3224       string
	SHA3256       string
	SHA3384       string
	SHA3512       string
	SHA384        string
	SHA512        string
	SHA512224     string
	SHA512256     string
	StartsWith    string
	Uppercase     string
	URI           string
	URL           string
	URNRFC2141    string
	UUID          string
	UUID3         string
	UUID4         string
	UUID5         string
}{
	ASCII:         "invalid ascii",
	Base64:        "invalid base64",
	Base64RawURL:  "invalid base64rawurl",
	Base64URL:     "invalid base64url",
	BLAKE2B256:    "invalid blake2b-256",
	BLAKE2B384:    "invalid blake2b-384",
	BLAKE2B512:    "invalid blake2b-512",
	BLAKE2S256:    "invalid blake2s-256",
	CIDR:          "invalid cidr",
	Contains:      "must contain",
	CreditCard:    "invalid credit card",
	CVE:           "invalid cve",
	DataURI:       "invalid data uri",
	Dir:           "invalid dir",
	DirPath:       "invalid dirpath",
	E164:          "invalid e164",
	Email:         "invalid email",
	EndsWith:      "must end with",
	Eq:            "must be equal",
	EqIgnoreCase:  "must be equal (ignore case)",
	Excludes:      "must not contain",
	File:          "invalid file",
	FilePath:      "invalid filepath",
	FQDN:          "invalid fqdn",
	HexColor:      "invalid hex color",
	Hexadecimal:   "invalid hexadecimal",
	Hostname:      "invalid hostname",
	HSL:           "invalid hsl",
	HSLA:          "invalid hsla",
	HTTPURL:       "invalid http url",
	Image:         "invalid image",
	IP:            "invalid ip",
	IPv4:          "invalid ipv4",
	IPv6:          "invalid ipv6",
	ISBN:          "invalid isbn",
	ISBN10:        "invalid isbn10",
	ISBN13:        "invalid isbn13",
	ISSN:          "invalid issn",
	Len:           "invalid length",
	Lowercase:     "must be lowercase",
	LuhnChecksum:  "invalid luhn checksum",
	MAC:           "invalid mac",
	Max:           "too long",
	MD4:           "invalid md4",
	MD5:           "invalid md5",
	Min:           "too short",
	Multibyte:     "invalid multibyte",
	Ne:            "must not be equal",
	NeIgnoreCase:  "must not be equal (ignore case)",
	NotEndsWith:   "must not end with",
	NotStartsWith: "must not start with",
	Number:        "invalid number",
	Numeric:       "invalid numeric",
	OneOf:         "not allowed",
	Pattern:       "does not match pattern",
	Port:          "invalid port",
	PrintASCII:    "invalid printascii",
	RGB:           "invalid rgb",
	RGBA:          "invalid rgba",
	RIPEMD160:     "invalid ripemd160",
	SemVer:        "invalid semver",
	SHA1:          "invalid sha1",
	SHA224:        "invalid sha224",
	SHA256:        "invalid sha256",
	SHA3224:       "invalid sha3-224",
	SHA3256:       "invalid sha3-256",
	SHA3384:       "invalid sha3-384",
	SHA3512:       "invalid sha3-512",
	SHA384:        "invalid sha384",
	SHA512:        "invalid sha512",
	SHA512224:     "invalid sha512/224",
	SHA512256:     "invalid sha512/256",
	StartsWith:    "must start with",
	Uppercase:     "must be uppercase",
	URI:           "invalid uri",
	URL:           "invalid url",
	URNRFC2141:    "invalid urn",
	UUID:          "invalid uuid",
	UUID3:         "invalid uuid3",
	UUID4:         "invalid uuid4",
	UUID5:         "invalid uuid5",
}

func newRule(code string, message string, validate func(actual string) (bool, types.AnyMap)) ruleset.Rule[string] {
	return ruleset.New(code, func(actual string, context *engine.Context) (string, bool) {
		ok, meta := validate(actual)
		if ok {
			return actual, false
		}
		stop := context.AddIssue(code, message, meta)
		return actual, stop
	})
}

func digestRule(code string, message string, sizeBytes int) ruleset.Rule[string] {
	return newRule(code, message, func(actual string) (bool, types.AnyMap) {
		if actual == "" || sizeBytes <= 0 {
			return false, types.AnyMap{"actual": actual}
		}
		if len(actual) == sizeBytes*2 {
			ok := true
			for i := 0; i < len(actual); i++ {
				if !isHexDigit(actual[i]) {
					ok = false
					break
				}
			}
			if ok {
				return true, types.AnyMap{"actual": actual}
			}
		}
		if decoded, err := base64.StdEncoding.DecodeString(actual); err == nil && len(decoded) == sizeBytes {
			return true, types.AnyMap{"actual": actual}
		}
		if decoded, err := base64.RawStdEncoding.DecodeString(actual); err == nil && len(decoded) == sizeBytes {
			return true, types.AnyMap{"actual": actual}
		}
		if decoded, err := base64.URLEncoding.DecodeString(actual); err == nil && len(decoded) == sizeBytes {
			return true, types.AnyMap{"actual": actual}
		}
		if decoded, err := base64.RawURLEncoding.DecodeString(actual); err == nil && len(decoded) == sizeBytes {
			return true, types.AnyMap{"actual": actual}
		}
		return false, types.AnyMap{"actual": actual}
	})
}

func digitsForChecksum(value string, allowSpaceDash bool) (string, bool) {
	if value == "" {
		return "", false
	}
	var builder strings.Builder
	builder.Grow(len(value))
	for i := 0; i < len(value); i++ {
		ch := value[i]
		if ch >= '0' && ch <= '9' {
			builder.WriteByte(ch)
			continue
		}
		if allowSpaceDash && (ch == ' ' || ch == '-') {
			continue
		}
		return "", false
	}
	out := builder.String()
	if out == "" {
		return "", false
	}
	return out, true
}

func isHexDigit(ch byte) bool {
	return (ch >= '0' && ch <= '9') || (ch >= 'a' && ch <= 'f') || (ch >= 'A' && ch <= 'F')
}

func parseIntRange(value string, min int, max int) bool {
	parsed, err := strconv.Atoi(value)
	return err == nil && parsed >= min && parsed <= max
}

func parseFloatRange(value string, min float64, max float64) bool {
	parsed, err := strconv.ParseFloat(value, 64)
	return err == nil && !math.IsNaN(parsed) && !math.IsInf(parsed, 0) && parsed >= min && parsed <= max
}

func parseTextValue(context *engine.Context, value ast.Value) (string, bool) {
	actual := value.Kind.String()
	switch value.Kind {
	case ast.KindString:
		return value.String, false

	case ast.KindNumber:
		if context.Options.Coerce {
			return value.Number, false
		}
		stop := context.AddIssue(CodeType, "expected string", map[string]any{"expected": "string", "actual": actual})
		return "", stop

	case ast.KindBoolean:
		if context.Options.Coerce {
			if value.Boolean {
				return "true", false
			}
			return "false", false
		}
		stop := context.AddIssue(CodeType, "expected string", map[string]any{"expected": "string", "actual": actual})
		return "", stop

	default:
		stop := context.AddIssue(CodeType, "expected string", map[string]any{"expected": "string", "actual": actual})
		return "", stop
	}
}

func luhnValid(digits string) bool {
	sum := 0
	double := false
	for i := len(digits) - 1; i >= 0; i-- {
		d := int(digits[i] - '0')
		if double {
			d = d * 2
			if d > 9 {
				d -= 9
			}
		}
		sum += d
		double = !double
	}
	return sum%10 == 0
}

func isHostname(value string, requireDot bool) bool {
	if value == "" || strings.TrimSpace(value) != value {
		return false
	}
	if before, ok := strings.CutSuffix(value, "."); ok {
		value = before
	}
	if value == "" || len(value) > 253 {
		return false
	}
	labels := strings.Split(value, ".")
	if requireDot && len(labels) < 2 {
		return false
	}
	for _, label := range labels {
		if !hostnameLabelRegex.MatchString(label) {
			return false
		}
	}
	return true
}

func stripISBN(value string) string {
	var b strings.Builder
	b.Grow(len(value))
	for i := 0; i < len(value); i++ {
		ch := value[i]
		if (ch >= '0' && ch <= '9') || ch == 'X' || ch == 'x' {
			b.WriteByte(ch)
		}
	}
	return b.String()
}

func stripDigitsOnly(value string) string {
	var b strings.Builder
	b.Grow(len(value))
	for i := 0; i < len(value); i++ {
		ch := value[i]
		if ch >= '0' && ch <= '9' {
			b.WriteByte(ch)
		}
	}
	return b.String()
}

func isISBN10(value string) bool {
	cleaned := stripISBN(value)
	if len(cleaned) != 10 {
		return false
	}
	total := 0
	for i := range 9 {
		ch := cleaned[i]
		if ch < '0' || ch > '9' {
			return false
		}
		total += (i + 1) * int(ch-'0')
	}
	last := cleaned[9]
	if last == 'X' || last == 'x' {
		total += 10 * 10
	} else if last >= '0' && last <= '9' {
		total += 10 * int(last-'0')
	} else {
		return false
	}
	return total%11 == 0
}

func isISBN13(value string) bool {
	cleaned := stripDigitsOnly(value)
	if len(cleaned) != 13 {
		return false
	}
	total := 0
	for i := range 12 {
		ch := cleaned[i]
		if ch < '0' || ch > '9' {
			return false
		}
		d := int(ch - '0')
		if i%2 == 0 {
			total += d
		} else {
			total += 3 * d
		}
	}
	check := (10 - (total % 10)) % 10
	last := cleaned[12]
	if last < '0' || last > '9' {
		return false
	}
	return check == int(last-'0')
}

func isValidPath(value string) bool {
	if value == "" || strings.TrimSpace(value) != value || strings.IndexByte(value, 0) >= 0 {
		return false
	}
	if runtime.GOOS == "windows" {
		for i := 0; i < len(value); i++ {
			ch := value[i]
			if ch < 0x20 {
				return false
			}
			switch ch {
			case '<', '>', '"', '|', '?', '*':
				return false
			case ':':
				if i == 1 {
					c := value[0]
					if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') {
						continue
					}
				}
				return false
			}
		}
	}
	_ = filepath.Clean(value)
	return true
}

func isFile(value string) bool {
	if !isValidPath(value) {
		return false
	}
	info, err := os.Stat(value)
	return err == nil && !info.IsDir()
}

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

func (s *Schema) Required() *Schema {
	s.required = true
	return s
}

func (s *Schema) IsDefault() *Schema {
	s.isDefault = true
	return s
}

func (s *Schema) Len(length int) *Schema {
	if length < 0 {
		s.rules.Remove(CodeLen)
		return s
	}
	s.rules.Put(Rule.Len(CodeLen, length))
	return s
}

func (s *Schema) Min(length int) *Schema {
	if length < 0 {
		s.rules.Remove(CodeMin)
		return s
	}
	s.rules.Put(Rule.Min(CodeMin, length))
	return s
}

func (s *Schema) Max(length int) *Schema {
	if length < 0 {
		s.rules.Remove(CodeMax)
		return s
	}
	s.rules.Put(Rule.Max(CodeMax, length))
	return s
}

func (s *Schema) Eq(value string) *Schema {
	s.rules.Put(Rule.Eq(CodeEq, value))
	return s
}

func (s *Schema) Ne(value string) *Schema {
	s.rules.Put(Rule.Ne(CodeNe, value))
	return s
}

func (s *Schema) EqIgnoreCase(value string) *Schema {
	s.rules.Put(Rule.EqIgnoreCase(CodeEqI, value))
	return s
}

func (s *Schema) NeIgnoreCase(value string) *Schema {
	s.rules.Put(Rule.NeIgnoreCase(CodeNeI, value))
	return s
}

func (s *Schema) OneOf(values ...string) *Schema {
	if len(values) == 0 {
		s.rules.Remove(CodeOneOf)
		return s
	}
	s.rules.Put(Rule.OneOf(CodeOneOf, values...))
	return s
}

func (s *Schema) Enum(values ...any) *Schema {
	if len(values) == 0 {
		return s
	}
	strValues := make([]string, len(values))
	for i, v := range values {
		rv := reflect.ValueOf(v)
		if rv.Kind() == reflect.String {
			strValues[i] = rv.String()
		} else {
			strValues[i] = fmt.Sprint(v)
		}
	}
	return s.OneOf(strValues...)
}

func (s *Schema) Contains(value string) *Schema {
	if value == "" {
		s.rules.Remove(CodeContains)
		return s
	}
	s.rules.Put(Rule.Contains(CodeContains, value))
	return s
}

func (s *Schema) Excludes(value string) *Schema {
	if value == "" {
		s.rules.Remove(CodeExcludes)
		return s
	}
	s.rules.Put(Rule.Excludes(CodeExcludes, value))
	return s
}

func (s *Schema) StartsWith(value string) *Schema {
	if value == "" {
		s.rules.Remove(CodeStartsWith)
		return s
	}
	s.rules.Put(Rule.StartsWith(CodeStartsWith, value))
	return s
}

func (s *Schema) NotStartsWith(value string) *Schema {
	if value == "" {
		s.rules.Remove(CodeNotStartsWith)
		return s
	}
	s.rules.Put(Rule.NotStartsWith(CodeNotStartsWith, value))
	return s
}

func (s *Schema) EndsWith(value string) *Schema {
	if value == "" {
		s.rules.Remove(CodeEndsWith)
		return s
	}
	s.rules.Put(Rule.EndsWith(CodeEndsWith, value))
	return s
}

func (s *Schema) NotEndsWith(value string) *Schema {
	if value == "" {
		s.rules.Remove(CodeNotEndsWith)
		return s
	}
	s.rules.Put(Rule.NotEndsWith(CodeNotEndsWith, value))
	return s
}

func (s *Schema) Lowercase() *Schema {
	s.rules.Put(Rule.Lowercase(CodeLowercase))
	return s
}

func (s *Schema) Uppercase() *Schema {
	s.rules.Put(Rule.Uppercase(CodeUppercase))
	return s
}

func (s *Schema) Pattern(value string) *Schema {
	if value == "" {
		s.rules.Remove(CodePattern)
		return s
	}
	compiled, err := regexp.Compile(value)
	if err != nil {
		return s
	}
	s.rules.Put(Rule.Pattern(CodePattern, compiled))
	return s
}

func (s *Schema) PatternRegexp(value *regexp.Regexp) *Schema {
	if value == nil {
		s.rules.Remove(CodePattern)
		return s
	}
	s.rules.Put(Rule.Pattern(CodePattern, value))
	return s
}

func (s *Schema) Email() *Schema {
	s.rules.Put(Rule.Email(CodeEmail))
	return s
}

func (s *Schema) URL() *Schema {
	s.rules.Put(Rule.URL(CodeURL))
	return s
}

func (s *Schema) HTTPURL() *Schema {
	s.rules.Put(Rule.HTTPURL(CodeHTTPURL))
	return s
}

func (s *Schema) HttpURL() *Schema {
	return s.HTTPURL()
}

func (s *Schema) URI() *Schema {
	s.rules.Put(Rule.URI(CodeURI))
	return s
}

func (s *Schema) URNRFC2141() *Schema {
	s.rules.Put(Rule.URNRFC2141(CodeURNRFC2141))
	return s
}

func (s *Schema) UrnRFC2141() *Schema {
	return s.URNRFC2141()
}

func (s *Schema) File() *Schema {
	s.rules.Put(Rule.File(CodeFile))
	return s
}

func (s *Schema) FilePath() *Schema {
	s.rules.Put(Rule.FilePath(CodeFilePath))
	return s
}

func (s *Schema) Filepath() *Schema {
	return s.FilePath()
}

func (s *Schema) Dir() *Schema {
	s.rules.Put(Rule.Dir(CodeDir))
	return s
}

func (s *Schema) DirPath() *Schema {
	s.rules.Put(Rule.DirPath(CodeDirPath))
	return s
}

func (s *Schema) Dirpath() *Schema {
	return s.DirPath()
}

func (s *Schema) Image() *Schema {
	s.rules.Put(Rule.Image(CodeImage))
	return s
}

func (s *Schema) UUID() *Schema {
	s.rules.Put(Rule.UUID(CodeUUID))
	return s
}

func (s *Schema) UUID3() *Schema {
	s.rules.Put(Rule.UUID3(CodeUUID3))
	return s
}

func (s *Schema) UUID4() *Schema {
	s.rules.Put(Rule.UUID4(CodeUUID4))
	return s
}

func (s *Schema) UUID5() *Schema {
	s.rules.Put(Rule.UUID5(CodeUUID5))
	return s
}

func (s *Schema) IP() *Schema {
	s.rules.Put(Rule.IP(CodeIP))
	return s
}

func (s *Schema) IPv4() *Schema {
	s.rules.Put(Rule.IPv4(CodeIPv4))
	return s
}

func (s *Schema) IPv6() *Schema {
	s.rules.Put(Rule.IPv6(CodeIPv6))
	return s
}

func (s *Schema) CIDR() *Schema {
	s.rules.Put(Rule.CIDR(CodeCIDR))
	return s
}

func (s *Schema) MAC() *Schema {
	s.rules.Put(Rule.MAC(CodeMAC))
	return s
}

func (s *Schema) Hostname() *Schema {
	s.rules.Put(Rule.Hostname(CodeHostname))
	return s
}

func (s *Schema) FQDN() *Schema {
	s.rules.Put(Rule.FQDN(CodeFQDN))
	return s
}

func (s *Schema) Port() *Schema {
	s.rules.Put(Rule.Port(CodePort))
	return s
}

func (s *Schema) Numeric() *Schema {
	s.rules.Put(Rule.Numeric(CodeNumeric))
	return s
}

func (s *Schema) Number() *Schema {
	s.rules.Put(Rule.Number(CodeNumber))
	return s
}

func (s *Schema) Hexadecimal() *Schema {
	s.rules.Put(Rule.Hexadecimal(CodeHexadecimal))
	return s
}

func (s *Schema) HexColor() *Schema {
	s.rules.Put(Rule.HexColor(CodeHexColor))
	return s
}

func (s *Schema) RGB() *Schema {
	s.rules.Put(Rule.RGB(CodeRGB))
	return s
}

func (s *Schema) RGBA() *Schema {
	s.rules.Put(Rule.RGBA(CodeRGBA))
	return s
}

func (s *Schema) HSL() *Schema {
	s.rules.Put(Rule.HSL(CodeHSL))
	return s
}

func (s *Schema) HSLA() *Schema {
	s.rules.Put(Rule.HSLA(CodeHSLA))
	return s
}

func (s *Schema) Base64() *Schema {
	s.rules.Put(Rule.Base64(CodeBase64))
	return s
}

func (s *Schema) Base64URL() *Schema {
	s.rules.Put(Rule.Base64URL(CodeBase64URL))
	return s
}

func (s *Schema) Base64Url() *Schema {
	return s.Base64URL()
}

func (s *Schema) Base64RawURL() *Schema {
	s.rules.Put(Rule.Base64RawURL(CodeBase64RawURL))
	return s
}

func (s *Schema) Base64RawUrl() *Schema {
	return s.Base64RawURL()
}

func (s *Schema) DataURI() *Schema {
	s.rules.Put(Rule.DataURI(CodeDataURI))
	return s
}

func (s *Schema) ASCII() *Schema {
	s.rules.Put(Rule.ASCII(CodeASCII))
	return s
}

func (s *Schema) PrintASCII() *Schema {
	s.rules.Put(Rule.PrintASCII(CodePrintASCII))
	return s
}

func (s *Schema) Multibyte() *Schema {
	s.rules.Put(Rule.Multibyte(CodeMultibyte))
	return s
}

func (s *Schema) CreditCard() *Schema {
	s.rules.Put(Rule.CreditCard(CodeCreditCard))
	return s
}

func (s *Schema) LuhnChecksum() *Schema {
	s.rules.Put(Rule.LuhnChecksum(CodeLuhnChecksum))
	return s
}

func (s *Schema) ISBN() *Schema {
	s.rules.Put(Rule.ISBN(CodeISBN))
	return s
}

func (s *Schema) ISBN10() *Schema {
	s.rules.Put(Rule.ISBN10(CodeISBN10))
	return s
}

func (s *Schema) ISBN13() *Schema {
	s.rules.Put(Rule.ISBN13(CodeISBN13))
	return s
}

func (s *Schema) ISSN() *Schema {
	s.rules.Put(Rule.ISSN(CodeISSN))
	return s
}

func (s *Schema) E164() *Schema {
	s.rules.Put(Rule.E164(CodeE164))
	return s
}

func (s *Schema) SemVer() *Schema {
	s.rules.Put(Rule.SemVer(CodeSemVer))
	return s
}

func (s *Schema) Semver() *Schema {
	return s.SemVer()
}

func (s *Schema) CVE() *Schema {
	s.rules.Put(Rule.CVE(CodeCVE))
	return s
}

func (s *Schema) MD4() *Schema {
	s.rules.Put(Rule.MD4(CodeMD4))
	return s
}

func (s *Schema) MD5() *Schema {
	s.rules.Put(Rule.MD5(CodeMD5))
	return s
}

func (s *Schema) SHA1() *Schema {
	s.rules.Put(Rule.SHA1(CodeSHA1))
	return s
}

func (s *Schema) SHA224() *Schema {
	s.rules.Put(Rule.SHA224(CodeSHA224))
	return s
}

func (s *Schema) SHA256() *Schema {
	s.rules.Put(Rule.SHA256(CodeSHA256))
	return s
}

func (s *Schema) SHA384() *Schema {
	s.rules.Put(Rule.SHA384(CodeSHA384))
	return s
}

func (s *Schema) SHA512() *Schema {
	s.rules.Put(Rule.SHA512(CodeSHA512))
	return s
}

func (s *Schema) SHA512_224() *Schema {
	s.rules.Put(Rule.SHA512_224(CodeSHA512224))
	return s
}

func (s *Schema) SHA512224() *Schema {
	return s.SHA512_224()
}

func (s *Schema) SHA512_256() *Schema {
	s.rules.Put(Rule.SHA512_256(CodeSHA512256))
	return s
}

func (s *Schema) SHA512256() *Schema {
	return s.SHA512_256()
}

func (s *Schema) SHA3_224() *Schema {
	s.rules.Put(Rule.SHA3_224(CodeSHA3224))
	return s
}

func (s *Schema) SHA3224() *Schema {
	return s.SHA3_224()
}

func (s *Schema) SHA3_256() *Schema {
	s.rules.Put(Rule.SHA3_256(CodeSHA3256))
	return s
}

func (s *Schema) SHA3256() *Schema {
	return s.SHA3_256()
}

func (s *Schema) SHA3_384() *Schema {
	s.rules.Put(Rule.SHA3_384(CodeSHA3384))
	return s
}

func (s *Schema) SHA3384() *Schema {
	return s.SHA3_384()
}

func (s *Schema) SHA3_512() *Schema {
	s.rules.Put(Rule.SHA3_512(CodeSHA3512))
	return s
}

func (s *Schema) SHA3512() *Schema {
	return s.SHA3_512()
}

func (s *Schema) RIPEMD160() *Schema {
	s.rules.Put(Rule.RIPEMD160(CodeRIPEMD160))
	return s
}

func (s *Schema) BLAKE2B_256() *Schema {
	s.rules.Put(Rule.BLAKE2B_256(CodeBLAKE2B256))
	return s
}

func (s *Schema) Blake2b256() *Schema {
	return s.BLAKE2B_256()
}

func (s *Schema) BLAKE2B_384() *Schema {
	s.rules.Put(Rule.BLAKE2B_384(CodeBLAKE2B384))
	return s
}

func (s *Schema) Blake2b384() *Schema {
	return s.BLAKE2B_384()
}

func (s *Schema) BLAKE2B_512() *Schema {
	s.rules.Put(Rule.BLAKE2B_512(CodeBLAKE2B512))
	return s
}

func (s *Schema) Blake2b512() *Schema {
	return s.BLAKE2B_512()
}

func (s *Schema) BLAKE2S_256() *Schema {
	s.rules.Put(Rule.BLAKE2S_256(CodeBLAKE2S256))
	return s
}

func (s *Schema) Blake2s256() *Schema {
	return s.BLAKE2S_256()
}

func (s *Schema) Default(value string) *Schema {
	s.defaultProvider = defaults.Value(value)
	return s
}

func (s *Schema) DefaultFunc(fn func() string) *Schema {
	s.defaultProvider = defaults.Func(fn)
	return s
}

func (s *Schema) Custom(ruleFn ruleset.RuleFn[string]) *Schema {
	if ruleFn != nil {
		s.customRules = append(s.customRules, ruleFn)
	}
	return s
}

func (s *Schema) Validate(input any, optionList ...schema.Option) (string, error) {
	options := schema.ApplyOptions(optionList...)
	return s.validateWithOptions(input, options)
}

func (s *Schema) ValidateAny(input any, options schema.Options) (any, error) {
	return s.validateWithOptions(input, options)
}

func (s *Schema) OutputType() reflect.Type {
	return reflect.TypeFor[string]()
}

func (s *Schema) validateWithOptions(input any, options schema.Options) (string, error) {
	context := engine.NewContext(options)

	value, err := engine.InputToASTWithOptions(input, options)
	if err != nil {
		return "", err
	}

	output, _ := s.validateAST(context, value)
	return output, context.Error()
}

func (s *Schema) validateAST(context *engine.Context, value ast.Value) (string, bool) {
	if defaultValue, ok := defaults.Apply(value.Presence, context.Options, s.defaultProvider); ok {
		return defaultValue, false
	}

	if value.IsMissing() || value.IsNull() {
		if s.required {
			stop := context.AddIssue(CodeRequired, "required")
			return "", stop
		}
		return "", false
	}

	output, stopParse := parseTextValue(context, value)
	if stopParse {
		return "", true
	}

	if s.isDefault && reflection.IsDefault(output) {
		return output, false
	}

	output, stopRules := s.rules.ApplyAll(output, context)
	if stopRules {
		return output, true
	}

	if len(s.customRules) > 0 {
		if ruleset.Apply(output, context, s.customRules...) {
			return output, true
		}
	}

	return output, false
}

// --- rule functions ---

var Rule = struct {
	ASCII         func(code string) ruleset.Rule[string]
	Base64        func(code string) ruleset.Rule[string]
	Base64RawURL  func(code string) ruleset.Rule[string]
	Base64URL     func(code string) ruleset.Rule[string]
	BLAKE2B_256   func(code string) ruleset.Rule[string]
	BLAKE2B_384   func(code string) ruleset.Rule[string]
	BLAKE2B_512   func(code string) ruleset.Rule[string]
	BLAKE2S_256   func(code string) ruleset.Rule[string]
	CIDR          func(code string) ruleset.Rule[string]
	Contains      func(code string, needle string) ruleset.Rule[string]
	CreditCard    func(code string) ruleset.Rule[string]
	CVE           func(code string) ruleset.Rule[string]
	DataURI       func(code string) ruleset.Rule[string]
	Dir           func(code string) ruleset.Rule[string]
	DirPath       func(code string) ruleset.Rule[string]
	E164          func(code string) ruleset.Rule[string]
	Email         func(code string) ruleset.Rule[string]
	EndsWith      func(code string, suffix string) ruleset.Rule[string]
	Eq            func(code string, expected string) ruleset.Rule[string]
	EqIgnoreCase  func(code string, expected string) ruleset.Rule[string]
	Excludes      func(code string, needle string) ruleset.Rule[string]
	File          func(code string) ruleset.Rule[string]
	FilePath      func(code string) ruleset.Rule[string]
	FQDN          func(code string) ruleset.Rule[string]
	HexColor      func(code string) ruleset.Rule[string]
	Hexadecimal   func(code string) ruleset.Rule[string]
	Hostname      func(code string) ruleset.Rule[string]
	HSL           func(code string) ruleset.Rule[string]
	HSLA          func(code string) ruleset.Rule[string]
	HTTPURL       func(code string) ruleset.Rule[string]
	Image         func(code string) ruleset.Rule[string]
	IP            func(code string) ruleset.Rule[string]
	IPv4          func(code string) ruleset.Rule[string]
	IPv6          func(code string) ruleset.Rule[string]
	ISBN          func(code string) ruleset.Rule[string]
	ISBN10        func(code string) ruleset.Rule[string]
	ISBN13        func(code string) ruleset.Rule[string]
	ISSN          func(code string) ruleset.Rule[string]
	Len           func(code string, expected int) ruleset.Rule[string]
	Lowercase     func(code string) ruleset.Rule[string]
	LuhnChecksum  func(code string) ruleset.Rule[string]
	MAC           func(code string) ruleset.Rule[string]
	Max           func(code string, max int) ruleset.Rule[string]
	MD4           func(code string) ruleset.Rule[string]
	MD5           func(code string) ruleset.Rule[string]
	Min           func(code string, min int) ruleset.Rule[string]
	Multibyte     func(code string) ruleset.Rule[string]
	Ne            func(code string, disallowed string) ruleset.Rule[string]
	NeIgnoreCase  func(code string, disallowed string) ruleset.Rule[string]
	NotEndsWith   func(code string, suffix string) ruleset.Rule[string]
	NotStartsWith func(code string, prefix string) ruleset.Rule[string]
	Number        func(code string) ruleset.Rule[string]
	Numeric       func(code string) ruleset.Rule[string]
	OneOf         func(code string, values ...string) ruleset.Rule[string]
	Pattern       func(code string, pattern *regexp.Regexp) ruleset.Rule[string]
	Port          func(code string) ruleset.Rule[string]
	PrintASCII    func(code string) ruleset.Rule[string]
	RGB           func(code string) ruleset.Rule[string]
	RGBA          func(code string) ruleset.Rule[string]
	RIPEMD160     func(code string) ruleset.Rule[string]
	SemVer        func(code string) ruleset.Rule[string]
	SHA1          func(code string) ruleset.Rule[string]
	SHA224        func(code string) ruleset.Rule[string]
	SHA256        func(code string) ruleset.Rule[string]
	SHA3_224      func(code string) ruleset.Rule[string]
	SHA3_256      func(code string) ruleset.Rule[string]
	SHA3_384      func(code string) ruleset.Rule[string]
	SHA3_512      func(code string) ruleset.Rule[string]
	SHA384        func(code string) ruleset.Rule[string]
	SHA512        func(code string) ruleset.Rule[string]
	SHA512_224    func(code string) ruleset.Rule[string]
	SHA512_256    func(code string) ruleset.Rule[string]
	StartsWith    func(code string, prefix string) ruleset.Rule[string]
	Uppercase     func(code string) ruleset.Rule[string]
	URI           func(code string) ruleset.Rule[string]
	URL           func(code string) ruleset.Rule[string]
	URNRFC2141    func(code string) ruleset.Rule[string]
	UUID          func(code string) ruleset.Rule[string]
	UUID3         func(code string) ruleset.Rule[string]
	UUID4         func(code string) ruleset.Rule[string]
	UUID5         func(code string) ruleset.Rule[string]
}{
	ASCII: func(code string) ruleset.Rule[string] {
		return newRule(code, Msg.ASCII, func(actual string) (bool, types.AnyMap) {
			return asciiRegex.MatchString(actual), types.AnyMap{"actual": actual}
		})
	},
	Base64: func(code string) ruleset.Rule[string] {
		return newRule(code, Msg.Base64, func(actual string) (bool, types.AnyMap) {
			_, err := base64.StdEncoding.DecodeString(actual)
			return err == nil, types.AnyMap{"actual": actual}
		})
	},
	Base64RawURL: func(code string) ruleset.Rule[string] {
		return newRule(code, Msg.Base64RawURL, func(actual string) (bool, types.AnyMap) {
			_, err := base64.RawURLEncoding.DecodeString(actual)
			return err == nil, types.AnyMap{"actual": actual}
		})
	},
	Base64URL: func(code string) ruleset.Rule[string] {
		return newRule(code, Msg.Base64URL, func(actual string) (bool, types.AnyMap) {
			_, err := base64.URLEncoding.DecodeString(actual)
			return err == nil, types.AnyMap{"actual": actual}
		})
	},
	BLAKE2B_256: func(code string) ruleset.Rule[string] {
		return digestRule(code, Msg.BLAKE2B256, 32)
	},
	BLAKE2B_384: func(code string) ruleset.Rule[string] {
		return digestRule(code, Msg.BLAKE2B384, 48)
	},
	BLAKE2B_512: func(code string) ruleset.Rule[string] {
		return digestRule(code, Msg.BLAKE2B512, 64)
	},
	BLAKE2S_256: func(code string) ruleset.Rule[string] {
		return digestRule(code, Msg.BLAKE2S256, 32)
	},
	CIDR: func(code string) ruleset.Rule[string] {
		return newRule(code, Msg.CIDR, func(actual string) (bool, types.AnyMap) {
			_, _, err := net.ParseCIDR(actual)
			return err == nil, types.AnyMap{"actual": actual}
		})
	},
	Contains: func(code string, needle string) ruleset.Rule[string] {
		return newRule(code, Msg.Contains, func(actual string) (bool, types.AnyMap) {
			if strings.Contains(actual, needle) {
				return true, nil
			}
			return false, types.AnyMap{"expected": needle, "actual": actual}
		})
	},
	CreditCard: func(code string) ruleset.Rule[string] {
		return newRule(code, Msg.CreditCard, func(actual string) (bool, types.AnyMap) {
			digits, ok := digitsForChecksum(actual, true)
			if !ok || len(digits) < 12 || len(digits) > 19 {
				return false, types.AnyMap{"actual": actual}
			}
			return luhnValid(digits), types.AnyMap{"actual": actual}
		})
	},
	CVE: func(code string) ruleset.Rule[string] {
		return newRule(code, Msg.CVE, func(actual string) (bool, types.AnyMap) {
			return cveRegex.MatchString(actual), types.AnyMap{"actual": actual}
		})
	},
	DataURI: func(code string) ruleset.Rule[string] {
		return newRule(code, Msg.DataURI, func(actual string) (bool, types.AnyMap) {
			if !strings.HasPrefix(actual, "data:") {
				return false, types.AnyMap{"actual": actual}
			}
			commaIndex := strings.IndexByte(actual, ',')
			if commaIndex < 0 {
				return false, types.AnyMap{"actual": actual}
			}
			meta := actual[len("data:"):commaIndex]
			data := actual[commaIndex+1:]
			var ok bool
			if strings.Contains(meta, ";base64") {
				_, err := base64.StdEncoding.DecodeString(data)
				if err == nil {
					ok = true
				} else {
					_, err = base64.RawStdEncoding.DecodeString(data)
					ok = err == nil
				}
			} else {
				_, err := url.PathUnescape(data)
				ok = err == nil
			}
			return ok, types.AnyMap{"actual": actual}
		})
	},
	Dir: func(code string) ruleset.Rule[string] {
		return newRule(code, Msg.Dir, func(actual string) (bool, types.AnyMap) {
			if !isValidPath(actual) {
				return false, types.AnyMap{"actual": actual}
			}
			info, err := os.Stat(actual)
			if err != nil {
				return false, types.AnyMap{"actual": actual}
			}
			return info.IsDir(), types.AnyMap{"actual": actual}
		})
	},
	DirPath: func(code string) ruleset.Rule[string] {
		return newRule(code, Msg.DirPath, func(actual string) (bool, types.AnyMap) {
			if !isValidPath(actual) {
				return false, types.AnyMap{"actual": actual}
			}
			info, err := os.Stat(actual)
			if err == nil && info.IsDir() {
				return true, types.AnyMap{"actual": actual}
			}
			if len(actual) > 0 {
				last := actual[len(actual)-1]
				if last == '/' || last == '\\' {
					return true, types.AnyMap{"actual": actual}
				}
			}
			return false, types.AnyMap{"actual": actual}
		})
	},
	E164: func(code string) ruleset.Rule[string] {
		return newRule(code, Msg.E164, func(actual string) (bool, types.AnyMap) {
			return e164Regex.MatchString(actual), types.AnyMap{"actual": actual}
		})
	},
	Email: func(code string) ruleset.Rule[string] {
		return newRule(code, Msg.Email, func(actual string) (bool, types.AnyMap) {
			if actual == "" || strings.TrimSpace(actual) != actual {
				return false, types.AnyMap{"actual": actual}
			}
			parsed, err := mail.ParseAddress(actual)
			ok := err == nil && parsed != nil && parsed.Address == actual
			return ok, types.AnyMap{"actual": actual}
		})
	},
	EndsWith: func(code string, suffix string) ruleset.Rule[string] {
		return newRule(code, Msg.EndsWith, func(actual string) (bool, types.AnyMap) {
			if strings.HasSuffix(actual, suffix) {
				return true, nil
			}
			return false, types.AnyMap{"expected": suffix, "actual": actual}
		})
	},
	Eq: func(code string, expected string) ruleset.Rule[string] {
		return newRule(code, Msg.Eq, func(actual string) (bool, types.AnyMap) {
			if actual == expected {
				return true, nil
			}
			return false, types.AnyMap{"expected": expected, "actual": actual}
		})
	},
	EqIgnoreCase: func(code string, expected string) ruleset.Rule[string] {
		return newRule(code, "must be equal (ignore case)", func(actual string) (bool, types.AnyMap) {
			if strings.EqualFold(actual, expected) {
				return true, nil
			}
			return false, types.AnyMap{"expected": expected, "actual": actual}
		})
	},
	Excludes: func(code string, needle string) ruleset.Rule[string] {
		return newRule(code, Msg.Excludes, func(actual string) (bool, types.AnyMap) {
			if !strings.Contains(actual, needle) {
				return true, nil
			}
			return false, types.AnyMap{"expected": needle, "actual": actual}
		})
	},
	File: func(code string) ruleset.Rule[string] {
		return newRule(code, Msg.File, func(actual string) (bool, types.AnyMap) {
			return isFile(actual), types.AnyMap{"actual": actual}
		})
	},
	FilePath: func(code string) ruleset.Rule[string] {
		return newRule(code, Msg.FilePath, func(actual string) (bool, types.AnyMap) {
			return isValidPath(actual), types.AnyMap{"actual": actual}
		})
	},
	FQDN: func(code string) ruleset.Rule[string] {
		return newRule(code, Msg.FQDN, func(actual string) (bool, types.AnyMap) {
			return isHostname(actual, true), types.AnyMap{"actual": actual}
		})
	},
	HexColor: func(code string) ruleset.Rule[string] {
		return newRule(code, Msg.HexColor, func(actual string) (bool, types.AnyMap) {
			if actual == "" {
				return false, types.AnyMap{"actual": actual}
			}
			return hexColorRegex.MatchString(actual), types.AnyMap{"actual": actual}
		})
	},
	Hexadecimal: func(code string) ruleset.Rule[string] {
		return newRule(code, Msg.Hexadecimal, func(actual string) (bool, types.AnyMap) {
			return hexadecimalRegex.MatchString(actual), types.AnyMap{"actual": actual}
		})
	},
	Hostname: func(code string) ruleset.Rule[string] {
		return newRule(code, Msg.Hostname, func(actual string) (bool, types.AnyMap) {
			return isHostname(actual, false), types.AnyMap{"actual": actual}
		})
	},
	HSL: func(code string) ruleset.Rule[string] {
		return newRule(code, Msg.HSL, func(actual string) (bool, types.AnyMap) {
			match := hslRegex.FindStringSubmatch(actual)
			if match == nil || len(match) != 4 {
				return false, types.AnyMap{"actual": actual}
			}
			if !parseIntRange(match[1], 0, 360) {
				return false, types.AnyMap{"actual": actual}
			}
			if !parseIntRange(match[2], 0, 100) {
				return false, types.AnyMap{"actual": actual}
			}
			if !parseIntRange(match[3], 0, 100) {
				return false, types.AnyMap{"actual": actual}
			}
			return true, types.AnyMap{"actual": actual}
		})
	},
	HSLA: func(code string) ruleset.Rule[string] {
		return newRule(code, Msg.HSLA, func(actual string) (bool, types.AnyMap) {
			match := hslaRegex.FindStringSubmatch(actual)
			if match == nil || len(match) != 5 {
				return false, types.AnyMap{"actual": actual}
			}
			if !parseIntRange(match[1], 0, 360) {
				return false, types.AnyMap{"actual": actual}
			}
			if !parseIntRange(match[2], 0, 100) {
				return false, types.AnyMap{"actual": actual}
			}
			if !parseIntRange(match[3], 0, 100) {
				return false, types.AnyMap{"actual": actual}
			}
			if !parseFloatRange(match[4], 0, 1) {
				return false, types.AnyMap{"actual": actual}
			}
			return true, types.AnyMap{"actual": actual}
		})
	},
	HTTPURL: func(code string) ruleset.Rule[string] {
		return newRule(code, Msg.HTTPURL, func(actual string) (bool, types.AnyMap) {
			if actual == "" {
				return false, types.AnyMap{"actual": actual}
			}
			parsed, err := url.Parse(actual)
			ok := err == nil && parsed != nil && (parsed.Scheme == "http" || parsed.Scheme == "https") && parsed.Host != ""
			return ok, types.AnyMap{"actual": actual}
		})
	},
	Image: func(code string) ruleset.Rule[string] {
		return newRule(code, Msg.Image, func(actual string) (bool, types.AnyMap) {
			if !isFile(actual) {
				return false, types.AnyMap{"actual": actual}
			}
			file, err := os.Open(actual)
			if err != nil {
				return false, types.AnyMap{"actual": actual}
			}
			defer file.Close()
			_, _, err = image.DecodeConfig(file)
			return err == nil, types.AnyMap{"actual": actual}
		})
	},
	IP: func(code string) ruleset.Rule[string] {
		return newRule(code, Msg.IP, func(actual string) (bool, types.AnyMap) {
			return net.ParseIP(actual) != nil, types.AnyMap{"actual": actual}
		})
	},
	IPv4: func(code string) ruleset.Rule[string] {
		return newRule(code, Msg.IPv4, func(actual string) (bool, types.AnyMap) {
			parsed := net.ParseIP(actual)
			return parsed != nil && parsed.To4() != nil, types.AnyMap{"actual": actual}
		})
	},
	IPv6: func(code string) ruleset.Rule[string] {
		return newRule(code, Msg.IPv6, func(actual string) (bool, types.AnyMap) {
			parsed := net.ParseIP(actual)
			return parsed != nil && parsed.To4() == nil, types.AnyMap{"actual": actual}
		})
	},
	ISBN: func(code string) ruleset.Rule[string] {
		return newRule(code, Msg.ISBN, func(actual string) (bool, types.AnyMap) {
			return isISBN10(actual) || isISBN13(actual), types.AnyMap{"actual": actual}
		})
	},
	ISBN10: func(code string) ruleset.Rule[string] {
		return newRule(code, Msg.ISBN10, func(actual string) (bool, types.AnyMap) {
			return isISBN10(actual), types.AnyMap{"actual": actual}
		})
	},
	ISBN13: func(code string) ruleset.Rule[string] {
		return newRule(code, Msg.ISBN13, func(actual string) (bool, types.AnyMap) {
			return isISBN13(actual), types.AnyMap{"actual": actual}
		})
	},
	ISSN: func(code string) ruleset.Rule[string] {
		return newRule(code, Msg.ISSN, func(actual string) (bool, types.AnyMap) {
			var b strings.Builder
			b.Grow(len(actual))
			for i := 0; i < len(actual); i++ {
				ch := actual[i]
				if (ch >= '0' && ch <= '9') || ch == 'X' || ch == 'x' {
					b.WriteByte(ch)
				} else if ch == '-' || ch == ' ' {
					continue
				}
			}
			cleaned := b.String()
			if len(cleaned) != 8 {
				return false, types.AnyMap{"actual": actual}
			}
			sum := 0
			for i := 0; i < 7; i++ {
				ch := cleaned[i]
				if ch < '0' || ch > '9' {
					return false, types.AnyMap{"actual": actual}
				}
				sum += int(ch-'0') * (8 - i)
			}
			remainder := sum % 11
			check := (11 - remainder) % 11
			var expected byte
			if check == 10 {
				expected = 'X'
			} else {
				expected = byte('0' + check)
			}
			last := cleaned[7]
			if last == 'x' {
				last = 'X'
			}
			return last == expected, types.AnyMap{"actual": actual}
		})
	},
	Len: func(code string, expected int) ruleset.Rule[string] {
		return newRule(code, Msg.Len, func(actual string) (bool, types.AnyMap) {
			actualLen := len(actual)
			if actualLen == expected {
				return true, nil
			}
			return false, types.AnyMap{"expected": expected, "actual": actualLen}
		})
	},
	Lowercase: func(code string) ruleset.Rule[string] {
		return newRule(code, Msg.Lowercase, func(actual string) (bool, types.AnyMap) {
			if actual == strings.ToLower(actual) {
				return true, nil
			}
			return false, types.AnyMap{"actual": actual}
		})
	},
	LuhnChecksum: func(code string) ruleset.Rule[string] {
		return newRule(code, Msg.LuhnChecksum, func(actual string) (bool, types.AnyMap) {
			digits, ok := digitsForChecksum(actual, true)
			if !ok || len(digits) < 2 {
				return false, types.AnyMap{"actual": actual}
			}
			return luhnValid(digits), types.AnyMap{"actual": actual}
		})
	},
	MAC: func(code string) ruleset.Rule[string] {
		return newRule(code, Msg.MAC, func(actual string) (bool, types.AnyMap) {
			_, err := net.ParseMAC(actual)
			return err == nil, types.AnyMap{"actual": actual}
		})
	},
	Max: func(code string, max int) ruleset.Rule[string] {
		return newRule(code, Msg.Max, func(actual string) (bool, types.AnyMap) {
			actualLen := len(actual)
			if actualLen <= max {
				return true, nil
			}
			return false, types.AnyMap{"max": max, "actual": actualLen}
		})
	},
	MD4: func(code string) ruleset.Rule[string] {
		return digestRule(code, Msg.MD4, 16)
	},
	MD5: func(code string) ruleset.Rule[string] {
		return digestRule(code, Msg.MD5, 16)
	},
	Min: func(code string, min int) ruleset.Rule[string] {
		return newRule(code, Msg.Min, func(actual string) (bool, types.AnyMap) {
			actualLen := len(actual)
			if actualLen >= min {
				return true, nil
			}
			return false, types.AnyMap{"min": min, "actual": actualLen}
		})
	},
	Multibyte: func(code string) ruleset.Rule[string] {
		return newRule(code, Msg.Multibyte, func(actual string) (bool, types.AnyMap) {
			return multibyteRegex.MatchString(actual), types.AnyMap{"actual": actual}
		})
	},
	Ne: func(code string, disallowed string) ruleset.Rule[string] {
		return newRule(code, Msg.Ne, func(actual string) (bool, types.AnyMap) {
			if actual != disallowed {
				return true, nil
			}
			return false, types.AnyMap{"expected": disallowed, "actual": actual}
		})
	},
	NeIgnoreCase: func(code string, disallowed string) ruleset.Rule[string] {
		return newRule(code, "must not be equal (ignore case)", func(actual string) (bool, types.AnyMap) {
			if !strings.EqualFold(actual, disallowed) {
				return true, nil
			}
			return false, types.AnyMap{"expected": disallowed, "actual": actual}
		})
	},
	NotEndsWith: func(code string, suffix string) ruleset.Rule[string] {
		return newRule(code, Msg.NotEndsWith, func(actual string) (bool, types.AnyMap) {
			if !strings.HasSuffix(actual, suffix) {
				return true, nil
			}
			return false, types.AnyMap{"expected": suffix, "actual": actual}
		})
	},
	NotStartsWith: func(code string, prefix string) ruleset.Rule[string] {
		return newRule(code, Msg.NotStartsWith, func(actual string) (bool, types.AnyMap) {
			if !strings.HasPrefix(actual, prefix) {
				return true, nil
			}
			return false, types.AnyMap{"expected": prefix, "actual": actual}
		})
	},
	Number: func(code string) ruleset.Rule[string] {
		return newRule(code, Msg.Number, func(actual string) (bool, types.AnyMap) {
			if actual == "" {
				return false, types.AnyMap{"actual": actual}
			}
			parsed, err := strconv.ParseFloat(actual, 64)
			ok := err == nil && !math.IsNaN(parsed) && !math.IsInf(parsed, 0)
			return ok, types.AnyMap{"actual": actual}
		})
	},
	Numeric: func(code string) ruleset.Rule[string] {
		return newRule(code, Msg.Numeric, func(actual string) (bool, types.AnyMap) {
			return numericRegex.MatchString(actual), types.AnyMap{"actual": actual}
		})
	},
	OneOf: func(code string, values ...string) ruleset.Rule[string] {
		allowed := make(map[string]struct{}, len(values))
		for _, v := range values {
			allowed[v] = struct{}{}
		}
		return newRule(code, Msg.OneOf, func(actual string) (bool, types.AnyMap) {
			if _, ok := allowed[actual]; ok {
				return true, nil
			}
			return false, types.AnyMap{"allowed": values, "actual": actual}
		})
	},
	Pattern: func(code string, pattern *regexp.Regexp) ruleset.Rule[string] {
		patternString := ""
		if pattern != nil {
			patternString = pattern.String()
		}
		return newRule(code, Msg.Pattern, func(actual string) (bool, types.AnyMap) {
			if pattern != nil && pattern.MatchString(actual) {
				return true, nil
			}
			return false, types.AnyMap{"pattern": patternString, "actual": actual}
		})
	},
	Port: func(code string) ruleset.Rule[string] {
		return newRule(code, Msg.Port, func(actual string) (bool, types.AnyMap) {
			if actual == "" || strings.TrimSpace(actual) != actual {
				return false, types.AnyMap{"actual": actual}
			}
			for i := 0; i < len(actual); i++ {
				if actual[i] < '0' || actual[i] > '9' {
					return false, types.AnyMap{"actual": actual}
				}
			}
			port, err := strconv.Atoi(actual)
			ok := err == nil && port >= 0 && port <= 65535
			return ok, types.AnyMap{"actual": actual}
		})
	},
	PrintASCII: func(code string) ruleset.Rule[string] {
		return newRule(code, Msg.PrintASCII, func(actual string) (bool, types.AnyMap) {
			return printASCIIRegex.MatchString(actual), types.AnyMap{"actual": actual}
		})
	},
	RGB: func(code string) ruleset.Rule[string] {
		return newRule(code, Msg.RGB, func(actual string) (bool, types.AnyMap) {
			match := rgbRegex.FindStringSubmatch(actual)
			if match == nil || len(match) != 4 {
				return false, types.AnyMap{"actual": actual}
			}
			if !parseIntRange(match[1], 0, 255) {
				return false, types.AnyMap{"actual": actual}
			}
			if !parseIntRange(match[2], 0, 255) {
				return false, types.AnyMap{"actual": actual}
			}
			if !parseIntRange(match[3], 0, 255) {
				return false, types.AnyMap{"actual": actual}
			}
			return true, types.AnyMap{"actual": actual}
		})
	},
	RGBA: func(code string) ruleset.Rule[string] {
		return newRule(code, Msg.RGBA, func(actual string) (bool, types.AnyMap) {
			match := rgbaRegex.FindStringSubmatch(actual)
			if match == nil || len(match) != 5 {
				return false, types.AnyMap{"actual": actual}
			}
			if !parseIntRange(match[1], 0, 255) {
				return false, types.AnyMap{"actual": actual}
			}
			if !parseIntRange(match[2], 0, 255) {
				return false, types.AnyMap{"actual": actual}
			}
			if !parseIntRange(match[3], 0, 255) {
				return false, types.AnyMap{"actual": actual}
			}
			if !parseFloatRange(match[4], 0, 1) {
				return false, types.AnyMap{"actual": actual}
			}
			return true, types.AnyMap{"actual": actual}
		})
	},
	RIPEMD160: func(code string) ruleset.Rule[string] {
		return digestRule(code, Msg.RIPEMD160, 20)
	},
	SemVer: func(code string) ruleset.Rule[string] {
		return newRule(code, Msg.SemVer, func(actual string) (bool, types.AnyMap) {
			match := semVerRegex.FindStringSubmatch(actual)
			if match == nil {
				return false, types.AnyMap{"actual": actual}
			}
			prerelease := match[4]
			if prerelease != "" {
				for _, part := range strings.Split(prerelease, ".") {
					if part == "" {
						return false, types.AnyMap{"actual": actual}
					}
					allDigits := len(part) > 0
					for i := 0; i < len(part); i++ {
						if part[i] < '0' || part[i] > '9' {
							allDigits = false
							break
						}
					}
					if allDigits && len(part) > 1 && part[0] == '0' {
						return false, types.AnyMap{"actual": actual}
					}
				}
			}
			return true, types.AnyMap{"actual": actual}
		})
	},
	SHA1: func(code string) ruleset.Rule[string] {
		return digestRule(code, Msg.SHA1, 20)
	},
	SHA224: func(code string) ruleset.Rule[string] {
		return digestRule(code, Msg.SHA224, 28)
	},
	SHA256: func(code string) ruleset.Rule[string] {
		return digestRule(code, Msg.SHA256, 32)
	},
	SHA3_224: func(code string) ruleset.Rule[string] {
		return digestRule(code, Msg.SHA3224, 28)
	},
	SHA3_256: func(code string) ruleset.Rule[string] {
		return digestRule(code, Msg.SHA3256, 32)
	},
	SHA3_384: func(code string) ruleset.Rule[string] {
		return digestRule(code, Msg.SHA3384, 48)
	},
	SHA3_512: func(code string) ruleset.Rule[string] {
		return digestRule(code, Msg.SHA3512, 64)
	},
	SHA384: func(code string) ruleset.Rule[string] {
		return digestRule(code, Msg.SHA384, 48)
	},
	SHA512: func(code string) ruleset.Rule[string] {
		return digestRule(code, Msg.SHA512, 64)
	},
	SHA512_224: func(code string) ruleset.Rule[string] {
		return digestRule(code, Msg.SHA512224, 28)
	},
	SHA512_256: func(code string) ruleset.Rule[string] {
		return digestRule(code, Msg.SHA512256, 32)
	},
	StartsWith: func(code string, prefix string) ruleset.Rule[string] {
		return newRule(code, Msg.StartsWith, func(actual string) (bool, types.AnyMap) {
			if strings.HasPrefix(actual, prefix) {
				return true, nil
			}
			return false, types.AnyMap{"expected": prefix, "actual": actual}
		})
	},
	Uppercase: func(code string) ruleset.Rule[string] {
		return newRule(code, Msg.Uppercase, func(actual string) (bool, types.AnyMap) {
			if actual == strings.ToUpper(actual) {
				return true, nil
			}
			return false, types.AnyMap{"actual": actual}
		})
	},
	URI: func(code string) ruleset.Rule[string] {
		return newRule(code, Msg.URI, func(actual string) (bool, types.AnyMap) {
			if actual == "" {
				return false, types.AnyMap{"actual": actual}
			}
			_, err := url.ParseRequestURI(actual)
			return err == nil, types.AnyMap{"actual": actual}
		})
	},
	URL: func(code string) ruleset.Rule[string] {
		return newRule(code, Msg.URL, func(actual string) (bool, types.AnyMap) {
			if actual == "" {
				return false, types.AnyMap{"actual": actual}
			}
			parsed, err := url.Parse(actual)
			ok := err == nil && parsed != nil && parsed.Scheme != "" && parsed.Host != ""
			return ok, types.AnyMap{"actual": actual}
		})
	},
	URNRFC2141: func(code string) ruleset.Rule[string] {
		return newRule(code, Msg.URNRFC2141, func(actual string) (bool, types.AnyMap) {
			if actual == "" {
				return false, types.AnyMap{"actual": actual}
			}
			return urnRFC2141Regex.MatchString(actual), types.AnyMap{"actual": actual}
		})
	},
	UUID: func(code string) ruleset.Rule[string] {
		return newRule(code, Msg.UUID, func(actual string) (bool, types.AnyMap) {
			return uuidRegex.MatchString(actual), types.AnyMap{"actual": actual}
		})
	},
	UUID3: func(code string) ruleset.Rule[string] {
		return newRule(code, Msg.UUID3, func(actual string) (bool, types.AnyMap) {
			return uuid3Regex.MatchString(actual), types.AnyMap{"actual": actual}
		})
	},
	UUID4: func(code string) ruleset.Rule[string] {
		return newRule(code, Msg.UUID4, func(actual string) (bool, types.AnyMap) {
			return uuid4Regex.MatchString(actual), types.AnyMap{"actual": actual}
		})
	},
	UUID5: func(code string) ruleset.Rule[string] {
		return newRule(code, Msg.UUID5, func(actual string) (bool, types.AnyMap) {
			return uuid5Regex.MatchString(actual), types.AnyMap{"actual": actual}
		})
	},
}
