package util

// Ptr returns the value asserted as type T.
func Ptr[T any](value any) T { return value.(T) }

// Must panics if err is non-nil, otherwise returns the value asserted as type T.
func Must[T any](value any, err error) T {
	if err != nil {
		panic(err)
	}
	return value.(T)
}
