// schema/options.go
package schema

import "time"

type Formatter = func(path string, code string, message string, meta map[string]any) string

type Options struct {
	FailFast                   bool
	MaxIssues                  int
	DefaultOnNull              bool
	Coerce                     bool
	OmitZero                   bool
	OmitNil                    bool
	OmitEmpty                  bool
	Formatter                  Formatter
	CoerceTrimSpace            bool
	CoerceNumberUnderscore     bool
	CoerceDateUnixSeconds      bool
	CoerceDateUnixMilliseconds bool
	CoerceDurationSeconds      bool
	CoerceDurationMilliseconds bool
	TimeLocation               *time.Location
	DateLayouts                []string
}

type Option func(*Options)

func DefaultOptions() Options {
	return Options{
		FailFast:                   false,
		MaxIssues:                  0,
		DefaultOnNull:              true,
		Coerce:                     false,
		OmitZero:                   false,
		OmitNil:                    false,
		OmitEmpty:                  false,
		Formatter:                  nil,
		CoerceTrimSpace:            false,
		CoerceNumberUnderscore:     false,
		CoerceDateUnixSeconds:      false,
		CoerceDateUnixMilliseconds: false,
		CoerceDurationSeconds:      false,
		CoerceDurationMilliseconds: false,
		TimeLocation:               time.UTC,
		DateLayouts: []string{
			time.RFC3339Nano,
			time.RFC3339,
			"2006-01-02",
			"2006-01-02T15:04:05",
		},
	}
}

func ApplyOptions(optionList ...Option) Options {
	options := DefaultOptions()
	for _, option := range optionList {
		if option != nil {
			option(&options)
		}
	}
	return options
}

func WithFailFast(value bool) Option {
	return func(options *Options) {
		options.FailFast = value
	}
}

func WithMaxIssues(value int) Option {
	return func(options *Options) {
		if value < 0 {
			value = 0
		}
		options.MaxIssues = value
	}
}

func WithDefaultOnNull(value bool) Option {
	return func(options *Options) {
		options.DefaultOnNull = value
	}
}

func WithCoerce(value bool) Option {
	return func(options *Options) {
		options.Coerce = value
	}
}

func WithOmitZero(value bool) Option {
	return func(options *Options) {
		options.OmitZero = value
	}
}

func WithOmitNil(value bool) Option {
	return func(options *Options) {
		options.OmitNil = value
	}
}

func WithOmitEmpty(value bool) Option {
	return func(options *Options) {
		options.OmitEmpty = value
	}
}

func WithFormatter(formatter Formatter) Option {
	return func(options *Options) {
		options.Formatter = formatter
	}
}

func WithCoerceTrimSpace(value bool) Option {
	return func(options *Options) {
		options.CoerceTrimSpace = value
	}
}

func WithCoerceNumberUnderscore(value bool) Option {
	return func(options *Options) {
		options.CoerceNumberUnderscore = value
	}
}

func WithCoerceDateUnixSeconds(value bool) Option {
	return func(options *Options) {
		options.CoerceDateUnixSeconds = value
	}
}

func WithCoerceDateUnixMilliseconds(value bool) Option {
	return func(options *Options) {
		options.CoerceDateUnixMilliseconds = value
	}
}

func WithCoerceDurationSeconds(value bool) Option {
	return func(options *Options) {
		options.CoerceDurationSeconds = value
	}
}

func WithCoerceDurationMilliseconds(value bool) Option {
	return func(options *Options) {
		options.CoerceDurationMilliseconds = value
	}
}

func WithTimeLocation(location *time.Location) Option {
	return func(options *Options) {
		if location == nil {
			location = time.UTC
		}
		options.TimeLocation = location
	}
}

func WithTimeZone(name string) Option {
	return func(options *Options) {
		if name == "" {
			options.TimeLocation = time.UTC
			return
		}
		location, err := time.LoadLocation(name)
		if err != nil || location == nil {
			options.TimeLocation = time.UTC
			return
		}
		options.TimeLocation = location
	}
}

func WithTimezone(name string) Option {
	return WithTimeZone(name)
}

func WithDateLayouts(layouts ...string) Option {
	return func(options *Options) {
		options.DateLayouts = append([]string(nil), layouts...)
	}
}

func WithAdditionalDateLayouts(layouts ...string) Option {
	return func(options *Options) {
		options.DateLayouts = append(options.DateLayouts, layouts...)
	}
}
