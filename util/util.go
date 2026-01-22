package util

// Ptr returns the value asserted as type T.
func Ptr[T any](v T) *T {
	return &v
}

// Must panics if err is non-nil, otherwise returns the value asserted as type T.
func Must[T any](value T, err error) T {
	if err != nil {
		panic(err)
	}
	return value
}

// Try executes the function fn and catches any panic that implements the error interface,
// returning it as an error. If the panic value is not an error, it is re-panicked.
func Try(fn func()) (err error) {
	defer func() {
		if r := recover(); r != nil {
			if e, ok := r.(error); ok {
				err = e
			} else {
				panic(r)
			}
		}
	}()
	fn()
	return nil
}

// Check panics if err is non-nil.
func Check(err error) {
	if err != nil {
		panic(err)
	}
}
