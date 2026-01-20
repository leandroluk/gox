// internal/engine/walk.go
package engine

func WithField[T any](context *Context, name string, run func() (T, bool)) (T, bool) {
	context.PushField(name)
	output, stop := run()
	context.Pop()
	return output, stop
}

func WithIndex[T any](context *Context, index int, run func() (T, bool)) (T, bool) {
	context.PushIndex(index)
	output, stop := run()
	context.Pop()
	return output, stop
}

func WithKey[T any](context *Context, key string, run func() (T, bool)) (T, bool) {
	context.PushKey(key)
	output, stop := run()
	context.Pop()
	return output, stop
}
