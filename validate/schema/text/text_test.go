// schema/text/text_test.go
package text_test

import (
	"encoding/base64"
	"errors"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/leandroluk/go/validate/internal/ast"
	"github.com/leandroluk/go/validate/internal/issues"
	"github.com/leandroluk/go/validate/internal/testkit"
	"github.com/leandroluk/go/validate/schema"
	"github.com/leandroluk/go/validate/schema/text"
)

func TestText_MissingAndNullAreIgnoredByDefault(t *testing.T) {
	s := text.New().Min(3)

	if _, err := s.Validate(ast.MissingValue()); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if _, err := s.Validate(ast.NullValue()); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}

func TestText_Required(t *testing.T) {
	s := text.New().Required()

	_, err := s.Validate(ast.MissingValue())
	validationError := testkit.RequireValidationError(t, err)

	if len(validationError.Issues) != 1 {
		t.Fatalf("expected 1 issue, got %d", len(validationError.Issues))
	}
	issue := validationError.Issues[0]
	if issue.Path != "" {
		t.Fatalf("expected empty path, got %q", issue.Path)
	}
	if issue.Message != "required" {
		t.Fatalf("expected message %q, got %q", "required", issue.Message)
	}
}

func TestText_Default(t *testing.T) {
	s := text.New().Default("x")

	got, err := s.Validate(ast.MissingValue())
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if got != "x" {
		t.Fatalf("expected %q, got %q", "x", got)
	}

	got, err = s.Validate(ast.NullValue(), schema.WithDefaultOnNull(false))
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if got != "" {
		t.Fatalf("expected zero value when DefaultOnNull=false, got %q", got)
	}

	got, err = s.Validate(ast.NullValue(), schema.WithDefaultOnNull(true))
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if got != "x" {
		t.Fatalf("expected %q, got %q", "x", got)
	}
}

func TestText_TypeMismatchMeta(t *testing.T) {
	s := text.New()

	_, err := s.Validate(ast.NumberValue("123"))
	validationError := testkit.RequireValidationError(t, err)

	issue := validationError.Issues[0]
	if issue.Meta["expected"] != "string" {
		t.Fatalf("expected meta.expected=%q, got %#v", "string", issue.Meta["expected"])
	}
	if issue.Meta["actual"] != "number" {
		t.Fatalf("expected meta.actual=%q, got %#v", "number", issue.Meta["actual"])
	}
}

func TestText_Coerce(t *testing.T) {
	s := text.New()

	got, err := s.Validate(ast.NumberValue("123"), schema.WithCoerce(true))
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if got != "123" {
		t.Fatalf("expected %q, got %q", "123", got)
	}
}

func TestText_OneOf(t *testing.T) {
	s := text.New().OneOf("a", "b")

	_, err := s.Validate("c")
	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	var validationError issues.ValidationError
	if !errors.As(err, &validationError) {
		t.Fatalf("expected ValidationError, got %T: %v", err, err)
	}

	issue := validationError.Issues[0]
	if issue.Code != text.CodeOneOf {
		t.Fatalf("expected code %q, got %q", text.CodeOneOf, issue.Code)
	}
}

func TestText_Numeric_Number_Hexadecimal_HexColor(t *testing.T) {
	sNumeric := text.New().Numeric()
	_, err := sNumeric.Validate("12345")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	_, err = sNumeric.Validate("12a")
	validationError := testkit.RequireValidationError(t, err)
	if validationError.Issues[0].Code != text.CodeNumeric {
		t.Fatalf("expected code %q, got %q", text.CodeNumeric, validationError.Issues[0].Code)
	}

	sNumber := text.New().Number()
	_, err = sNumber.Validate("-10.5e2")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	_, err = sNumber.Validate("NaN")
	validationError = testkit.RequireValidationError(t, err)
	if validationError.Issues[0].Code != text.CodeNumber {
		t.Fatalf("expected code %q, got %q", text.CodeNumber, validationError.Issues[0].Code)
	}

	sHex := text.New().Hexadecimal()
	_, err = sHex.Validate("0x1A2b")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	_, err = sHex.Validate("xz")
	validationError = testkit.RequireValidationError(t, err)
	if validationError.Issues[0].Code != text.CodeHexadecimal {
		t.Fatalf("expected code %q, got %q", text.CodeHexadecimal, validationError.Issues[0].Code)
	}

	sHexColor := text.New().HexColor()
	_, err = sHexColor.Validate("#fff")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	_, err = sHexColor.Validate("fff")
	validationError = testkit.RequireValidationError(t, err)
	if validationError.Issues[0].Code != text.CodeHexColor {
		t.Fatalf("expected code %q, got %q", text.CodeHexColor, validationError.Issues[0].Code)
	}
}

func TestText_CreditCard(t *testing.T) {
	s := text.New().CreditCard()

	_, err := s.Validate("4111111111111111")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	_, err = s.Validate("4111 1111 1111 1111")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	_, err = s.Validate("4111111111111112")
	validationError := testkit.RequireValidationError(t, err)
	if validationError.Issues[0].Code != text.CodeCreditCard {
		t.Fatalf("expected code %q, got %q", text.CodeCreditCard, validationError.Issues[0].Code)
	}
}

func TestText_LuhnChecksum(t *testing.T) {
	s := text.New().LuhnChecksum()

	_, err := s.Validate("79927398713")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	_, err = s.Validate("79927398710")
	validationError := testkit.RequireValidationError(t, err)
	if validationError.Issues[0].Code != text.CodeLuhnChecksum {
		t.Fatalf("expected code %q, got %q", text.CodeLuhnChecksum, validationError.Issues[0].Code)
	}
}

func TestText_ISBN_ISBN10_ISBN13(t *testing.T) {
	sIsbn := text.New().ISBN()
	sIsbn10 := text.New().ISBN10()
	sIsbn13 := text.New().ISBN13()

	_, err := sIsbn10.Validate("0-306-40615-2")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	_, err = sIsbn10.Validate("0-306-40615-3")
	validationError := testkit.RequireValidationError(t, err)
	if validationError.Issues[0].Code != text.CodeISBN10 {
		t.Fatalf("expected code %q, got %q", text.CodeISBN10, validationError.Issues[0].Code)
	}

	_, err = sIsbn13.Validate("978-3-16-148410-0")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	_, err = sIsbn13.Validate("978-3-16-148410-1")
	validationError = testkit.RequireValidationError(t, err)
	if validationError.Issues[0].Code != text.CodeISBN13 {
		t.Fatalf("expected code %q, got %q", text.CodeISBN13, validationError.Issues[0].Code)
	}

	_, err = sIsbn.Validate("0-306-40615-2")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	_, err = sIsbn.Validate("978-3-16-148410-0")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	_, err = sIsbn.Validate("123")
	validationError = testkit.RequireValidationError(t, err)
	if validationError.Issues[0].Code != text.CodeISBN {
		t.Fatalf("expected code %q, got %q", text.CodeISBN, validationError.Issues[0].Code)
	}
}

func TestText_ISSN(t *testing.T) {
	s := text.New().ISSN()

	_, err := s.Validate("0378-5955")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	_, err = s.Validate("2434-561X")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	_, err = s.Validate("0378-5954")
	validationError := testkit.RequireValidationError(t, err)
	if validationError.Issues[0].Code != text.CodeISSN {
		t.Fatalf("expected code %q, got %q", text.CodeISSN, validationError.Issues[0].Code)
	}
}

func TestText_E164(t *testing.T) {
	s := text.New().E164()

	_, err := s.Validate("+14155552671")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	_, err = s.Validate("14155552671")
	validationError := testkit.RequireValidationError(t, err)
	if validationError.Issues[0].Code != text.CodeE164 {
		t.Fatalf("expected code %q, got %q", text.CodeE164, validationError.Issues[0].Code)
	}

	_, err = s.Validate("+0123")
	validationError = testkit.RequireValidationError(t, err)
	if validationError.Issues[0].Code != text.CodeE164 {
		t.Fatalf("expected code %q, got %q", text.CodeE164, validationError.Issues[0].Code)
	}
}

func TestText_SemVer(t *testing.T) {
	s := text.New().SemVer()

	_, err := s.Validate("1.2.3")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	_, err = s.Validate("1.0.0-alpha.1+build.9")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	_, err = s.Validate("1.0")
	validationError := testkit.RequireValidationError(t, err)
	if validationError.Issues[0].Code != text.CodeSemVer {
		t.Fatalf("expected code %q, got %q", text.CodeSemVer, validationError.Issues[0].Code)
	}

	_, err = s.Validate("01.2.3")
	validationError = testkit.RequireValidationError(t, err)
	if validationError.Issues[0].Code != text.CodeSemVer {
		t.Fatalf("expected code %q, got %q", text.CodeSemVer, validationError.Issues[0].Code)
	}

	_, err = s.Validate("1.0.0-01")
	validationError = testkit.RequireValidationError(t, err)
	if validationError.Issues[0].Code != text.CodeSemVer {
		t.Fatalf("expected code %q, got %q", text.CodeSemVer, validationError.Issues[0].Code)
	}
}

func TestText_CVE(t *testing.T) {
	s := text.New().CVE()

	_, err := s.Validate("CVE-2021-44228")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	_, err = s.Validate("cve-2021-44228")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	_, err = s.Validate("CVE-2021-1")
	validationError := testkit.RequireValidationError(t, err)
	if validationError.Issues[0].Code != text.CodeCVE {
		t.Fatalf("expected code %q, got %q", text.CodeCVE, validationError.Issues[0].Code)
	}
}

func TestText_Hashes(t *testing.T) {
	makeHex := func(sizeBytes int) string {
		return strings.Repeat("a", sizeBytes*2)
	}
	makeBase64 := func(sizeBytes int) string {
		return base64.StdEncoding.EncodeToString(make([]byte, sizeBytes))
	}

	cases := []struct {
		name  string
		build func() *text.Schema
		size  int
		code  string
	}{
		{name: "MD4", build: func() *text.Schema { return text.New().MD4() }, size: 16, code: text.CodeMD4},
		{name: "MD5", build: func() *text.Schema { return text.New().MD5() }, size: 16, code: text.CodeMD5},
		{name: "SHA1", build: func() *text.Schema { return text.New().SHA1() }, size: 20, code: text.CodeSHA1},
		{name: "SHA224", build: func() *text.Schema { return text.New().SHA224() }, size: 28, code: text.CodeSHA224},
		{name: "SHA256", build: func() *text.Schema { return text.New().SHA256() }, size: 32, code: text.CodeSHA256},
		{name: "SHA384", build: func() *text.Schema { return text.New().SHA384() }, size: 48, code: text.CodeSHA384},
		{name: "SHA512", build: func() *text.Schema { return text.New().SHA512() }, size: 64, code: text.CodeSHA512},
		{name: "SHA512_224", build: func() *text.Schema { return text.New().SHA512_224() }, size: 28, code: text.CodeSHA512224},
		{name: "SHA512_256", build: func() *text.Schema { return text.New().SHA512_256() }, size: 32, code: text.CodeSHA512256},
		{name: "SHA3_224", build: func() *text.Schema { return text.New().SHA3_224() }, size: 28, code: text.CodeSHA3224},
		{name: "SHA3_256", build: func() *text.Schema { return text.New().SHA3_256() }, size: 32, code: text.CodeSHA3256},
		{name: "SHA3_384", build: func() *text.Schema { return text.New().SHA3_384() }, size: 48, code: text.CodeSHA3384},
		{name: "SHA3_512", build: func() *text.Schema { return text.New().SHA3_512() }, size: 64, code: text.CodeSHA3512},
		{name: "RIPEMD160", build: func() *text.Schema { return text.New().RIPEMD160() }, size: 20, code: text.CodeRIPEMD160},
		{name: "BLAKE2B_256", build: func() *text.Schema { return text.New().BLAKE2B_256() }, size: 32, code: text.CodeBLAKE2B256},
		{name: "BLAKE2B_384", build: func() *text.Schema { return text.New().BLAKE2B_384() }, size: 48, code: text.CodeBLAKE2B384},
		{name: "BLAKE2B_512", build: func() *text.Schema { return text.New().BLAKE2B_512() }, size: 64, code: text.CodeBLAKE2B512},
		{name: "BLAKE2S_256", build: func() *text.Schema { return text.New().BLAKE2S_256() }, size: 32, code: text.CodeBLAKE2S256},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := tc.build()

			_, err := s.Validate(makeHex(tc.size))
			if err != nil {
				t.Fatalf("expected nil error, got %v", err)
			}

			base64Value := makeBase64(tc.size)
			_, err = s.Validate(base64Value)
			if err != nil {
				t.Fatalf("expected nil error, got %v", err)
			}

			_, err = s.Validate(strings.TrimRight(base64Value, "="))
			if err != nil {
				t.Fatalf("expected nil error, got %v", err)
			}

			_, err = s.Validate(strings.Repeat("g", tc.size*2))
			validationError := testkit.RequireValidationError(t, err)
			if validationError.Issues[0].Code != tc.code {
				t.Fatalf("expected code %q, got %q", tc.code, validationError.Issues[0].Code)
			}
		})
	}
}

func TestText_File_Dir_FilePath_DirPath_Image(t *testing.T) {
	baseDir := t.TempDir()

	filePath := filepath.Join(baseDir, "file.txt")
	if err := os.WriteFile(filePath, []byte("ok"), 0o644); err != nil {
		t.Fatalf("write temp file: %v", err)
	}

	sFile := text.New().File()
	if _, err := sFile.Validate(filePath); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	_, err := sFile.Validate(baseDir)
	validationError := testkit.RequireValidationError(t, err)
	if validationError.Issues[0].Code != text.CodeFile {
		t.Fatalf("expected code %q, got %q", text.CodeFile, validationError.Issues[0].Code)
	}
	_, err = sFile.Validate(filepath.Join(baseDir, "missing.txt"))
	validationError = testkit.RequireValidationError(t, err)
	if validationError.Issues[0].Code != text.CodeFile {
		t.Fatalf("expected code %q, got %q", text.CodeFile, validationError.Issues[0].Code)
	}

	sDir := text.New().Dir()
	if _, err := sDir.Validate(baseDir); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	_, err = sDir.Validate(filePath)
	validationError = testkit.RequireValidationError(t, err)
	if validationError.Issues[0].Code != text.CodeDir {
		t.Fatalf("expected code %q, got %q", text.CodeDir, validationError.Issues[0].Code)
	}
	_, err = sDir.Validate(filepath.Join(baseDir, "missing-dir"))
	validationError = testkit.RequireValidationError(t, err)
	if validationError.Issues[0].Code != text.CodeDir {
		t.Fatalf("expected code %q, got %q", text.CodeDir, validationError.Issues[0].Code)
	}

	sFilePath := text.New().FilePath()
	if _, err := sFilePath.Validate(filepath.Join("a", "b", "c.txt")); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	_, err = sFilePath.Validate("a\x00b")
	validationError = testkit.RequireValidationError(t, err)
	if validationError.Issues[0].Code != text.CodeFilePath {
		t.Fatalf("expected code %q, got %q", text.CodeFilePath, validationError.Issues[0].Code)
	}

	sDirPath := text.New().DirPath()
	if _, err := sDirPath.Validate(baseDir); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	missingDir := filepath.Join(baseDir, "missing-dirpath")
	_, err = sDirPath.Validate(missingDir)
	validationError = testkit.RequireValidationError(t, err)
	if validationError.Issues[0].Code != text.CodeDirPath {
		t.Fatalf("expected code %q, got %q", text.CodeDirPath, validationError.Issues[0].Code)
	}
	if _, err := sDirPath.Validate(missingDir + string(os.PathSeparator)); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	imagePath := filepath.Join(baseDir, "img.png")
	imgFile, err := os.Create(imagePath)
	if err != nil {
		t.Fatalf("create image file: %v", err)
	}
	if err := png.Encode(imgFile, image.NewRGBA(image.Rect(0, 0, 1, 1))); err != nil {
		imgFile.Close()
		t.Fatalf("encode png: %v", err)
	}
	if err := imgFile.Close(); err != nil {
		t.Fatalf("close image file: %v", err)
	}

	sImage := text.New().Image()
	if _, err := sImage.Validate(imagePath); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	notImagePath := filepath.Join(baseDir, "not-image.png")
	if err := os.WriteFile(notImagePath, []byte("nope"), 0o644); err != nil {
		t.Fatalf("write not-image: %v", err)
	}
	_, err = sImage.Validate(notImagePath)
	validationError = testkit.RequireValidationError(t, err)
	if validationError.Issues[0].Code != text.CodeImage {
		t.Fatalf("expected code %q, got %q", text.CodeImage, validationError.Issues[0].Code)
	}

	_, err = sImage.Validate(filepath.Join(baseDir, "missing.png"))
	validationError = testkit.RequireValidationError(t, err)
	if validationError.Issues[0].Code != text.CodeImage {
		t.Fatalf("expected code %q, got %q", text.CodeImage, validationError.Issues[0].Code)
	}
}
