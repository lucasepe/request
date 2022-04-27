package reply

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"
)

// Handler is used to validate or handle the response to a request.
type Handler func(*http.Response) error

// ToAny decodes a response as a JSON object.
func ToAny(v any) Handler {
	return func(res *http.Response) error {
		return json.NewDecoder(res.Body).Decode(v)
	}
}

// ToBytesBuffer writes the response body to the provided bytes.Buffer.
func ToBytesBuffer(buf *bytes.Buffer) Handler {
	return func(res *http.Response) error {
		_, err := io.Copy(buf, res.Body)
		return err
	}
}

// ToWriter copies the response body to w.
func ToWriter(w io.Writer) Handler {
	return ToBufioReader(func(r *bufio.Reader) error {
		_, err := io.Copy(w, r)

		return err
	})
}

// ToBufioReader takes a callback which wraps the response body in a bufio.Reader.
func ToBufioReader(f func(r *bufio.Reader) error) Handler {
	return func(res *http.Response) error {
		return f(bufio.NewReader(res.Body))
	}
}

func ToString(sp *string) Handler {
	return func(res *http.Response) error {
		var buf strings.Builder
		_, err := io.Copy(&buf, res.Body)
		if err == nil {
			*sp = buf.String()
		}
		return err
	}
}
