// schema/text/rule/helpers.go
package rule

import (
	"encoding/base64"
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
	"regexp"
	"runtime"
	"strconv"
	"strings"

	"github.com/leandroluk/go/validate/internal/engine"
	"github.com/leandroluk/go/validate/internal/ruleset"
)

var (
	uuidRegex  = regexp.MustCompile(`(?i)^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)
	uuid3Regex = regexp.MustCompile(`(?i)^[0-9a-f]{8}-[0-9a-f]{4}-3[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)
	uuid4Regex = regexp.MustCompile(`(?i)^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)
	uuid5Regex = regexp.MustCompile(`(?i)^[0-9a-f]{8}-[0-9a-f]{4}-5[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)

	urnRFC2141Regex = regexp.MustCompile(`(?i)^urn:[a-z0-9][a-z0-9-]{1,31}:[^\s]+$`)

	hexColorRegex = regexp.MustCompile(`(?i)^#[0-9a-f]{3}([0-9a-f]{3})?$`)

	rgbRegex  = regexp.MustCompile(`(?i)^rgb\(\s*([0-9]{1,3})\s*,\s*([0-9]{1,3})\s*,\s*([0-9]{1,3})\s*\)$`)
	rgbaRegex = regexp.MustCompile(`(?i)^rgba\(\s*([0-9]{1,3})\s*,\s*([0-9]{1,3})\s*,\s*([0-9]{1,3})\s*,\s*([0-9]+(?:\.[0-9]+)?|\.[0-9]+)\s*\)$`)

	hslRegex  = regexp.MustCompile(`(?i)^hsl\(\s*([0-9]{1,3})\s*,\s*([0-9]{1,3})%\s*,\s*([0-9]{1,3})%\s*\)$`)
	hslaRegex = regexp.MustCompile(`(?i)^hsla\(\s*([0-9]{1,3})\s*,\s*([0-9]{1,3})%\s*,\s*([0-9]{1,3})%\s*,\s*([0-9]+(?:\.[0-9]+)?|\.[0-9]+)\s*\)$`)

	e164Regex = regexp.MustCompile(`^\+[1-9]\d{1,14}$`)

	semVerRegex = regexp.MustCompile(`^(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)(?:-([0-9A-Za-z-]+(?:\.[0-9A-Za-z-]+)*))?(?:\+([0-9A-Za-z-]+(?:\.[0-9A-Za-z-]+)*))?$`)
	cveRegex    = regexp.MustCompile(`(?i)^CVE-\d{4}-\d{4,}$`)
)

func newRule(code string, message string, validate func(actual string) (bool, map[string]any)) ruleset.Rule[string] {
	return ruleset.New(code, func(actual string, context *engine.Context) (string, bool) {
		ok, meta := validate(actual)
		if ok {
			return actual, false
		}
		stop := context.AddIssueWithMeta(code, message, meta)
		return actual, stop
	})
}

func digestRule(code string, message string, sizeBytes int) ruleset.Rule[string] {
	return newRule(code, message, func(actual string) (bool, map[string]any) {
		return isDigest(actual, sizeBytes), map[string]any{"actual": actual}
	})
}

func isEmail(value string) bool {
	if value == "" {
		return false
	}
	if strings.TrimSpace(value) != value {
		return false
	}
	parsed, err := mail.ParseAddress(value)
	if err != nil {
		return false
	}
	return parsed != nil && parsed.Address == value
}

func isURL(value string) bool {
	if value == "" {
		return false
	}
	parsed, err := url.Parse(value)
	if err != nil || parsed == nil {
		return false
	}
	if parsed.Scheme == "" || parsed.Host == "" {
		return false
	}
	return true
}

func isHTTPURL(value string) bool {
	if value == "" {
		return false
	}
	parsed, err := url.Parse(value)
	if err != nil || parsed == nil {
		return false
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return false
	}
	if parsed.Host == "" {
		return false
	}
	return true
}

func isURI(value string) bool {
	if value == "" {
		return false
	}
	_, err := url.ParseRequestURI(value)
	return err == nil
}

func isURNRFC2141(value string) bool {
	if value == "" {
		return false
	}
	return urnRFC2141Regex.MatchString(value)
}

func isHostname(value string, requireDot bool) bool {
	if value == "" {
		return false
	}
	if strings.TrimSpace(value) != value {
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
		if label == "" || len(label) > 63 {
			return false
		}
		if label[0] == '-' || label[len(label)-1] == '-' {
			return false
		}
		for i := 0; i < len(label); i++ {
			ch := label[i]
			isDigit := ch >= '0' && ch <= '9'
			isLower := ch >= 'a' && ch <= 'z'
			isUpper := ch >= 'A' && ch <= 'Z'
			if isDigit || isLower || isUpper || ch == '-' {
				continue
			}
			return false
		}
	}

	return true
}

func isPort(value string) bool {
	if value == "" {
		return false
	}
	if strings.TrimSpace(value) != value {
		return false
	}
	for i := 0; i < len(value); i++ {
		ch := value[i]
		if ch < '0' || ch > '9' {
			return false
		}
	}
	port, err := strconv.Atoi(value)
	if err != nil {
		return false
	}
	return port >= 0 && port <= 65535
}

func isNumeric(value string) bool {
	if value == "" {
		return false
	}
	for i := 0; i < len(value); i++ {
		ch := value[i]
		if ch < '0' || ch > '9' {
			return false
		}
	}
	return true
}

func isNumber(value string) bool {
	if value == "" {
		return false
	}
	parsed, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return false
	}
	if math.IsNaN(parsed) || math.IsInf(parsed, 0) {
		return false
	}
	return true
}

func isHexadecimal(value string) bool {
	if value == "" {
		return false
	}
	if strings.HasPrefix(value, "0x") || strings.HasPrefix(value, "0X") {
		value = value[2:]
	}
	if value == "" {
		return false
	}
	for i := 0; i < len(value); i++ {
		if !isHexDigit(value[i]) {
			return false
		}
	}
	return true
}

func isHexDigit(ch byte) bool {
	isDigit := ch >= '0' && ch <= '9'
	isLower := ch >= 'a' && ch <= 'f'
	isUpper := ch >= 'A' && ch <= 'F'
	return isDigit || isLower || isUpper
}

func isHexColor(value string) bool {
	if value == "" {
		return false
	}
	return hexColorRegex.MatchString(value)
}

func isRGB(value string) bool {
	match := rgbRegex.FindStringSubmatch(value)
	if match == nil || len(match) != 4 {
		return false
	}

	if _, ok := parseIntRange(match[1], 0, 255); !ok {
		return false
	}
	if _, ok := parseIntRange(match[2], 0, 255); !ok {
		return false
	}
	if _, ok := parseIntRange(match[3], 0, 255); !ok {
		return false
	}

	return true
}

func isRGBA(value string) bool {
	match := rgbaRegex.FindStringSubmatch(value)
	if match == nil || len(match) != 5 {
		return false
	}

	if _, ok := parseIntRange(match[1], 0, 255); !ok {
		return false
	}
	if _, ok := parseIntRange(match[2], 0, 255); !ok {
		return false
	}
	if _, ok := parseIntRange(match[3], 0, 255); !ok {
		return false
	}
	if _, ok := parseFloatRange(match[4], 0, 1); !ok {
		return false
	}

	return true
}

func isHSL(value string) bool {
	match := hslRegex.FindStringSubmatch(value)
	if match == nil || len(match) != 4 {
		return false
	}

	if _, ok := parseIntRange(match[1], 0, 360); !ok {
		return false
	}
	if _, ok := parseIntRange(match[2], 0, 100); !ok {
		return false
	}
	if _, ok := parseIntRange(match[3], 0, 100); !ok {
		return false
	}

	return true
}

func isHSLA(value string) bool {
	match := hslaRegex.FindStringSubmatch(value)
	if match == nil || len(match) != 5 {
		return false
	}

	if _, ok := parseIntRange(match[1], 0, 360); !ok {
		return false
	}
	if _, ok := parseIntRange(match[2], 0, 100); !ok {
		return false
	}
	if _, ok := parseIntRange(match[3], 0, 100); !ok {
		return false
	}
	if _, ok := parseFloatRange(match[4], 0, 1); !ok {
		return false
	}

	return true
}

func parseIntRange(value string, min int, max int) (int, bool) {
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, false
	}
	if parsed < min || parsed > max {
		return 0, false
	}
	return parsed, true
}

func parseFloatRange(value string, min float64, max float64) (float64, bool) {
	parsed, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0, false
	}
	if math.IsNaN(parsed) || math.IsInf(parsed, 0) {
		return 0, false
	}
	if parsed < min || parsed > max {
		return 0, false
	}
	return parsed, true
}

func isBase64(value string) bool {
	_, err := base64.StdEncoding.DecodeString(value)
	return err == nil
}

func isBase64URL(value string) bool {
	_, err := base64.URLEncoding.DecodeString(value)
	return err == nil
}

func isBase64RawURL(value string) bool {
	_, err := base64.RawURLEncoding.DecodeString(value)
	return err == nil
}

func isDataURI(value string) bool {
	if !strings.HasPrefix(value, "data:") {
		return false
	}

	commaIndex := strings.IndexByte(value, ',')
	if commaIndex < 0 {
		return false
	}

	meta := value[len("data:"):commaIndex]
	data := value[commaIndex+1:]

	if strings.Contains(meta, ";base64") {
		_, err := base64.StdEncoding.DecodeString(data)
		if err == nil {
			return true
		}
		_, err = base64.RawStdEncoding.DecodeString(data)
		return err == nil
	}

	_, err := url.PathUnescape(data)
	return err == nil
}

func isASCII(value string) bool {
	for i := 0; i < len(value); i++ {
		if value[i] > 0x7F {
			return false
		}
	}
	return true
}

func isPrintASCII(value string) bool {
	for i := 0; i < len(value); i++ {
		ch := value[i]
		if ch < 0x20 || ch > 0x7E {
			return false
		}
	}
	return true
}

func isMultibyte(value string) bool {
	for i := 0; i < len(value); i++ {
		if value[i] > 0x7F {
			return true
		}
	}
	return false
}

func isCreditCard(value string) bool {
	digits, ok := digitsForChecksum(value, true)
	if !ok {
		return false
	}
	if len(digits) < 12 || len(digits) > 19 {
		return false
	}
	return luhnValid(digits)
}

func isLuhnChecksum(value string) bool {
	digits, ok := digitsForChecksum(value, true)
	if !ok {
		return false
	}
	if len(digits) < 2 {
		return false
	}
	return luhnValid(digits)
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

func luhnValid(digits string) bool {
	sum := 0
	double := false

	for i := len(digits) - 1; i >= 0; i-- {
		d := int(digits[i] - '0')
		if double {
			d = d * 2
			if d > 9 {
				d = d - 9
			}
		}
		sum += d
		double = !double
	}

	return sum%10 == 0
}

func isISBN(value string) bool {
	return isISBN10(value) || isISBN13(value)
}

func isISBN10(value string) bool {
	cleaned := stripISBN(value)
	if len(cleaned) != 10 {
		return false
	}

	total := 0
	for i := 0; i < 9; i++ {
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
	for i := 0; i < 12; i++ {
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

func stripISBN(value string) string {
	var builder strings.Builder
	builder.Grow(len(value))

	for i := 0; i < len(value); i++ {
		ch := value[i]
		if ch >= '0' && ch <= '9' {
			builder.WriteByte(ch)
			continue
		}
		if ch == 'X' || ch == 'x' {
			builder.WriteByte(ch)
			continue
		}
		if ch == '-' || ch == ' ' {
			continue
		}
	}

	return builder.String()
}

func stripDigitsOnly(value string) string {
	var builder strings.Builder
	builder.Grow(len(value))

	for i := 0; i < len(value); i++ {
		ch := value[i]
		if ch >= '0' && ch <= '9' {
			builder.WriteByte(ch)
		}
	}

	return builder.String()
}

func isISSN(value string) bool {
	cleaned := stripISSN(value)
	if len(cleaned) != 8 {
		return false
	}

	sum := 0
	for i := 0; i < 7; i++ {
		ch := cleaned[i]
		if ch < '0' || ch > '9' {
			return false
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

	return last == expected
}

func stripISSN(value string) string {
	var builder strings.Builder
	builder.Grow(len(value))

	for i := 0; i < len(value); i++ {
		ch := value[i]
		if ch >= '0' && ch <= '9' {
			builder.WriteByte(ch)
			continue
		}
		if ch == 'X' || ch == 'x' {
			builder.WriteByte(ch)
			continue
		}
		if ch == '-' || ch == ' ' {
			continue
		}
	}

	return builder.String()
}

func isE164(value string) bool {
	return e164Regex.MatchString(value)
}

func isSemVer(value string) bool {
	match := semVerRegex.FindStringSubmatch(value)
	if match == nil {
		return false
	}

	prerelease := match[4]
	if prerelease == "" {
		return true
	}

	parts := strings.Split(prerelease, ".")
	for _, part := range parts {
		if part == "" {
			return false
		}
		if isDigits(part) && len(part) > 1 && part[0] == '0' {
			return false
		}
	}

	return true
}

func isDigits(value string) bool {
	for i := 0; i < len(value); i++ {
		ch := value[i]
		if ch < '0' || ch > '9' {
			return false
		}
	}
	return value != ""
}

func isCVE(value string) bool {
	return cveRegex.MatchString(value)
}

func isDigest(value string, sizeBytes int) bool {
	if value == "" || sizeBytes <= 0 {
		return false
	}

	if isFixedHexLen(value, sizeBytes*2) {
		return true
	}

	if decodedLength, ok := base64DecodedLength(value); ok && decodedLength == sizeBytes {
		return true
	}

	return false
}

func isFixedHexLen(value string, length int) bool {
	if len(value) != length {
		return false
	}
	for i := 0; i < len(value); i++ {
		if !isHexDigit(value[i]) {
			return false
		}
	}
	return true
}

func base64DecodedLength(value string) (int, bool) {
	if value == "" {
		return 0, false
	}

	if decoded, err := base64.StdEncoding.DecodeString(value); err == nil {
		return len(decoded), true
	}
	if decoded, err := base64.RawStdEncoding.DecodeString(value); err == nil {
		return len(decoded), true
	}
	if decoded, err := base64.URLEncoding.DecodeString(value); err == nil {
		return len(decoded), true
	}
	if decoded, err := base64.RawURLEncoding.DecodeString(value); err == nil {
		return len(decoded), true
	}

	return 0, false
}

func isValidPath(value string) bool {
	if value == "" {
		return false
	}
	if strings.TrimSpace(value) != value {
		return false
	}
	if strings.IndexByte(value, 0) >= 0 {
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
				if i == 1 && isAlpha(value[0]) {
					continue
				}
				return false
			}
		}
	}

	_ = filepath.Clean(value)
	return true
}

func isAlpha(ch byte) bool {
	isLower := ch >= 'a' && ch <= 'z'
	isUpper := ch >= 'A' && ch <= 'Z'
	return isLower || isUpper
}

func isFilePath(value string) bool {
	return isValidPath(value)
}

func hasTrailingSeparator(value string) bool {
	if value == "" {
		return false
	}
	last := value[len(value)-1]
	return last == '/' || last == '\\'
}

func isDirPath(value string) bool {
	if !isValidPath(value) {
		return false
	}

	info, err := os.Stat(value)
	if err == nil && info.IsDir() {
		return true
	}

	return hasTrailingSeparator(value)
}

func isFile(value string) bool {
	if !isValidPath(value) {
		return false
	}
	info, err := os.Stat(value)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

func isDir(value string) bool {
	if !isValidPath(value) {
		return false
	}
	info, err := os.Stat(value)
	if err != nil {
		return false
	}
	return info.IsDir()
}

func isImage(value string) bool {
	if !isFile(value) {
		return false
	}

	file, err := os.Open(value)
	if err != nil {
		return false
	}
	defer file.Close()

	_, _, err = image.DecodeConfig(file)
	return err == nil
}

func isIP(value string) bool {
	return net.ParseIP(value) != nil
}

func isIPv4(value string) bool {
	parsed := net.ParseIP(value)
	return parsed != nil && parsed.To4() != nil
}

func isIPv6(value string) bool {
	parsed := net.ParseIP(value)
	return parsed != nil && parsed.To4() == nil
}

func isCIDR(value string) bool {
	_, _, err := net.ParseCIDR(value)
	return err == nil
}

func isMAC(value string) bool {
	_, err := net.ParseMAC(value)
	return err == nil
}
