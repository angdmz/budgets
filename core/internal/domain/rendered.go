package domain

import "encoding/json"

// apiDateFormat is the date layout used in API wire shapes.
const apiDateFormat = "2006-01-02"

// Rendered is the opaque, final rendering of a domain object.
// Handlers pass Rendered[T] values straight to the JSON encoder;
// they must not inspect or mutate the wrapped value.
//
// Only Persisted* objects and rendering decorator types produce Rendered.
// Render() must be pure: all fallible work (decryption, conversion, I/O)
// happens while constructing the rendering object, never inside Render().
type Rendered[T any] struct {
	v T
}

// Render wraps a wire-shape value into its final renderable form.
func Render[T any](v T) Rendered[T] {
	return Rendered[T]{v: v}
}

// MarshalJSON serializes the wrapped wire shape.
func (r Rendered[T]) MarshalJSON() ([]byte, error) {
	return json.Marshal(r.v)
}

// Value exposes the wrapped wire shape. For internal domain use only —
// e.g. composing one rendering inside another.
func (r Rendered[T]) Value() T {
	return r.v
}
