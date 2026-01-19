// oas/oas.go
package oas

import (
	"github.com/leandroluk/go/oas/builder"
	"github.com/leandroluk/go/oas/types"
)

type OpenAPI = types.OpenAPI
type Info = types.Info
type Contact = types.Contact
type License = types.License
type Server = types.Server
type ServerVariable = types.ServerVariable
type Paths = types.Paths
type PathItem = types.PathItem
type PathOperation = types.PathOperation
type Parameter = types.Parameter
type RequestBody = types.RequestBody
type Response = types.Response
type MediaType = types.MediaType
type Encoding = types.Encoding
type ExampleObject = types.ExampleObject
type Header = types.Header
type Link = types.Link
type Callback = types.Callback
type Components = types.Components
type Tag = types.Tag
type ExternalDocs = types.ExternalDocs
type SecurityRequirement = types.SecurityRequirement
type SecurityScheme = types.SecurityScheme
type OAuthFlows = types.OAuthFlows
type OAuthFlow = types.OAuthFlow
type Discriminator = types.Discriminator
type XML = types.XML
type Schema = types.Schema
type SchemaType = types.SchemaType
type ContentType = types.ContentType

const (
	SchemaType_Object  = types.SchemaType_Object
	SchemaType_String  = types.SchemaType_String
	SchemaType_Integer = types.SchemaType_Integer
	SchemaType_Number  = types.SchemaType_Number
	SchemaType_Boolean = types.SchemaType_Boolean
	SchemaType_Array   = types.SchemaType_Array
	SchemaType_Null    = types.SchemaType_Null
)

const (
	ContentType_ApplicationJson = types.ContentType_ApplicationJson
	ContentType_TextPlain       = types.ContentType_TextPlain
	ContentType_TextHtml        = types.ContentType_TextHtml
	ContentType_TextXml         = types.ContentType_TextXml
	ContentType_TextCsv         = types.ContentType_TextCsv
	ContentType_ImageJpeg       = types.ContentType_ImageJpeg
	ContentType_ImagePng        = types.ContentType_ImagePng
	ContentType_ImageGif        = types.ContentType_ImageGif
	ContentType_ImageSvg        = types.ContentType_ImageSvg
	ContentType_ImageWebp       = types.ContentType_ImageWebp
)

type OpenAPIBuilder = builder.OpenAPIBuilder
type PathItemBuilder = builder.PathItemBuilder
type OperationBuilder = builder.OperationBuilder
type ResponseBuilder = builder.ResponseBuilder
type RequestBodyBuilder = builder.RequestBodyBuilder
type ParameterBuilder = builder.ParameterBuilder
type MediaTypeBuilder = builder.MediaTypeBuilder
type SchemaBuilder = builder.SchemaBuilder
type ContentBuilder = builder.ContentBuilder
type HeaderBuilder = builder.HeaderBuilder
type LinkBuilder = builder.LinkBuilder
type ExampleObjectBuilder = builder.ExampleObjectBuilder
type EncodingBuilder = builder.EncodingBuilder

func New() *builder.OpenAPIBuilder { return builder.New() }

type NewBuilder = *builder.OpenAPIBuilder

// Schema construtores por tipo
var (
	String  = builder.String
	Integer = builder.Integer
	Number  = builder.Number
	Boolean = builder.Boolean
	Array   = builder.Array
	Object  = builder.Object
	Ref     = builder.Ref
)

// Operation, RequestBody, Response e Parameter construtores
var (
	Operation       = builder.Operation // Renamed to avoid conflict with Operation type
	Body            = builder.Body
	ResponseCode    = builder.ResponseCode
	ResponseRange   = builder.ResponseRange
	ResponseDefault = builder.ResponseDefault
	InPath          = builder.InPath
	InQuery         = builder.InQuery
	InHeader        = builder.InHeader
	InCookie        = builder.InCookie
)
