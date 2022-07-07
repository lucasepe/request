package request

import (
	"bytes"
	"encoding/json"
	"io"
)

// BodyProvider provides data for a request body.
type BodyProvider func() (io.ReadCloser, error)

// WithBytes is a BodyProvider that returns the provided raw bytes.
func WithBytes(b []byte) BodyProvider {
	return func() (io.ReadCloser, error) {
		return io.NopCloser(bytes.NewReader(b)), nil
	}
}

// WithAny is a BodyProvider that marshals a JSON object.
func WithAny(v any) BodyProvider {
	return func() (io.ReadCloser, error) {
		b, err := json.Marshal(v)
		if err != nil {
			return nil, err
		}
		return io.NopCloser(bytes.NewReader(b)), nil
	}
}
