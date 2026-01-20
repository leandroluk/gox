// schema/text/rules.go
package text

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
