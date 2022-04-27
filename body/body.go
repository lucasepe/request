package body

import (
	"bytes"
	"encoding/json"
	"io"
	"net/url"
	"os"
	"strings"
)

// Getter provides a Builder with a source for a request body.
type Getter func() (io.ReadCloser, error)

// Reader is a BodyGetter that returns an io.Reader.
func Reader(r io.Reader) Getter {
	return func() (io.ReadCloser, error) {
		if rc, ok := r.(io.ReadCloser); ok {
			return rc, nil
		}
		return io.NopCloser(r), nil
	}
}

// Writer is a Getter that pipes writes into a request body.
func Writer(f func(w io.Writer) error) Getter {
	return func() (io.ReadCloser, error) {
		r, w := io.Pipe()
		go func() {
			var err error
			defer func() {
				w.CloseWithError(err)
			}()
			err = f(w)
		}()
		return r, nil
	}
}

// Bytes is a Getter that returns the provided raw bytes.
func Bytes(b []byte) Getter {
	return func() (io.ReadCloser, error) {
		return io.NopCloser(bytes.NewReader(b)), nil
	}
}

// JSON is a Getter that marshals a JSON object.
func JSON(v any) Getter {
	return func() (io.ReadCloser, error) {
		b, err := json.Marshal(v)
		if err != nil {
			return nil, err
		}
		return io.NopCloser(bytes.NewReader(b)), nil
	}
}

// Form is a Getter that builds an encoded form body.
func Form(data url.Values) Getter {
	return func() (r io.ReadCloser, err error) {
		return io.NopCloser(strings.NewReader(data.Encode())), nil
	}
}

// File is a BodyGetter that reads the provided file path.
func File(name string) Getter {
	return func() (r io.ReadCloser, err error) {
		return os.Open(name)
	}
}
